package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"sezzle.com/calculator/internal/domain/dto"
	"sezzle.com/calculator/internal/domain/service"
)

type CalculationHandler struct {
	calculator *service.CalculatorService
}

func NewCalculationHandler(calculator *service.CalculatorService) *CalculationHandler {
	return &CalculationHandler{calculator: calculator}
}

// Pointers distinguish a missing or null field (invalid_request) from an empty one (left to the domain).
type calculationRequest struct {
	Operation *string    `json:"operation"`
	Operands  *[]operand `json:"operands"`
}

// operand accepts a JSON string or number and keeps its raw text, so a number never passes through float64.
type operand string

var errOperandType = errors.New("operand must be a string or a number")

// encoding/json hands over exactly one valid literal, so its first byte identifies the type.
func (o *operand) UnmarshalJSON(data []byte) error {
	switch first := data[0]; {
	case first == '"':
		return json.Unmarshal(data, (*string)(o))
	case first == '-' || (first >= '0' && first <= '9'):
		*o = operand(data)
		return nil
	default:
		return errOperandType
	}
}

type calculationResponse struct {
	Operation string   `json:"operation"`
	Operands  []string `json:"operands"`
	Result    string   `json:"result"`
}

type operationResponse struct {
	Name   string `json:"name"`
	Arity  int    `json:"arity"`
	Symbol string `json:"symbol"`
}

type operationsResponse struct {
	Operations []operationResponse `json:"operations"`
}

func (h *CalculationHandler) Calculate(c *gin.Context) {
	req, err := decodeCalculationRequest(c.Request.Body)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			RespondError(c, http.StatusRequestEntityTooLarge, CodePayloadTooLarge, MessagePayloadTooLarge)
			return
		}
		RespondError(c, http.StatusBadRequest, CodeInvalidRequest, MessageInvalidRequest)
		return
	}

	result, err := h.calculator.Calculate(c.Request.Context(), req)
	if err != nil {
		respondDomainError(c, err)
		return
	}

	c.JSON(http.StatusOK, calculationResponse{
		Operation: result.Operation,
		Operands:  result.Operands,
		Result:    result.Result,
	})
}

func (h *CalculationHandler) Operations(c *gin.Context) {
	infos := h.calculator.Operations(c.Request.Context())
	operations := make([]operationResponse, 0, len(infos))
	for _, info := range infos {
		operations = append(operations, operationResponse{Name: info.Name, Arity: info.Arity, Symbol: info.Symbol})
	}
	c.JSON(http.StatusOK, operationsResponse{Operations: operations})
}

var errMissingField = errors.New("operation and operands are required")

func decodeCalculationRequest(body io.Reader) (dto.CalculationRequest, error) {
	decoder := json.NewDecoder(body)
	var req calculationRequest
	if err := decoder.Decode(&req); err != nil {
		return dto.CalculationRequest{}, err
	}
	// Exactly one JSON value: anything after the object makes the request ambiguous.
	if err := decoder.Decode(&json.RawMessage{}); !errors.Is(err, io.EOF) {
		if err == nil {
			err = errors.New("unexpected data after the JSON object")
		}
		return dto.CalculationRequest{}, err
	}
	if req.Operation == nil || req.Operands == nil {
		return dto.CalculationRequest{}, errMissingField
	}

	operands := make([]string, 0, len(*req.Operands))
	for _, o := range *req.Operands {
		operands = append(operands, string(o))
	}
	return dto.CalculationRequest{Operation: *req.Operation, Operands: operands}, nil
}
