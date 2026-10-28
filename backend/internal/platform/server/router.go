package server

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"sezzle.com/calculator/internal/platform/config"
	"sezzle.com/calculator/internal/platform/handler"
	"sezzle.com/calculator/internal/platform/middleware"
)

const maxBodyBytes = 1 << 20

type Handlers struct {
	Health      *handler.HealthHandler
	Calculation *handler.CalculationHandler
}

func NewRouter(cfg config.Config, logger *slog.Logger, h Handlers) *gin.Engine {
	gin.SetMode(cfg.GinMode)

	r := gin.New()
	r.HandleMethodNotAllowed = true
	r.Use(
		middleware.RequestID(),
		middleware.AccessLog(logger),
		middleware.Recovery(logger),
		middleware.BodyLimit(maxBodyBytes),
	)
	r.NoRoute(handler.NotFound)
	r.NoMethod(handler.MethodNotAllowed)

	r.GET("/health", h.Health.Check)

	v1 := r.Group("/api/v1")
	v1.POST("/calculations", h.Calculation.Calculate)
	v1.GET("/operations", h.Calculation.Operations)

	return r
}
