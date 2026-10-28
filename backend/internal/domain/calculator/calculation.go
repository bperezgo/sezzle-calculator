package calculator

import (
	"strconv"

	"github.com/shopspring/decimal"

	domainerror "sezzle.com/calculator/internal/domain/error"
)

// Calculation is an operation paired with the right number of operands; valid by construction.
type Calculation struct {
	operation Operation
	operands  []Operand
}

func NewCalculation(operation Operation, operands []Operand) (Calculation, error) {
	if operation.evaluate == nil {
		return Calculation{}, domainerror.ErrUnsupportedOperation
	}
	if len(operands) != int(operation.arity) {
		return Calculation{}, domainerror.ErrInvalidOperandCount.WithFields(map[string]string{
			"operation": operation.name,
			"expected":  strconv.Itoa(int(operation.arity)),
			"received":  strconv.Itoa(len(operands)),
		})
	}
	return Calculation{operation: operation, operands: append([]Operand(nil), operands...)}, nil
}

func (c Calculation) Evaluate() (decimal.Decimal, error) {
	values := make([]decimal.Decimal, len(c.operands))
	for i, operand := range c.operands {
		values[i] = operand.value
	}
	return c.operation.evaluate(values)
}
