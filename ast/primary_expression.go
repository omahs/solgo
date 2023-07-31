package ast

import (
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"

	ast_pb "github.com/txpull/protos/dist/go/ast"
	"github.com/txpull/solgo/parser"
)

// PrimaryExpression represents a primary expression node in the AST.
type PrimaryExpression struct {
	*ASTBuilder

	Id                     int64              `json:"id"`                         // Unique identifier for the node.
	NodeType               ast_pb.NodeType    `json:"node_type"`                  // Type of the node.
	Kind                   ast_pb.NodeType    `json:"kind,omitempty"`             // Kind of the node.
	Value                  string             `json:"value,omitempty"`            // Value of the node.
	HexValue               string             `json:"hex_value,omitempty"`        // Hexadecimal value of the node.
	Src                    SrcNode            `json:"src"`                        // Source location of the node.
	Name                   string             `json:"name,omitempty"`             // Name of the node.
	TypeDescription        *TypeDescription   `json:"type_description,omitempty"` // Type description of the node.
	OverloadedDeclarations []int64            `json:"overloaded_declarations"`    // Overloaded declarations of the node.
	ReferencedDeclaration  int64              `json:"referenced_declaration"`     // Referenced declaration of the node.
	IsPure                 bool               `json:"is_pure"`                    // Indicates if the node is pure.
	ArgumentTypes          []*TypeDescription `json:"argument_types,omitempty"`   // Argument types of the node.
}

// NewPrimaryExpression creates a new PrimaryExpression node with a given ASTBuilder.
// It initializes the OverloadedDeclarations slice and sets the NodeType to IDENTIFIER.
func NewPrimaryExpression(b *ASTBuilder) *PrimaryExpression {
	return &PrimaryExpression{
		ASTBuilder:             b,
		Id:                     b.GetNextID(),
		OverloadedDeclarations: make([]int64, 0),
		NodeType:               ast_pb.NodeType_IDENTIFIER,
	}
}

// SetReferenceDescriptor sets the reference descriptions of the PrimaryExpression node.
func (p *PrimaryExpression) SetReferenceDescriptor(refId int64, refDesc *TypeDescription) bool {
	p.ReferencedDeclaration = refId
	p.TypeDescription = refDesc
	return false
}

// GetId returns the unique identifier of the PrimaryExpression node.
func (p *PrimaryExpression) GetId() int64 {
	return p.Id
}

// GetType returns the type of the PrimaryExpression node.
func (p *PrimaryExpression) GetType() ast_pb.NodeType {
	return p.NodeType
}

// GetSrc returns the source location of the PrimaryExpression node.
func (p *PrimaryExpression) GetSrc() SrcNode {
	return p.Src
}

// GetName returns the name of the PrimaryExpression node.
func (p *PrimaryExpression) GetName() string {
	return p.Name
}

// GetTypeDescription returns the type description of the PrimaryExpression node.
func (p *PrimaryExpression) GetTypeDescription() *TypeDescription {
	return p.TypeDescription
}

// GetArgumentTypes returns the argument types of the PrimaryExpression node.
func (p *PrimaryExpression) GetArgumentTypes() []*TypeDescription {
	return p.ArgumentTypes
}

// GetReferencedDeclaration returns the referenced declaration of the PrimaryExpression node.
func (p *PrimaryExpression) GetReferencedDeclaration() int64 {
	return p.ReferencedDeclaration
}

// GetNodes returns a slice of nodes that includes the expression of the PrimaryExpression node.
func (p *PrimaryExpression) GetNodes() []Node[NodeType] {
	return nil
}

// GetKind returns the kind of the PrimaryExpression node.
func (p *PrimaryExpression) GetKind() ast_pb.NodeType {
	return p.Kind
}

// ToProto returns a protobuf representation of the PrimaryExpression node.
// Currently, it returns an empty PrimaryExpression and needs to be implemented.
func (p *PrimaryExpression) ToProto() NodeType {
	return ast_pb.PrimaryExpression{}
}

