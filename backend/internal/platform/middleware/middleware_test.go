package middleware_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sezzle.com/calculator/internal/platform/middleware"
)

func newEngine(mw ...gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(mw...)
	return r
}

func TestRequestID_EchoesValidIncomingID(t *testing.T) {
	r := newEngine(middleware.RequestID())
	var seen string
	r.GET("/", func(c *gin.Context) { seen = middleware.RequestIDFrom(c) })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(middleware.RequestIDHeader, "abc-123_DEF.4")
	r.ServeHTTP(rec, req)

	assert.Equal(t, "abc-123_DEF.4", rec.Header().Get(middleware.RequestIDHeader))
	assert.Equal(t, "abc-123_DEF.4", seen)
}

func TestRequestID_GeneratesIDWhenMissingOrUnsafe(t *testing.T) {
	tests := []struct {
		name     string
		incoming string
	}{
		{name: "missing", incoming: ""},
		{name: "contains spaces and newlines", incoming: "evil\nlog line"},
		{name: "too long", incoming: strings.Repeat("a", 129)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newEngine(middleware.RequestID())
			r.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.incoming != "" {
				req.Header.Set(middleware.RequestIDHeader, tt.incoming)
			}
			r.ServeHTTP(rec, req)

			got := rec.Header().Get(middleware.RequestIDHeader)
			assert.NotEmpty(t, got)
			assert.NotEqual(t, tt.incoming, got)
		})
	}
}

func TestAccessLog_LogsRequestAtInfo(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	r := newEngine(middleware.RequestID(), middleware.AccessLog(logger))
	r.GET("/thing", func(c *gin.Context) { c.Status(http.StatusTeapot) })

	req := httptest.NewRequest(http.MethodGet, "/thing", nil)
	req.Header.Set(middleware.RequestIDHeader, "req-1")
	r.ServeHTTP(httptest.NewRecorder(), req)

	entry := decodeLog(t, &buf)
	assert.Equal(t, "INFO", entry["level"])
	assert.Equal(t, "request", entry["msg"])
	assert.Equal(t, "GET", entry["method"])
	assert.Equal(t, "/thing", entry["path"])
	assert.InDelta(t, http.StatusTeapot, entry["status"], 0)
	assert.Equal(t, "req-1", entry["request_id"])
	assert.Contains(t, entry, "duration_ms")
	assert.NotContains(t, entry, "error")
}

func TestAccessLog_LogsServerErrorsWithTheirCause(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	r := newEngine(middleware.AccessLog(logger))
	r.GET("/fail", func(c *gin.Context) {
		_ = c.Error(errors.New("db exploded"))
		c.Status(http.StatusInternalServerError)
	})

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/fail", nil))

	entry := decodeLog(t, &buf)
	assert.Equal(t, "ERROR", entry["level"])
	assert.Contains(t, entry["error"], "db exploded")
}

func TestRecovery_PanicBecomesGenericJSON500AndIsLogged(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	r := newEngine(middleware.RequestID(), middleware.Recovery(logger))
	r.GET("/panic", func(_ *gin.Context) { panic("secret panic detail") })

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/panic", nil))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"error":{"code":"internal_error","message":"an unexpected error occurred"}}`, rec.Body.String())
	assert.NotEmpty(t, rec.Header().Get(middleware.RequestIDHeader))

	entry := decodeLog(t, &buf)
	assert.Equal(t, "ERROR", entry["level"])
	assert.Equal(t, "panic recovered", entry["msg"])
	assert.Contains(t, entry["panic"], "secret panic detail")
	assert.Contains(t, entry, "stack")
}

func TestRecovery_RepanicsAbortHandler(t *testing.T) {
	r := newEngine(middleware.Recovery(slog.New(slog.NewTextHandler(io.Discard, nil))))
	r.GET("/abort", func(_ *gin.Context) { panic(http.ErrAbortHandler) })

	assert.PanicsWithValue(t, http.ErrAbortHandler, func() {
		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/abort", nil))
	})
}

func TestBodyLimit_RejectsReadsBeyondTheLimit(t *testing.T) {
	r := newEngine(middleware.BodyLimit(4))
	var readErr error
	r.POST("/", func(c *gin.Context) { _, readErr = io.ReadAll(c.Request.Body) })

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", strings.NewReader("12345")))

	var maxBytesErr *http.MaxBytesError
	assert.ErrorAs(t, readErr, &maxBytesErr)
}

func decodeLog(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	var entry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
	return entry
}
