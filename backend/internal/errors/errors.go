package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Common errors
var (
	// ErrNotFound - error when entity is not found
	ErrNotFound = errors.New("entity not found")

	// ErrInvalidInput - input validation error
	ErrInvalidInput = errors.New("invalid input")

	// ErrUnauthorized - authorization error
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden - access error
	ErrForbidden = errors.New("forbidden")

	// ErrConflict - data conflict error
	ErrConflict = errors.New("conflict")

	// ErrInternalServer - internal server error
	ErrInternalServer = errors.New("internal server error")

	// ErrServiceUnavailable - service unavailable
	ErrServiceUnavailable = errors.New("service unavailable")

	// ErrTimeout - timeout error
	ErrTimeout = errors.New("timeout")

	// ErrRateLimited - rate limit exceeded error
	ErrRateLimited = errors.New("rate limited")
)

// AppError represents application error with additional information
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
	StatusCode int    `json:"-"`
	Err        error  `json:"-"`
}

// Error implements error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates new application error
func NewAppError(code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// WrapError wraps error in AppError
func WrapError(err error, code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

// WithDetails adds details to error
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// Predefined application errors
var (
	ErrValidationFailed = NewAppError("VALIDATION_FAILED", "Validation failed", http.StatusBadRequest)
	ErrBrokerNotFound   = NewAppError("BROKER_NOT_FOUND", "Broker not found", http.StatusNotFound)
	ErrBrokerOffline    = NewAppError("BROKER_OFFLINE", "Broker is offline", http.StatusServiceUnavailable)
	ErrInvalidSymbol    = NewAppError("INVALID_SYMBOL", "Invalid symbol", http.StatusBadRequest)
	ErrInvalidTimeframe = NewAppError("INVALID_TIMEFRAME", "Invalid timeframe", http.StatusBadRequest)
	ErrDataNotFound     = NewAppError("DATA_NOT_FOUND", "Data not found", http.StatusNotFound)
	ErrStorageError     = NewAppError("STORAGE_ERROR", "Storage error", http.StatusInternalServerError)
	ErrConnectionError  = NewAppError("CONNECTION_ERROR", "Connection error", http.StatusServiceUnavailable)
)
