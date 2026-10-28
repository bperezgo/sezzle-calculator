package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	version string
}

type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

func NewHealthHandler(version string) *HealthHandler {
	return &HealthHandler{version: version}
}

func (h *HealthHandler) Check(c *gin.Context) {
	c.JSON(http.StatusOK, healthResponse{Status: "ok", Version: h.version})
}
