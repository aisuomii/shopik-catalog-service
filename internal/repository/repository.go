package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/aisuomii/shopik-catalog-service/internal/config"
)

// Open builds a database/sql pool on top of the pgx driver explicitly, without
// relying on the driver-name registry, applies the pool limits from the config
// and verifies the connection.
func Open(ctx context.Context, cfg config.Postgres) (*sql.DB, error) {
	pgxCfg, err := pgx.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse postgres dsn: %w", err)
	}

	db := stdlib.OpenDB(*pgxCfg)
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	if err := db.PingContext(ctx); err != nil {
		db.Close()

		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return db, nil
}
