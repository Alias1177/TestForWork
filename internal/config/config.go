package config

import (
	"fmt"
	"strings"
	"time"

	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config holds application configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Grinex   GrinexConfig   `mapstructure:"grinex"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Tracing  TracingConfig  `mapstructure:"tracing"`
	Metrics  MetricsConfig  `mapstructure:"metrics"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port              int           `mapstructure:"port"`
	GracefulTimeout   time.Duration `mapstructure:"graceful_timeout"`
	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
	MaxConnectionIdle time.Duration `mapstructure:"max_connection_idle"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// GrinexConfig holds Grinex API configuration
type GrinexConfig struct {
	BaseURL string        `mapstructure:"base_url"`
	Timeout time.Duration `mapstructure:"timeout"`
	Market  string        `mapstructure:"market"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// TracingConfig holds tracing configuration
type TracingConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	JaegerURL   string `mapstructure:"jaeger_url"`
	ServiceName string `mapstructure:"service_name"`
}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
	Port    int    `mapstructure:"port"`
}

// Load loads configuration from flags and environment variables
func Load() (*Config, error) {
	// Define command line flags
	flag.Int("server.port", 8080, "Server port")
	flag.Duration("server.graceful-timeout", 30*time.Second, "Graceful shutdown timeout")
	flag.Duration("server.read-timeout", 10*time.Second, "Server read timeout")
	flag.Duration("server.write-timeout", 10*time.Second, "Server write timeout")
	flag.Duration("server.max-connection-idle", 2*time.Minute, "Max connection idle time")

	flag.String("database.host", "localhost", "Database host")
	flag.Int("database.port", 5432, "Database port")
	flag.String("database.user", "postgres", "Database user")
	flag.String("database.password", "postgres", "Database password")
	flag.String("database.database", "usdt_rates", "Database name")
	flag.String("database.ssl-mode", "disable", "Database SSL mode")
	flag.Int("database.max-open-conns", 25, "Database max open connections")
	flag.Int("database.max-idle-conns", 25, "Database max idle connections")
	flag.Duration("database.conn-max-lifetime", 5*time.Minute, "Database connection max lifetime")

	flag.String("grinex.base_url", "https://grinex.io", "Grinex API base URL")
	flag.Duration("grinex.timeout", 10*time.Second, "Grinex API timeout")
	flag.String("grinex.market", "usdtrub", "Trading market pair")

	flag.String("logging.level", "info", "Logging level")
	flag.String("logging.format", "json", "Logging format")

	flag.Bool("tracing.enabled", false, "Enable tracing")
	flag.String("tracing.jaeger-url", "http://localhost:14268/api/traces", "Jaeger URL")
	flag.String("tracing.service-name", "usdt-rates-service", "Service name for tracing")

	flag.Bool("metrics.enabled", true, "Enable metrics")
	flag.String("metrics.path", "/metrics", "Metrics endpoint path")
	flag.Int("metrics.port", 9090, "Metrics server port")

	flag.Parse()

	// Configure viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")

	// Enable environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("USDT")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind flags to viper
	if err := viper.BindPFlags(flag.CommandLine); err != nil {
		return nil, fmt.Errorf("failed to bind flags: %w", err)
	}

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal config
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// DatabaseDSN returns the database connection string
func (c *DatabaseConfig) DatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable connect_timeout=10",
		c.Host, c.Port, c.User, c.Password, c.Database)
}
