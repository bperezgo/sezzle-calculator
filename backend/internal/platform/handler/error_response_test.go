package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	domainerror "sezzle.com/calculator/internal/domain/error"
)

// The service never returns an unknown Kind today, so this branch is exercised directly.
func TestRespondDomainError_UnknownKindIsGeneric500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	respondDomainError(c, &domainerror.Error{Kind: "surprise", Code: "secret_code", Message: "secret detail"})

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.JSONEq(t, `{"error":{"code":"internal_error","message":"an unexpected error occurred"}}`, rec.Body.String())
	assert.Equal(t, []string{"secret detail"}, c.Errors.Errors())
}
