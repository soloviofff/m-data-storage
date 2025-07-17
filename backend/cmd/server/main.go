package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"m-data-storage/internal/infrastructure/config"
	"m-data-storage/internal/infrastructure/container"
	"m-data-storage/internal/infrastructure/logger"
	"m-data-storage/internal/infrastructure/server"
)

func main() {
	// Parse command line flags
	var configPath = flag.String("config", "", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.New(cfg.Logging)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	log.WithComponent("server").Info("Starting M-Data-Storage server",
		"version", cfg.App.Version,
		"environment", cfg.App.Environment)

	// Create context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.WithComponent("server").Info("Received shutdown signal", "signal", sig.String())
		cancel()
	}()

	// Initialize and start server
	if err := runServer(ctx, cfg, log); err != nil {
		log.WithComponent("server").WithError(err).Fatal("Server failed")
	}

	log.WithComponent("server").Info("Server stopped")
}

// runServer starts server with full initialization
func runServer(ctx context.Context, cfg *config.Config, log *logger.Logger) error {
	log.WithComponent("server").Info("Initializing server components")

	// 1. Initialize dependency injection container
	container, err := initializeContainer(ctx, cfg, log)
	if err != nil {
		return fmt.Errorf("failed to initialize container: %w", err)
	}
	defer container.Shutdown()

	// 2. Initialize HTTP server
	httpServer, err := initializeHTTPServer(cfg, container, log)
	if err != nil {
		return fmt.Errorf("failed to initialize HTTP server: %w", err)
	}

	// 3. Start HTTP server
	serverErrChan := make(chan error, 1)
	go func() {
		if err := httpServer.Start(); err != nil {
			serverErrChan <- fmt.Errorf("HTTP server failed: %w", err)
		}
	}()

	log.WithComponent("server").Info("All components initialized successfully")

	// Wait for shutdown signal or server error
	select {
	case <-ctx.Done():
		log.WithComponent("server").Info("Received shutdown signal")
	case err := <-serverErrChan:
		log.WithComponent("server").WithError(err).Error("Server error occurred")
		return err
	}

	// Graceful shutdown
	log.WithComponent("server").Info("Starting graceful shutdown")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.API.ShutdownTimeout)
	defer shutdownCancel()

	// Stop HTTP server
	if err := httpServer.Stop(shutdownCtx); err != nil {
		log.WithComponent("server").WithError(err).Error("Failed to stop HTTP server gracefully")
		return err
	}

	log.WithComponent("server").Info("Graceful shutdown completed")
	return nil
}

// initializeContainer initializes dependency injection container
func initializeContainer(ctx context.Context, cfg *config.Config, log *logger.Logger) (*container.Container, error) {
	log.WithComponent("container").Info("Initializing dependency injection container")

	// Create container
	c := container.NewContainer(cfg, log)

	// Initialize services
	if err := c.InitializeServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	log.WithComponent("container").Info("Dependency injection container initialized successfully")
	return c, nil
}

// initializeHTTPServer initializes HTTP server
func initializeHTTPServer(cfg *config.Config, container *container.Container, log *logger.Logger) (*server.Server, error) {
	log.WithComponent("http").Info("Initializing HTTP server")

	// Create HTTP server
	httpServer := server.NewServer(cfg, container, log)

	// Setup middleware
	httpServer.SetupMiddleware()

	// Setup routes
	httpServer.SetupRoutes()

	log.WithComponent("http").Info("HTTP server initialized successfully")
	return httpServer, nil
}
