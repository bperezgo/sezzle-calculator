package handler_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"sezzle.com/calculator/internal/domain/dto"
	domainerror "sezzle.com/calculator/internal/domain/error"
	"sezzle.com/calculator/internal/domain/ports/mocks"
	"sezzle.com/calculator/internal/platform/handler"
)

func newCalculationEngine(calc *mocks.MockCalculator) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := handler.NewCalculationHandler(calc)
	r := gin.New()
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
		name    string
		body    string
		wantReq dto.CalculationRequest
		result  dto.CalculationResult
		want    string
	}{
		{
			name:    "string operands are passed through",
			body:    `{"operation":"add","operands":["2","3"]}`,
			wantReq: dto.CalculationRequest{Operation: "add", Operands: []string{"2", "3"}},
			result:  dto.CalculationResult{Operation: "add", Operands: []string{"2", "3"}, Result: "5"},
			want:    `{"operation":"add","operands":["2","3"],"result":"5"}`,
		},
		{
			name:    "JSON number operands are read from their raw text, never float64",
			body:    `{"operation":"add","operands":[0.1, 0.2]}`,
			wantReq: dto.CalculationRequest{Operation: "add", Operands: []string{"0.1", "0.2"}},
			result:  dto.CalculationResult{Operation: "add", Operands: []string{"0.1", "0.2"}, Result: "0.3"},
			want:    `{"operation":"add","operands":["0.1","0.2"],"result":"0.3"}`,
		},
		{
			name:    "number operands keep precision beyond float64",
			body:    `{"operation":"add","operands":[12345678901234567890.123456789, 1e-30]}`,
			wantReq: dto.CalculationRequest{Operation: "add", Operands: []string{"12345678901234567890.123456789", "1e-30"}},
			result:  dto.CalculationResult{Operation: "add", Operands: []string{"x", "y"}, Result: "z"},
			want:    `{"operation":"add","operands":["x","y"],"result":"z"}`,
		},
		{
			name:    "mixed string and number operands",
			body:    `{"operation":"divide","operands":["10", 2]}`,
			wantReq: dto.CalculationRequest{Operation: "divide", Operands: []string{"10", "2"}},
			result:  dto.CalculationResult{Operation: "divide", Operands: []string{"10", "2"}, Result: "5"},
			want:    `{"operation":"divide","operands":["10","2"],"result":"5"}`,
		},
		{
			name:    "empty operands array reaches the domain for its arity check",
			body:    `{"operation":"add","operands":[]}`,
			wantReq: dto.CalculationRequest{Operation: "add", Operands: []string{}},
			result:  dto.CalculationResult{Operation: "add", Operands: []string{}, Result: "0"},
			want:    `{"operation":"add","operands":[],"result":"0"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := mocks.NewMockCalculator(t)
			calc.EXPECT().Calculate(mock.Anything, tt.wantReq).Return(tt.result, nil)

			rec := post(newCalculationEngine(calc), tt.body)

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
			calc := mocks.NewMockCalculator(t)

			rec := post(newCalculationEngine(calc), tt.body)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
			assert.JSONEq(t, `{"error":{"code":"invalid_request","message":"request body must be a JSON object with an operation string and an operands array of numbers or numeric strings"}}`, rec.Body.String())
		})
	}
}

func TestCalculationHandler_Calculate_DomainErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		want       string
	}{
		{
			name:       "invalid operand is 400",
			err:        domainerror.ErrInvalidOperand.WithFields(map[string]string{"operand": "a"}),
			wantStatus: http.StatusBadRequest,
			want:       `{"error":{"code":"invalid_operand","message":"operand is not a valid decimal number"}}`,
		},
		{
			name:       "unsupported operation is 400",
			err:        domainerror.ErrUnsupportedOperation,
			wantStatus: http.StatusBadRequest,
			want:       `{"error":{"code":"unsupported_operation","message":"operation is not supported"}}`,
		},
		{
			name:       "wrong arity is 400",
			err:        domainerror.ErrInvalidOperandCount,
			wantStatus: http.StatusBadRequest,
			want:       `{"error":{"code":"invalid_operand_count","message":"wrong number of operands for operation"}}`,
		},
		{
			name:       "operand too large is 400",
			err:        domainerror.ErrOperandTooLarge,
			wantStatus: http.StatusBadRequest,
			want:       `{"error":{"code":"operand_too_large","message":"operand exceeds the supported size"}}`,
		},
		{
			name:       "division by zero is 422",
			err:        domainerror.ErrDivisionByZero,
			wantStatus: http.StatusUnprocessableEntity,
			want:       `{"error":{"code":"division_by_zero","message":"division by zero is undefined"}}`,
		},
		{
			name:       "domain error with an unknown kind is a generic 500",
			err:        &domainerror.Error{Kind: "surprise", Code: "secret_code", Message: "secret detail"},
			wantStatus: http.StatusInternalServerError,
			want:       `{"error":{"code":"internal_error","message":"an unexpected error occurred"}}`,
		},
		{
			name:       "non-domain error is a generic 500 that hides the cause",
			err:        errors.New("secret detail"),
			wantStatus: http.StatusInternalServerError,
			want:       `{"error":{"code":"internal_error","message":"an unexpected error occurred"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := mocks.NewMockCalculator(t)
			calc.EXPECT().Calculate(mock.Anything, mock.Anything).Return(dto.CalculationResult{}, tt.err)

			rec := post(newCalculationEngine(calc), `{"operation":"add","operands":["1","2"]}`)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.want, rec.Body.String())
		})
	}
}

func TestCalculationHandler_Calculate_InternalErrorIsRecordedForTheAccessLog(t *testing.T) {
	calc := mocks.NewMockCalculator(t)
	calc.EXPECT().Calculate(mock.Anything, mock.Anything).Return(dto.CalculationResult{}, errors.New("boom"))

	var recorded []string
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Next()
		recorded = c.Errors.Errors()
	})
	r.POST("/calculations", handler.NewCalculationHandler(calc).Calculate)

	rec := post(r, `{"operation":"add","operands":["1","2"]}`)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, []string{"boom"}, recorded)
}

func TestCalculationHandler_Operations_ListsOperationsInOrder(t *testing.T) {
	calc := mocks.NewMockCalculator(t)
	calc.EXPECT().Operations(mock.Anything).Return([]dto.OperationInfo{
		{Name: "add", Arity: 2, Symbol: "+"},
		{Name: "divide", Arity: 2, Symbol: "÷"},
	})

	rec := httptest.NewRecorder()
	newCalculationEngine(calc).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/operations", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"operations":[{"name":"add","arity":2,"symbol":"+"},{"name":"divide","arity":2,"symbol":"÷"}]}`, rec.Body.String())
}

func TestCalculationHandler_Calculate_BodyOverLimitIs413(t *testing.T) {
	calc := mocks.NewMockCalculator(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16)
	})
	r.POST("/calculations", handler.NewCalculationHandler(calc).Calculate)

	rec := post(r, `{"operation":"add","operands":["1","2"]}`)

	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
	assert.JSONEq(t, `{"error":{"code":"payload_too_large","message":"request body exceeds 1 MiB"}}`, rec.Body.String())
}
