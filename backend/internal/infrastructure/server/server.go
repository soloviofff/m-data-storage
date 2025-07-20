package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"m-data-storage/api/handlers"
	"m-data-storage/api/middleware"
	"m-data-storage/api/routes"
	"m-data-storage/internal/infrastructure/config"
	"m-data-storage/internal/infrastructure/container"
	"m-data-storage/internal/infrastructure/logger"
)

// Server represents HTTP server
type Server struct {
	httpServer          *http.Server
	router              *mux.Router
	container           *container.Container
	logger              *logger.Logger
	config              *config.Config
	instrumentHandler   *handlers.InstrumentHandler
	subscriptionHandler *handlers.SubscriptionHandler
	dataHandler         *handlers.DataHandler
	authHandler         *handlers.AuthHandler
	authMiddleware      *middleware.AuthMiddleware
}

// NewServer creates a new HTTP server
func NewServer(cfg *config.Config, container *container.Container, logger *logger.Logger) *Server {
	router := mux.NewRouter()

	server := &Server{
		router:    router,
		container: container,
		logger:    logger,
		config:    cfg,
	}

	// Configure HTTP server
	server.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.API.Host, cfg.API.Port),
		Handler:      router,
		ReadTimeout:  cfg.API.ReadTimeout,
		WriteTimeout: cfg.API.WriteTimeout,
	}

	return server
}

// SetupHandlers initializes HTTP handlers
func (s *Server) SetupHandlers() error {
	// Initialize authentication handlers
	if err := s.setupAuthHandlers(); err != nil {
		return fmt.Errorf("failed to setup authentication handlers: %w", err)
	}

	// Get InstrumentManager from container
	instrumentManager, err := s.container.GetInstrumentManager()
	if err != nil {
		s.logger.WithError(err).Warn("InstrumentManager not available, instrument endpoints will be disabled")
	} else {
		s.instrumentHandler = handlers.NewInstrumentHandler(instrumentManager, s.logger)
		s.subscriptionHandler = handlers.NewSubscriptionHandler(instrumentManager, s.logger)
		s.logger.Info("Instrument and subscription handlers initialized")
	}

	// Get DataQuery service from container
	dataQuery, err := s.container.GetDataQuery()
	if err != nil {
		s.logger.WithError(err).Warn("DataQuery service not available, data endpoints will be disabled")
	} else {
		s.dataHandler = handlers.NewDataHandler(dataQuery, s.logger)
		s.logger.Info("Data handler initialized")
	}

	return nil
}

// setupAuthHandlers initializes authentication handlers and middleware
func (s *Server) setupAuthHandlers() error {
	// Get authentication services from container
	authService, err := s.container.GetAuthService()
	if err != nil {
		s.logger.WithError(err).Warn("AuthService not available, authentication endpoints will be disabled")
		return nil // Continue without authentication for now
	}

	tokenService, err := s.container.GetTokenService()
	if err != nil {
		s.logger.WithError(err).Warn("TokenService not available, authentication endpoints will be disabled")
		return nil // Continue without authentication for now
	}

	apiKeyService, err := s.container.GetAPIKeyService()
	if err != nil {
		s.logger.WithError(err).Warn("APIKeyService not available, authentication endpoints will be disabled")
		return nil // Continue without authentication for now
	}

	authorizationService, err := s.container.GetAuthorizationService()
	if err != nil {
		s.logger.WithError(err).Warn("AuthorizationService not available, authentication endpoints will be disabled")
		return nil // Continue without authentication for now
	}

	securityService, err := s.container.GetSecurityService()
	if err != nil {
		s.logger.WithError(err).Warn("SecurityService not available, authentication endpoints will be disabled")
		return nil // Continue without authentication for now
	}

	passwordService, err := s.container.GetPasswordService()
	if err != nil {
		s.logger.WithError(err).Warn("PasswordService not available, authentication endpoints will be disabled")
		return nil // Continue without authentication for now
	}

	// For services that are not implemented yet, we'll use placeholders
	// UserService and PermissionService are not implemented as separate services
	// They are handled through other interfaces in the current architecture

	// Create authentication handler with available services
	// Note: We're passing nil for userService and permissionService since they're not implemented
	s.authHandler = handlers.NewAuthHandler(
		authService,
		nil, // userService - not implemented as separate service
		tokenService,
		apiKeyService,
		authorizationService,
		nil, // permissionService - not implemented as separate service
		securityService,
		passwordService,
		*s.logger, // Dereference the logger pointer
	)

	// Create authentication middleware
	s.authMiddleware = middleware.NewAuthMiddleware(
		tokenService,
		apiKeyService,
		authorizationService, // Use authorizationService instead of separate authzService
		nil,                  // permissionService - not implemented as separate service
		securityService,
		s.logger,
	)

	s.logger.Info("Authentication handlers and middleware initialized")
	return nil
}

