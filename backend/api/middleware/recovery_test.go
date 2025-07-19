package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"m-data-storage/api/dto"
	"m-data-storage/internal/infrastructure/config"
	"m-data-storage/internal/infrastructure/logger"
)

// MockLogrusEntry is a mock implementation of logrus.Entry
type MockLogrusEntry struct {
	mock.Mock
}

func (m *MockLogrusEntry) Error(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogrusEntry) WithFields(fields logrus.Fields) *logrus.Entry {
	m.Called(fields)
	return &logrus.Entry{}
}

// MockLoggerForRecovery is a mock implementation of logger.Logger for recovery tests
type MockLoggerForRecovery struct {
	mock.Mock
}

func (m *MockLoggerForRecovery) WithFields(fields map[string]interface{}) *MockLogrusEntry {
	args := m.Called(fields)
	return args.Get(0).(*MockLogrusEntry)
}

func (m *MockLoggerForRecovery) LogAPIRequest(requestID, method, path, userAgent, duration string, statusCode int) {
	m.Called(requestID, method, path, userAgent, duration, statusCode)
}

func (m *MockLoggerForRecovery) Info(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLoggerForRecovery) Error(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLoggerForRecovery) Warning(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLoggerForRecovery) Debug(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func TestNewRecoveryMiddleware(t *testing.T) {
	// Create a real logger for this test
	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	realLogger, err := logger.New(cfg)
	assert.NoError(t, err)

	middleware := NewRecoveryMiddleware(realLogger)

	assert.NotNil(t, middleware)
	assert.Equal(t, realLogger, middleware.logger)
}

func TestRecoveryMiddleware_Recover_NoPanic(t *testing.T) {
	// Create a real logger for this test
	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	realLogger, err := logger.New(cfg)
	assert.NoError(t, err)

	middleware := NewRecoveryMiddleware(realLogger)

	// Create a test handler that doesn't panic
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Wrap with recovery middleware
	recoveryHandler := middleware.Recover(handler)

	// Create request
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	// Execute request
	recoveryHandler.ServeHTTP(rr, req)

	// Check that normal response is preserved
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "success", rr.Body.String())
}

func TestRecoveryMiddleware_Recover_WithPanic(t *testing.T) {
	// Create a real logger for this test
	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	realLogger, err := logger.New(cfg)
	assert.NoError(t, err)

	middleware := NewRecoveryMiddleware(realLogger)

	// Create a test handler that panics
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Wrap with recovery middleware
	recoveryHandler := middleware.Recover(handler)

	// Create request
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("User-Agent", "test-agent")
	rr := httptest.NewRecorder()

	// Execute request
	recoveryHandler.ServeHTTP(rr, req)

	// Check that panic was recovered and proper error response returned
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	// Parse response body
	var response dto.Response
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check response structure
	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", response.Error.Code)
	assert.Equal(t, "Internal server error occurred", response.Error.Message)
}

func TestRecoveryMiddleware_Recover_WithStringPanic(t *testing.T) {
	// Create a real logger for this test
	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	realLogger, err := logger.New(cfg)
	assert.NoError(t, err)

	middleware := NewRecoveryMiddleware(realLogger)

	// Create a test handler that panics with a string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("string panic message")
	})

	// Wrap with recovery middleware
	recoveryHandler := middleware.Recover(handler)

	// Create request
	req := httptest.NewRequest("POST", "/api/v1/test", nil)
	rr := httptest.NewRecorder()

	// Execute request
	recoveryHandler.ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	var response dto.Response
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", response.Error.Code)
}

func TestRecoveryMiddleware_Recover_WithErrorPanic(t *testing.T) {
	// Create a real logger for this test
	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	realLogger, err := logger.New(cfg)
	assert.NoError(t, err)

	middleware := NewRecoveryMiddleware(realLogger)

	// Create a test handler that panics with an error
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(assert.AnError)
	})

	// Wrap with recovery middleware
	recoveryHandler := middleware.Recover(handler)

	// Create request
	req := httptest.NewRequest("PUT", "/api/v1/update", nil)
	rr := httptest.NewRecorder()

	// Execute request
	recoveryHandler.ServeHTTP(rr, req)

	// Check response
	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	var response dto.Response
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", response.Error.Code)
}

func TestRecoveryMiddleware_Recover_PreservesHeaders(t *testing.T) {
	// Create a real logger for this test
	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	realLogger, err := logger.New(cfg)
	assert.NoError(t, err)

	middleware := NewRecoveryMiddleware(realLogger)

	// Create a test handler that sets headers before panicking
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "custom-value")
		panic("test panic")
	})

	// Wrap with recovery middleware
	recoveryHandler := middleware.Recover(handler)

	// Create request
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	// Execute request
	recoveryHandler.ServeHTTP(rr, req)

	// Check that custom header is preserved and Content-Type is set
	assert.Equal(t, "custom-value", rr.Header().Get("X-Custom-Header"))
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
