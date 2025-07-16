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
	// Парсим флаги командной строки
	var configPath = flag.String("config", "", "Path to configuration file")
	flag.Parse()

	// Загружаем конфигурацию
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Инициализируем логгер
	log, err := logger.New(cfg.Logging)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	log.WithComponent("server").Info("Starting M-Data-Storage server",
		"version", cfg.App.Version,
		"environment", cfg.App.Environment)

	// Создаем контекст с отменой для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Настраиваем обработку сигналов
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.WithComponent("server").Info("Received shutdown signal", "signal", sig.String())
		cancel()
	}()

	// Инициализация и запуск сервера
	if err := runServer(ctx, cfg, log); err != nil {
		log.WithComponent("server").WithError(err).Fatal("Server failed")
	}

	log.WithComponent("server").Info("Server stopped")
}

// runServer запускает сервер с полной инициализацией
func runServer(ctx context.Context, cfg *config.Config, log *logger.Logger) error {
	log.WithComponent("server").Info("Initializing server components")

	// 1. Инициализация контейнера зависимостей
	container, err := initializeContainer(ctx, cfg, log)
	if err != nil {
		return fmt.Errorf("failed to initialize container: %w", err)
	}
	defer container.Shutdown()

	// 2. Инициализация HTTP сервера
	httpServer, err := initializeHTTPServer(cfg, container, log)
	if err != nil {
		return fmt.Errorf("failed to initialize HTTP server: %w", err)
	}

	// 3. Запуск HTTP сервера
	serverErrChan := make(chan error, 1)
	go func() {
		if err := httpServer.Start(); err != nil {
			serverErrChan <- fmt.Errorf("HTTP server failed: %w", err)
		}
	}()

	log.WithComponent("server").Info("All components initialized successfully")

	// Ждем сигнала завершения или ошибки сервера
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

	// Остановка HTTP сервера
	if err := httpServer.Stop(shutdownCtx); err != nil {
		log.WithComponent("server").WithError(err).Error("Failed to stop HTTP server gracefully")
		return err
	}

	log.WithComponent("server").Info("Graceful shutdown completed")
	return nil
}

// initializeContainer инициализирует контейнер зависимостей
func initializeContainer(ctx context.Context, cfg *config.Config, log *logger.Logger) (*container.Container, error) {
	log.WithComponent("container").Info("Initializing dependency injection container")

	// Создаем контейнер
	c := container.NewContainer(cfg, log)

	// Инициализируем сервисы
	if err := c.InitializeServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	log.WithComponent("container").Info("Dependency injection container initialized successfully")
	return c, nil
}

// initializeHTTPServer инициализирует HTTP сервер
func initializeHTTPServer(cfg *config.Config, container *container.Container, log *logger.Logger) (*server.Server, error) {
	log.WithComponent("http").Info("Initializing HTTP server")

	// Создаем HTTP сервер
	httpServer := server.NewServer(cfg, container, log)

	// Настраиваем middleware
	httpServer.SetupMiddleware()

	// Настраиваем маршруты
	httpServer.SetupRoutes()

	log.WithComponent("http").Info("HTTP server initialized successfully")
	return httpServer, nil
}
