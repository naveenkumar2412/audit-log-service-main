package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server       ServerConfig       `mapstructure:"server"`
	Database     DatabaseConfig     `mapstructure:"database"`
	Auth         AuthConfig         `mapstructure:"auth"`
	Audit        AuditConfig        `mapstructure:"audit"`
	Notification NotificationConfig `mapstructure:"notification"`
	Logging      LoggingConfig      `mapstructure:"logging"`
	Redis        RedisConfig        `mapstructure:"redis"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	IdleTimeout  int    `mapstructure:"idle_timeout"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	Name            string `mapstructure:"name"`
	SSLMode         string `mapstructure:"ssl_mode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	JWTSecret     string   `mapstructure:"jwt_secret"`
	JWTExpiration int      `mapstructure:"jwt_expiration"`
	APIKeys       []string `mapstructure:"api_keys"`
}

// AuditConfig represents audit configuration
type AuditConfig struct {
	Enabled       bool     `mapstructure:"enabled"`
	DefaultStatus string   `mapstructure:"default_status"`
	StatusValues  []string `mapstructure:"status_values"`
	ErrorStatus   string   `mapstructure:"error_status"`
}

// NotificationConfig represents notification configuration
type NotificationConfig struct {
	Email   EmailConfig   `mapstructure:"email"`
	Slack   SlackConfig   `mapstructure:"slack"`
	Webhook WebhookConfig `mapstructure:"webhook"`
}

// EmailConfig represents email notification configuration
type EmailConfig struct {
	Enabled  bool     `mapstructure:"enabled"`
	SMTPHost string   `mapstructure:"smtp_host"`
	SMTPPort int      `mapstructure:"smtp_port"`
	Username string   `mapstructure:"username"`
	Password string   `mapstructure:"password"`
	From     string   `mapstructure:"from"`
	To       []string `mapstructure:"to"`
}

// SlackConfig represents Slack notification configuration
type SlackConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	WebhookURL string `mapstructure:"webhook_url"`
	Channel    string `mapstructure:"channel"`
	Username   string `mapstructure:"username"`
}

// WebhookConfig represents webhook notification configuration
type WebhookConfig struct {
	Enabled bool     `mapstructure:"enabled"`
	URLs    []string `mapstructure:"urls"`
	Timeout int      `mapstructure:"timeout"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// Set defaults
	setDefaults()

	// Read environment variables
	viper.AutomaticEnv()

	// Override with environment variables
	bindEnvVars()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, use defaults and env vars
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 9025)
	viper.SetDefault("server.read_timeout", 10)
	viper.SetDefault("server.write_timeout", 10)
	viper.SetDefault("server.idle_timeout", 120)

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "password")
	viper.SetDefault("database.name", "audit_logs")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 25)
	viper.SetDefault("database.conn_max_lifetime", 300)

	viper.SetDefault("auth.jwt_expiration", 3600)

	// Audit defaults
	viper.SetDefault("audit.enabled", true)
	viper.SetDefault("audit.default_status", "pending")
	viper.SetDefault("audit.status_values", []string{"pending", "processing", "completed", "failed", "archived"})
	viper.SetDefault("audit.error_status", "failed")

	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")

	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
}

// bindEnvVars binds environment variables to configuration
func bindEnvVars() {
	// Server
	viper.BindEnv("server.host", "SERVER_HOST")
	viper.BindEnv("server.port", "SERVER_PORT")

	// Database
	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.user", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.name", "DB_NAME")
	viper.BindEnv("database.ssl_mode", "DB_SSL_MODE")

	// Auth
	viper.BindEnv("auth.jwt_secret", "JWT_SECRET")

	// Email
	viper.BindEnv("notification.email.smtp_host", "EMAIL_SMTP_HOST")
	viper.BindEnv("notification.email.smtp_port", "EMAIL_SMTP_PORT")
	viper.BindEnv("notification.email.username", "EMAIL_USERNAME")
	viper.BindEnv("notification.email.password", "EMAIL_PASSWORD")

	// Slack
	viper.BindEnv("notification.slack.webhook_url", "SLACK_WEBHOOK_URL")

	// Redis
	viper.BindEnv("redis.host", "REDIS_HOST")
	viper.BindEnv("redis.port", "REDIS_PORT")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode)
}

// GetServerAddress returns the server address
func (c *ServerConfig) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// IsValidStatus checks if the given status is valid according to configuration
func (c *AuditConfig) IsValidStatus(status string) bool {
	if !c.Enabled {
		return true // If audit is disabled, any status is valid
	}

	for _, validStatus := range c.StatusValues {
		if validStatus == status {
			return true
		}
	}
	return false
}

// GetStatusValidationString returns a comma-separated string of valid status values for validation
func (c *AuditConfig) GetStatusValidationString() string {
	if !c.Enabled {
		return ""
	}

	result := ""
	for i, status := range c.StatusValues {
		if i > 0 {
			result += " "
		}
		result += status
	}
	return result
}
