package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"m-data-storage/internal/infrastructure/config"
	"m-data-storage/internal/infrastructure/container"
	"m-data-storage/internal/infrastructure/logger"
)

func TestNewServer(t *testing.T) {
	cfg := createTestConfig()
	log := createTestLogger()
	c := container.NewContainer(cfg, log)

	server := NewServer(cfg, c, log)

	assert.NotNil(t, server)
	assert.NotNil(t, server.router)
	assert.NotNil(t, server.httpServer)
	assert.Equal(t, "localhost:8080", server.httpServer.Addr)
}

func TestServer_SetupMiddleware(t *testing.T) {
	cfg := createTestConfig()
	log := createTestLogger()
	c := container.NewContainer(cfg, log)

	server := NewServer(cfg, c, log)
	server.SetupMiddleware()

	// Middleware configured, check that router is not nil
	assert.NotNil(t, server.router)
}

func TestServer_SetupRoutes(t *testing.T) {
	cfg := createTestConfig()
	log := createTestLogger()
	c := container.NewContainer(cfg, log)

	server := NewServer(cfg, c, log)
	server.SetupMiddleware()
	server.SetupRoutes()

	// Check that routes are configured
	assert.NotNil(t, server.router)
}

func TestHealthCheckHandler(t *testing.T) {
	cfg := createTestConfig()
	log := createTestLogger()
	c := container.NewContainer(cfg, log)

	server := NewServer(cfg, c, log)
	server.SetupMiddleware()
	server.SetupRoutes()

	req, err := http.NewRequest("GET", "/health", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "ok")
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
}

func TestReadinessHandler(t *testing.T) {
	cfg := createTestConfig()
	log := createTestLogger()
	c := container.NewContainer(cfg, log)

	server := NewServer(cfg, c, log)
	server.SetupMiddleware()
	server.SetupRoutes()

	req, err := http.NewRequest("GET", "/ready", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "ready")
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
}

func TestNotImplementedEndpoints(t *testing.T) {
	cfg := createTestConfig()
	log := createTestLogger()
	c := container.NewContainer(cfg, log)

	server := NewServer(cfg, c, log)
	server.SetupMiddleware()
	server.SetupRoutes()

	testCases := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/brokers"},
		{"GET", "/api/v1/brokers/test-id"},
		{"GET", "/api/v1/brokers/test-id/status"},
		{"GET", "/api/v1/data/tickers"},
		{"GET", "/api/v1/data/candles"},
		{"GET", "/api/v1/data/orderbooks"},
		{"GET", "/api/v1/config"},
	}

	for _, tc := range testCases {
		t.Run(tc.method+"_"+tc.path, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, tc.path, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			server.router.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusNotImplemented, rr.Code)
			assert.Contains(t, rr.Body.String(), "NOT_IMPLEMENTED")
			assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
		})
	}
}

// TestServer_StartStop test for server start and stop
// Skip this test as it requires real network connection
func TestServer_StartStop(t *testing.T) {
	t.Skip("Skipping server start/stop test - requires network setup")
}

func TestCORSHeaders(t *testing.T) {
	cfg := createTestConfig()
	log := createTestLogger()
	c := container.NewContainer(cfg, log)

	server := NewServer(cfg, c, log)
	server.SetupMiddleware()
	server.SetupRoutes()

	// Test regular GET request with Origin header
	req, err := http.NewRequest("GET", "/health", nil)
	require.NoError(t, err)
	req.Header.Set("Origin", "http://localhost:3000")

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	// CORS middleware should add headers
	assert.NotEmpty(t, rr.Header().Get("Access-Control-Allow-Origin"))
}

func TestRequestIDMiddleware(t *testing.T) {
	cfg := createTestConfig()
	log := createTestLogger()
	c := container.NewContainer(cfg, log)

	server := NewServer(cfg, c, log)
	server.SetupMiddleware()
	server.SetupRoutes()

	req, err := http.NewRequest("GET", "/health", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.NotEmpty(t, rr.Header().Get("X-Request-ID"))
}

// Helper functions

func createTestConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Name:        "m-data-storage-test",
			Version:     "1.0.0",
			Environment: "test",
		},
		API: config.APIConfig{
			Host:            "localhost",
			Port:            8080,
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			ShutdownTimeout: 10 * time.Second,
			CORS: config.CORSConfig{
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders: []string{"*"},
			},
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}
}

func createTestLogger() *logger.Logger {
	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}
	log, _ := logger.New(cfg)
	return log
}
