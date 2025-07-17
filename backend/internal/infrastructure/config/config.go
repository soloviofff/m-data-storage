package config

import (
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Config - main application configuration
type Config struct {
	App      AppConfig      `yaml:"app" env-prefix:"APP_"`
	Database DatabaseConfig `yaml:"database" env-prefix:"DB_"`
	API      APIConfig      `yaml:"api" env-prefix:"API_"`
	Logging  LoggingConfig  `yaml:"logging" env-prefix:"LOG_"`
	Storage  StorageConfig  `yaml:"storage" env-prefix:"STORAGE_"`
	Brokers  BrokersConfig  `yaml:"brokers" env-prefix:"BROKERS_"`
}

// AppConfig - application configuration
type AppConfig struct {
	Name        string        `yaml:"name" env:"NAME" envDefault:"m-data-storage"`
	Version     string        `yaml:"version" env:"VERSION" envDefault:"1.0.0"`
	Environment string        `yaml:"environment" env:"ENVIRONMENT" envDefault:"development"`
	Debug       bool          `yaml:"debug" env:"DEBUG" envDefault:"false"`
	Timeout     time.Duration `yaml:"timeout" env:"TIMEOUT" envDefault:"30s"`
}

// DatabaseConfig - database configuration
type DatabaseConfig struct {
	SQLite  SQLiteConfig  `yaml:"sqlite" env-prefix:"SQLITE_"`
	QuestDB QuestDBConfig `yaml:"questdb" env-prefix:"QUESTDB_"`
}

// SQLiteConfig - SQLite configuration
type SQLiteConfig struct {
	Path            string        `yaml:"path" env:"PATH" envDefault:"./data/metadata.db"`
	MaxOpenConns    int           `yaml:"max_open_conns" env:"MAX_OPEN_CONNS" envDefault:"10"`
	MaxIdleConns    int           `yaml:"max_idle_conns" env:"MAX_IDLE_CONNS" envDefault:"5"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env:"CONN_MAX_LIFETIME" envDefault:"1h"`
	WALMode         bool          `yaml:"wal_mode" env:"WAL_MODE" envDefault:"true"`
	ForeignKeys     bool          `yaml:"foreign_keys" env:"FOREIGN_KEYS" envDefault:"true"`
}

// QuestDBConfig - QuestDB configuration
type QuestDBConfig struct {
	Host            string        `yaml:"host" env:"HOST" envDefault:"localhost"`
	Port            int           `yaml:"port" env:"PORT" envDefault:"8812"`
	Database        string        `yaml:"database" env:"DATABASE" envDefault:"qdb"`
	Username        string        `yaml:"username" env:"USERNAME"`
	Password        string        `yaml:"password" env:"PASSWORD"`
	MaxOpenConns    int           `yaml:"max_open_conns" env:"MAX_OPEN_CONNS" envDefault:"20"`
	MaxIdleConns    int           `yaml:"max_idle_conns" env:"MAX_IDLE_CONNS" envDefault:"10"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env:"CONN_MAX_LIFETIME" envDefault:"1h"`
	QueryTimeout    time.Duration `yaml:"query_timeout" env:"QUERY_TIMEOUT" envDefault:"30s"`
}

// APIConfig - API configuration
type APIConfig struct {
	Host            string        `yaml:"host" env:"HOST" envDefault:"0.0.0.0"`
	Port            int           `yaml:"port" env:"PORT" envDefault:"8080"`
	ReadTimeout     time.Duration `yaml:"read_timeout" env:"READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env:"WRITE_TIMEOUT" envDefault:"10s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"SHUTDOWN_TIMEOUT" envDefault:"30s"`
	CORS            CORSConfig    `yaml:"cors" env-prefix:"CORS_"`
	Auth            AuthConfig    `yaml:"auth" env-prefix:"AUTH_"`
}

// CORSConfig - CORS configuration
type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins" env:"ALLOWED_ORIGINS" envSeparator:","`
	AllowedMethods []string `yaml:"allowed_methods" env:"ALLOWED_METHODS" envSeparator:"," envDefault:"GET,POST,PUT,DELETE,OPTIONS"`
	AllowedHeaders []string `yaml:"allowed_headers" env:"ALLOWED_HEADERS" envSeparator:"," envDefault:"Content-Type,Authorization,X-API-Key"`
}

