package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"m-data-storage/api/middleware"
	"m-data-storage/internal/infrastructure/config"
	"m-data-storage/internal/infrastructure/container"
	"m-data-storage/internal/infrastructure/logger"
)

// Server представляет HTTP сервер
type Server struct {
	httpServer *http.Server
	router     *mux.Router
	container  *container.Container
	logger     *logger.Logger
	config     *config.Config
}

// NewServer создает новый HTTP сервер
func NewServer(cfg *config.Config, container *container.Container, logger *logger.Logger) *Server {
	router := mux.NewRouter()

	server := &Server{
		router:    router,
		container: container,
		logger:    logger,
		config:    cfg,
	}

	// Настраиваем HTTP сервер
	server.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.API.Host, cfg.API.Port),
		Handler:      router,
		ReadTimeout:  cfg.API.ReadTimeout,
		WriteTimeout: cfg.API.WriteTimeout,
	}

	return server
}

// SetupMiddleware настраивает middleware
func (s *Server) SetupMiddleware() {
	// Request ID middleware (должен быть первым)
	requestIDMiddleware := middleware.NewRequestIDMiddleware()
	s.router.Use(requestIDMiddleware.RequestID)

	// Recovery middleware (должен быть вторым для перехвата паник)
	recoveryMiddleware := middleware.NewRecoveryMiddleware(s.logger)
	s.router.Use(recoveryMiddleware.Recover)

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

// SetupRoutes настраивает маршруты
func (s *Server) SetupRoutes() {
	// API версионирование
	apiV1 := s.router.PathPrefix("/api/v1").Subrouter()

	// Health check endpoint
	s.router.HandleFunc("/health", s.healthCheckHandler).Methods("GET")
	s.router.HandleFunc("/ready", s.readinessHandler).Methods("GET")

	// Broker endpoints (заглушки для будущей реализации)
	// TODO: Реализовать BrokerService и подключить его
	apiV1.HandleFunc("/brokers", s.notImplementedHandler).Methods("GET")
	apiV1.HandleFunc("/brokers/{id}", s.notImplementedHandler).Methods("GET")
	apiV1.HandleFunc("/brokers/{id}/status", s.notImplementedHandler).Methods("GET")

	// Data endpoints (заглушки для будущей реализации)
	apiV1.HandleFunc("/data/tickers", s.notImplementedHandler).Methods("GET")
	apiV1.HandleFunc("/data/candles", s.notImplementedHandler).Methods("GET")
	apiV1.HandleFunc("/data/orderbooks", s.notImplementedHandler).Methods("GET")

	// Configuration endpoints
	apiV1.HandleFunc("/config", s.notImplementedHandler).Methods("GET", "PUT")

	s.logger.Info("Routes configured successfully")
}

// Start запускает сервер
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

// Stop останавливает сервер
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping HTTP server")

	return s.httpServer.Shutdown(ctx)
}

// healthCheckHandler обрабатывает запросы проверки здоровья
func (s *Server) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","service":"m-data-storage"}`))
}

// readinessHandler обрабатывает запросы готовности
func (s *Server) readinessHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Добавить проверки готовности (БД, брокеры и т.д.)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready","service":"m-data-storage"}`))
}

// notImplementedHandler возвращает ошибку "не реализовано"
func (s *Server) notImplementedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"success":false,"error":{"code":"NOT_IMPLEMENTED","message":"Endpoint not implemented yet"}}`))
}
