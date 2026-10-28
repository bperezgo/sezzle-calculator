package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sezzle.com/calculator/internal/domain/dto"
	domainerror "sezzle.com/calculator/internal/domain/error"
	"sezzle.com/calculator/internal/domain/service"
)

func TestCalculatorService_Calculate(t *testing.T) {
	tests := []struct {
		name    string
		req     dto.CalculationRequest
		want    dto.CalculationResult
		wantErr *domainerror.Error
	}{
		// Service scenarios mirror T01's: success
		{
			name: "add 2 3 is 5",
			req:  dto.CalculationRequest{Operation: "add", Operands: []string{"2", "3"}},
			want: dto.CalculationResult{Operation: "add", Operands: []string{"2", "3"}, Result: "5"},
		},
		{
			name: "subtract 5 8 is -3",
			req:  dto.CalculationRequest{Operation: "subtract", Operands: []string{"5", "8"}},
			want: dto.CalculationResult{Operation: "subtract", Operands: []string{"5", "8"}, Result: "-3"},
		},
		{
			name: "multiply 4 2.5 is 10",
			req:  dto.CalculationRequest{Operation: "multiply", Operands: []string{"4", "2.5"}},
			want: dto.CalculationResult{Operation: "multiply", Operands: []string{"4", "2.5"}, Result: "10"},
		},
		{
			name: "divide 1 3 rounds to 16 places",
			req:  dto.CalculationRequest{Operation: "divide", Operands: []string{"1", "3"}},
			want: dto.CalculationResult{Operation: "divide", Operands: []string{"1", "3"}, Result: "0.3333333333333333"},
		},
		{
			name: "add 0.1 0.2 is exactly 0.3",
			req:  dto.CalculationRequest{Operation: "add", Operands: []string{"0.1", "0.2"}},
			want: dto.CalculationResult{Operation: "add", Operands: []string{"0.1", "0.2"}, Result: "0.3"},
		},

		// Operands: ["2.50", "1"] → result DTO operands ["2.5", "1"]
		{
			name: "operands are echoed in canonical form",
			req:  dto.CalculationRequest{Operation: "add", Operands: []string{"2.50", "1"}},
			want: dto.CalculationResult{Operation: "add", Operands: []string{"2.5", "1"}, Result: "3.5"},
		},
		{
			name: "exponent operands are echoed in canonical form",
			req:  dto.CalculationRequest{Operation: "add", Operands: []string{"1e3", "-0.0"}},
			want: dto.CalculationResult{Operation: "add", Operands: []string{"1000", "0"}, Result: "1000"},
		},

		// Service scenarios mirror T01's: each error code passes through with errors.Is
		{
			name:    "unsupported operation",
			req:     dto.CalculationRequest{Operation: "modulo", Operands: []string{"5", "2"}},
			wantErr: domainerror.ErrUnsupportedOperation,
		},
		{
			name:    "invalid operand",
			req:     dto.CalculationRequest{Operation: "add", Operands: []string{"abc", "1"}},
			wantErr: domainerror.ErrInvalidOperand,
		},
		{
			name:    "invalid operand count",
			req:     dto.CalculationRequest{Operation: "add", Operands: []string{"1"}},
			wantErr: domainerror.ErrInvalidOperandCount,
		},
		{
			name:    "nil operands is an invalid operand count",
			req:     dto.CalculationRequest{Operation: "add"},
			wantErr: domainerror.ErrInvalidOperandCount,
		},
		{
			name:    "operand too large",
			req:     dto.CalculationRequest{Operation: "add", Operands: []string{"1e1001", "1"}},
			wantErr: domainerror.ErrOperandTooLarge,
		},
		{
			name:    "division by zero",
			req:     dto.CalculationRequest{Operation: "divide", Operands: []string{"1", "0"}},
			wantErr: domainerror.ErrDivisionByZero,
		},
	}

	svc := service.NewCalculatorService()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.Calculate(context.Background(), tt.req)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, dto.CalculationResult{}, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCalculatorService_Calculate_ErrorsKeepDomainFields(t *testing.T) {
	_, err := service.NewCalculatorService().Calculate(context.Background(), dto.CalculationRequest{
		Operation: "add",
		Operands:  []string{"x", "1"},
	})

	var domainErr *domainerror.Error
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, domainerror.KindInvalidInput, domainErr.Kind)
	assert.Equal(t, "x", domainErr.Fields["operand"])
}

func TestCalculatorService_Calculate_DoesNotAliasRequestOperands(t *testing.T) {
	req := dto.CalculationRequest{Operation: "add", Operands: []string{"2", "3"}}

	got, err := service.NewCalculatorService().Calculate(context.Background(), req)
	require.NoError(t, err)
	got.Operands[0] = "changed"

	assert.Equal(t, []string{"2", "3"}, req.Operands)
}

// Cancelled context → returns ctx.Err() without calculating
func TestCalculatorService_Calculate_CancelledContext(t *testing.T) {
	tests := []struct {
		name string
		req  dto.CalculationRequest
	}{
		{name: "valid request", req: dto.CalculationRequest{Operation: "add", Operands: []string{"2", "3"}}},
		{name: "invalid request is not evaluated", req: dto.CalculationRequest{Operation: "divide", Operands: []string{"1", "0"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			got, err := service.NewCalculatorService().Calculate(ctx, tt.req)

			require.ErrorIs(t, err, context.Canceled)
			require.NotErrorIs(t, err, domainerror.ErrDivisionByZero)
			assert.Equal(t, dto.CalculationResult{}, got)
		})
	}
}

// Decision 1: the port lists the supported operations so the API and UI can be data-driven.
func TestCalculatorService_Operations_ListsSupportedOperationsInOrder(t *testing.T) {
	got := service.NewCalculatorService().Operations(context.Background())

	assert.Equal(t, []dto.OperationInfo{
		{Name: "add", Arity: 2, Symbol: "+"},
		{Name: "subtract", Arity: 2, Symbol: "−"},
		{Name: "multiply", Arity: 2, Symbol: "×"},
		{Name: "divide", Arity: 2, Symbol: "÷"},
	}, got)
}
