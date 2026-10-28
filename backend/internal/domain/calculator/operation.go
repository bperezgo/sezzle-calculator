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
	symbol   string
	arity    Arity
	evaluate evaluator
}

// ordered is the display order; byName is derived from it.
var ordered = []Operation{
	{name: "add", symbol: "+", arity: Binary, evaluate: add},
	{name: "subtract", symbol: "−", arity: Binary, evaluate: subtract},
	{name: "multiply", symbol: "×", arity: Binary, evaluate: multiply},
	{name: "divide", symbol: "÷", arity: Binary, evaluate: divide},
}

var byName = indexByName(ordered)

func indexByName(ops []Operation) map[string]Operation {
	index := make(map[string]Operation, len(ops))
	for _, op := range ops {
		index[op.name] = op
	}
	return index
}

func ParseOperation(name string) (Operation, error) {
	op, ok := byName[name]
	if !ok {
		return Operation{}, domainerror.ErrUnsupportedOperation.WithFields(map[string]string{"operation": name})
	}
	return op, nil
}

// Operations returns every supported operation in display order.
func Operations() []Operation {
	return append([]Operation(nil), ordered...)
}

func (o Operation) Name() string { return o.name }

func (o Operation) Symbol() string { return o.symbol }

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
