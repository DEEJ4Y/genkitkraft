//go:build integration

package postgresdb_test

import (
	"database/sql"
	"os"
	"strings"
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
