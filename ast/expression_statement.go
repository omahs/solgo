package ast

import (
	"fmt"

	"github.com/antlr4-go/antlr/v4"
	ast_pb "github.com/txpull/protos/dist/go/ast"
	"github.com/txpull/solgo/parser"
)

func parseExpressionStatement(
	b *ASTBuilder,
	unit *SourceUnit[Node[ast_pb.SourceUnit]],
	contractNode Node[NodeType],
	fnNode Node[NodeType],
	bodyNode *BodyNode,
	parentNode Node[NodeType],
	ctx *parser.ExpressionStatementContext,
) Node[NodeType] {
	for _, child := range ctx.GetChildren() {
		switch childCtx := child.(type) {
		case *parser.FunctionCallContext:
			statementNode := NewFunctionCall(b)
			statementNode.Parse(unit, contractNode, fnNode, bodyNode, nil, parentNode, childCtx)
			return statementNode
		case *parser.AssignmentContext:
			assignment := NewAssignment(b)
			assignment.ParseStatement(unit, contractNode, fnNode, bodyNode, parentNode, ctx, childCtx)
			return assignment
		case *parser.PrimaryExpressionContext:
			primaryExpression := NewPrimaryExpression(b)
			return primaryExpression.Parse(unit, contractNode, fnNode, bodyNode, nil, parentNode, childCtx)
		case *parser.UnarySuffixOperationContext:
			unarySuffixOperation := NewUnarySuffixExpression(b)
			return unarySuffixOperation.Parse(unit, contractNode, fnNode, bodyNode, nil, parentNode, childCtx)
		case *parser.UnaryPrefixOperationContext:
			unaryPrefixOperation := NewUnaryPrefixExpression(b)
			return unaryPrefixOperation.Parse(unit, contractNode, fnNode, bodyNode, nil, parentNode, childCtx)
		case *parser.OrderComparisonContext:
			binaryExp := NewBinaryOperationExpression(b)
			return binaryExp.ParseOrderComparison(unit, contractNode, fnNode, bodyNode, nil, parentNode, childCtx)
		case *antlr.TerminalNodeImpl:
			// @TODO: Not sure what to do with this... It's usually just a semicolon. Perhaps to
			// add to each expression statement semicolon_found?
			// Not important right now at all...
			continue
		default:
			panic(
				fmt.Sprintf(
					"Expression statement child not recognized @ ExpressionStatement.Parse: %T",
					childCtx,
				),
			)
		}
	}

	return nil
}