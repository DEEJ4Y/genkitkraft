//go:build integration

package mcpdiscoveryadapter_test

import (
	"context"
	"net"
	"testing"
	"time"

	mcpdiscoveryadapter "github.com/DEEJ4Y/genkitkraft/internal/adapters/mcp_discovery"
)

// TestListToolsTimesOutAgainstUnresponsiveServer guards against a regression
// where connecting to a user-configured MCP server had no timeout: the
// underlying genkit MCP client always dials with context.Background()
// internally, so a transport-level HTTP timeout is the only thing that can
// bound this call. Runs against a listener that accepts the TCP connection
// but never responds, and expects ListTools to fail near the 30s client
// timeout (with a further 30s for the SSE/StreamableHTTP fallback attempt)
// rather than hanging indefinitely. Takes ~60s, hence gated behind the
// integration build tag alongside the other slow/networked tests.
func TestListToolsTimesOutAgainstUnresponsiveServer(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			// Accept the TCP connection but never write a response, simulating
			// an unresponsive/hung MCP server.
			_ = conn
		}
	}()

	url := "http://" + ln.Addr().String()
	a := mcpdiscoveryadapter.New()

	start := time.Now()
	_, err = a.ListTools(context.Background(), "unresponsive-server", "sse", url, nil)
	elapsed := time.Since(start)

	t.Logf("ListTools(sse, unresponsive server) returned err=%v after %s", err, elapsed)

	if err == nil {
		t.Fatalf("expected an error from an unresponsive server, got nil after %s", elapsed)
	}
	if elapsed > 75*time.Second {
		t.Fatalf("expected failure within ~60s (30s primary + 30s fallback), took %s — looks like it hung", elapsed)
	}
	if elapsed < 25*time.Second {
		t.Fatalf("failed suspiciously fast (%s) — timeout may not actually be wired up", elapsed)
	}
}
