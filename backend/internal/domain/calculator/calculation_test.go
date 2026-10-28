package calculator_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sezzle.com/calculator/internal/domain/calculator"
	domainerror "sezzle.com/calculator/internal/domain/error"
)

// calculate mirrors how the use case drives the domain: parse, build, evaluate.
func calculate(operation string, operands []string) (string, error) {
	op, err := calculator.ParseOperation(operation)
	if err != nil {
		return "", err
	}
	parsed := make([]calculator.Operand, 0, len(operands))
	for _, raw := range operands {
		operand, parseErr := calculator.NewOperand(raw)
		if parseErr != nil {
			return "", parseErr
		}
		parsed = append(parsed, operand)
	}
	calc, err := calculator.NewCalculation(op, parsed)
	if err != nil {
		return "", err
	}
	result, err := calc.Evaluate()
	if err != nil {
		return "", err
	}
	return result.String(), nil
}

func TestCalculation_Evaluate(t *testing.T) {
	fiftyDigits := strings.Repeat("9", 50)

	tests := []struct {
		name      string
		operation string
		operands  []string
		want      string
		wantErr   *domainerror.Error
	}{
		// add ["2", "3"] → 5; subtract ["5", "8"] → -3; multiply ["4", "2.5"] → 10; divide ["10", "4"] → 2.5
		{name: "add 2 3 is 5", operation: "add", operands: []string{"2", "3"}, want: "5"},
		{name: "subtract 5 8 is -3", operation: "subtract", operands: []string{"5", "8"}, want: "-3"},
		{name: "multiply 4 2.5 is 10", operation: "multiply", operands: []string{"4", "2.5"}, want: "10"},
		{name: "divide 10 4 is 2.5", operation: "divide", operands: []string{"10", "4"}, want: "2.5"},

		// add ["0.1", "0.2"] → exactly 0.3
		{name: "add 0.1 0.2 is exactly 0.3", operation: "add", operands: []string{"0.1", "0.2"}, want: "0.3"},

		// divide ["1", "3"] → 0.3333333333333333 (16 places)
		{name: "divide 1 3 rounds to 16 places", operation: "divide", operands: []string{"1", "3"}, want: "0.3333333333333333"},

		// divide ["x", "0"] → division_by_zero (including ["0", "0"])
		{name: "divide by zero", operation: "divide", operands: []string{"7", "0"}, wantErr: domainerror.ErrDivisionByZero},
		{name: "divide zero by zero", operation: "divide", operands: []string{"0", "0"}, wantErr: domainerror.ErrDivisionByZero},
		{name: "divide by negative zero decimal", operation: "divide", operands: []string{"1", "-0.0"}, wantErr: domainerror.ErrDivisionByZero},

		// modulo → unsupported_operation
		{name: "modulo is unsupported", operation: "modulo", operands: []string{"5", "2"}, wantErr: domainerror.ErrUnsupportedOperation},
		{name: "empty operation is unsupported", operation: "", operands: []string{"5", "2"}, wantErr: domainerror.ErrUnsupportedOperation},

		// Binary op with 1 or 3 operands → invalid_operand_count
		{name: "binary op with 1 operand", operation: "add", operands: []string{"1"}, wantErr: domainerror.ErrInvalidOperandCount},
		{name: "binary op with 3 operands", operation: "multiply", operands: []string{"1", "2", "3"}, wantErr: domainerror.ErrInvalidOperandCount},
		{name: "binary op with no operands", operation: "subtract", operands: nil, wantErr: domainerror.ErrInvalidOperandCount},

		// "abc", "", "1,5" → invalid_operand; "1e3" and "-0.5" accepted
		{name: "operand abc is invalid", operation: "add", operands: []string{"abc", "1"}, wantErr: domainerror.ErrInvalidOperand},
		{name: "empty operand is invalid", operation: "add", operands: []string{"", "1"}, wantErr: domainerror.ErrInvalidOperand},
		{name: "operand with comma is invalid", operation: "add", operands: []string{"1,5", "1"}, wantErr: domainerror.ErrInvalidOperand},
		{name: "operand 1e3 accepted", operation: "add", operands: []string{"1e3", "1"}, want: "1001"},
		{name: "operand -0.5 accepted", operation: "add", operands: []string{"-0.5", "1"}, want: "0.5"},

		// divide ["2", "3"] → 0.6666666666666667 (rounded half away from zero, not truncated)
		{name: "divide 2 3 rounds half away from zero", operation: "divide", operands: []string{"2", "3"}, want: "0.6666666666666667"},
		{name: "divide -2 3 rounds half away from zero", operation: "divide", operands: []string{"-2", "3"}, want: "-0.6666666666666667"},

		// 50-significant-digit operand accepted; 51-digit operand → operand_too_large
		{name: "50 significant digits accepted", operation: "add", operands: []string{fiftyDigits, "0"}, want: fiftyDigits},
		{name: "51 significant digits too large", operation: "add", operands: []string{fiftyDigits + "9", "0"}, wantErr: domainerror.ErrOperandTooLarge},
		{name: "51 significant digits with decimal point too large", operation: "add", operands: []string{"0." + fiftyDigits + "1", "0"}, wantErr: domainerror.ErrOperandTooLarge},
		{name: "leading and trailing zeros are not significant", operation: "add", operands: []string{"000123000", "0.000100"}, want: "123000.0001"},

		// "1e1000" and "1e-1000" accepted; "1e1001" and "1e-1001" → operand_too_large
		{name: "exponent 1000 accepted", operation: "multiply", operands: []string{"1e1000", "0"}, want: "0"},
		{name: "exponent -1000 accepted", operation: "multiply", operands: []string{"1e-1000", "1e1000"}, want: "1"},
		{name: "exponent 1001 too large", operation: "add", operands: []string{"1e1001", "0"}, wantErr: domainerror.ErrOperandTooLarge},
		{name: "exponent -1001 too large", operation: "add", operands: []string{"1e-1001", "0"}, wantErr: domainerror.ErrOperandTooLarge},
		{name: "exponent is scientific not raw", operation: "add", operands: []string{"10e1000", "0"}, wantErr: domainerror.ErrOperandTooLarge},

		// Results print without trailing zeros ("10", not "10.0")
		{name: "result has no trailing zeros", operation: "add", operands: []string{"7.50", "2.50"}, want: "10"},
		{name: "exact division has no trailing zeros", operation: "divide", operands: []string{"1", "8"}, want: "0.125"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculate(tt.operation, tt.operands)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseOperation_ExposesNameAndArity(t *testing.T) {
	for _, name := range []string{"add", "subtract", "multiply", "divide"} {
		t.Run(name, func(t *testing.T) {
			op, err := calculator.ParseOperation(name)

			require.NoError(t, err)
			assert.Equal(t, name, op.Name())
			assert.Equal(t, calculator.Binary, op.Arity())
		})
	}
}

func TestParseOperation_UnknownCarriesOperationField(t *testing.T) {
	_, err := calculator.ParseOperation("modulo")

	var domainErr *domainerror.Error
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "modulo", domainErr.Fields["operation"])
}

func TestNewOperand_InvalidCarriesOperandField(t *testing.T) {
	_, err := calculator.NewOperand("abc")

	var domainErr *domainerror.Error
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, "abc", domainErr.Fields["operand"])
}

func TestNewOperand_KeepsExactValue(t *testing.T) {
	operand, err := calculator.NewOperand("2.50")

	require.NoError(t, err)
	assert.Equal(t, "2.5", operand.Decimal().String())
}

func TestNewCalculation_ZeroValueOperationIsUnsupported(t *testing.T) {
	_, err := calculator.NewCalculation(calculator.Operation{}, nil)

	require.ErrorIs(t, err, domainerror.ErrUnsupportedOperation)
}