// Parse takes a parser.PrimaryExpressionContext and parses it into a PrimaryExpression node.
// It sets the Src, Name, NodeType, Kind, Value, HexValue, TypeDescription, and other properties of the PrimaryExpression node.
// It returns the created PrimaryExpression node.
func (p *PrimaryExpression) Parse(
	unit *SourceUnit[Node[ast_pb.SourceUnit]],
	contractNode Node[NodeType],
	fnNode Node[NodeType],
	bodyNode *BodyNode,
	vDeclar *VariableDeclaration,
	expNode Node[NodeType],
	ctx *parser.PrimaryExpressionContext,
) Node[NodeType] {
	p.Src = SrcNode{
		Id:     p.GetNextID(),
		Line:   int64(ctx.GetStart().GetLine()),
		Column: int64(ctx.GetStart().GetColumn()),
		Start:  int64(ctx.GetStart().GetStart()),
		End:    int64(ctx.GetStop().GetStop()),
		Length: int64(ctx.GetStop().GetStop() - ctx.GetStart().GetStart() + 1),
		ParentIndex: func() int64 {
			if expNode != nil {
				return expNode.GetId()
			}

			if vDeclar != nil {
				return vDeclar.GetId()
			}

			return bodyNode.GetId()
		}(),
	}

	// This is a magic message type and should be treated accordingly...
	if ctx.GetText() == "msg" {
		p.TypeDescription = &TypeDescription{
			TypeIdentifier: "t_magic_message",
			TypeString:     "msg",
		}
	}

	// This is a magic block type and should be treated accordingly...
	if ctx.GetText() == "block" {
		p.TypeDescription = &TypeDescription{
			TypeIdentifier: "t_magic_block",
			TypeString:     "block",
		}
	}

	// This is a magic this type and should be treated by setting type description to the contract type
	if ctx.GetText() == "this" {
		p.TypeDescription = unit.GetTypeDescription()
	}

	if expNode != nil {
		switch expNodeCtx := expNode.(type) {
		case *FunctionCall:
			for _, argument := range expNodeCtx.GetArguments() {
				p.ArgumentTypes = append(p.ArgumentTypes, argument.GetTypeDescription())
			}
		}
	}

	if ctx.Identifier() != nil {
		p.Name = ctx.Identifier().GetText()

		// We cannot reach all of the parameter type description by simply discoveryReference
		// as some of the nodes such as this one is not yet written and is not accessible by
		// the discoverReferenceByCtxName()
		if fnNode != nil {
			switch fnNodeCtx := fnNode.(type) {
			case *Constructor:
				for _, param := range fnNodeCtx.GetParameters().GetParameters() {
					if param.GetName() == p.Name {
						if param.GetTypeName() != nil {
							p.TypeDescription = param.GetTypeName().GetTypeDescription()
							p.ReferencedDeclaration = p.GetId()
						}
						break
					}
				}
			case *Function:
				if fnNodeCtx.GetParameters() != nil {
					for _, param := range fnNodeCtx.GetParameters().GetParameters() {
						if param.GetName() == p.Name {
							if param.GetTypeName() != nil {
								p.TypeDescription = param.GetTypeName().GetTypeDescription()
								p.ReferencedDeclaration = p.GetId()
							}
							break
						}
					}
				}
			}
		}

		if bodyNode != nil {
			for _, statement := range bodyNode.GetStatements() {
				if statement.GetType() == ast_pb.NodeType_VARIABLE_DECLARATION {
					vDeclar := statement.(*VariableDeclaration)
					for _, declaration := range vDeclar.GetDeclarations() {
						if declaration.GetName() == p.Name {
							p.TypeDescription = declaration.GetTypeName().GetTypeDescription()
							p.ReferencedDeclaration = vDeclar.GetId()
							break
						}
					}
				}
			}
		}

		if p.TypeDescription == nil {
			if ref, refTypeDescription := p.GetResolver().ResolveByNode(p, p.Name); ref != nil {
				p.ReferencedDeclaration = ref.GetId()
				p.TypeDescription = refTypeDescription
			}
		}
	}

	// There is a case where we get PlaceholderStatement as a PrimaryExpression and this one does nothing really...
	// So we are going to do some hack here to make it work properly...
	if p.Name == "_" {
		p.NodeType = ast_pb.NodeType_PLACEHOLDER_STATEMENT
		return p
	}

	literalCtx := ctx.Literal()
	if literalCtx != nil {
		p.NodeType = ast_pb.NodeType_LITERAL
		p.IsPure = true

		if literalCtx.BooleanLiteral() != nil {
			if p.Name == "true" || p.Name == "false" {
				p.Name = ""
			}

			p.Kind = ast_pb.NodeType_BOOLEAN
			p.Value = strings.TrimSpace(
				// There can be hex 22 at beginning and end of literal.
				// We should drop it as that's ASCII for double quote.
				strings.ReplaceAll(literalCtx.BooleanLiteral().GetText(), "\"", ""),
			)
			p.HexValue = hex.EncodeToString([]byte(p.Value))

			p.TypeDescription = &TypeDescription{
				TypeIdentifier: "t_bool",
				TypeString:     "bool",
			}
		} else if literalCtx.StringLiteral() != nil {
			p.Name = ""
			p.Kind = ast_pb.NodeType_STRING

			p.Value = strings.TrimSpace(
				// There can be hex 22 at beginning and end of literal.
				// We should drop it as that's ASCII for double quote.
				strings.ReplaceAll(literalCtx.StringLiteral().GetText(), "\"", ""),
			)
			p.HexValue = hex.EncodeToString([]byte(p.Value))

			p.TypeDescription = &TypeDescription{
				TypeIdentifier: "t_string_literal",
				TypeString: fmt.Sprintf(
					"literal_string %s",
					literalCtx.StringLiteral().GetText(),
				),
			}
		} else if literalCtx.NumberLiteral() != nil {
			p.Kind = ast_pb.NodeType_NUMBER

			p.Value = strings.TrimSpace(
				// There can be hex 22 at beginning and end of literal.
				// We should drop it as that's ASCII for double quote.
				strings.ReplaceAll(literalCtx.NumberLiteral().GetText(), "\"", ""),
			)
			p.HexValue = hex.EncodeToString([]byte(p.Value))

			// Check if the number is a floating-point number
			if strings.Contains(p.Value, ".") {
				parts := strings.Split(p.Value, ".")

				// The numerator is the number without the decimal point
				numerator, _ := strconv.Atoi(parts[0] + parts[1])

				// The denominator is a power of 10 equal to the number of digits in the fractional part
				denominator := int(math.Pow(10, float64(len(parts[1]))))

				p.TypeDescription = &TypeDescription{
					TypeIdentifier: fmt.Sprintf("t_rational_%d_by_%d", numerator, denominator),
					TypeString: fmt.Sprintf(
						"fixed_const %s",
						literalCtx.NumberLiteral().GetText(),
					),
				}
			} else {
				numerator, _ := strconv.Atoi(p.Value)
				denominator := 1
				p.TypeDescription = &TypeDescription{
					TypeIdentifier: fmt.Sprintf("t_rational_%d_by_%d", numerator, denominator),
					TypeString: fmt.Sprintf(
						"int_const %s",
						literalCtx.NumberLiteral().GetText(),
					),
				}
			}
		} else if literalCtx.HexStringLiteral() != nil {
			p.Kind = ast_pb.NodeType_HEX_STRING

			p.Value = strings.TrimSpace(
				// There can be hex 22 at beginning and end of literal.
				// We should drop it as that's ASCII for double quote.
				strings.ReplaceAll(literalCtx.StringLiteral().GetText(), "\"", ""),
			)
			p.HexValue = hex.EncodeToString([]byte(p.Value))

			p.TypeDescription = &TypeDescription{
				TypeIdentifier: "t_string_hex_literal",
				TypeString: fmt.Sprintf(
					"literal_hex_string %s",
					literalCtx.StringLiteral().GetText(),
				),
			}
		} else {
			if ctx.GetText() == "msg" {
				p.TypeDescription = &TypeDescription{
					TypeIdentifier: "t_magic_message",
					TypeString:     "msg",
				}
			}

			if p.TypeDescription.TypeString == "" {
				if expNode != nil {
					p.TypeDescription = expNode.GetTypeDescription()
				}
			}
		}
	}

	return p
}