// SetupMiddleware configures middleware
func (s *Server) SetupMiddleware() {
	// Request ID middleware (should be first)
	requestIDMiddleware := middleware.NewRequestIDMiddleware()
	s.router.Use(requestIDMiddleware.RequestID)

	// Recovery middleware (should be second to catch panics)
	recoveryMiddleware := middleware.NewRecoveryMiddleware(s.logger)
	s.router.Use(recoveryMiddleware.Recover)

	// Security headers middleware (optional)
	securityService, err := s.container.GetSecurityService()
	if err != nil {
		s.logger.Error("Failed to get security service", "error", err.Error())
		// Continue without security middleware for now
	} else {
		securityMiddleware := middleware.NewSecurityMiddleware(
			securityService,
			s.logger,
		)
		s.router.Use(securityMiddleware.SecurityHeaders)
		// Security logging middleware
		s.router.Use(securityMiddleware.SecurityLogging)
	}

	// Logging middleware
	loggingMiddleware := middleware.NewLoggingMiddleware(s.logger)
	s.router.Use(loggingMiddleware.Log)

	// CORS middleware
	corsMiddleware := middleware.NewCORSMiddleware(
		s.config.API.CORS.AllowedOrigins,
		s.config.API.CORS.AllowedMethods,
		s.config.API.CORS.AllowedHeaders,
	)
	s.router.Use(corsMiddleware.CORS)

	// Rate limiting middleware
	rateLimiter := middleware.NewRateLimiter(100, time.Minute) // 100 requests per minute
	s.router.Use(rateLimiter.RateLimit)

	// Error handling middleware
	errorMiddleware := middleware.NewErrorHandlerMiddleware(s.logger)
	s.router.Use(errorMiddleware.ErrorHandler)
}

