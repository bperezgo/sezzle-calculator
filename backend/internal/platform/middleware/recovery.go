package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"sezzle.com/calculator/internal/platform/handler"
)

// Recovery turns a panic into the standard JSON 500; the panic value and stack are logged, never sent.
func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}
			// net/http uses ErrAbortHandler to drop the connection on purpose.
			if recovered == http.ErrAbortHandler { //nolint:errorlint // sentinel compared as a panic value, not a wrapped error
				panic(recovered)
			}
			logger.ErrorContext(c.Request.Context(), "panic recovered",
				slog.String("panic", fmt.Sprint(recovered)),
				slog.String("stack", string(debug.Stack())),
				slog.String("request_id", RequestIDFrom(c)),
			)
			handler.RespondError(c, http.StatusInternalServerError, handler.CodeInternalError, handler.MessageInternalError)
		}()
		c.Next()
	}
}
