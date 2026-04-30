package entity

import (
	"fmt"
)

// Code type (string for readability)
type Code string

const (
	// domain/application codes
	ErrorCodeInternal     Code = "internal_error"
	ErrorCodeNotFound     Code = "not_found"
	ErrorCodeUnauthorized Code = "unauthorized"
	ErrorCodeBadRequest   Code = "bad_request"
)

type AppError struct {
	Code    Code   `json:"code"`
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

func NewError(code Code, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

func WrapError(err error, code Code, message string) *AppError {
	if app, ok := err.(*AppError); ok {
		return &AppError{Code: app.Code, Message: app.Message, Err: app.Err, Details: app.Details}
	}
	return &AppError{Code: code, Message: message, Err: err}
}

// Convenience constructors
func ErrorNotFound(msg string) *AppError     { return NewError(ErrorCodeNotFound, msg) }
func ErrorUnauthorized(msg string) *AppError { return NewError(ErrorCodeUnauthorized, msg) }
func ErrorInternal(msg string) *AppError     { return NewError(ErrorCodeInternal, msg) }
func ErrorBadRequest(msg string) *AppError   { return NewError(ErrorCodeBadRequest, msg) }
