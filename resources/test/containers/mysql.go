//go:build integration

package containers

import (
	"context"
	"fmt"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/wait"
)

// StartMySQLDSN starts a MySQL 8 container and returns a DSN compatible with
// mysqldb.Open(). The container is terminated on t.Cleanup.
func StartMySQLDSN(t *testing.T) string {
	t.Helper()
	ctx := context.Background()

	c, err := tcmysql.Run(ctx, "mysql:8.0",
		tcmysql.WithDatabase("genkitkraft"),
		tcmysql.WithUsername("test"),
		tcmysql.WithPassword("test"),
	)
	if err != nil {
		t.Fatalf("start mysql container: %v", err)
	}
	t.Cleanup(func() {
		if err := c.Terminate(ctx); err != nil {
			t.Logf("terminate mysql container: %v", err)
		}
	})

	dsn, err := c.ConnectionString(ctx, "parseTime=true")
	if err != nil {
		t.Fatalf("mysql connection string: %v", err)
	}
	return dsn
}

// StartMariaDBDSN starts a MariaDB 11 container and returns a DSN compatible
// with mysqldb.Open(). The container is terminated on t.Cleanup.
func StartMariaDBDSN(t *testing.T) string {
	t.Helper()
	ctx := context.Background()

	c, err := testcontainers.Run(ctx, "mariadb:11",
		testcontainers.WithExposedPorts("3306/tcp"),
		testcontainers.WithEnv(map[string]string{
			"MARIADB_ROOT_PASSWORD": "test",
			"MARIADB_DATABASE":      "genkitkraft",
			"MARIADB_USER":          "test",
			"MARIADB_PASSWORD":      "test",
		}),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("3306/tcp")),
	)
	if err != nil {
		t.Fatalf("start mariadb container: %v", err)
	}
	t.Cleanup(func() {
		if err := c.Terminate(ctx); err != nil {
			t.Logf("terminate mariadb container: %v", err)
		}
	})

	host, err := c.Host(ctx)
	if err != nil {
		t.Fatalf("mariadb host: %v", err)
	}
	port, err := c.MappedPort(ctx, "3306/tcp")
	if err != nil {
		t.Fatalf("mariadb port: %v", err)
	}

	return fmt.Sprintf("test:test@tcp(%s:%s)/genkitkraft?parseTime=true", host, port.Port())
}
