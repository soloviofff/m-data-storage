package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"m-data-storage/internal/infrastructure/config"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		config      config.LoggingConfig
		expectError bool
	}{
		{
			name: "valid config - json stdout",
			config: config.LoggingConfig{
				Level:  "info",
				Format: "json",
				Output: "stdout",
			},
			expectError: false,
		},
		{
			name: "valid config - text stderr",
			config: config.LoggingConfig{
				Level:  "debug",
				Format: "text",
				Output: "stderr",
			},
			expectError: false,
		},
		{
			name: "invalid log level",
			config: config.LoggingConfig{
				Level:  "invalid",
				Format: "json",
				Output: "stdout",
			},
			expectError: true,
		},
		{
			name: "file output without file path",
			config: config.LoggingConfig{
				Level:  "info",
				Format: "json",
				Output: "file",
				File:   "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, logger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
				assert.Equal(t, tt.config, logger.config)
			}
		})
	}
}

func TestNew_FileOutput(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	config := config.LoggingConfig{
		Level:      "info",
		Format:     "json",
		Output:     "file",
		File:       logFile,
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	}

	logger, err := New(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Test logging to file
	logger.Info("test message")

	// Close logger to flush
	err = logger.Close()
	assert.NoError(t, err)

	// Check if file was created
	assert.FileExists(t, logFile)
}

func TestNew_BothOutput(t *testing.T) {
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	config := config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "both",
		File:   logFile,
	}

	logger, err := New(config)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Close logger immediately to avoid stdout interference
	err = logger.Close()
	assert.NoError(t, err)

	// Check if file was created (directory should exist)
	assert.DirExists(t, tempDir)
}

func TestGetOutput(t *testing.T) {
	tests := []struct {
		name        string
		config      config.LoggingConfig
		expectError bool
	}{
		{
			name: "stdout output",
			config: config.LoggingConfig{
				Output: "stdout",
			},
			expectError: false,
		},
		{
			name: "stderr output",
			config: config.LoggingConfig{
				Output: "stderr",
			},
			expectError: false,
		},
		{
			name: "default output",
			config: config.LoggingConfig{
				Output: "unknown",
			},
			expectError: false,
		},
		{
			name: "file output without path",
			config: config.LoggingConfig{
				Output: "file",
				File:   "",
			},
			expectError: true,
		},
		{
			name: "both output without path",
			config: config.LoggingConfig{
				Output: "both",
				File:   "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := getOutput(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, output)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
			}
		})
	}
}

func TestLogger_WithMethods(t *testing.T) {
	var buf bytes.Buffer

	// Create logger with buffer output for testing
	logger := &Logger{
		Logger: &logrus.Logger{
			Out:       &buf,
			Formatter: &logrus.JSONFormatter{},
			Level:     logrus.InfoLevel,
		},
		config: config.LoggingConfig{},
	}

	t.Run("WithComponent", func(t *testing.T) {
		buf.Reset()
		entry := logger.WithComponent("test-component")
		entry.Info("test message")

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "test-component", logEntry["component"])
		assert.Equal(t, "test message", logEntry["msg"])
	})

	t.Run("WithBroker", func(t *testing.T) {
		buf.Reset()
		entry := logger.WithBroker("broker-123")
		entry.Info("broker message")

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "broker-123", logEntry["broker_id"])
	})

	t.Run("WithSymbol", func(t *testing.T) {
		buf.Reset()
		entry := logger.WithSymbol("BTCUSD")
		entry.Info("symbol message")

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "BTCUSD", logEntry["symbol"])
	})

	t.Run("WithRequest", func(t *testing.T) {
		buf.Reset()
		entry := logger.WithRequest("req-123", "GET", "/api/v1/test")
		entry.Info("request message")

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "req-123", logEntry["request_id"])
		assert.Equal(t, "GET", logEntry["method"])
		assert.Equal(t, "/api/v1/test", logEntry["path"])
	})

	t.Run("WithError", func(t *testing.T) {
		buf.Reset()
		testErr := errors.New("test error")
		entry := logger.WithError(testErr)
		entry.Error("error message")

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "test error", logEntry["error"])
	})

	t.Run("WithDuration", func(t *testing.T) {
		buf.Reset()
		entry := logger.WithDuration("100ms")
		entry.Info("duration message")

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "100ms", logEntry["duration"])
	})

	t.Run("WithCount", func(t *testing.T) {
		buf.Reset()
		entry := logger.WithCount(42)
		entry.Info("count message")

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, float64(42), logEntry["count"]) // JSON numbers are float64
	})

	t.Run("WithDataType", func(t *testing.T) {
		buf.Reset()
		entry := logger.WithDataType("candles")
		entry.Info("data type message")

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "candles", logEntry["data_type"])
	})
}

