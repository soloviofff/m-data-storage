package container

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"m-data-storage/internal/application/services"
	"m-data-storage/internal/domain/interfaces"
	"m-data-storage/internal/infrastructure/config"
	"m-data-storage/internal/infrastructure/logger"
	"m-data-storage/internal/infrastructure/storage"
	configservice "m-data-storage/internal/service/config"
)

// Container represents DI container
type Container struct {
	mu       sync.RWMutex
	services map[string]interface{}
	config   *config.Config
	logger   *logger.Logger
}

// NewContainer creates a new DI container
func NewContainer(cfg *config.Config, log *logger.Logger) *Container {
	return &Container{
		services: make(map[string]interface{}),
		config:   cfg,
		logger:   log,
	}
}

// Register registers a service in the container
func (c *Container) Register(name string, service interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.services[name] = service
}

// Get retrieves a service from the container
func (c *Container) Get(name string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	service, exists := c.services[name]
	if !exists {
		return nil, fmt.Errorf("service %s not found", name)
	}

	return service, nil
}

// GetConfig returns the configuration
func (c *Container) GetConfig() *config.Config {
	return c.config
}

// GetLogger returns the logger
func (c *Container) GetLogger() *logger.Logger {
	return c.logger
}

// InitializeServices initializes all services
func (c *Container) InitializeServices() error {
	// Initialize database (stub)
	// TODO: Implement SQLite connection
	var db *sql.DB // Currently nil, will be implemented in future tasks

	// Register repositories
	configRepo := configservice.NewRepository(db)
	c.Register("config.repository", configRepo)

	// Register services
	configSvc, err := configservice.NewService(configRepo)
	if err != nil {
		return fmt.Errorf("failed to create config service: %w", err)
	}
	c.Register("config.service", configSvc)

	dataValidator := services.NewDataValidatorService()
	c.Register("data.validator", dataValidator)

	// Initialize StorageManager
	storageConfig := storage.Config{
		SQLite: storage.SQLiteConfig{
			DatabasePath: c.config.Database.SQLite.Path,
		},
		QuestDB: storage.QuestDBConfig{
			Host:     c.config.Database.QuestDB.Host,
			Port:     c.config.Database.QuestDB.Port,
			Database: c.config.Database.QuestDB.Database,
			Username: c.config.Database.QuestDB.Username,
			Password: c.config.Database.QuestDB.Password,
			SSLMode:  "disable", // Default SSL mode
		},
	}

	storageManager, err := storage.NewManager(storageConfig, c.logger.Logger)
	if err != nil {
		return fmt.Errorf("failed to create storage manager: %w", err)
	}

	// Initialize storage manager
	ctx := context.Background()
	if err := storageManager.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize storage manager: %w", err)
	}

	c.Register("storage.manager", storageManager)

	// Create DataProcessor service
	dataProcessor := services.NewDataProcessorService(
		nil, // storageManager - currently nil, will be added later
		dataValidator,
		c.logger.Logger,
	)
	c.Register("data.processor", dataProcessor)

	// Create DataPipeline service
	dataPipeline := services.NewDataPipelineService(
		nil, // brokerManager - currently nil, will be added later
		nil, // storageIntegration - currently nil, will be added later
		c.logger.Logger,
		services.DefaultDataPipelineConfig(),
	)
	c.Register("data.pipeline", dataPipeline)

	// Create InstrumentManager for API endpoints
	instrumentManager := services.NewInstrumentManagerService(
		nil, // metadataStorage - currently nil, will be added later
		dataPipeline,
		dataValidator,
		c.logger.Logger, // Use the underlying logrus.Logger
	)
	c.Register("instrument.manager", instrumentManager)

	// Create DataCollector service
	dataCollector := services.NewDataCollectorService(
		dataPipeline,
		instrumentManager,
		dataProcessor,
		c.logger.Logger,
	)
	c.Register("data.collector", dataCollector)

	// Create stub DataQuery service for API endpoints
	// In the future this will be replaced with full implementation with StorageManager
	dataQuery := services.NewDataQueryService(
		nil,             // storageManager - currently nil, will be added later
		nil,             // dateFilter - currently nil, will be added later
		c.logger.Logger, // Use the underlying logrus.Logger
	)
	c.Register("data.query", dataQuery)

	// Initialize authentication services
	if err := c.initializeAuthServices(); err != nil {
		return fmt.Errorf("failed to initialize authentication services: %w", err)
	}

	c.logger.Info("All services initialized successfully")
	return nil
}

