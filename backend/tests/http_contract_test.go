package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"sezzle.com/calculator/internal/platform/middleware"
)

func TestOperationsAPI_Get_ListsTheDomainRegistryInOrder(t *testing.T) {
	rec := newApp(t).serve(http.MethodGet, "/api/v1/operations", "")

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
func TestHTTPContract_EveryResponseIsJSONWithRequestID(t *testing.T) {
	a := newApp(t)
	a.GET("/test-panic", func(_ *gin.Context) { panic("secret panic detail") })

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
		want       string
	}{
		{name: "200 health", method: http.MethodGet, path: "/health", wantStatus: http.StatusOK, want: `{"status":"ok","version":"test"}`},
		{name: "400 invalid request", method: http.MethodPost, path: "/api/v1/calculations", body: `{`, wantStatus: http.StatusBadRequest, want: invalidRequestBody},
		{
			name: "404 unknown route", method: http.MethodGet, path: "/nope", wantStatus: http.StatusNotFound,
			want: `{"error":{"code":"not_found","message":"route not found"}}`,
		},
		{
			name: "405 wrong method", method: http.MethodGet, path: "/api/v1/calculations", wantStatus: http.StatusMethodNotAllowed,
			want: `{"error":{"code":"method_not_allowed","message":"method not allowed for this route"}}`,
		},
		{name: "500 panic", method: http.MethodGet, path: "/test-panic", wantStatus: http.StatusInternalServerError, want: internalErrorBody},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := a.serve(tt.method, tt.path, tt.body)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, wantContentType, rec.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.want, rec.Body.String())
			assert.NotEmpty(t, rec.Header().Get(middleware.RequestIDHeader))
		})
	}
}

func TestHTTPContract_WrongMethodAdvertisesAllowedMethods(t *testing.T) {
	rec := newApp(t).serve(http.MethodDelete, "/api/v1/calculations", "")

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.Equal(t, http.MethodPost, rec.Header().Get("Allow"))
}

func TestHTTPContract_EchoesIncomingRequestID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set(middleware.RequestIDHeader, "trace-42")

	rec := newApp(t).do(req)

	assert.Equal(t, "trace-42", rec.Header().Get(middleware.RequestIDHeader))
}
