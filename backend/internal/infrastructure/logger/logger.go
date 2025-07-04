package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/lumberjack.v2"

	"m-data-storage/internal/infrastructure/config"
)

// Logger - обертка над logrus с дополнительной функциональностью
type Logger struct {
	*logrus.Logger
	config config.LoggingConfig
}

// New создает новый логгер на основе конфигурации
func New(cfg config.LoggingConfig) (*Logger, error) {
	logger := logrus.New()

	// Устанавливаем уровень логирования
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}
	logger.SetLevel(level)

	// Устанавливаем формат
	switch strings.ToLower(cfg.Format) {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "function",
				logrus.FieldKeyFile:  "file",
			},
		})
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	default:
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	}

	// Устанавливаем вывод
	output, err := getOutput(cfg)
	if err != nil {
		return nil, err
	}
	logger.SetOutput(output)

	// Включаем информацию о вызывающей функции в development режиме
	logger.SetReportCaller(cfg.Level == "debug" || cfg.Level == "trace")

	return &Logger{
		Logger: logger,
		config: cfg,
	}, nil
}

// getOutput определяет куда направлять логи
func getOutput(cfg config.LoggingConfig) (io.Writer, error) {
	switch strings.ToLower(cfg.Output) {
	case "stdout":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	case "file":
		if cfg.File == "" {
			return nil, fmt.Errorf("file path is required when output is 'file'")
		}

		// Создаем директорию если она не существует
		dir := filepath.Dir(cfg.File)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}

		// Используем lumberjack для ротации логов
		return &lumberjack.Logger{
			Filename:   cfg.File,
			MaxSize:    cfg.MaxSize, // MB
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge, // дни
			Compress:   cfg.Compress,
		}, nil
	case "both":
		if cfg.File == "" {
			return nil, fmt.Errorf("file path is required when output is 'both'")
		}

		// Создаем директорию если она не существует
		dir := filepath.Dir(cfg.File)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}

		fileWriter := &lumberjack.Logger{
			Filename:   cfg.File,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}

		return io.MultiWriter(os.Stdout, fileWriter), nil
	default:
		return os.Stdout, nil
	}
}

// WithComponent добавляет компонент к логгеру
func (l *Logger) WithComponent(component string) *logrus.Entry {
	return l.WithField("component", component)
}

// WithBroker добавляет информацию о брокере к логгеру
func (l *Logger) WithBroker(brokerID string) *logrus.Entry {
	return l.WithField("broker_id", brokerID)
}

// WithSymbol добавляет символ к логгеру
func (l *Logger) WithSymbol(symbol string) *logrus.Entry {
	return l.WithField("symbol", symbol)
}

// WithRequest добавляет информацию о запросе к логгеру
func (l *Logger) WithRequest(requestID, method, path string) *logrus.Entry {
	return l.WithFields(logrus.Fields{
		"request_id": requestID,
		"method":     method,
		"path":       path,
	})
}

// WithError добавляет ошибку к логгеру
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.Logger.WithError(err)
}

// WithDuration добавляет продолжительность операции к логгеру
func (l *Logger) WithDuration(duration string) *logrus.Entry {
	return l.WithField("duration", duration)
}

// WithCount добавляет количество к логгеру
func (l *Logger) WithCount(count int64) *logrus.Entry {
	return l.WithField("count", count)
}

// WithDataType добавляет тип данных к логгеру
func (l *Logger) WithDataType(dataType string) *logrus.Entry {
	return l.WithField("data_type", dataType)
}

// LogDataReceived логирует получение данных
func (l *Logger) LogDataReceived(brokerID, symbol, dataType string, count int64) {
	l.WithFields(logrus.Fields{
		"broker_id": brokerID,
		"symbol":    symbol,
		"data_type": dataType,
		"count":     count,
	}).Debug("Data received")
}

// LogDataProcessed логирует обработку данных
func (l *Logger) LogDataProcessed(dataType string, count int64, duration string) {
	l.WithFields(logrus.Fields{
		"data_type": dataType,
		"count":     count,
		"duration":  duration,
	}).Info("Data processed")
}

// LogBrokerConnected логирует подключение к брокеру
func (l *Logger) LogBrokerConnected(brokerID string) {
	l.WithBroker(brokerID).Info("Broker connected")
}

// LogBrokerDisconnected логирует отключение от брокера
func (l *Logger) LogBrokerDisconnected(brokerID string, reason string) {
	l.WithFields(logrus.Fields{
		"broker_id": brokerID,
		"reason":    reason,
	}).Warn("Broker disconnected")
}

// LogBrokerError логирует ошибку брокера
func (l *Logger) LogBrokerError(brokerID string, err error) {
	l.WithBroker(brokerID).WithError(err).Error("Broker error")
}

// LogAPIRequest логирует API запрос
func (l *Logger) LogAPIRequest(requestID, method, path, userAgent string, duration string, statusCode int) {
	entry := l.WithFields(logrus.Fields{
		"request_id":  requestID,
		"method":      method,
		"path":        path,
		"user_agent":  userAgent,
		"duration":    duration,
		"status_code": statusCode,
	})

	if statusCode >= 400 {
		entry.Warn("API request completed with error")
	} else {
		entry.Info("API request completed")
	}
}

// LogStorageOperation логирует операцию с хранилищем
func (l *Logger) LogStorageOperation(operation, table string, count int64, duration string) {
	l.WithFields(logrus.Fields{
		"operation": operation,
		"table":     table,
		"count":     count,
		"duration":  duration,
	}).Debug("Storage operation completed")
}

// LogHealthCheck логирует проверку здоровья
func (l *Logger) LogHealthCheck(component string, healthy bool, details string) {
	entry := l.WithFields(logrus.Fields{
		"component": component,
		"healthy":   healthy,
		"details":   details,
	})

	if healthy {
		entry.Debug("Health check passed")
	} else {
		entry.Warn("Health check failed")
	}
}

// Close закрывает логгер и освобождает ресурсы
func (l *Logger) Close() error {
	// Если используется файловый вывод, закрываем его
	if closer, ok := l.Logger.Out.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
