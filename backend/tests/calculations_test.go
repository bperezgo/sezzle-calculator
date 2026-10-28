package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculationsAPI_Post_Success(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "add 2 3 is 200 result 5",
			body: `{"operation":"add","operands":["2","3"]}`,
			want: `{"operation":"add","operands":["2","3"],"result":"5"}`,
		},
		{
			name: "JSON numbers 0.1 + 0.2 is exactly 0.3",
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
			rec := newApp(t).calculate(tt.body)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, wantContentType, rec.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.want, rec.Body.String())
		})
	}
}

// Not JSON / empty body / operands not an array / an operand that is neither string nor number → 400 invalid_request.
func TestCalculationsAPI_Post_InvalidRequest(t *testing.T) {
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
			rec := newApp(t).calculate(tt.body)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, wantContentType, rec.Header().Get("Content-Type"))
			assert.JSONEq(t, invalidRequestBody, rec.Body.String())
		})
	}
}

func TestCalculationsAPI_Post_DomainErrors(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		want       string
	}{
		{
			name:       "divide by zero is 422 division_by_zero",
			body:       `{"operation":"divide","operands":["10","0"]}`,
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
			name:       "empty operation is 400 unsupported_operation",
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
			name:       "empty operands array is 400 invalid_operand_count",
			body:       `{"operation":"add","operands":[]}`,
			wantStatus: http.StatusBadRequest,
			want:       `{"error":{"code":"invalid_operand_count","message":"wrong number of operands for operation"}}`,
		},
		{
			name:       "operand too large is 400 operand_too_large",
			body:       `{"operation":"add","operands":["1e1001","1"]}`,
			wantStatus: http.StatusBadRequest,
			want:       `{"error":{"code":"operand_too_large","message":"operand exceeds the supported size"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := newApp(t).calculate(tt.body)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, wantContentType, rec.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.want, rec.Body.String())
		})
	}
}

func TestCalculationsAPI_Post_BodyOver1MiBIs413(t *testing.T) {
	body := `{"operation":"add","operands":["` + strings.Repeat("1", 1<<20) + `","1"]}`

	rec := newApp(t).calculate(body)

	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
	assert.Equal(t, wantContentType, rec.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"error":{"code":"payload_too_large","message":"request body exceeds 1 MiB"}}`, rec.Body.String())
}

// Unknown errors (here the service's ctx.Err()) → 500 with a generic message; the cause is logged, not shown.
func TestCalculationsAPI_Post_UnexpectedErrorIsGeneric500AndLogged(t *testing.T) {
	a := newApp(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/calculations",
		strings.NewReader(`{"operation":"add","operands":["1","2"]}`))

	rec := a.do(req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.JSONEq(t, internalErrorBody, rec.Body.String())
	assert.Contains(t, a.logs.String(), "context canceled")
}
