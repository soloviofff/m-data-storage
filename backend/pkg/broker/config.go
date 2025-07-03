package broker

import "time"

// BrokerConfig represents the configuration for a broker
type BrokerConfig struct {
	ID      string     `yaml:"id"`
	Name    string     `yaml:"name"`
	Type    BrokerType `yaml:"type"`
	Enabled bool       `yaml:"enabled"`

	// Connection settings
	Connection ConnectionConfig `yaml:"connection"`

	// Authentication settings
	Auth AuthConfig `yaml:"auth"`

	// Default settings
	Defaults DefaultsConfig `yaml:"defaults"`

	// Broker-specific settings
	Settings map[string]interface{} `yaml:"settings"`

	// Rate limits
	Limits LimitsConfig `yaml:"limits"`
}

// ConnectionConfig represents connection settings
type ConnectionConfig struct {
	WebSocketURL         string        `yaml:"websocket_url"`
	RestAPIURL           string        `yaml:"rest_api_url"`
	Timeout              time.Duration `yaml:"timeout"`
	PingInterval         time.Duration `yaml:"ping_interval"`
	ReconnectDelay       time.Duration `yaml:"reconnect_delay"`
	MaxReconnectAttempts int           `yaml:"max_reconnect_attempts"`
}

// AuthConfig represents authentication settings
type AuthConfig struct {
	APIKey     string `yaml:"api_key"`
	SecretKey  string `yaml:"secret_key"`
	Passphrase string `yaml:"passphrase,omitempty"`
	Sandbox    bool   `yaml:"sandbox"`
}

// DefaultsConfig represents default settings
type DefaultsConfig struct {
	OrderBookDepth    int           `yaml:"orderbook_depth"`
	OrderBookInterval time.Duration `yaml:"orderbook_interval"`
	BufferSize        int           `yaml:"buffer_size"`
	BatchSize         int           `yaml:"batch_size"`
}

// LimitsConfig represents rate limits
type LimitsConfig struct {
	MaxSubscriptions  int `yaml:"max_subscriptions"`
	RequestsPerSecond int `yaml:"requests_per_second"`
	RequestsPerMinute int `yaml:"requests_per_minute"`
}
