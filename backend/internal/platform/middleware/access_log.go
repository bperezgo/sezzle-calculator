package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"sezzle.com/calculator/internal/platform/handler"
)

// AccessLog logs one line per request with the API error code, if any; server errors are logged at ERROR with the causes handlers recorded via c.Error.
func AccessLog(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		attrs := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", status),
			slog.Float64("duration_ms", float64(time.Since(start).Microseconds())/1000),
			slog.Int("bytes", c.Writer.Size()),
			slog.String("request_id", RequestIDFrom(c)),
		}
		level := slog.LevelInfo
		if status >= http.StatusInternalServerError {
			level = slog.LevelError
		}
		if code := handler.ErrorCodeFrom(c); code != "" {
			attrs = append(attrs, slog.String("error_code", code))
		}
		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("error", c.Errors.String()))
		}
		logger.LogAttrs(c.Request.Context(), level, "request", attrs...)
	}
}