// AuthConfig - authentication configuration
type AuthConfig struct {
	JWTSecret    string        `yaml:"jwt_secret" env:"JWT_SECRET"`
	JWTExpiry    time.Duration `yaml:"jwt_expiry" env:"JWT_EXPIRY" envDefault:"24h"`
	APIKeyHeader string        `yaml:"api_key_header" env:"API_KEY_HEADER" envDefault:"X-API-Key"`
	Enabled      bool          `yaml:"enabled" env:"ENABLED" envDefault:"false"`
}

// LoggingConfig - logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level" env:"LEVEL" envDefault:"info"`
	Format     string `yaml:"format" env:"FORMAT" envDefault:"json"`
	Output     string `yaml:"output" env:"OUTPUT" envDefault:"stdout"`
	File       string `yaml:"file" env:"FILE"`
	MaxSize    int    `yaml:"max_size" env:"MAX_SIZE" envDefault:"100"`
	MaxBackups int    `yaml:"max_backups" env:"MAX_BACKUPS" envDefault:"3"`
	MaxAge     int    `yaml:"max_age" env:"MAX_AGE" envDefault:"28"`
	Compress   bool   `yaml:"compress" env:"COMPRESS" envDefault:"true"`
}

// StorageConfig - storage configuration
type StorageConfig struct {
	RetentionPeriod time.Duration `yaml:"retention_period" env:"RETENTION_PERIOD" envDefault:"720h"` // 30 days
	VacuumInterval  time.Duration `yaml:"vacuum_interval" env:"VACUUM_INTERVAL" envDefault:"24h"`
	MaxStorageSize  int64         `yaml:"max_storage_size" env:"MAX_STORAGE_SIZE" envDefault:"10737418240"` // 10GB
	BatchSize       int           `yaml:"batch_size" env:"BATCH_SIZE" envDefault:"1000"`
	FlushInterval   time.Duration `yaml:"flush_interval" env:"FLUSH_INTERVAL" envDefault:"5s"`
}

// BrokersConfig - brokers configuration
type BrokersConfig struct {
	ConfigPath          string        `yaml:"config_path" env:"CONFIG_PATH" envDefault:"./configs/brokers"`
	ReconnectDelay      time.Duration `yaml:"reconnect_delay" env:"RECONNECT_DELAY" envDefault:"5s"`
	MaxReconnects       int           `yaml:"max_reconnects" env:"MAX_RECONNECTS" envDefault:"10"`
	HealthCheckInterval time.Duration `yaml:"health_check_interval" env:"HEALTH_CHECK_INTERVAL" envDefault:"30s"`
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	config := &Config{}

	// Load from file if it exists
	if configPath != "" {
		if _, err := os.Stat(configPath); err == nil {
			data, err := os.ReadFile(configPath)
			if err != nil {
				return nil, errors.Wrap(err, "failed to read config file")
			}

			if err := yaml.Unmarshal(data, config); err != nil {
				return nil, errors.Wrap(err, "failed to parse config file")
			}
		}
	}

	// Override with environment variables
	if err := env.Parse(config); err != nil {
		return nil, errors.Wrap(err, "failed to parse environment variables")
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "config validation failed")
	}

	return config, nil
}

// Validate checks configuration validity
func (c *Config) Validate() error {
	if c.App.Name == "" {
		return errors.New("app name is required")
	}

	if c.API.Port <= 0 || c.API.Port > 65535 {
		return errors.New("invalid API port")
	}

	if c.Database.QuestDB.Port <= 0 || c.Database.QuestDB.Port > 65535 {
		return errors.New("invalid QuestDB port")
	}

	if c.Storage.BatchSize <= 0 {
		return errors.New("batch size must be positive")
	}

	if c.Storage.RetentionPeriod <= 0 {
		return errors.New("retention period must be positive")
	}

	return nil
}

// IsDevelopment checks if the application is running in development mode
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// IsProduction checks if the application is running in production
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

// GetDatabaseURL returns URL for QuestDB connection
func (c *Config) GetDatabaseURL() string {
	if c.Database.QuestDB.Username != "" && c.Database.QuestDB.Password != "" {
		return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
			c.Database.QuestDB.Username,
			c.Database.QuestDB.Password,
			c.Database.QuestDB.Host,
			c.Database.QuestDB.Port,
			c.Database.QuestDB.Database)
	}

	return fmt.Sprintf("postgresql://%s:%d/%s",
		c.Database.QuestDB.Host,
		c.Database.QuestDB.Port,
		c.Database.QuestDB.Database)
}
