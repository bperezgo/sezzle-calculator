package error_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainerror "sezzle.com/calculator/internal/domain/error"
)

func TestError_Sentinels_HaveStableCodesKindsAndMessages(t *testing.T) {
	tests := []struct {
		name    string
		err     *domainerror.Error
		code    domainerror.Code
		kind    domainerror.Kind
		message string
	}{
		{"unsupported operation", domainerror.ErrUnsupportedOperation, "unsupported_operation", domainerror.KindInvalidInput, "operation is not supported"},
		{"invalid operand", domainerror.ErrInvalidOperand, "invalid_operand", domainerror.KindInvalidInput, "operand is not a valid decimal number"},
		{"invalid operand count", domainerror.ErrInvalidOperandCount, "invalid_operand_count", domainerror.KindInvalidInput, "wrong number of operands for operation"},
		{"operand too large", domainerror.ErrOperandTooLarge, "operand_too_large", domainerror.KindInvalidInput, "operand exceeds the supported size"},
		{"division by zero", domainerror.ErrDivisionByZero, "division_by_zero", domainerror.KindMath, "division by zero is undefined"},
		{"negative square root", domainerror.ErrNegativeSquareRoot, "negative_square_root", domainerror.KindMath, "square root of a negative number is undefined"},
		{"undefined power", domainerror.ErrUndefinedPower, "undefined_power", domainerror.KindMath, "power is undefined for these operands"},
		{"result too large", domainerror.ErrResultTooLarge, "result_too_large", domainerror.KindMath, "result exceeds the supported size"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.code, tt.err.Code)
			assert.Equal(t, tt.kind, tt.err.Kind)
			assert.Equal(t, tt.message, tt.err.Message)
			assert.Equal(t, tt.message, tt.err.Error())
		})
	}
}

func TestError_WithFields_KeepsIdentityAndDoesNotMutateSentinel(t *testing.T) {
	err := domainerror.ErrInvalidOperand.WithFields(map[string]string{"operand": "abc"})

	require.ErrorIs(t, err, domainerror.ErrInvalidOperand)
	assert.Equal(t, map[string]string{"operand": "abc"}, err.Fields)
	assert.Nil(t, domainerror.ErrInvalidOperand.Fields)
}

func TestError_Is_MatchesByCodeThroughWrapping(t *testing.T) {
	wrapped := fmt.Errorf("context: %w", domainerror.ErrDivisionByZero.WithFields(nil))

	require.ErrorIs(t, wrapped, domainerror.ErrDivisionByZero)
	require.NotErrorIs(t, wrapped, domainerror.ErrInvalidOperand)
	require.NotErrorIs(t, domainerror.ErrDivisionByZero, errors.New("division by zero is undefined"))
}

func TestError_As_ExposesKindAndCode(t *testing.T) {
	var target *domainerror.Error
	err := fmt.Errorf("wrapped: %w", domainerror.ErrOperandTooLarge)

	require.ErrorAs(t, err, &target)
	assert.Equal(t, domainerror.KindInvalidInput, target.Kind)
	assert.Equal(t, domainerror.CodeOperandTooLarge, target.Code)
}
