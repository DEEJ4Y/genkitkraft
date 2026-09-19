package mysqldb

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pressly/goose/v3/lock"
)

const (
	migrationLockName    = "genkitkraft:goose:migrations"
	migrationLockTimeout = 5 * time.Minute // matches goose's Postgres locker default
)

// mysqlSessionLocker implements goose's lock.SessionLocker using MySQL/MariaDB's
// GET_LOCK()/RELEASE_LOCK() advisory lock functions, since goose ships a
// SessionLocker for Postgres but not for MySQL/MariaDB.
type mysqlSessionLocker struct {
	lockName string
	timeout  time.Duration
}

var _ lock.SessionLocker = (*mysqlSessionLocker)(nil)

func newMySQLSessionLocker() lock.SessionLocker {
	return &mysqlSessionLocker{lockName: migrationLockName, timeout: migrationLockTimeout}
}

func (l *mysqlSessionLocker) SessionLock(ctx context.Context, conn *sql.Conn) error {
	var result sql.NullInt64
	if err := conn.QueryRowContext(ctx, "SELECT GET_LOCK(?, ?)", l.lockName, int(l.timeout.Seconds())).Scan(&result); err != nil {
		return fmt.Errorf("executing GET_LOCK: %w", err)
	}
	if !result.Valid || result.Int64 != 1 {
		return fmt.Errorf("failed to acquire migration lock %q within %s", l.lockName, l.timeout)
	}
	return nil
}

func (l *mysqlSessionLocker) SessionUnlock(ctx context.Context, conn *sql.Conn) error {
	var result sql.NullInt64
	if err := conn.QueryRowContext(ctx, "SELECT RELEASE_LOCK(?)", l.lockName).Scan(&result); err != nil {
		return fmt.Errorf("executing RELEASE_LOCK: %w", err)
	}
	if !result.Valid || result.Int64 != 1 {
		return fmt.Errorf("failed to release migration lock %q", l.lockName)
	}
	return nil
}
