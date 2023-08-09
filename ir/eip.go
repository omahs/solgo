package ir

import (
	ir_pb "github.com/txpull/protos/dist/go/ir"
	"github.com/txpull/solgo/eip"
)

// EIP represents a specific Ethereum Improvement Proposal standard that a contract may adhere to.
type EIP struct {
	// ContractId is the unique identifier for the contract.
	ContractId int64 `json:"contract_id"`

	// ContractName is the name of the contract.
	ContractName string `json:"contract_name"`

	// Confidence represents the confidence level of the contract adhering to a specific EIP standard.
	Confidence eip.Discovery `json:"confidences"`

	// Standard provides details about the specific EIP standard.
	Standard eip.ContractStandard `json:"standards"`
}

// GetContractId returns the unique identifier for the contract.
func (e *EIP) GetContractId() int64 {
	return e.ContractId
}

// GetContractName returns the name of the contract.
func (e *EIP) GetContractName() string {
	return e.ContractName
}

// GetConfidence returns the confidence level of the contract adhering to a specific EIP standard.
func (e *EIP) GetConfidence() eip.Discovery {
	return e.Confidence
}

// GetStandard returns the EIP standard.
func (e *EIP) GetStandard() eip.ContractStandard {
	return e.Standard
}

// ToProto converts the EIP to its protobuf representation.
func (e *EIP) ToProto() *ir_pb.EIP {
	return &ir_pb.EIP{
		ContractId:   e.ContractId,
		ContractName: e.ContractName,
		Confidence:   e.Confidence.ToProto(),
		Standard:     e.Standard.ToProto(),
	}
}

// processEips processes the given RootSourceUnit to identify and associate it with known EIP standards.
// It extracts functions and events from the contract and sends them to the EIP discovery package
// to determine if the contract matches any known EIPs.
func (b *Builder) processEips(root *RootSourceUnit) {
	// Extracting functions and events to build actual contract that is going to
	// be sent towards the EIP discovery package.

	contract := &eip.Contract{
		Name:      root.GetEntryName(),
		Functions: make([]eip.Function, 0),
		Events:    make([]eip.Event, 0),
	}

	for _, unit := range root.GetContracts() {
		for _, function := range unit.GetFunctions() {
			inputs := make([]eip.Input, 0)
			outputs := make([]eip.Output, 0)

			for _, param := range function.GetParameters() {
				inputs = append(inputs, eip.Input{
					Type:    param.GetTypeDescription().GetString(),
					Indexed: false, // Specific to events...
				})
			}

			for _, ret := range function.GetReturnStatements() {
				outputs = append(outputs, eip.Output{
					Type: ret.GetTypeDescription().GetString(),
				})
			}

			contract.Functions = append(contract.Functions, eip.Function{
				Name:    function.GetName(),
				Inputs:  inputs,
				Outputs: outputs,
			})
		}

		for _, event := range unit.GetEvents() {
			inputs := make([]eip.Input, 0)

			for _, param := range event.GetParameters() {
				inputs = append(inputs, eip.Input{
					Type:    param.GetTypeDescription().GetString(),
					Indexed: param.IsIndexed(),
				})
			}

			contract.Events = append(contract.Events, eip.Event{
				Name:    event.GetName(),
				Inputs:  inputs,
				Outputs: make([]eip.Output, 0),
			})
		}
	}

	// Now when we have full contract functions and events we can send it to the
	// EIP discovery package to find out if it matches any of the EIPs.
	for _, standard := range eip.GetSortedRegisteredStandards() {
		if !root.HasEIP(standard.GetType()) {
			if confidence, found := eip.ConfidenceCheck(standard, contract); found {
				root.Eips = append(root.Eips, &EIP{
					ContractName: contract.Name,
					ContractId:   root.GetEntryId(),
					Confidence:   confidence,
					Standard:     standard.GetStandard(),
				})

				// Fuck finally, lets check if this contract is proxy, upgradeable,
				// nft or token or any other really that we support...
				root.SetContractType(standard.GetType())
			}
		}
	}

}