package server

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"sezzle.com/calculator/internal/domain/service"
	"sezzle.com/calculator/internal/platform/config"
	"sezzle.com/calculator/internal/platform/handler"
	"sezzle.com/calculator/internal/platform/middleware"
)

const wantContentType = "application/json; charset=utf-8"

func newTestEngine() *gin.Engine {
	return NewRouter(config.Config{Port: "8080", GinMode: "test"}, slog.New(slog.NewTextHandler(io.Discard, nil)), Handlers{
		Health:      handler.NewHealthHandler("test"),
		Calculation: handler.NewCalculationHandler(service.NewCalculatorService()),
	})
}

func newTestRouter() http.Handler {
	return newTestEngine()
}

func serve(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	r.ServeHTTP(rec, req)
	return rec
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

func TestRouter_Calculations_EndToEnd(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		want       string
	}{
		{
			name:       "add 2 3 is 200 result 5",
			body:       `{"operation":"add","operands":["2","3"]}`,
			wantStatus: http.StatusOK,
			want:       `{"operation":"add","operands":["2","3"],"result":"5"}`,
		},
		{
			name:       "JSON numbers 0.1 + 0.2 is exactly 0.3",
			body:       `{"operation":"add","operands":[0.1, 0.2]}`,
			wantStatus: http.StatusOK,
			want:       `{"operation":"add","operands":["0.1","0.2"],"result":"0.3"}`,
		},
		{
			name:       "divide 10 2 echoes the request shape",
			body:       `{"operation":"divide","operands":["10","2"]}`,
			wantStatus: http.StatusOK,
			want:       `{"operation":"divide","operands":["10","2"],"result":"5"}`,
		},
		{
			name:       "divide by zero is 422 division_by_zero",
			body:       `{"operation":"divide","operands":["1","0"]}`,
			wantStatus: http.StatusUnprocessableEntity,
			want:       `{"error":{"code":"division_by_zero","message":"division by zero is undefined"}}`,
		},
		{
			name:       "non-numeric operand is 400 invalid_operand",
			body:       `{"operation":"add","operands":["a","1"]}`,
			wantStatus: http.StatusBadRequest,
			want:       `{"error":{"code":"invalid_operand","message":"operand is not a valid decimal number"}}`,
		},
		{
			name:       "unknown operation is 400 unsupported_operation",
			body:       `{"operation":"modulo","operands":["1","2"]}`,
			wantStatus: http.StatusBadRequest,
			want:       `{"error":{"code":"unsupported_operation","message":"operation is not supported"}}`,
		},
		{
			name:       "wrong arity is 400 invalid_operand_count",
			body:       `{"operation":"add","operands":["1"]}`,
			wantStatus: http.StatusBadRequest,
			want:       `{"error":{"code":"invalid_operand_count","message":"wrong number of operands for operation"}}`,
		},
		{
			name:       "not JSON is 400 invalid_request",
			body:       `hello`,
			wantStatus: http.StatusBadRequest,
			want:       `{"error":{"code":"invalid_request","message":"request body must be a JSON object with an operation string and an operands array of numbers or numeric strings"}}`,
		},
		{
			name:       "empty body is 400 invalid_request",
			body:       ``,
			wantStatus: http.StatusBadRequest,
			want:       `{"error":{"code":"invalid_request","message":"request body must be a JSON object with an operation string and an operands array of numbers or numeric strings"}}`,
		},
		{
			name:       "body over 1 MiB is 413 payload_too_large",
			body:       `{"operation":"add","operands":["` + strings.Repeat("1", maxBodyBytes) + `","1"]}`,
			wantStatus: http.StatusRequestEntityTooLarge,
			want:       `{"error":{"code":"payload_too_large","message":"request body exceeds 1 MiB"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(newTestRouter(), http.MethodPost, "/api/v1/calculations", tt.body)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, wantContentType, rec.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.want, rec.Body.String())
		})
	}
}

func TestRouter_Operations_ListsTheDomainRegistry(t *testing.T) {
	rec := serve(newTestRouter(), http.MethodGet, "/api/v1/operations", "")

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, wantContentType, rec.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"operations":[
		{"name":"add","arity":2,"symbol":"+"},
		{"name":"subtract","arity":2,"symbol":"−"},
		{"name":"multiply","arity":2,"symbol":"×"},
		{"name":"divide","arity":2,"symbol":"÷"}
	]}`, rec.Body.String())
}

// Every response (including 404/405/500) is application/json, and X-Request-ID is present on every response.
func TestRouter_EveryResponseIsJSONWithRequestID(t *testing.T) {
	r := newTestEngine()
	r.GET("/test-panic", func(_ *gin.Context) { panic("boom") })

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
		want       string
	}{
		{name: "200 health", method: http.MethodGet, path: "/health", wantStatus: http.StatusOK, want: `{"status":"ok","version":"test"}`},
		{
			name: "400 invalid request", method: http.MethodPost, path: "/api/v1/calculations", body: `{`, wantStatus: http.StatusBadRequest,
			want: `{"error":{"code":"invalid_request","message":"request body must be a JSON object with an operation string and an operands array of numbers or numeric strings"}}`,
		},
		{
			name: "404 unknown route", method: http.MethodGet, path: "/nope", wantStatus: http.StatusNotFound,
			want: `{"error":{"code":"not_found","message":"route not found"}}`,
		},
		{
			name: "405 wrong method", method: http.MethodGet, path: "/api/v1/calculations", wantStatus: http.StatusMethodNotAllowed,
			want: `{"error":{"code":"method_not_allowed","message":"method not allowed for this route"}}`,
		},
		{
			name: "500 panic", method: http.MethodGet, path: "/test-panic", wantStatus: http.StatusInternalServerError,
			want: `{"error":{"code":"internal_error","message":"an unexpected error occurred"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(r, tt.method, tt.path, tt.body)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, wantContentType, rec.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.want, rec.Body.String())
			assert.NotEmpty(t, rec.Header().Get(middleware.RequestIDHeader))
		})
	}
}

func TestRouter_WrongMethod_AdvertisesAllowedMethods(t *testing.T) {
	rec := serve(newTestRouter(), http.MethodDelete, "/api/v1/calculations", "")

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.Equal(t, http.MethodPost, rec.Header().Get("Allow"))
}

func TestRouter_EchoesIncomingRequestID(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set(middleware.RequestIDHeader, "trace-42")

	newTestRouter().ServeHTTP(rec, req)

	assert.Equal(t, "trace-42", rec.Header().Get(middleware.RequestIDHeader))
}
