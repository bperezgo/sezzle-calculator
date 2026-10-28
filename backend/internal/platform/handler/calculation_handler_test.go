package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"sezzle.com/calculator/internal/domain/service"
	"sezzle.com/calculator/internal/platform/handler"
)

const invalidRequestBody = `{"error":{"code":"invalid_request","message":"request body must be a JSON object with an operation string and an operands array of numbers or numeric strings"}}`

func newCalculationEngine(mw ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := handler.NewCalculationHandler(service.NewCalculatorService())
	r := gin.New()
	r.Use(mw...)
	r.POST("/calculations", h.Calculate)
	r.GET("/operations", h.Operations)
	return r
}

func post(r http.Handler, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/calculations", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	return rec
}

func TestCalculationHandler_Calculate_Success(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "string operands",
			body: `{"operation":"add","operands":["2","3"]}`,
			want: `{"operation":"add","operands":["2","3"],"result":"5"}`,
		},
		{
			name: "JSON number operands are read from their raw text, never float64",
			body: `{"operation":"add","operands":[0.1, 0.2]}`,
			want: `{"operation":"add","operands":["0.1","0.2"],"result":"0.3"}`,
		},
		{
			name: "number operands keep precision beyond float64",
			body: `{"operation":"add","operands":[12345678901234567890.123456789, 0.000000001]}`,
			want: `{"operation":"add","operands":["12345678901234567890.123456789","0.000000001"],"result":"12345678901234567890.12345679"}`,
		},
		{
			name: "mixed string and number operands",
			body: `{"operation":"divide","operands":["10", 2]}`,
			want: `{"operation":"divide","operands":["10","2"],"result":"5"}`,
		},
		{
			name: "operands are echoed in canonical form",
			body: `{"operation":"add","operands":["2.50", 1e1]}`,
			want: `{"operation":"add","operands":["2.5","10"],"result":"12.5"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := post(newCalculationEngine(), tt.body)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.want, rec.Body.String())
		})
	}
}

// Not JSON / empty body / operands not an array / an operand that is neither string nor number → 400 invalid_request.
func TestCalculationHandler_Calculate_InvalidRequest(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "not JSON", body: `operation=add`},
		{name: "empty body", body: ``},
		{name: "truncated JSON", body: `{"operation":"add","operands":["1"`},
		{name: "trailing data after the object", body: `{"operation":"add","operands":["1","2"]} {}`},
		{name: "top-level array", body: `[]`},
		{name: "top-level null", body: `null`},
		{name: "operands not an array", body: `{"operation":"add","operands":"1,2"}`},
		{name: "operand is true", body: `{"operation":"add","operands":[true,"1"]}`},
		{name: "operand is null", body: `{"operation":"add","operands":[null,"1"]}`},
		{name: "operand is an object", body: `{"operation":"add","operands":[{},"1"]}`},
		{name: "operand is an array", body: `{"operation":"add","operands":[["1"],"1"]}`},
		{name: "operation not a string", body: `{"operation":1,"operands":["1","2"]}`},
		{name: "missing operation", body: `{"operands":["1","2"]}`},
		{name: "null operation", body: `{"operation":null,"operands":["1","2"]}`},
		{name: "missing operands", body: `{"operation":"add"}`},
		{name: "null operands", body: `{"operation":"add","operands":null}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := post(newCalculationEngine(), tt.body)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
			assert.JSONEq(t, invalidRequestBody, rec.Body.String())
		})
	}
}

func TestCalculationHandler_Calculate_DomainErrors(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		want       string
	}{
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
			name:       "empty operation reaches the domain as unsupported_operation",
			body:       `{"operation":"","operands":["1","2"]}`,
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
			name:       "empty operands array reaches the domain as invalid_operand_count",
			body:       `{"operation":"add","operands":[]}`,
			wantStatus: http.StatusBadRequest,
			want:       `{"error":{"code":"invalid_operand_count","message":"wrong number of operands for operation"}}`,
		},
		{
			name:       "operand too large is 400",
			body:       `{"operation":"add","operands":["1e1001","1"]}`,
			wantStatus: http.StatusBadRequest,
			want:       `{"error":{"code":"operand_too_large","message":"operand exceeds the supported size"}}`,
		},
		{
			name:       "division by zero is 422",
			body:       `{"operation":"divide","operands":["10","0"]}`,
			wantStatus: http.StatusUnprocessableEntity,
			want:       `{"error":{"code":"division_by_zero","message":"division by zero is undefined"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := post(newCalculationEngine(), tt.body)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.want, rec.Body.String())
		})
	}
}

// A non-domain error (here the service's ctx.Err()) is a generic 500; the cause is recorded for the access log only.
func TestCalculationHandler_Calculate_NonDomainErrorIsGeneric500(t *testing.T) {
	var recorded []string
	r := newCalculationEngine(func(c *gin.Context) {
		ctx, cancel := context.WithCancel(c.Request.Context())
		cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
		recorded = c.Errors.Errors()
	})

	rec := post(r, `{"operation":"add","operands":["1","2"]}`)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.JSONEq(t, `{"error":{"code":"internal_error","message":"an unexpected error occurred"}}`, rec.Body.String())
	assert.Equal(t, []string{context.Canceled.Error()}, recorded)
}

func TestCalculationHandler_Calculate_BodyOverLimitIs413(t *testing.T) {
	r := newCalculationEngine(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16)
	})

	rec := post(r, `{"operation":"add","operands":["1","2"]}`)

	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
	assert.JSONEq(t, `{"error":{"code":"payload_too_large","message":"request body exceeds 1 MiB"}}`, rec.Body.String())
}

func TestCalculationHandler_Operations_ListsOperationsInOrder(t *testing.T) {
	rec := httptest.NewRecorder()
	newCalculationEngine().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/operations", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"operations":[
		{"name":"add","arity":2,"symbol":"+"},
		{"name":"subtract","arity":2,"symbol":"−"},
		{"name":"multiply","arity":2,"symbol":"×"},
		{"name":"divide","arity":2,"symbol":"÷"}
	]}`, rec.Body.String())
}