// initializeAuthServices initializes all authentication-related services
func (c *Container) initializeAuthServices() error {
	// Get storage manager
	storageManagerInterface, err := c.Get("storage.manager")
	if err != nil {
		return fmt.Errorf("failed to get storage manager: %w", err)
	}
	storageManager := storageManagerInterface.(*storage.Manager)

	// Get user and role storage from storage manager
	userStorage := storageManager.GetUserStorage()
	roleStorage := storageManager.GetRoleStorage()

	// Initialize authentication repositories
	// userRepo := sqlite.NewUserRepository(db)
	// roleRepo := sqlite.NewRoleRepository(db)

	// Initialize authentication services
	passwordService := services.NewPasswordService()
	c.Register("auth.password", passwordService)

	tokenService := services.NewTokenService(
		c.config.API.Auth.JWT.Secret,
		c.config.API.Auth.JWT.AccessTokenExpiry,
		c.config.API.Auth.JWT.RefreshTokenExpiry,
		c.config.API.Auth.JWT.Issuer,
	)
	c.Register("auth.token", tokenService)

	apiKeyService := services.NewAPIKeyService(
		userStorage,
		roleStorage,
	)
	c.Register("auth.apikey", apiKeyService)

	// Note: UserService doesn't exist as a separate service in the current implementation
	// The user operations are handled directly through UserStorage interface
	// We'll register nil for now and implement when needed
	c.Register("auth.user", nil)

	// Note: PermissionService doesn't exist as a separate service in the current implementation
	// The permission operations are handled through AuthorizationService
	// We'll register nil for now and implement when needed
	c.Register("auth.permission", nil)

	authorizationService := services.NewAuthorizationService(
		userStorage,
		roleStorage,
	)
	c.Register("auth.authorization", authorizationService)

	securityService := services.NewSecurityService(
		userStorage,
	)
	c.Register("auth.security", securityService)

	authService := services.NewAuthService(
		userStorage,
		tokenService,
		passwordService,
		securityService,
	)
	c.Register("auth.service", authService)

	c.logger.Info("Authentication services initialized successfully")
	return nil
}

// GetConfigService returns the configuration service
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

// GetDataValidator returns the data validator
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

// GetInstrumentManager returns the instrument manager service
func (c *Container) GetInstrumentManager() (interfaces.InstrumentManager, error) {
	svc, err := c.Get("instrument.manager")
	if err != nil {
		return nil, err
	}

	instrumentManager, ok := svc.(interfaces.InstrumentManager)
	if !ok {
		return nil, fmt.Errorf("service is not an InstrumentManager")
	}

	return instrumentManager, nil
}

// GetDataQuery returns the data query service
func (c *Container) GetDataQuery() (interfaces.DataQuery, error) {
	svc, err := c.Get("data.query")
	if err != nil {
		return nil, err
	}

	dataQuery, ok := svc.(interfaces.DataQuery)
	if !ok {
		return nil, fmt.Errorf("service is not a DataQuery")
	}

	return dataQuery, nil
}

// GetDataProcessor returns the data processor service
func (c *Container) GetDataProcessor() (interfaces.DataProcessor, error) {
	svc, err := c.Get("data.processor")
	if err != nil {
		return nil, err
	}

	dataProcessor, ok := svc.(interfaces.DataProcessor)
	if !ok {
		return nil, fmt.Errorf("service is not a DataProcessor")
	}

	return dataProcessor, nil
}

