package apperror

import (
	"fmt"
	"net/http"
	"strings"
)

// AppError represents a domain error with HTTP status code and NestJS-compatible error envelope data.
type AppError struct {
	StatusCode int
	ErrorName  string
	Message    any // string or []string
	Err        error
}

func (e *AppError) Error() string {
	switch m := e.Message.(type) {
	case string:
		return m
	case []string:
		return strings.Join(m, "; ")
	default:
		return fmt.Sprintf("%v", e.Message)
	}
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a custom AppError.
func New(statusCode int, errorName string, message any) *AppError {
	return &AppError{
		StatusCode: statusCode,
		ErrorName:  errorName,
		Message:    message,
	}
}

// NotFound returns a 404 AppError.
func NotFound(message string) *AppError {
	if message == "" {
		message = "Requested record was not found."
	}
	return &AppError{
		StatusCode: http.StatusNotFound,
		ErrorName:  "NotFound",
		Message:    message,
	}
}

// Conflict returns a 409 AppError.
func Conflict(message string) *AppError {
	if message == "" {
		message = "A record with this unique value already exists."
	}
	return &AppError{
		StatusCode: http.StatusConflict,
		ErrorName:  "Conflict",
		Message:    message,
	}
}

// BadRequest returns a 400 AppError.
func BadRequest(message string) *AppError {
	return &AppError{
		StatusCode: http.StatusBadRequest,
		ErrorName:  "BadRequest",
		Message:    message,
	}
}

// Validation returns a 400 AppError with an array of validation failure strings.
func Validation(messages []string) *AppError {
	return &AppError{
		StatusCode: http.StatusBadRequest,
		ErrorName:  "BadRequest",
		Message:    messages,
	}
}

// Unauthorized returns a 401 AppError.
func Unauthorized(message string) *AppError {
	if message == "" {
		message = "Unauthorized"
	}
	return &AppError{
		StatusCode: http.StatusUnauthorized,
		ErrorName:  "Unauthorized",
		Message:    message,
	}
}

// Forbidden returns a 403 AppError.
func Forbidden(message string) *AppError {
	if message == "" {
		message = "Forbidden resource"
	}
	return &AppError{
		StatusCode: http.StatusForbidden,
		ErrorName:  "Forbidden",
		Message:    message,
	}
}

// Internal returns a 500 AppError.
func Internal(message string, originalErr ...error) *AppError {
	if message == "" {
		message = "An unexpected internal error occurred."
	}
	var err error
	if len(originalErr) > 0 {
		err = originalErr[0]
	}
	return &AppError{
		StatusCode: http.StatusInternalServerError,
		ErrorName:  "InternalServerError",
		Message:    message,
		Err:        err,
	}
}
