package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"m-data-storage/api/dto"
	internalerrors "m-data-storage/internal/errors"
	"m-data-storage/internal/infrastructure/logger"
)

// ErrorHandlerMiddleware handles application errors
type ErrorHandlerMiddleware struct {
	logger *logger.Logger
}

// NewErrorHandlerMiddleware creates a new ErrorHandlerMiddleware instance
func NewErrorHandlerMiddleware(logger *logger.Logger) *ErrorHandlerMiddleware {
	return &ErrorHandlerMiddleware{
		logger: logger,
	}
}

// ErrorHandler handles errors and converts them to proper HTTP responses
func (m *ErrorHandlerMiddleware) ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a custom response writer to capture errors
		ew := &errorResponseWriter{
			ResponseWriter: w,
			request:        r,
			logger:         m.logger,
		}

		next.ServeHTTP(ew, r)
	})
}

// errorResponseWriter wraps http.ResponseWriter to handle errors
type errorResponseWriter struct {
	http.ResponseWriter
	request *http.Request
	logger  *logger.Logger
	written bool
}

// WriteError writes an error response
func (ew *errorResponseWriter) WriteError(err error) {
	if ew.written {
		return
	}

	ew.written = true

	var appErr *internalerrors.AppError
	var statusCode int
	var response dto.Response

	// Check if it's an AppError
	if errors.As(err, &appErr) {
		statusCode = appErr.StatusCode
		response = dto.Response{
			Success: false,
			Error: &dto.Error{
				Code:    appErr.Code,
				Message: appErr.Message,
				Details: appErr.Details,
			},
		}

		// Log error with appropriate level
		if statusCode >= 500 {
			ew.logger.WithError(err).WithFields(map[string]interface{}{
				"method":     ew.request.Method,
				"path":       ew.request.URL.Path,
				"remote_ip":  ew.request.RemoteAddr,
				"user_agent": ew.request.UserAgent(),
			}).Error("Internal server error")
		} else {
			ew.logger.WithError(err).WithFields(map[string]interface{}{
				"method":    ew.request.Method,
				"path":      ew.request.URL.Path,
				"remote_ip": ew.request.RemoteAddr,
			}).Warn("Client error")
		}
	} else {
		// Handle standard errors
		statusCode = http.StatusInternalServerError
		response = dto.Response{
			Success: false,
			Error: &dto.Error{
				Code:    "INTERNAL_SERVER_ERROR",
				Message: "Internal server error occurred",
			},
		}

		ew.logger.WithError(err).WithFields(map[string]interface{}{
			"method":     ew.request.Method,
			"path":       ew.request.URL.Path,
			"remote_ip":  ew.request.RemoteAddr,
			"user_agent": ew.request.UserAgent(),
		}).Error("Unhandled error")
	}

	ew.Header().Set("Content-Type", "application/json")
	ew.WriteHeader(statusCode)
	json.NewEncoder(ew).Encode(response)
}

// WriteHeader captures the status code
func (ew *errorResponseWriter) WriteHeader(statusCode int) {
	ew.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the response body
func (ew *errorResponseWriter) Write(data []byte) (int, error) {
	ew.written = true
	return ew.ResponseWriter.Write(data)
}

// ErrorFromContext extracts error from context and writes it
func ErrorFromContext(ctx context.Context, w http.ResponseWriter) {
	if err := ctx.Err(); err != nil {
		if ew, ok := w.(*errorResponseWriter); ok {
			ew.WriteError(err)
		}
	}
}
