package calculator

import (
	"strings"

	"github.com/shopspring/decimal"

	domainerror "sezzle.com/calculator/internal/domain/error"
)

// Arbitrary precision makes huge operands valid but expensive, so their size is bounded.
const (
	maxSignificantDigits = 50
	maxExponent          = 1000
)

// Operand is an exact decimal input, parsed from its string form (never from float64).
type Operand struct {
	value decimal.Decimal
}

func NewOperand(raw string) (Operand, error) {
	value, err := decimal.NewFromString(raw)
	if err != nil {
		return Operand{}, domainerror.ErrInvalidOperand.WithFields(map[string]string{"operand": raw})
	}
	if !value.IsZero() {
		digits, exponent := scientific(value)
		if digits > maxSignificantDigits || exponent > maxExponent || exponent < -maxExponent {
			return Operand{}, domainerror.ErrOperandTooLarge.WithFields(map[string]string{"operand": raw})
		}
	}
	return Operand{value: value}, nil
}

func (o Operand) Decimal() decimal.Decimal { return o.value }

// scientific returns the significant-digit count and the exponent of value in scientific
// notation (d.ddd × 10^exponent), so "10e1000" and "1e1001" share exponent 1001.
func scientific(value decimal.Decimal) (digits, exponent int) {
	coefficient := value.Coefficient()
	coefficient.Abs(coefficient)
	text := coefficient.String()
	significant := strings.TrimRight(text, "0")
	trailingZeros := len(text) - len(significant)
	digits = len(significant)
	exponent = int(value.Exponent()) + trailingZeros + digits - 1
	return digits, exponent
}