// SetupRoutes configures routes
func (s *Server) SetupRoutes() {
	// Setup handlers first
	if err := s.SetupHandlers(); err != nil {
		s.logger.WithError(err).Error("Failed to setup handlers")
	}

	// API versioning
	apiV1 := s.router.PathPrefix("/api/v1").Subrouter()

	// Health check endpoint
	s.router.HandleFunc("/health", s.healthCheckHandler).Methods("GET")
	s.router.HandleFunc("/ready", s.readinessHandler).Methods("GET")

	// Authentication routes
	if s.authHandler != nil && s.authMiddleware != nil {
		routes.RegisterAuthRoutes(apiV1, s.authHandler, s.authMiddleware)
		s.logger.Info("Authentication routes registered")
	} else {
		s.logger.Warn("Authentication routes not registered - handlers not available")
	}

	// Broker endpoints (stubs for future implementation)
	// TODO: Implement BrokerService and connect it
	apiV1.HandleFunc("/brokers", s.notImplementedHandler).Methods("GET")
	apiV1.HandleFunc("/brokers/{id}", s.notImplementedHandler).Methods("GET")
	apiV1.HandleFunc("/brokers/{id}/status", s.notImplementedHandler).Methods("GET")

	// Instrument management endpoints
	if s.instrumentHandler != nil {
		apiV1.HandleFunc("/instruments", s.instrumentHandler.ListInstruments).Methods("GET")
		apiV1.HandleFunc("/instruments", s.instrumentHandler.CreateInstrument).Methods("POST")
		apiV1.HandleFunc("/instruments/{symbol}", s.instrumentHandler.GetInstrument).Methods("GET")
		apiV1.HandleFunc("/instruments/{symbol}", s.instrumentHandler.UpdateInstrument).Methods("PUT")
		apiV1.HandleFunc("/instruments/{symbol}", s.instrumentHandler.DeleteInstrument).Methods("DELETE")

		// Subscription endpoints for specific instruments
		apiV1.HandleFunc("/instruments/{symbol}/subscriptions", s.instrumentHandler.CreateSubscription).Methods("POST")
	} else {
		apiV1.HandleFunc("/instruments", s.notImplementedHandler).Methods("GET")
		apiV1.HandleFunc("/instruments", s.notImplementedHandler).Methods("POST")
		apiV1.HandleFunc("/instruments/{symbol}", s.notImplementedHandler).Methods("GET")
		apiV1.HandleFunc("/instruments/{symbol}", s.notImplementedHandler).Methods("PUT")
		apiV1.HandleFunc("/instruments/{symbol}", s.notImplementedHandler).Methods("DELETE")
	}

	// Subscription management endpoints
	if s.subscriptionHandler != nil {
		apiV1.HandleFunc("/subscriptions", s.subscriptionHandler.ListSubscriptions).Methods("GET")
		apiV1.HandleFunc("/subscriptions/{id}", s.subscriptionHandler.GetSubscription).Methods("GET")
		apiV1.HandleFunc("/subscriptions/{id}", s.subscriptionHandler.UpdateSubscription).Methods("PUT")
		apiV1.HandleFunc("/subscriptions/{id}", s.subscriptionHandler.DeleteSubscription).Methods("DELETE")

		// Subscription control endpoints
		apiV1.HandleFunc("/subscriptions/{id}/start", s.subscriptionHandler.StartTracking).Methods("POST")
		apiV1.HandleFunc("/subscriptions/{id}/stop", s.subscriptionHandler.StopTracking).Methods("POST")
	} else {
		apiV1.HandleFunc("/subscriptions", s.notImplementedHandler).Methods("GET")
		apiV1.HandleFunc("/subscriptions/{id}", s.notImplementedHandler).Methods("GET")
		apiV1.HandleFunc("/subscriptions/{id}", s.notImplementedHandler).Methods("PUT")
		apiV1.HandleFunc("/subscriptions/{id}", s.notImplementedHandler).Methods("DELETE")
	}

	// Data endpoints
	if s.dataHandler != nil {
		apiV1.HandleFunc("/data/tickers", s.dataHandler.GetTickers).Methods("GET")
		apiV1.HandleFunc("/data/candles", s.dataHandler.GetCandles).Methods("GET")
		apiV1.HandleFunc("/data/orderbooks", s.dataHandler.GetOrderBooks).Methods("GET")
	} else {
		apiV1.HandleFunc("/data/tickers", s.notImplementedHandler).Methods("GET")
		apiV1.HandleFunc("/data/candles", s.notImplementedHandler).Methods("GET")
		apiV1.HandleFunc("/data/orderbooks", s.notImplementedHandler).Methods("GET")
	}

	// Configuration endpoints
	apiV1.HandleFunc("/config", s.notImplementedHandler).Methods("GET", "PUT")

	s.logger.Info("Routes configured successfully")
}

// Start starts the server
func (s *Server) Start() error {
	s.logger.WithFields(map[string]interface{}{
		"address": s.httpServer.Addr,
		"env":     s.config.App.Environment,
	}).Info("Starting HTTP server")

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// Stop stops the server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping HTTP server")

	return s.httpServer.Shutdown(ctx)
}

// healthCheckHandler handles health check requests
func (s *Server) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","service":"m-data-storage"}`))
}

// readinessHandler handles readiness requests
func (s *Server) readinessHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Add readiness checks (DB, brokers, etc.)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready","service":"m-data-storage"}`))
}

// notImplementedHandler returns "not implemented" error
func (s *Server) notImplementedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"success":false,"error":{"code":"NOT_IMPLEMENTED","message":"Endpoint not implemented yet"}}`))
}
