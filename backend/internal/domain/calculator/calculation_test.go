package calculator_test

import (
	"math/big"
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

		// T04: power
		{name: "power 2 10 is 1024", operation: "power", operands: []string{"2", "10"}, want: "1024"},
		{name: "power 2 -1 is 0.5", operation: "power", operands: []string{"2", "-1"}, want: "0.5"},
		{name: "power 4 0.5 is 2", operation: "power", operands: []string{"4", "0.5"}, want: "2"},
		{name: "power 2 0.5 rounds to 16 places like sqrt", operation: "power", operands: []string{"2", "0.5"}, want: "1.414213562373095"},
		{name: "power -2 3 is -8", operation: "power", operands: []string{"-2", "3"}, want: "-8"},
		{name: "power x 0 is 1", operation: "power", operands: []string{"-7.5", "0"}, want: "1"},
		{name: "power 0 positive is 0", operation: "power", operands: []string{"0", "3"}, want: "0"},
		{name: "power 0 -1 is undefined", operation: "power", operands: []string{"0", "-1"}, wantErr: domainerror.ErrUndefinedPower},
		{name: "power 0 0 is undefined", operation: "power", operands: []string{"0", "0"}, wantErr: domainerror.ErrUndefinedPower},
		{name: "power -8 0.5 is undefined", operation: "power", operands: []string{"-8", "0.5"}, wantErr: domainerror.ErrUndefinedPower},
		{name: "power 10 100000 is too large", operation: "power", operands: []string{"10", "100000"}, wantErr: domainerror.ErrResultTooLarge},
		{name: "power exponent 1000 accepted", operation: "power", operands: []string{"1", "1000"}, want: "1"},
		{name: "power exponent 1001 too large", operation: "power", operands: []string{"1", "1001"}, wantErr: domainerror.ErrResultTooLarge},
		{name: "power exponent -1001 too large", operation: "power", operands: []string{"1", "-1001"}, wantErr: domainerror.ErrResultTooLarge},
		{name: "power result 1e1000 accepted", operation: "power", operands: []string{"10", "1000"}, want: "1" + strings.Repeat("0", 1000)},
		{name: "power result 1e-1000 accepted", operation: "power", operands: []string{"10", "-1000"}, want: "0." + strings.Repeat("0", 999) + "1"},
		{name: "power result beyond 1e1000 too large", operation: "power", operands: []string{"11", "1000"}, wantErr: domainerror.ErrResultTooLarge},
		{name: "power result beyond 1e-1000 too large", operation: "power", operands: []string{"0.001", "500"}, wantErr: domainerror.ErrResultTooLarge},
		{name: "power fractional result beyond 1e1000 too large", operation: "power", operands: []string{"1e999", "1.5"}, wantErr: domainerror.ErrResultTooLarge},
		{name: "power result of 1000 digits accepted", operation: "power", operands: []string{"1.1", "960"}, want: onePointOnePow(960)},
		{name: "power result of 1001 digits too large", operation: "power", operands: []string{"1.1", "961"}, wantErr: domainerror.ErrResultTooLarge},
		{name: "power result of 50000 digits too large", operation: "power", operands: []string{"1." + fiftyDigits[1:], "1000"}, wantErr: domainerror.ErrResultTooLarge},

		// T04: sqrt
		{name: "sqrt 16 is 4", operation: "sqrt", operands: []string{"16"}, want: "4"},
		{name: "sqrt 2 rounds to 16 places with trailing zero trimmed", operation: "sqrt", operands: []string{"2"}, want: "1.414213562373095"},
		{name: "sqrt 0 is 0", operation: "sqrt", operands: []string{"0"}, want: "0"},
		{name: "sqrt 0.0625 is 0.25", operation: "sqrt", operands: []string{"0.0625"}, want: "0.25"},
		{name: "sqrt -1 is negative_square_root", operation: "sqrt", operands: []string{"-1"}, wantErr: domainerror.ErrNegativeSquareRoot},
		{name: "sqrt with 2 operands", operation: "sqrt", operands: []string{"1", "2"}, wantErr: domainerror.ErrInvalidOperandCount},
		{name: "sqrt with no operands", operation: "sqrt", operands: nil, wantErr: domainerror.ErrInvalidOperandCount},

		// T04: percentage
		{name: "percentage 20 of 50 is 10", operation: "percentage", operands: []string{"20", "50"}, want: "10"},
		{name: "percentage 12.5 of 80 is 10", operation: "percentage", operands: []string{"12.5", "80"}, want: "10"},
		{name: "percentage is exact", operation: "percentage", operands: []string{"1", "0.01"}, want: "0.0001"},
		{name: "percentage of negative", operation: "percentage", operands: []string{"-15", "200"}, want: "-30"},
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
	tests := []struct {
		name  string
		arity calculator.Arity
	}{
		{name: "add", arity: calculator.Binary},
		{name: "subtract", arity: calculator.Binary},
		{name: "multiply", arity: calculator.Binary},
		{name: "divide", arity: calculator.Binary},
		{name: "power", arity: calculator.Binary},
		{name: "sqrt", arity: calculator.Unary},
		{name: "percentage", arity: calculator.Binary},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op, err := calculator.ParseOperation(tt.name)

			require.NoError(t, err)
			assert.Equal(t, tt.name, op.Name())
			assert.Equal(t, tt.arity, op.Arity())
		})
	}
}

// onePointOnePow returns 1.1^n exactly: 11^n with n decimal places (11^960 has 1000 digits).
func onePointOnePow(n int64) string {
	digits := new(big.Int).Exp(big.NewInt(11), big.NewInt(n), nil).String()
	return digits[:len(digits)-int(n)] + "." + digits[len(digits)-int(n):]
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

func TestOperations_ListsEveryOperationInDisplayOrder(t *testing.T) {
	ops := calculator.Operations()

	names := make([]string, 0, len(ops))
	for _, op := range ops {
		parsed, err := calculator.ParseOperation(op.Name())
		require.NoError(t, err)
		assert.Equal(t, op.Symbol(), parsed.Symbol())
		names = append(names, op.Name())
	}
	assert.Equal(t, []string{"add", "subtract", "multiply", "divide", "power", "sqrt", "percentage"}, names)
}
