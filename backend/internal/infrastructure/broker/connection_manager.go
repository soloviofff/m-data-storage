package broker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"m-data-storage/internal/domain/interfaces"
)

// ConnectionState представляет состояние соединения
type ConnectionState string

const (
	StateDisconnected ConnectionState = "disconnected"
	StateConnecting   ConnectionState = "connecting"
	StateConnected    ConnectionState = "connected"
	StateReconnecting ConnectionState = "reconnecting"
	StateError        ConnectionState = "error"
)

// ConnectionInfo содержит информацию о соединении
type ConnectionInfo struct {
	State           ConnectionState `json:"state"`
	ConnectedAt     time.Time       `json:"connected_at"`
	LastError       string          `json:"last_error"`
	ReconnectCount  int             `json:"reconnect_count"`
	LastReconnectAt time.Time       `json:"last_reconnect_at"`
}

// ConnectionManager управляет соединениями брокеров
type ConnectionManager struct {
	config      interfaces.ConnectionConfig
	logger      *logrus.Logger
	state       ConnectionState
	info        ConnectionInfo
	mu          sync.RWMutex
	
	// Управление переподключением
	reconnectTimer *time.Timer
	ctx            context.Context
	cancel         context.CancelFunc
	
	// Колбэки
	onConnect    func() error
	onDisconnect func() error
	onError      func(error)
}

// NewConnectionManager создает новый менеджер соединений
func NewConnectionManager(config interfaces.ConnectionConfig, logger *logrus.Logger) *ConnectionManager {
	if logger == nil {
		logger = logrus.New()
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	return &ConnectionManager{
		config: config,
		logger: logger,
		state:  StateDisconnected,
		info: ConnectionInfo{
			State: StateDisconnected,
		},
		ctx:    ctx,
		cancel: cancel,
	}
}

// SetCallbacks устанавливает колбэки для событий соединения
func (cm *ConnectionManager) SetCallbacks(onConnect, onDisconnect func() error, onError func(error)) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	cm.onConnect = onConnect
	cm.onDisconnect = onDisconnect
	cm.onError = onError
}

// Connect устанавливает соединение
func (cm *ConnectionManager) Connect(ctx context.Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	if cm.state == StateConnected {
		return nil
	}
	
	cm.setState(StateConnecting)
	cm.logger.Info("Establishing connection")
	
	// Вызываем колбэк подключения
	if cm.onConnect != nil {
		if err := cm.onConnect(); err != nil {
			cm.setError(fmt.Errorf("connection callback failed: %w", err))
			return err
		}
	}
	
	cm.setState(StateConnected)
	cm.info.ConnectedAt = time.Now()
	cm.logger.Info("Connection established successfully")
	
	return nil
}

// Disconnect разрывает соединение
func (cm *ConnectionManager) Disconnect() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	if cm.state == StateDisconnected {
		return nil
	}
	
	cm.logger.Info("Disconnecting")
	
	// Останавливаем таймер переподключения
	if cm.reconnectTimer != nil {
		cm.reconnectTimer.Stop()
		cm.reconnectTimer = nil
	}
	
	// Отменяем контекст
	cm.cancel()
	
	// Вызываем колбэк отключения
	if cm.onDisconnect != nil {
		if err := cm.onDisconnect(); err != nil {
			cm.logger.WithError(err).Warn("Disconnect callback failed")
		}
	}
	
	cm.setState(StateDisconnected)
	cm.logger.Info("Disconnected successfully")
	
	return nil
}

// IsConnected возвращает статус соединения
func (cm *ConnectionManager) IsConnected() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.state == StateConnected
}

// GetState возвращает текущее состояние соединения
func (cm *ConnectionManager) GetState() ConnectionState {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.state
}

// GetInfo возвращает информацию о соединении
func (cm *ConnectionManager) GetInfo() ConnectionInfo {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.info
}

// HandleError обрабатывает ошибку соединения
func (cm *ConnectionManager) HandleError(err error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	cm.logger.WithError(err).Error("Connection error occurred")
	
	// Вызываем колбэк ошибки
	if cm.onError != nil {
		cm.onError(err)
	}
	
	cm.setError(err)
	
	// Запускаем переподключение, если настроено
	if cm.config.MaxReconnectAttempts > 0 && cm.info.ReconnectCount < cm.config.MaxReconnectAttempts {
		cm.scheduleReconnect()
	}
}

// setState устанавливает состояние соединения
func (cm *ConnectionManager) setState(state ConnectionState) {
	cm.state = state
	cm.info.State = state
}

// setError устанавливает ошибку
func (cm *ConnectionManager) setError(err error) {
	cm.setState(StateError)
	cm.info.LastError = err.Error()
}

// scheduleReconnect планирует переподключение
func (cm *ConnectionManager) scheduleReconnect() {
	if cm.reconnectTimer != nil {
		cm.reconnectTimer.Stop()
	}
	
	cm.setState(StateReconnecting)
	cm.info.ReconnectCount++
	cm.info.LastReconnectAt = time.Now()
	
	delay := cm.config.ReconnectDelay
	if delay == 0 {
		delay = 5 * time.Second // Значение по умолчанию
	}
	
	cm.logger.WithFields(logrus.Fields{
		"attempt": cm.info.ReconnectCount,
		"delay":   delay,
	}).Info("Scheduling reconnection")
	
	cm.reconnectTimer = time.AfterFunc(delay, func() {
		cm.mu.Lock()
		defer cm.mu.Unlock()
		
		if cm.state != StateReconnecting {
			return
		}
		
		cm.logger.WithField("attempt", cm.info.ReconnectCount).Info("Attempting to reconnect")
		
		// Пытаемся переподключиться
		if err := cm.Connect(cm.ctx); err != nil {
			cm.logger.WithError(err).Error("Reconnection failed")
			
			// Планируем следующую попытку, если не достигли лимита
			if cm.info.ReconnectCount < cm.config.MaxReconnectAttempts {
				cm.scheduleReconnect()
			} else {
				cm.logger.Error("Maximum reconnection attempts reached")
				cm.setState(StateError)
			}
		} else {
			cm.logger.Info("Reconnection successful")
			cm.info.ReconnectCount = 0 // Сбрасываем счетчик при успешном подключении
		}
	})
}

// StartHealthCheck запускает периодическую проверку здоровья соединения
func (cm *ConnectionManager) StartHealthCheck(healthCheck func() error) {
	if cm.config.PingInterval == 0 {
		return // Проверка здоровья отключена
	}
	
	go func() {
		ticker := time.NewTicker(cm.config.PingInterval)
		defer ticker.Stop()
		
		for {
			select {
			case <-cm.ctx.Done():
				return
			case <-ticker.C:
				if cm.IsConnected() && healthCheck != nil {
					if err := healthCheck(); err != nil {
						cm.HandleError(fmt.Errorf("health check failed: %w", err))
					}
				}
			}
		}
	}()
}

// Shutdown корректно завершает работу менеджера соединений
func (cm *ConnectionManager) Shutdown() error {
	return cm.Disconnect()
}
