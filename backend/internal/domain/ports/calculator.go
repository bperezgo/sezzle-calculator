package ports

import (
	"context"

	"sezzle.com/calculator/internal/domain/dto"
)

// Calculator is the driving port: the use case the HTTP adapter depends on.
type Calculator interface {
	Calculate(ctx context.Context, req dto.CalculationRequest) (dto.CalculationResult, error)
	Operations(ctx context.Context) []dto.OperationInfo
}
