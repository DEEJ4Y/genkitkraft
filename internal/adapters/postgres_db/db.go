package postgresdb

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Open connects to a PostgreSQL database at the given URL, runs all pending
// migrations, and returns the ready-to-use *sql.DB.
func Open(url string) (*sql.DB, error) {
	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, fmt.Errorf("opening postgres: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
	migFS, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("creating migrations sub-FS: %w", err)
	}

	provider, err := goose.NewProvider(goose.DialectPostgres, db, migFS)
	if err != nil {
		return fmt.Errorf("setting up goose provider: %w", err)
	}

	if _, err = provider.Up(context.Background()); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	return nil
}
