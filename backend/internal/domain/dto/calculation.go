// Package dto holds the plain structs that cross layer boundaries. Decimals travel as canonical strings.
package dto

type CalculationRequest struct {
	Operation string
	Operands  []string
}

type CalculationResult struct {
	Operation string
	Operands  []string
	Result    string
}

type OperationInfo struct {
	Name   string
	Arity  int
	Symbol string
}
