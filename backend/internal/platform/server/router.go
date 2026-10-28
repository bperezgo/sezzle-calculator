package server

import (
	"github.com/gin-gonic/gin"

	"sezzle.com/calculator/internal/platform/config"
	"sezzle.com/calculator/internal/platform/handler"
)

type Handlers struct {
	Health *handler.HealthHandler
}

func NewRouter(cfg config.Config, h Handlers) *gin.Engine {
	gin.SetMode(cfg.GinMode)

	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", h.Health.Check)

	return r
}
