//go:build integration

package containers

import (
	"context"
	"fmt"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// StartRedisURL starts a Redis container and returns its connection URL.
func StartRedisURL(t *testing.T) string {
	t.Helper()
	return startRedisCompatible(t, "redis:7-alpine")
}

// StartValkeyURL starts a Valkey container and returns its connection URL.
// Valkey speaks the Redis protocol, so the URL uses the redis:// scheme.
func StartValkeyURL(t *testing.T) string {
	t.Helper()
	return startRedisCompatible(t, "valkey/valkey:8-alpine")
}

// startRedisCompatible runs image and waits for it to accept connections.
// Redis and Valkey share an image contract (port 6379, same ready log), so one
// helper covers both.
func startRedisCompatible(t *testing.T, image string) string {
	t.Helper()
	ctx := context.Background()

	c, err := testcontainers.Run(ctx, image,
		testcontainers.WithExposedPorts("6379/tcp"),
		testcontainers.WithWaitStrategy(wait.ForLog("Ready to accept connections")),
	)
	if err != nil {
		t.Fatalf("starting %s container: %v", image, err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(c); err != nil {
			t.Logf("terminating %s container: %v", image, err)
		}
	})

	host, err := c.Host(ctx)
	if err != nil {
		t.Fatalf("getting %s host: %v", image, err)
	}
	port, err := c.MappedPort(ctx, "6379/tcp")
	if err != nil {
		t.Fatalf("getting %s port: %v", image, err)
	}

	return fmt.Sprintf("redis://%s:%s", host, port.Port())
}
