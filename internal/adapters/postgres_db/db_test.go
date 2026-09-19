//go:build integration

package postgresdb_test

import (
	"database/sql"
	"os"
	"strings"
	"sync"
	"testing"

	postgresdb "github.com/DEEJ4Y/genkitkraft/internal/adapters/postgres_db"
	"github.com/DEEJ4Y/genkitkraft/resources/test/containers"
)

var expectedTables = []string{
	"providers",
	"prompts",
	"agents",
	"playground_sessions",
	"playground_messages",
	"http_tools",
	"mcp_servers",
	"agent_http_tools",
	"agent_mcp_servers",
	"agent_mcp_server_tools",
	"agent_builtin_tools",
}

func TestOpen(t *testing.T) {
	url := containers.StartPostgresDSN(t)

	db, err := postgresdb.Open(url)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}

	assertTablesExist(t, db)
	assertMigrationCount(t, db)
}

// TestOpenConcurrent guards against #37: multiple instances starting
// simultaneously against a database with pending migrations must all
// succeed, with no duplicate goose_db_version rows.
func TestOpenConcurrent(t *testing.T) {
	url := containers.StartPostgresDSN(t)

	const n = 20
	var wg sync.WaitGroup
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			db, err := postgresdb.Open(url)
			errs[i] = err
			if db != nil {
				db.Close()
			}
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("instance %d: Open: %v", i, err)
		}
	}

	db, err := postgresdb.Open(url)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	var total, distinct int
	if err := db.QueryRow("SELECT COUNT(*), COUNT(DISTINCT version_id) FROM goose_db_version").Scan(&total, &distinct); err != nil {
		t.Fatalf("querying goose_db_version: %v", err)
	}
	if total != distinct {
		t.Errorf("goose_db_version has duplicate versions: %d rows, %d distinct", total, distinct)
	}
}

func assertTablesExist(t *testing.T, db *sql.DB) {
	t.Helper()
	for _, table := range expectedTables {
		t.Run("table/"+table, func(t *testing.T) {
			var n int
			err := db.QueryRow(
				"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1",
				table,
			).Scan(&n)
			if err != nil {
				t.Fatalf("checking table %q: %v", table, err)
			}
			if n == 0 {
				t.Errorf("table %q does not exist", table)
			}
		})
	}
}

func countMigrationFiles(t *testing.T) int {
	t.Helper()
	entries, err := os.ReadDir("migrations")
	if err != nil {
		t.Fatalf("reading migrations dir: %v", err)
	}
	var n int
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			n++
		}
	}
	return n
}

func assertMigrationCount(t *testing.T, db *sql.DB) {
	t.Helper()
	t.Run("migration_count", func(t *testing.T) {
		want := countMigrationFiles(t)
		var got int
		if err := db.QueryRow("SELECT COUNT(*) FROM goose_db_version WHERE is_applied = true AND version_id > 0").Scan(&got); err != nil {
			t.Fatalf("querying goose_db_version: %v", err)
		}
		if got != want {
			t.Errorf("expected %d applied migrations, got %d", want, got)
		}
	})
}
