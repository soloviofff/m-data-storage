package container

import (
	"database/sql"
	"fmt"
	"sync"

	"m-data-storage/internal/application/services"
	"m-data-storage/internal/infrastructure/config"
	"m-data-storage/internal/infrastructure/logger"
	configservice "m-data-storage/internal/service/config"
)

// Container представляет DI контейнер
type Container struct {
	mu       sync.RWMutex
	services map[string]interface{}
	config   *config.Config
	logger   *logger.Logger
}

// NewContainer создает новый DI контейнер
func NewContainer(cfg *config.Config, log *logger.Logger) *Container {
	return &Container{
		services: make(map[string]interface{}),
		config:   cfg,
		logger:   log,
	}
}

// Register регистрирует сервис в контейнере
func (c *Container) Register(name string, service interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[name] = service
}

// Get получает сервис из контейнера
func (c *Container) Get(name string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	service, exists := c.services[name]
	if !exists {
		return nil, fmt.Errorf("service %s not found", name)
	}

	return service, nil
}

// GetConfig возвращает конфигурацию
func (c *Container) GetConfig() *config.Config {
	return c.config
}

// GetLogger возвращает логгер
func (c *Container) GetLogger() *logger.Logger {
	return c.logger
}

// InitializeServices инициализирует все сервисы
func (c *Container) InitializeServices() error {
	// Инициализируем базу данных (заглушка)
	// TODO: Реализовать подключение к SQLite
	var db *sql.DB // Пока nil, будет реализовано в следующих задачах

	// Регистрируем репозитории
	configRepo := configservice.NewRepository(db)
	c.Register("config.repository", configRepo)

	// Регистрируем сервисы
	configSvc, err := configservice.NewService(configRepo)
	if err != nil {
		return fmt.Errorf("failed to create config service: %w", err)
	}
	c.Register("config.service", configSvc)

	dataValidator := services.NewDataValidatorService()
	c.Register("data.validator", dataValidator)

	// TODO: Добавить другие сервисы по мере их реализации

	c.logger.Info("All services initialized successfully")
	return nil
}

// GetConfigService возвращает сервис конфигурации
func (c *Container) GetConfigService() (*configservice.Service, error) {
	svc, err := c.Get("config.service")
	if err != nil {
		return nil, err
	}

	configSvc, ok := svc.(*configservice.Service)
	if !ok {
		return nil, fmt.Errorf("invalid config service type")
	}

	return configSvc, nil
}

// GetDataValidator возвращает валидатор данных
func (c *Container) GetDataValidator() (*services.DataValidatorService, error) {
	svc, err := c.Get("data.validator")
	if err != nil {
		return nil, err
	}

	validator, ok := svc.(*services.DataValidatorService)
	if !ok {
		return nil, fmt.Errorf("invalid data validator type")
	}

	return validator, nil
}

// Shutdown корректно завершает работу всех сервисов
func (c *Container) Shutdown() error {
	c.logger.Info("Shutting down container services")

	// TODO: Добавить корректное завершение работы сервисов
	// Например, закрытие соединений с БД, остановка воркеров и т.д.

	c.mu.Lock()
	defer c.mu.Unlock()

	// Очищаем контейнер
	c.services = make(map[string]interface{})

	c.logger.Info("Container shutdown completed")
	return nil
}
