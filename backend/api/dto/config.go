package dto

import "time"

// SystemConfig represents system configuration in API requests/responses
type SystemConfig struct {
	// Storage settings
	StorageRetention time.Duration `json:"storage_retention" validate:"required"`
	VacuumInterval   time.Duration `json:"vacuum_interval" validate:"required"`
	MaxStorageSize   int64         `json:"max_storage_size" validate:"required,min=1"`

	// API settings
	APIPort         int           `json:"api_port" validate:"required,min=1,max=65535"`
	APIHost         string        `json:"api_host" validate:"required"`
	ReadTimeout     time.Duration `json:"read_timeout" validate:"required"`
	WriteTimeout    time.Duration `json:"write_timeout" validate:"required"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout" validate:"required"`

	// Security settings
	JWTSecret      string   `json:"jwt_secret" validate:"required,min=32"`
	APIKeyHeader   string   `json:"api_key_header" validate:"required"`
	AllowedOrigins []string `json:"allowed_origins" validate:"required,min=1"`

	// Monitoring settings
	MetricsEnabled bool `json:"metrics_enabled"`
	MetricsPort    int  `json:"metrics_port" validate:"required_if=MetricsEnabled true,min=1,max=65535"`
	TracingEnabled bool `json:"tracing_enabled"`
}
