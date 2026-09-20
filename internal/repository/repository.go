package repository

import (
	"database/sql"

	// pgx registers itself as the "pgx" driver for database/sql.
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Open opens a database/sql pool backed by the pgx driver.
func Open(dsn string) (*sql.DB, error) {
	return sql.Open("pgx", dsn)
}
