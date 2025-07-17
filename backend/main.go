package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"m-data-storage/internal/infrastructure/config"
	"m-data-storage/internal/infrastructure/logger"
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

	log.WithComponent("main").Info("Starting M-Data-Storage application",
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
		log.WithComponent("main").Info("Received shutdown signal", "signal", sig.String())
		cancel()
	}()

	// TODO: Initialize application components
	if err := initializeApplication(ctx, cfg, log); err != nil {
		log.WithComponent("main").WithError(err).Fatal("Failed to initialize application")
	}

	log.WithComponent("main").Info("Application started successfully")

	// Wait for shutdown signal
	<-ctx.Done()

	// Graceful shutdown
	log.WithComponent("main").Info("Starting graceful shutdown")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.API.ShutdownTimeout)
	defer shutdownCancel()

	if err := shutdownApplication(shutdownCtx, log); err != nil {
		log.WithComponent("main").WithError(err).Error("Error during shutdown")
	}

	log.WithComponent("main").Info("Application stopped")
}

// initializeApplication initializes all application components
func initializeApplication(ctx context.Context, cfg *config.Config, log *logger.Logger) error {
	log.WithComponent("init").Info("Initializing application components")

	// TODO: Initialize components
	// 1. Connect to databases (SQLite + QuestDB)
	// 2. Initialize storage repositories
	// 3. Initialize services
	// 4. Load broker configurations
	// 5. Initialize brokers
	// 6. Start web server
	// 7. Start data collection service

	log.WithComponent("init").Info("Application components initialized successfully")
	return nil
}

// shutdownApplication performs graceful shutdown of all components
func shutdownApplication(ctx context.Context, log *logger.Logger) error {
	log.WithComponent("shutdown").Info("Shutting down application components")

	// TODO: Graceful shutdown of components in reverse order
	// 1. Stop data collection
	// 2. Stop web server
	// 3. Stop brokers
	// 4. Stop services
	// 5. Close database connections

	// Give time for operations to complete
	select {
	case <-time.After(1 * time.Second):
	case <-ctx.Done():
		return ctx.Err()
	}

	log.WithComponent("shutdown").Info("Application components shut down successfully")
	return nil
}
