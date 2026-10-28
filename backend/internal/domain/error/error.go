// Package error holds the domain's categorised errors. Codes and messages are API contract.
package error

import "errors"

// Kind categorises an error so adapters can map it (e.g. to an HTTP status) without knowing codes.
type Kind string

const (
	// KindInvalidInput: the request itself is unusable.
	KindInvalidInput Kind = "invalid_input"
	// KindMath: the request is well-formed but mathematically undefined.
	KindMath Kind = "math"
)

type Code string

const (
	CodeUnsupportedOperation Code = "unsupported_operation"
	CodeInvalidOperand       Code = "invalid_operand"
	CodeInvalidOperandCount  Code = "invalid_operand_count"
	CodeOperandTooLarge      Code = "operand_too_large"
	CodeDivisionByZero       Code = "division_by_zero"
	CodeNegativeSquareRoot   Code = "negative_square_root"
	CodeUndefinedPower       Code = "undefined_power"
	CodeResultTooLarge       Code = "result_too_large"
)

const (
	MessageUnsupportedOperation = "operation is not supported"
	MessageInvalidOperand       = "operand is not a valid decimal number"
	MessageInvalidOperandCount  = "wrong number of operands for operation"
	MessageOperandTooLarge      = "operand exceeds the supported size"
	MessageDivisionByZero       = "division by zero is undefined"
	MessageNegativeSquareRoot   = "square root of a negative number is undefined"
	MessageUndefinedPower       = "power is undefined for these operands"
	MessageResultTooLarge       = "result exceeds the supported size"
)

var (
	ErrUnsupportedOperation = &Error{Kind: KindInvalidInput, Code: CodeUnsupportedOperation, Message: MessageUnsupportedOperation}
	ErrInvalidOperand       = &Error{Kind: KindInvalidInput, Code: CodeInvalidOperand, Message: MessageInvalidOperand}
	ErrInvalidOperandCount  = &Error{Kind: KindInvalidInput, Code: CodeInvalidOperandCount, Message: MessageInvalidOperandCount}
	ErrOperandTooLarge      = &Error{Kind: KindInvalidInput, Code: CodeOperandTooLarge, Message: MessageOperandTooLarge}
	ErrDivisionByZero       = &Error{Kind: KindMath, Code: CodeDivisionByZero, Message: MessageDivisionByZero}
	ErrNegativeSquareRoot   = &Error{Kind: KindMath, Code: CodeNegativeSquareRoot, Message: MessageNegativeSquareRoot}
	ErrUndefinedPower       = &Error{Kind: KindMath, Code: CodeUndefinedPower, Message: MessageUndefinedPower}
	ErrResultTooLarge       = &Error{Kind: KindMath, Code: CodeResultTooLarge, Message: MessageResultTooLarge}
)

type Error struct {
	Kind    Kind
	Code    Code
	Message string
	Fields  map[string]string
}

func (e *Error) Error() string {
	return e.Message
}

// Is matches by code, so a sentinel matches any copy carrying request-specific fields.
func (e *Error) Is(target error) bool {
	var other *Error
	if !errors.As(target, &other) {
		return false
	}
	return e.Code == other.Code
}

// WithFields returns a copy of the error with request details; sentinels are never mutated.
func (e *Error) WithFields(fields map[string]string) *Error {
	clone := *e
	clone.Fields = fields
	return &clone
}
