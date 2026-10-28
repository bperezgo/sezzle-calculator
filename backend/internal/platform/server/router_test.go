package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"sezzle.com/calculator/internal/platform/config"
	"sezzle.com/calculator/internal/platform/handler"
)

func newTestRouter() http.Handler {
	return NewRouter(config.Config{Port: "8080", GinMode: "test"}, Handlers{
		Health: handler.NewHealthHandler("test"),
	})
}

func TestRouter_Health_ReturnsOK(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	newTestRouter().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"status":"ok","version":"test"}`, rec.Body.String())
}

func TestRouter_UnknownRoute_ReturnsNotFound(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/nope", nil)

	newTestRouter().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
