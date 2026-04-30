package entity

import "fmt"

// Code is a domain-level error code; HTTP mapping lives in transport.
type Code string

const (
	ErrorCodeInternal     Code = "internal_error"
	ErrorCodeBadRequest   Code = "bad_request"
	ErrorCodeUnauthorized Code = "unauthorized"
	ErrorCodeForbidden    Code = "forbidden"
	ErrorCodeNotFound     Code = "not_found"
	ErrorCodeConflict     Code = "conflict"
)

type AppError struct {
	Code    Code   `json:"-"`
	Message string `json:"message"`
	Err     error  `json:"-"`
	Details any    `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Err }

func NewError(code Code, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func WrapError(err error, code Code, message string) *AppError {
	if app, ok := err.(*AppError); ok {
		return &AppError{Code: app.Code, Message: app.Message, Err: app.Err, Details: app.Details}
	}
	return &AppError{Code: code, Message: message, Err: err}
}

// Convenience constructors.
func ErrorBadRequest(msg string) *AppError   { return NewError(ErrorCodeBadRequest, msg) }
func ErrorUnauthorized(msg string) *AppError { return NewError(ErrorCodeUnauthorized, msg) }
func ErrorForbidden(msg string) *AppError    { return NewError(ErrorCodeForbidden, msg) }
func ErrorNotFound(msg string) *AppError     { return NewError(ErrorCodeNotFound, msg) }
func ErrorConflict(msg string) *AppError     { return NewError(ErrorCodeConflict, msg) }
func ErrorInternal(msg string) *AppError     { return NewError(ErrorCodeInternal, msg) }
