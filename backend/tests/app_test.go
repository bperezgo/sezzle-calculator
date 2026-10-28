// Package tests holds the acceptance tests: black-box HTTP against the fully wired router and real service.
// Each test or table row is named after a ticket acceptance criterion. Unit tests live next to the code.
package tests

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"sezzle.com/calculator/internal/domain/service"
	"sezzle.com/calculator/internal/platform/config"
	"sezzle.com/calculator/internal/platform/handler"
	"sezzle.com/calculator/internal/platform/server"
)

const (
	wantContentType    = "application/json; charset=utf-8"
	invalidRequestBody = `{"error":{"code":"invalid_request","message":"request body must be a JSON object with an operation string and an operands array of numbers or numeric strings"}}`
	internalErrorBody  = `{"error":{"code":"internal_error","message":"an unexpected error occurred"}}`
)

type app struct {
	*gin.Engine
	logs *bytes.Buffer
}

// newApp wires the API the way cmd/main.go does, with logs captured for assertions.
func newApp(t *testing.T) app {
	t.Helper()
	logs := &bytes.Buffer{}
	engine := server.NewRouter(
		config.Config{Port: "8080", GinMode: gin.TestMode},
		slog.New(slog.NewJSONHandler(logs, nil)),
		server.Handlers{
			Health:      handler.NewHealthHandler("test"),
			Calculation: handler.NewCalculationHandler(service.NewCalculatorService()),
		},
	)
	return app{Engine: engine, logs: logs}
}

func (a app) do(req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	a.ServeHTTP(rec, req)
	return rec
}

func (a app) serve(method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	return a.do(req)
}

func (a app) calculate(body string) *httptest.ResponseRecorder {
	return a.serve(http.MethodPost, "/api/v1/calculations", body)
}
