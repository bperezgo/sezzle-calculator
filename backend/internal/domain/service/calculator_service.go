package service

import (
	"context"

	"sezzle.com/calculator/internal/domain/calculator"
	"sezzle.com/calculator/internal/domain/dto"
	"sezzle.com/calculator/internal/domain/ports"
)

type CalculatorService struct{}

func NewCalculatorService() ports.Calculator {
	return &CalculatorService{}
}

func (s *CalculatorService) Calculate(ctx context.Context, req dto.CalculationRequest) (dto.CalculationResult, error) {
	if err := ctx.Err(); err != nil {
		return dto.CalculationResult{}, err
	}

	operation, err := calculator.ParseOperation(req.Operation)
	if err != nil {
		return dto.CalculationResult{}, err
	}

	operands := make([]calculator.Operand, 0, len(req.Operands))
	canonical := make([]string, 0, len(req.Operands))
	for _, raw := range req.Operands {
		operand, parseErr := calculator.NewOperand(raw)
		if parseErr != nil {
			return dto.CalculationResult{}, parseErr
		}
		operands = append(operands, operand)
		canonical = append(canonical, operand.Decimal().String())
	}

	calculation, err := calculator.NewCalculation(operation, operands)
	if err != nil {
		return dto.CalculationResult{}, err
	}
	result, err := calculation.Evaluate()
	if err != nil {
		return dto.CalculationResult{}, err
	}

	return dto.CalculationResult{
		Operation: operation.Name(),
		Operands:  canonical,
		Result:    result.String(),
	}, nil
}

func (s *CalculatorService) Operations(_ context.Context) []dto.OperationInfo {
	operations := calculator.Operations()
	infos := make([]dto.OperationInfo, 0, len(operations))
	for _, op := range operations {
		infos = append(infos, dto.OperationInfo{Name: op.Name(), Arity: int(op.Arity()), Symbol: op.Symbol()})
	}
	return infos
}
