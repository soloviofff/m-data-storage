package container

import (
	"testing"

	"github.com/sirupsen/logrus"

	"m-data-storage/internal/infrastructure/config"
	"m-data-storage/internal/infrastructure/logger"
)

// createTestLogger creates logger for tests
func createTestLogger() *logger.Logger {
	logrusLogger := logrus.New()
	logrusLogger.SetLevel(logrus.InfoLevel)
	return &logger.Logger{Logger: logrusLogger}
}

func TestNewContainer(t *testing.T) {
	cfg := &config.Config{}
	log := createTestLogger()

	container := NewContainer(cfg, log)

	if container == nil {
		t.Fatal("NewContainer() returned nil")
	}

	if container.config != cfg {
		t.Errorf("Container.config = %v, want %v", container.config, cfg)
	}

	if container.logger != log {
		t.Errorf("Container.logger = %v, want %v", container.logger, log)
	}

	if container.services == nil {
		t.Error("Container.services should be initialized")
	}
}

func TestContainer_RegisterAndGet(t *testing.T) {
	container := NewContainer(&config.Config{}, createTestLogger())

	// Test registering a service
	testService := "test-service-instance"
	container.Register("test-service", testService)

	// Test getting the service
	service, err := container.Get("test-service")
	if err != nil {
		t.Errorf("Container.Get() error = %v, want nil", err)
	}

	if service != testService {
		t.Errorf("Container.Get() = %v, want %v", service, testService)
	}
}

func TestContainer_GetNonExistent(t *testing.T) {
	container := NewContainer(&config.Config{}, createTestLogger())

	// Test getting a non-existent service
	service, err := container.Get("non-existent")
	if err == nil {
		t.Error("Container.Get() error = nil, want error")
	}

	if service != nil {
		t.Errorf("Container.Get() = %v, want nil", service)
	}

	expectedError := "service non-existent not found"
	if err.Error() != expectedError {
		t.Errorf("Container.Get() error = %v, want %v", err.Error(), expectedError)
	}
}

func TestContainer_GetConfig(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Name: "test-app",
		},
	}
	container := NewContainer(cfg, createTestLogger())

	result := container.GetConfig()
	if result != cfg {
		t.Errorf("Container.GetConfig() = %v, want %v", result, cfg)
	}
}

func TestContainer_GetLogger(t *testing.T) {
	log := &logger.Logger{}
	container := NewContainer(&config.Config{}, log)

	result := container.GetLogger()
	if result != log {
		t.Errorf("Container.GetLogger() = %v, want %v", result, log)
	}
}

func TestContainer_InitializeServices(t *testing.T) {
	// Create a minimal config for testing
	cfg := &config.Config{
		App: config.AppConfig{
			Name:        "test-app",
			Environment: "test",
		},
	}

	// Create a test logger
	log := createTestLogger()

	container := NewContainer(cfg, log)

	// Test service initialization
	err := container.InitializeServices()
	if err != nil {
		t.Errorf("Container.InitializeServices() error = %v, want nil", err)
	}

	// Check if services were registered
	_, err = container.Get("config.repository")
	if err != nil {
		t.Errorf("config.repository service not found after initialization")
	}

	_, err = container.Get("config.service")
	if err != nil {
		t.Errorf("config.service service not found after initialization")
	}

	_, err = container.Get("data.validator")
	if err != nil {
		t.Errorf("data.validator service not found after initialization")
	}
}

func TestContainer_GetConfigService(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Name:        "test-app",
			Environment: "test",
		},
	}

	log := createTestLogger()

	container := NewContainer(cfg, log)

	// Initialize services first
	err := container.InitializeServices()
	if err != nil {
		t.Fatalf("Failed to initialize services: %v", err)
	}

	// Test getting config service
	configSvc, err := container.GetConfigService()
	if err != nil {
		t.Errorf("Container.GetConfigService() error = %v, want nil", err)
	}

	if configSvc == nil {
		t.Error("Container.GetConfigService() returned nil service")
	}
}

func TestContainer_GetDataValidator(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Name:        "test-app",
			Environment: "test",
		},
	}

	log := createTestLogger()

	container := NewContainer(cfg, log)

	// Initialize services first
	err := container.InitializeServices()
	if err != nil {
		t.Fatalf("Failed to initialize services: %v", err)
	}

	// Test getting data validator
	validator, err := container.GetDataValidator()
	if err != nil {
		t.Errorf("Container.GetDataValidator() error = %v, want nil", err)
	}

	if validator == nil {
		t.Error("Container.GetDataValidator() returned nil validator")
	}
}

func TestContainer_Shutdown(t *testing.T) {
	container := NewContainer(&config.Config{}, createTestLogger())

	// Register some services
	container.Register("test1", "service1")
	container.Register("test2", "service2")

	// Test shutdown
	err := container.Shutdown()
	if err != nil {
		t.Errorf("Container.Shutdown() error = %v, want nil", err)
	}

	// Verify services are cleared
	_, err = container.Get("test1")
	if err == nil {
		t.Error("Services should be cleared after shutdown")
	}
}
