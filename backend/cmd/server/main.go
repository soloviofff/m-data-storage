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

	// TODO: Полная инициализация компонентов
	// 1. Инициализация хранилищ (SQLite + QuestDB)
	// 2. Инициализация сервисов (валидатор, процессор данных)
	// 3. Инициализация менеджера брокеров
	// 4. Инициализация HTTP сервера
	// 5. Запуск сбора данных

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.API.ShutdownTimeout)
	defer shutdownCancel()

	// Ждем сигнала завершения
	<-ctx.Done()

	log.WithComponent("server").Info("Starting graceful shutdown")

	// TODO: Graceful shutdown всех компонентов
	// 1. Остановка сбора данных
	// 2. Остановка HTTP сервера
	// 3. Остановка брокеров
	// 4. Закрытие хранилищ

	// Даем время на завершение операций
	select {
	case <-time.After(1 * time.Second):
	case <-shutdownCtx.Done():
		return shutdownCtx.Err()
	}

	return nil
}
