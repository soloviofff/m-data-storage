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

	log.WithComponent("main").Info("Starting M-Data-Storage application",
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
		log.WithComponent("main").Info("Received shutdown signal", "signal", sig.String())
		cancel()
	}()

	// TODO: Инициализация компонентов приложения
	if err := initializeApplication(ctx, cfg, log); err != nil {
		log.WithComponent("main").WithError(err).Fatal("Failed to initialize application")
	}

	log.WithComponent("main").Info("Application started successfully")

	// Ждем сигнала завершения
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

// initializeApplication инициализирует все компоненты приложения
func initializeApplication(ctx context.Context, cfg *config.Config, log *logger.Logger) error {
	log.WithComponent("init").Info("Initializing application components")

	// TODO: Инициализация компонентов
	// 1. Подключение к базам данных (SQLite + QuestDB)
	// 2. Инициализация хранилищ
	// 3. Инициализация сервисов
	// 4. Загрузка конфигураций брокеров
	// 5. Инициализация брокеров
	// 6. Запуск веб-сервера
	// 7. Запуск сервиса сбора данных

	log.WithComponent("init").Info("Application components initialized successfully")
	return nil
}

// shutdownApplication выполняет graceful shutdown всех компонентов
func shutdownApplication(ctx context.Context, log *logger.Logger) error {
	log.WithComponent("shutdown").Info("Shutting down application components")

	// TODO: Graceful shutdown компонентов в обратном порядке
	// 1. Остановка сбора данных
	// 2. Остановка веб-сервера
	// 3. Остановка брокеров
	// 4. Остановка сервисов
	// 5. Закрытие соединений с базами данных

	// Даем время на завершение операций
	select {
	case <-time.After(1 * time.Second):
	case <-ctx.Done():
		return ctx.Err()
	}

	log.WithComponent("shutdown").Info("Application components shut down successfully")
	return nil
}
