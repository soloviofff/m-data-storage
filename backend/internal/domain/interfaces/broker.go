package interfaces

import (
	"context"
	"time"

	"m-data-storage/internal/domain/entities"
)

// BrokerType - тип брокера
type BrokerType string

const (
	BrokerTypeCrypto BrokerType = "crypto"
	BrokerTypeStock  BrokerType = "stock"
)

// BrokerInfo - информация о брокере
type BrokerInfo struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Type        BrokerType `json:"type"`
	Description string     `json:"description"`
	Website     string     `json:"website"`
	Status      string     `json:"status"`
	Features    []string   `json:"features"`
}

// Broker - базовый интерфейс для всех типов брокеров
type Broker interface {
	// Основные методы подключения
	Connect(ctx context.Context) error
	Disconnect() error
	IsConnected() bool

	// Информация о брокере
	GetInfo() BrokerInfo
	GetSupportedInstruments() []entities.InstrumentInfo

	// Управление подписками
	Subscribe(ctx context.Context, instruments []entities.InstrumentSubscription) error
	Unsubscribe(ctx context.Context, instruments []entities.InstrumentSubscription) error

	// Получение каналов данных
	GetTickerChannel() <-chan entities.Ticker
	GetCandleChannel() <-chan entities.Candle
	GetOrderBookChannel() <-chan entities.OrderBook

	// Управление жизненным циклом
	Start(ctx context.Context) error
	Stop() error

	// Проверка здоровья
	Health() error
}

// ConnectionConfig - настройки подключения
type ConnectionConfig struct {
	WebSocketURL         string        `yaml:"websocket_url"`
	RestAPIURL           string        `yaml:"rest_api_url"`
	Timeout              time.Duration `yaml:"timeout"`
	PingInterval         time.Duration `yaml:"ping_interval"`
	ReconnectDelay       time.Duration `yaml:"reconnect_delay"`
	MaxReconnectAttempts int           `yaml:"max_reconnect_attempts"`
}

// AuthConfig - настройки аутентификации
type AuthConfig struct {
	APIKey     string `yaml:"api_key"`
	SecretKey  string `yaml:"secret_key"`
	Passphrase string `yaml:"passphrase,omitempty"`
	Sandbox    bool   `yaml:"sandbox"`
}

// DefaultsConfig - настройки по умолчанию
type DefaultsConfig struct {
	OrderBookDepth    int           `yaml:"orderbook_depth"`
	OrderBookInterval time.Duration `yaml:"orderbook_interval"`
	BufferSize        int           `yaml:"buffer_size"`
	BatchSize         int           `yaml:"batch_size"`
}

// LimitsConfig - лимиты брокера
type LimitsConfig struct {
	MaxSubscriptions  int `yaml:"max_subscriptions"`
	RequestsPerSecond int `yaml:"requests_per_second"`
	RequestsPerMinute int `yaml:"requests_per_minute"`
}

// BrokerConfig - конфигурация брокера
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

// BrokerFactory - фабрика для создания брокеров
type BrokerFactory interface {
	CreateBroker(config BrokerConfig) (Broker, error)
	GetSupportedTypes() []BrokerType
}

// BrokerManager - менеджер для управления брокерами
type BrokerManager interface {
	Initialize(ctx context.Context) error
	GetBroker(id string) (Broker, error)
	GetAllBrokers() map[string]Broker
	StartAll(ctx context.Context) error
	StopAll() error
	HealthCheck() map[string]error
}
