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
	Path            string        `yaml:"path" env:"SQLITE_PATH" envDefault:"./data/metadata.db"`
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
	Enabled  bool           `yaml:"enabled" env:"ENABLED" envDefault:"true"`
	JWT      JWTConfig      `yaml:"jwt" env-prefix:"JWT_"`
	APIKey   APIKeyConfig   `yaml:"api_key" env-prefix:"API_KEY_"`
	Password PasswordConfig `yaml:"password" env-prefix:"PASSWORD_"`
	Security SecurityConfig `yaml:"security" env-prefix:"SECURITY_"`
}

// JWTConfig - JWT configuration
type JWTConfig struct {
	Secret             string        `yaml:"secret" env:"SECRET" envDefault:"your-secret-key-change-in-production"`
	AccessTokenExpiry  time.Duration `yaml:"access_token_expiry" env:"ACCESS_TOKEN_EXPIRY" envDefault:"1h"`
	RefreshTokenExpiry time.Duration `yaml:"refresh_token_expiry" env:"REFRESH_TOKEN_EXPIRY" envDefault:"168h"` // 7 days
	Issuer             string        `yaml:"issuer" env:"ISSUER" envDefault:"m-data-storage"`
	Audience           string        `yaml:"audience" env:"AUDIENCE" envDefault:"m-data-storage-api"`
}

// APIKeyConfig - API key configuration
type APIKeyConfig struct {
	Header        string        `yaml:"header" env:"HEADER" envDefault:"X-API-Key"`
	Prefix        string        `yaml:"prefix" env:"PREFIX" envDefault:"mds_"`
	Length        int           `yaml:"length" env:"LENGTH" envDefault:"32"`
	DefaultExpiry time.Duration `yaml:"default_expiry" env:"DEFAULT_EXPIRY" envDefault:"8760h"` // 1 year
}

// PasswordConfig - password configuration
type PasswordConfig struct {
	MinLength      int    `yaml:"min_length" env:"MIN_LENGTH" envDefault:"8"`
	RequireUpper   bool   `yaml:"require_upper" env:"REQUIRE_UPPER" envDefault:"true"`
	RequireLower   bool   `yaml:"require_lower" env:"REQUIRE_LOWER" envDefault:"true"`
	RequireDigit   bool   `yaml:"require_digit" env:"REQUIRE_DIGIT" envDefault:"true"`
	RequireSpecial bool   `yaml:"require_special" env:"REQUIRE_SPECIAL" envDefault:"false"`
	Argon2Memory   uint32 `yaml:"argon2_memory" env:"ARGON2_MEMORY" envDefault:"65536"` // 64MB
	Argon2Time     uint32 `yaml:"argon2_time" env:"ARGON2_TIME" envDefault:"3"`         // 3 iterations
	Argon2Threads  uint8  `yaml:"argon2_threads" env:"ARGON2_THREADS" envDefault:"2"`   // 2 threads
	Argon2KeyLen   uint32 `yaml:"argon2_keylen" env:"ARGON2_KEYLEN" envDefault:"32"`    // 32 bytes
}

// SecurityConfig - security configuration
type SecurityConfig struct {
	MaxFailedAttempts int           `yaml:"max_failed_attempts" env:"MAX_FAILED_ATTEMPTS" envDefault:"5"`
	LockoutDuration   time.Duration `yaml:"lockout_duration" env:"LOCKOUT_DURATION" envDefault:"15m"`
	SessionTimeout    time.Duration `yaml:"session_timeout" env:"SESSION_TIMEOUT" envDefault:"24h"`
	RequireHTTPS      bool          `yaml:"require_https" env:"REQUIRE_HTTPS" envDefault:"false"`
	CSRFProtection    bool          `yaml:"csrf_protection" env:"CSRF_PROTECTION" envDefault:"true"`
	RateLimitRequests int           `yaml:"rate_limit_requests" env:"RATE_LIMIT_REQUESTS" envDefault:"100"`
	RateLimitWindow   time.Duration `yaml:"rate_limit_window" env:"RATE_LIMIT_WINDOW" envDefault:"1m"`
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
