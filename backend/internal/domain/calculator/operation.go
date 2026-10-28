package calculator

import (
	"math"
	"strconv"

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

// maxResultDigits caps exact power results, which arbitrary precision would otherwise let grow without bound.
const maxResultDigits = 1000

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
	{name: "power", symbol: "^", arity: Binary, evaluate: power},
	{name: "sqrt", symbol: "√", arity: Unary, evaluate: sqrt},
	{name: "percentage", symbol: "%", arity: Binary, evaluate: percentage},
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

func power(operands []decimal.Decimal) (decimal.Decimal, error) {
	base, exponent := operands[0], operands[1]
	if (base.IsZero() && exponent.Sign() <= 0) || (base.Sign() < 0 && !exponent.IsInteger()) {
		return decimal.Decimal{}, domainerror.ErrUndefinedPower
	}
	if !powerMayFit(base, exponent) {
		return decimal.Decimal{}, domainerror.ErrResultTooLarge
	}
	result, err := base.PowWithPrecision(exponent, resultScale)
	if err != nil || !resultFits(result) {
		return decimal.Decimal{}, domainerror.ErrResultTooLarge
	}
	return result, nil
}

// powerMayFit rejects, before any big-number work, powers whose result is certainly out of bounds:
// its magnitude from |exponent| × log10|base|, and its digit count, since an exact c^n has at least
// n × (digits(c) − 1) + 1 digits. resultFits makes the exact call afterwards.
func powerMayFit(base, exponent decimal.Decimal) bool {
	if exponent.Abs().GreaterThan(decimal.NewFromInt(maxExponent)) {
		return false
	}
	if base.IsZero() {
		return true
	}
	if math.Abs(exponent.InexactFloat64()*log10Abs(base)) > maxExponent+1 {
		return false
	}
	if exponent.IsInteger() && exponent.Sign() > 0 {
		digits, _ := scientific(base)
		return exponent.IntPart()*int64(digits-1)+1 <= maxResultDigits
	}
	return true
}

func resultFits(result decimal.Decimal) bool {
	if result.IsZero() {
		return true
	}
	digits, exponent := scientific(result)
	return digits <= maxResultDigits && exponent <= maxExponent && exponent >= -maxExponent
}

// log10Abs works from the leading digits, since a bounded operand can still overflow float64.
func log10Abs(value decimal.Decimal) float64 {
	_, exponent := scientific(value)
	text := value.Coefficient().String()
	if text[0] == '-' {
		text = text[1:]
	}
	leading, _ := strconv.ParseFloat(text[:1]+"."+text[1:min(len(text), 17)], 64)
	return float64(exponent) + math.Log10(leading)
}

func sqrt(operands []decimal.Decimal) (decimal.Decimal, error) {
	if operands[0].Sign() < 0 {
		return decimal.Decimal{}, domainerror.ErrNegativeSquareRoot
	}
	return operands[0].SqrtRound(resultScale)
}

// percentage is a% of b; dividing by 100 is an exact decimal shift, so no rounding is needed.
func percentage(operands []decimal.Decimal) (decimal.Decimal, error) {
	return operands[0].Mul(operands[1]).Shift(-2), nil
}