func TestLogger_LogMethods(t *testing.T) {
	var buf bytes.Buffer

	// Create logger with buffer output for testing
	logger := &Logger{
		Logger: &logrus.Logger{
			Out:       &buf,
			Formatter: &logrus.JSONFormatter{},
			Level:     logrus.DebugLevel, // Set to debug to capture all logs
		},
		config: config.LoggingConfig{},
	}

	t.Run("LogDataReceived", func(t *testing.T) {
		buf.Reset()
		logger.LogDataReceived("broker-1", "BTCUSD", "candles", 100)

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "broker-1", logEntry["broker_id"])
		assert.Equal(t, "BTCUSD", logEntry["symbol"])
		assert.Equal(t, "candles", logEntry["data_type"])
		assert.Equal(t, float64(100), logEntry["count"])
		assert.Equal(t, "Data received", logEntry["msg"])
		assert.Equal(t, "debug", logEntry["level"])
	})

	t.Run("LogDataProcessed", func(t *testing.T) {
		buf.Reset()
		logger.LogDataProcessed("ticks", 50, "200ms")

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "ticks", logEntry["data_type"])
		assert.Equal(t, float64(50), logEntry["count"])
		assert.Equal(t, "200ms", logEntry["duration"])
		assert.Equal(t, "Data processed", logEntry["msg"])
		assert.Equal(t, "info", logEntry["level"])
	})

	t.Run("LogBrokerConnected", func(t *testing.T) {
		buf.Reset()
		logger.LogBrokerConnected("broker-2")

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "broker-2", logEntry["broker_id"])
		assert.Equal(t, "Broker connected", logEntry["msg"])
		assert.Equal(t, "info", logEntry["level"])
	})

	t.Run("LogBrokerDisconnected", func(t *testing.T) {
		buf.Reset()
		logger.LogBrokerDisconnected("broker-3", "connection timeout")

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "broker-3", logEntry["broker_id"])
		assert.Equal(t, "connection timeout", logEntry["reason"])
		assert.Equal(t, "Broker disconnected", logEntry["msg"])
		assert.Equal(t, "warning", logEntry["level"])
	})

	t.Run("LogBrokerError", func(t *testing.T) {
		buf.Reset()
		testErr := errors.New("connection failed")
		logger.LogBrokerError("broker-4", testErr)

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "broker-4", logEntry["broker_id"])
		assert.Equal(t, "connection failed", logEntry["error"])
		assert.Equal(t, "Broker error", logEntry["msg"])
		assert.Equal(t, "error", logEntry["level"])
	})

	t.Run("LogAPIRequest - success", func(t *testing.T) {
		buf.Reset()
		logger.LogAPIRequest("req-123", "GET", "/api/v1/data", "curl/7.68.0", "150ms", 200)

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "req-123", logEntry["request_id"])
		assert.Equal(t, "GET", logEntry["method"])
		assert.Equal(t, "/api/v1/data", logEntry["path"])
		assert.Equal(t, "curl/7.68.0", logEntry["user_agent"])
		assert.Equal(t, "150ms", logEntry["duration"])
		assert.Equal(t, float64(200), logEntry["status_code"])
		assert.Equal(t, "API request completed", logEntry["msg"])
		assert.Equal(t, "info", logEntry["level"])
	})

	t.Run("LogAPIRequest - error", func(t *testing.T) {
		buf.Reset()
		logger.LogAPIRequest("req-456", "POST", "/api/v1/invalid", "browser/1.0", "50ms", 404)

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, float64(404), logEntry["status_code"])
		assert.Equal(t, "API request completed with error", logEntry["msg"])
		assert.Equal(t, "warning", logEntry["level"])
	})

	t.Run("LogStorageOperation", func(t *testing.T) {
		buf.Reset()
		logger.LogStorageOperation("INSERT", "candles", 25, "300ms")

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "INSERT", logEntry["operation"])
		assert.Equal(t, "candles", logEntry["table"])
		assert.Equal(t, float64(25), logEntry["count"])
		assert.Equal(t, "300ms", logEntry["duration"])
		assert.Equal(t, "Storage operation completed", logEntry["msg"])
		assert.Equal(t, "debug", logEntry["level"])
	})

	t.Run("LogHealthCheck - healthy", func(t *testing.T) {
		buf.Reset()
		logger.LogHealthCheck("database", true, "connection ok")

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "database", logEntry["component"])
		assert.Equal(t, true, logEntry["healthy"])
		assert.Equal(t, "connection ok", logEntry["details"])
		assert.Equal(t, "Health check passed", logEntry["msg"])
		assert.Equal(t, "debug", logEntry["level"])
	})

	t.Run("LogHealthCheck - unhealthy", func(t *testing.T) {
		buf.Reset()
		logger.LogHealthCheck("redis", false, "connection timeout")

		var logEntry map[string]interface{}
		err := json.Unmarshal(buf.Bytes(), &logEntry)
		require.NoError(t, err)

		assert.Equal(t, "redis", logEntry["component"])
		assert.Equal(t, false, logEntry["healthy"])
		assert.Equal(t, "connection timeout", logEntry["details"])
		assert.Equal(t, "Health check failed", logEntry["msg"])
		assert.Equal(t, "warning", logEntry["level"])
	})
}

func TestLogger_Close(t *testing.T) {
	t.Run("close with stdout", func(t *testing.T) {
		var buf bytes.Buffer
		logger := &Logger{
			Logger: &logrus.Logger{
				Out: &buf, // Use buffer instead of stdout
			},
		}

		err := logger.Close()
		assert.NoError(t, err)
	})

	t.Run("close with closable output", func(t *testing.T) {
		// Create a mock closer
		mockCloser := &mockCloser{}

		logger := &Logger{
			Logger: &logrus.Logger{
				Out:       mockCloser,
				Formatter: &logrus.JSONFormatter{},
				Level:     logrus.InfoLevel,
			},
		}

		err := logger.Close()
		assert.NoError(t, err)
		assert.True(t, mockCloser.closed)
	})
}

// mockCloser implements io.Writer and io.Closer for testing
type mockCloser struct {
	closed bool
}

func (m *mockCloser) Write(p []byte) (int, error) {
	return len(p), nil
}

func (m *mockCloser) Close() error {
	m.closed = true
	return nil
}
