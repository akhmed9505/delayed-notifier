package postgres

import (
	"errors"
	"fmt"

	"github.com/akhmed9505/delayed-notifier/internal/config"
	"github.com/wb-go/wbf/dbpg"
)

func New(cfg *config.Postgres) (*dbpg.DB, error) {
	if cfg == nil {
		return nil, errors.New("postgres: nil config")
	}
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
	)
	opts := &dbpg.Options{
		MaxOpenConns:    int(cfg.Pool.MaxConns),
		MaxIdleConns:    int(cfg.Pool.MinConns),
		ConnMaxLifetime: cfg.Pool.MaxConnLifetime,
	}
	return dbpg.New(dsn, nil, opts)
}
