// Package config provides structures and loading logic for application configuration.
package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/zlog"
)

// Config represents the root application configuration.
type Config struct {
	HTTPServer HTTPServer `mapstructure:"http_server"`
	Postgres   Postgres   `mapstructure:"postgres"`
	RabbitMQ   RabbitMQ   `mapstructure:"rabbitmq"`
	Redis      Redis      `mapstructure:"redis"`
	Retry      Retry      `mapstructure:"retry"`
	Logging    Logging    `mapstructure:"logging"`
	SMTP       SMTP       `mapstructure:"smtp"`
	Telegram   Telegram   `mapstructure:"telegram"`
}

// HTTPServer defines configuration for the HTTP server.
type HTTPServer struct {
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// Postgres defines PostgreSQL connection settings.
type Postgres struct {
	Host     string     `mapstructure:"host"`
	Port     int        `mapstructure:"port"`
	SSLMode  string     `mapstructure:"ssl_mode"`
	Pool     PoolConfig `mapstructure:"pool"`
	User     string     `mapstructure:"user"`
	Password string     `mapstructure:"password"`
	Database string     `mapstructure:"database"`
}

// PoolConfig defines connection pool settings for database drivers.
type PoolConfig struct {
	MaxConns        int32         `mapstructure:"max_conns"`
	MinConns        int32         `mapstructure:"min_conns"`
	MaxConnLifetime time.Duration `mapstructure:"max_conn_lifetime"`
}

// RabbitMQ defines message broker connection and queue settings.
type RabbitMQ struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	User       string `mapstructure:"user"`
	Password   string `mapstructure:"password"`
	VHost      string `mapstructure:"vhost"`
	Exchange   string `mapstructure:"exchange"`
	RoutingKey string `mapstructure:"routing_key"`
	Queue      string `mapstructure:"queue"`
	DLQ        string `mapstructure:"dlq"`
}

// Redis defines caching configuration.
type Redis struct {
	Host     string        `mapstructure:"host"`
	Port     int           `mapstructure:"port"`
	Password string        `mapstructure:"password"`
	DB       int           `mapstructure:"db"`
	TTL      time.Duration `mapstructure:"ttl"`
}

// Retry defines strategy settings for retry mechanisms.
type Retry struct {
	Attempts int           `mapstructure:"attempts"`
	Delay    time.Duration `mapstructure:"delay"`
	Backoff  float64       `mapstructure:"backoff"`
	MaxDelay time.Duration `mapstructure:"max_delay"`
}

// Logging defines application logging settings.
type Logging struct {
	Level string `mapstructure:"level"`
}

// SMTP defines mail server settings.
type SMTP struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"pass"`
	From     string `mapstructure:"from"`
	UseTLS   bool   `mapstructure:"use_tls"`
}

// Telegram defines notification channel settings for Telegram.
type Telegram struct {
	Token  string `mapstructure:"token"`
	ChatID string `mapstructure:"chat_id"`
}

// Must loads the configuration from files and environment variables.
// It panics if the configuration cannot be loaded.
func Must() *Config {
	c := config.New()

	if err := c.LoadConfigFiles("./config/config.yaml"); err != nil {
		zlog.Logger.Panic().Err(err).Msg("failed to read config")
	}

	if err := c.LoadEnvFiles(".env"); err != nil {
		zlog.Logger.Warn().Err(err).Msg(".env not found")
	}

	c.EnableEnv("")

	var cfg Config
	if err := c.Unmarshal(&cfg); err != nil {
		zlog.Logger.Panic().Err(err).Msg("unmarshal failed")
	}

	if val, ok := os.LookupEnv("DB_USER"); ok {
		cfg.Postgres.User = val
	}

	if val, ok := os.LookupEnv("DB_PASSWORD"); ok {
		cfg.Postgres.Password = val
	}

	if val, ok := os.LookupEnv("SMTP_HOST"); ok {
		cfg.SMTP.Host = val
	}

	if val, ok := os.LookupEnv("SMTP_PORT"); ok {
		if port, err := strconv.Atoi(val); err == nil {
			cfg.SMTP.Port = port
		}
	}

	if val, ok := os.LookupEnv("SMTP_USER"); ok {
		cfg.SMTP.User = val
	}

	if val, ok := os.LookupEnv("SMTP_PASS"); ok {
		cfg.SMTP.Password = val
	}

	if val, ok := os.LookupEnv("SMTP_FROM"); ok {
		cfg.SMTP.From = val
	}

	if val, ok := os.LookupEnv("SMTP_USE_TLS"); ok {
		cfg.SMTP.UseTLS = strings.EqualFold(val, "true") || val == "1"
	}

	if val, ok := os.LookupEnv("TELEGRAM_TOKEN"); ok {
		cfg.Telegram.Token = val
	}

	if val, ok := os.LookupEnv("TELEGRAM_CHAT_ID"); ok {
		cfg.Telegram.ChatID = val
	}

	return &cfg
}
