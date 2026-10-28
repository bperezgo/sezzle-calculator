package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	domainerror "sezzle.com/calculator/internal/domain/error"
)

// Transport-level error codes and messages; like the domain's, they are API contract.
const (
	CodeInvalidRequest   = "invalid_request"
	CodePayloadTooLarge  = "payload_too_large"
	CodeNotFound         = "not_found"
	CodeMethodNotAllowed = "method_not_allowed"
	CodeInternalError    = "internal_error"
)

const (
	MessageInvalidRequest   = "request body must be a JSON object with an operation string and an operands array of numbers or numeric strings"
	MessagePayloadTooLarge  = "request body exceeds 1 MiB"
	MessageNotFound         = "route not found"
	MessageMethodNotAllowed = "method not allowed for this route"
	MessageInternalError    = "an unexpected error occurred"
)

type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// RespondError aborts the request with the standard error envelope.
func RespondError(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, errorResponse{Error: errorBody{Code: code, Message: message}})
}

// RespondInternalError hides the cause from the client and records it on the context for the access log.
func RespondInternalError(c *gin.Context, cause error) {
	_ = c.Error(cause)
	RespondError(c, http.StatusInternalServerError, CodeInternalError, MessageInternalError)
}

// respondDomainError is the single domain-error → HTTP status mapper.
func respondDomainError(c *gin.Context, err error) {
	var domainErr *domainerror.Error
	if !errors.As(err, &domainErr) {
		RespondInternalError(c, err)
		return
	}
	switch domainErr.Kind {
	case domainerror.KindInvalidInput:
		RespondError(c, http.StatusBadRequest, string(domainErr.Code), domainErr.Message)
	case domainerror.KindMath:
		RespondError(c, http.StatusUnprocessableEntity, string(domainErr.Code), domainErr.Message)
	default:
		RespondInternalError(c, err)
	}
}

func NotFound(c *gin.Context) {
	RespondError(c, http.StatusNotFound, CodeNotFound, MessageNotFound)
}

func MethodNotAllowed(c *gin.Context) {
	RespondError(c, http.StatusMethodNotAllowed, CodeMethodNotAllowed, MessageMethodNotAllowed)
}