// GetDataPipeline returns the data pipeline service
func (c *Container) GetDataPipeline() (interfaces.DataPipeline, error) {
	svc, err := c.Get("data.pipeline")
	if err != nil {
		return nil, err
	}

	dataPipeline, ok := svc.(interfaces.DataPipeline)
	if !ok {
		return nil, fmt.Errorf("service is not a DataPipeline")
	}

	return dataPipeline, nil
}

// GetDataCollector returns the data collector service
func (c *Container) GetDataCollector() (interfaces.DataCollector, error) {
	svc, err := c.Get("data.collector")
	if err != nil {
		return nil, err
	}

	dataCollector, ok := svc.(interfaces.DataCollector)
	if !ok {
		return nil, fmt.Errorf("service is not a DataCollector")
	}

	return dataCollector, nil
}

// GetSecurityService returns the security service
func (c *Container) GetSecurityService() (interfaces.SecurityService, error) {
	svc, err := c.Get("auth.security")
	if err != nil {
		return nil, err
	}

	securityService, ok := svc.(interfaces.SecurityService)
	if !ok {
		return nil, fmt.Errorf("service is not a SecurityService")
	}

	return securityService, nil
}

// GetAuthService returns the authentication service
func (c *Container) GetAuthService() (interfaces.AuthService, error) {
	svc, err := c.Get("auth.service")
	if err != nil {
		return nil, err
	}

	authService, ok := svc.(interfaces.AuthService)
	if !ok {
		return nil, fmt.Errorf("service is not an AuthService")
	}

	return authService, nil
}

// GetUserService returns the user service (placeholder for now)
func (c *Container) GetUserService() (interfaces.UserService, error) {
	// For now, return an error since UserService is not implemented as a separate service
	// In the current architecture, user operations are handled through UserStorage interface
	return nil, fmt.Errorf("UserService not implemented - use UserStorage interface directly")
}

// GetTokenService returns the token service
func (c *Container) GetTokenService() (interfaces.TokenService, error) {
	svc, err := c.Get("auth.token")
	if err != nil {
		return nil, err
	}

	tokenService, ok := svc.(interfaces.TokenService)
	if !ok {
		return nil, fmt.Errorf("service is not a TokenService")
	}

	return tokenService, nil
}

// GetAPIKeyService returns the API key service
func (c *Container) GetAPIKeyService() (interfaces.APIKeyService, error) {
	svc, err := c.Get("auth.apikey")
	if err != nil {
		return nil, err
	}

	apiKeyService, ok := svc.(interfaces.APIKeyService)
	if !ok {
		return nil, fmt.Errorf("service is not an APIKeyService")
	}

	return apiKeyService, nil
}

// GetAuthorizationService returns the authorization service
func (c *Container) GetAuthorizationService() (interfaces.AuthorizationService, error) {
	svc, err := c.Get("auth.authorization")
	if err != nil {
		return nil, err
	}

	authorizationService, ok := svc.(interfaces.AuthorizationService)
	if !ok {
		return nil, fmt.Errorf("service is not an AuthorizationService")
	}

	return authorizationService, nil
}

// GetPermissionService returns the permission service (placeholder for now)
func (c *Container) GetPermissionService() (interfaces.PermissionService, error) {
	// For now, return an error since PermissionService is not implemented as a separate service
	// In the current architecture, permission operations are handled through AuthorizationService
	return nil, fmt.Errorf("PermissionService not implemented - use AuthorizationService interface directly")
}

// GetPasswordService returns the password service
func (c *Container) GetPasswordService() (interfaces.PasswordService, error) {
	svc, err := c.Get("auth.password")
	if err != nil {
		return nil, err
	}

	passwordService, ok := svc.(interfaces.PasswordService)
	if !ok {
		return nil, fmt.Errorf("service is not a PasswordService")
	}

	return passwordService, nil
}

// Shutdown properly shuts down all services
func (c *Container) Shutdown() error {
	c.logger.Info("Shutting down container services")

	// TODO: Add proper service shutdown
	// For example, closing database connections, stopping workers, etc.

	c.mu.Lock()
	defer c.mu.Unlock()

	// Clear the container
	c.services = make(map[string]interface{})

	c.logger.Info("Container shutdown completed")
	return nil
}
