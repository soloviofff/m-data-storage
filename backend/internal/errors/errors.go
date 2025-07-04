package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Common errors
var (
	// ErrNotFound - ошибка, когда сущность не найдена
	ErrNotFound = errors.New("entity not found")

	// ErrInvalidInput - ошибка валидации входных данных
	ErrInvalidInput = errors.New("invalid input")

	// ErrUnauthorized - ошибка авторизации
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden - ошибка доступа
	ErrForbidden = errors.New("forbidden")

	// ErrConflict - ошибка конфликта данных
	ErrConflict = errors.New("conflict")

	// ErrInternalServer - внутренняя ошибка сервера
	ErrInternalServer = errors.New("internal server error")

	// ErrServiceUnavailable - сервис недоступен
	ErrServiceUnavailable = errors.New("service unavailable")

	// ErrTimeout - ошибка таймаута
	ErrTimeout = errors.New("timeout")

	// ErrRateLimited - ошибка превышения лимита запросов
	ErrRateLimited = errors.New("rate limited")
)

// AppError представляет ошибку приложения с дополнительной информацией
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
	StatusCode int    `json:"-"`
	Err        error  `json:"-"`
}

// Error реализует интерфейс error
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap возвращает обернутую ошибку
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError создает новую ошибку приложения
func NewAppError(code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// WrapError оборачивает ошибку в AppError
func WrapError(err error, code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

// WithDetails добавляет детали к ошибке
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// Предопределенные ошибки приложения
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
