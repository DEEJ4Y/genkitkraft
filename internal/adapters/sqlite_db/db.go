package sqlitedb

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Open creates a SQLite database connection at the given path,
// runs all pending migrations, and returns the ready-to-use *sql.DB.
func Open(dbPath string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("creating database directory: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_foreign_keys=on", dbPath)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite: %w", err)
	}

	// SQLite is single-writer; limit to one connection.
	db.SetMaxOpenConns(1)

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

	provider, err := goose.NewProvider(goose.DialectSQLite3, db, migFS)
	if err != nil {
		return fmt.Errorf("setting up goose provider: %w", err)
	}

	if _, err = provider.Up(context.Background()); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	return nil
}
