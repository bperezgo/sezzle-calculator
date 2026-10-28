package calculator

import (
	"github.com/shopspring/decimal"

	domainerror "sezzle.com/calculator/internal/domain/error"
)

// Arity is the number of operands an operation takes.
type Arity int

const (
	Unary  Arity = 1
	Binary Arity = 2
)

// resultScale is the number of decimal places kept when a result is not exact.
const resultScale = 16

// evaluator receives exactly Arity operands; NewCalculation guarantees it.
type evaluator func(operands []decimal.Decimal) (decimal.Decimal, error)

// Operation is a supported arithmetic operation. Obtain one through ParseOperation.
type Operation struct {
	name     string
	arity    Arity
	evaluate evaluator
}

var operations = map[string]Operation{
	"add":      {name: "add", arity: Binary, evaluate: add},
	"subtract": {name: "subtract", arity: Binary, evaluate: subtract},
	"multiply": {name: "multiply", arity: Binary, evaluate: multiply},
	"divide":   {name: "divide", arity: Binary, evaluate: divide},
}

func ParseOperation(name string) (Operation, error) {
	op, ok := operations[name]
	if !ok {
		return Operation{}, domainerror.ErrUnsupportedOperation.WithFields(map[string]string{"operation": name})
	}
	return op, nil
}

func (o Operation) Name() string { return o.name }

func (o Operation) Arity() Arity { return o.arity }

func add(operands []decimal.Decimal) (decimal.Decimal, error) {
	return operands[0].Add(operands[1]), nil
}

func subtract(operands []decimal.Decimal) (decimal.Decimal, error) {
	return operands[0].Sub(operands[1]), nil
}

func multiply(operands []decimal.Decimal) (decimal.Decimal, error) {
	return operands[0].Mul(operands[1]), nil
}

func divide(operands []decimal.Decimal) (decimal.Decimal, error) {
	if operands[1].IsZero() {
		return decimal.Decimal{}, domainerror.ErrDivisionByZero
	}
	// DivRound rounds half away from zero; the package-global DivisionPrecision is left untouched.
	return operands[0].DivRound(operands[1], resultScale), nil
}
