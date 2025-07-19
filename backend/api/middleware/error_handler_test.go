package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"m-data-storage/api/dto"
	internalerrors "m-data-storage/internal/errors"
	"m-data-storage/internal/infrastructure/config"
	"m-data-storage/internal/infrastructure/logger"
)

func TestNewErrorHandlerMiddleware(t *testing.T) {
	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	realLogger, err := logger.New(cfg)
	assert.NoError(t, err)

	middleware := NewErrorHandlerMiddleware(realLogger)

	assert.NotNil(t, middleware)
	assert.Equal(t, realLogger, middleware.logger)
}

func TestErrorHandlerMiddleware_ErrorHandler(t *testing.T) {
	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	realLogger, err := logger.New(cfg)
	assert.NoError(t, err)

	middleware := NewErrorHandlerMiddleware(realLogger)

	// Test normal request (no error)
	t.Run("normal request", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

		errorHandler := middleware.ErrorHandler(handler)

		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		errorHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "success", rr.Body.String())
	})
}

func TestErrorResponseWriter_WriteError(t *testing.T) {
	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	realLogger, err := logger.New(cfg)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedCode   string
	}{
		{
			name: "app error - validation",
			err: internalerrors.NewAppError(
				"INVALID_INPUT",
				"Invalid input provided",
				http.StatusBadRequest,
			).WithDetails("field: name"),
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_INPUT",
		},
		{
			name: "app error - not found",
			err: internalerrors.NewAppError(
				"RESOURCE_NOT_FOUND",
				"Resource not found",
				http.StatusNotFound,
			),
			expectedStatus: http.StatusNotFound,
			expectedCode:   "RESOURCE_NOT_FOUND",
		},
		{
			name: "app error - internal",
			err: internalerrors.NewAppError(
				"DATABASE_ERROR",
				"Database connection failed",
				http.StatusInternalServerError,
			),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "DATABASE_ERROR",
		},
		{
			name:           "standard error",
			err:            errors.New("standard error"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "INTERNAL_SERVER_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			ew := &errorResponseWriter{
				ResponseWriter: rr,
				request:        req,
				logger:         realLogger,
			}

			ew.WriteError(tt.err)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

			var response dto.Response
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.False(t, response.Success)
			assert.NotNil(t, response.Error)
			assert.Equal(t, tt.expectedCode, response.Error.Code)
		})
	}
}

func TestErrorResponseWriter_WriteError_AlreadyWritten(t *testing.T) {
	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	realLogger, err := logger.New(cfg)
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	ew := &errorResponseWriter{
		ResponseWriter: rr,
		request:        req,
		logger:         realLogger,
		written:        true, // Already written
	}

	ew.WriteError(errors.New("test error"))

	// Should not write anything since already written
	assert.Equal(t, http.StatusOK, rr.Code) // Default status
	assert.Empty(t, rr.Body.String())
}

func TestErrorResponseWriter_Write(t *testing.T) {
	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	realLogger, err := logger.New(cfg)
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	ew := &errorResponseWriter{
		ResponseWriter: rr,
		request:        req,
		logger:         realLogger,
	}

	data := []byte("test data")
	n, err := ew.Write(data)

	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.True(t, ew.written)
	assert.Equal(t, "test data", rr.Body.String())
}

func TestErrorFromContext(t *testing.T) {
	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	realLogger, err := logger.New(cfg)
	assert.NoError(t, err)

	t.Run("context with error", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		ew := &errorResponseWriter{
			ResponseWriter: rr,
			request:        req,
			logger:         realLogger,
		}

		// Create context with timeout that's already expired
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		ErrorFromContext(ctx, ew)

		// Should write error response
		assert.True(t, ew.written)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("context without error", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		ew := &errorResponseWriter{
			ResponseWriter: rr,
			request:        req,
			logger:         realLogger,
		}

		ctx := context.Background()

		ErrorFromContext(ctx, ew)

		// Should not write anything
		assert.False(t, ew.written)
		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("wrong response writer type", func(t *testing.T) {
		rr := httptest.NewRecorder()

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Should not panic with regular ResponseWriter
		ErrorFromContext(ctx, rr)

		// Should not write error response
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
