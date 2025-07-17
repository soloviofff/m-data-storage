package interfaces

import (
	"context"
	"time"

	"m-data-storage/internal/domain/entities"
)

// BrokerType - broker type
type BrokerType string

const (
	BrokerTypeCrypto BrokerType = "crypto"
	BrokerTypeStock  BrokerType = "stock"
)

// BrokerInfo - broker information
type BrokerInfo struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Type        BrokerType `json:"type"`
	Description string     `json:"description"`
	Website     string     `json:"website"`
	Status      string     `json:"status"`
	Features    []string   `json:"features"`
}

// Broker - base interface for all broker types
type Broker interface {
	// Basic connection methods
	Connect(ctx context.Context) error
	Disconnect() error
	IsConnected() bool

	// Broker information
	GetInfo() BrokerInfo
	GetSupportedInstruments() []entities.InstrumentInfo

	// Subscription management
	Subscribe(ctx context.Context, instruments []entities.InstrumentSubscription) error
	Unsubscribe(ctx context.Context, instruments []entities.InstrumentSubscription) error

	// Data channel access
	GetTickerChannel() <-chan entities.Ticker
	GetCandleChannel() <-chan entities.Candle
	GetOrderBookChannel() <-chan entities.OrderBook

	// Lifecycle management
	Start(ctx context.Context) error
	Stop() error

	// Health check
	Health() error
}

// ConnectionConfig - connection settings
type ConnectionConfig struct {
	WebSocketURL         string        `yaml:"websocket_url"`
	RestAPIURL           string        `yaml:"rest_api_url"`
	Timeout              time.Duration `yaml:"timeout"`
	PingInterval         time.Duration `yaml:"ping_interval"`
	ReconnectDelay       time.Duration `yaml:"reconnect_delay"`
	MaxReconnectAttempts int           `yaml:"max_reconnect_attempts"`
}

// AuthConfig - authentication settings
type AuthConfig struct {
	APIKey     string `yaml:"api_key"`
	SecretKey  string `yaml:"secret_key"`
	Passphrase string `yaml:"passphrase,omitempty"`
	Sandbox    bool   `yaml:"sandbox"`
}

// DefaultsConfig - default settings
type DefaultsConfig struct {
	OrderBookDepth    int           `yaml:"orderbook_depth"`
	OrderBookInterval time.Duration `yaml:"orderbook_interval"`
	BufferSize        int           `yaml:"buffer_size"`
	BatchSize         int           `yaml:"batch_size"`
}

// LimitsConfig - broker limits
type LimitsConfig struct {
	MaxSubscriptions  int `yaml:"max_subscriptions"`
	RequestsPerSecond int `yaml:"requests_per_second"`
	RequestsPerMinute int `yaml:"requests_per_minute"`
}

// BrokerConfig - broker configuration
type BrokerConfig struct {
	ID         string                 `yaml:"id"`
	Name       string                 `yaml:"name"`
	Type       BrokerType             `yaml:"type"`
	Enabled    bool                   `yaml:"enabled"`
	Connection ConnectionConfig       `yaml:"connection"`
	Auth       AuthConfig             `yaml:"auth"`
	Defaults   DefaultsConfig         `yaml:"defaults"`
	Settings   map[string]interface{} `yaml:"settings"`
	Limits     LimitsConfig           `yaml:"limits"`
}

// BrokerFactory - factory for creating brokers
type BrokerFactory interface {
	CreateBroker(config BrokerConfig) (Broker, error)
	GetSupportedTypes() []BrokerType
}

// BrokerManager - manager for broker management
type BrokerManager interface {
	Initialize(ctx context.Context) error

	// Broker management
	AddBroker(ctx context.Context, config BrokerConfig) error
	RemoveBroker(ctx context.Context, brokerID string) error
	GetBroker(id string) (Broker, error)
	GetAllBrokers() map[string]Broker
	ListBrokers() []BrokerInfo

	// Lifecycle management
	StartAll(ctx context.Context) error
	StopAll() error

	// Monitoring
	Health() map[string]error
	HealthCheck() map[string]error // Deprecated: use Health() instead
}
