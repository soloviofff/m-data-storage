package broker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"m-data-storage/internal/domain/interfaces"
)

// ConnectionState represents connection state
type ConnectionState string

const (
	StateDisconnected ConnectionState = "disconnected"
	StateConnecting   ConnectionState = "connecting"
	StateConnected    ConnectionState = "connected"
	StateReconnecting ConnectionState = "reconnecting"
	StateError        ConnectionState = "error"
)

// ConnectionInfo contains connection information
type ConnectionInfo struct {
	State           ConnectionState `json:"state"`
	ConnectedAt     time.Time       `json:"connected_at"`
	LastError       string          `json:"last_error"`
	ReconnectCount  int             `json:"reconnect_count"`
	LastReconnectAt time.Time       `json:"last_reconnect_at"`
}

// ConnectionManager manages broker connections
type ConnectionManager struct {
	config interfaces.ConnectionConfig
	logger *logrus.Logger
	state  ConnectionState
	info   ConnectionInfo
	mu     sync.RWMutex

	// Reconnection management
	reconnectTimer *time.Timer
	ctx            context.Context
	cancel         context.CancelFunc

	// Callbacks
	onConnect    func() error
	onDisconnect func() error
	onError      func(error)
}

// NewConnectionManager creates a new connection manager
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

// SetCallbacks sets callbacks for connection events
func (cm *ConnectionManager) SetCallbacks(onConnect, onDisconnect func() error, onError func(error)) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.onConnect = onConnect
	cm.onDisconnect = onDisconnect
	cm.onError = onError
}

// Connect establishes connection
func (cm *ConnectionManager) Connect(ctx context.Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.state == StateConnected {
		return nil
	}

	cm.setState(StateConnecting)
	cm.logger.Info("Establishing connection")

	// Call connection callback
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

// Disconnect breaks connection
func (cm *ConnectionManager) Disconnect() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.state == StateDisconnected {
		return nil
	}

	cm.logger.Info("Disconnecting")

	// Stop reconnection timer
	if cm.reconnectTimer != nil {
		cm.reconnectTimer.Stop()
		cm.reconnectTimer = nil
	}

	// Cancel context
	cm.cancel()

	// Call disconnection callback
	if cm.onDisconnect != nil {
		if err := cm.onDisconnect(); err != nil {
			cm.logger.WithError(err).Warn("Disconnect callback failed")
		}
	}

	cm.setState(StateDisconnected)
	cm.logger.Info("Disconnected successfully")

	return nil
}

// IsConnected returns connection status
func (cm *ConnectionManager) IsConnected() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.state == StateConnected
}

// GetState returns current connection state
func (cm *ConnectionManager) GetState() ConnectionState {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.state
}

// GetInfo returns connection information
func (cm *ConnectionManager) GetInfo() ConnectionInfo {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.info
}

// HandleError handles connection error
func (cm *ConnectionManager) HandleError(err error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.logger.WithError(err).Error("Connection error occurred")

	// Call error callback
	if cm.onError != nil {
		cm.onError(err)
	}

	cm.setError(err)

	// Start reconnection if configured
	if cm.config.MaxReconnectAttempts > 0 && cm.info.ReconnectCount < cm.config.MaxReconnectAttempts {
		cm.scheduleReconnect()
	}
}

// setState sets connection state
func (cm *ConnectionManager) setState(state ConnectionState) {
	cm.state = state
	cm.info.State = state
}

// setError sets error
func (cm *ConnectionManager) setError(err error) {
	cm.setState(StateError)
	cm.info.LastError = err.Error()
}

// scheduleReconnect schedules reconnection
func (cm *ConnectionManager) scheduleReconnect() {
	if cm.reconnectTimer != nil {
		cm.reconnectTimer.Stop()
	}

	cm.setState(StateReconnecting)
	cm.info.ReconnectCount++
	cm.info.LastReconnectAt = time.Now()

	delay := cm.config.ReconnectDelay
	if delay == 0 {
		delay = 5 * time.Second // Default value
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

		// Try to reconnect
		if err := cm.Connect(cm.ctx); err != nil {
			cm.logger.WithError(err).Error("Reconnection failed")

			// Schedule next attempt if limit not reached
			if cm.info.ReconnectCount < cm.config.MaxReconnectAttempts {
				cm.scheduleReconnect()
			} else {
				cm.logger.Error("Maximum reconnection attempts reached")
				cm.setState(StateError)
			}
		} else {
			cm.logger.Info("Reconnection successful")
			cm.info.ReconnectCount = 0 // Reset counter on successful connection
		}
	})
}

// StartHealthCheck starts periodic connection health check
func (cm *ConnectionManager) StartHealthCheck(healthCheck func() error) {
	if cm.config.PingInterval == 0 {
		return // Health check disabled
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

// Shutdown gracefully shuts down the connection manager
func (cm *ConnectionManager) Shutdown() error {
	return cm.Disconnect()
}
