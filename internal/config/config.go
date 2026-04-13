package config

import (
	"os"
	"time"

	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/zlog"
)

type Config struct {
	HTTPServer HTTPServer `mapstructure:"http_server"`
	Postgres   Postgres   `mapstructure:"postgres"`
}

type HTTPServer struct {
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type Postgres struct {
	Host     string     `mapstructure:"host"`
	Port     int        `mapstructure:"port"`
	SSLMode  string     `mapstructure:"ssl_mode"`
	Pool     PoolConfig `mapstructure:"pool"`
	User     string     `mapstructure:"user"`
	Password string     `mapstructure:"password"`
	Database string     `mapstructure:"database"`
}

type PoolConfig struct {
	MaxConns        int32         `mapstructure:"max_conns"`
	MinConns        int32         `mapstructure:"min_conns"`
	MaxConnLifetime time.Duration `mapstructure:"max_conn_lifetime"`
}

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

	return &cfg
}
