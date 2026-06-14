package mysqldb

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Open connects to a MySQL or MariaDB database at the given DSN, runs all
// pending migrations, and returns the ready-to-use *sql.DB.
// DSN format: user:password@tcp(host:port)/database?parseTime=true
func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening mysql: %w", err)
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

	provider, err := goose.NewProvider(goose.DialectMySQL, db, migFS)
	if err != nil {
		return fmt.Errorf("setting up goose provider: %w", err)
	}

	if _, err = provider.Up(context.Background()); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	return nil
}
