package appserver

import (
	"context"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// startTestServer serves s on a random local port until the returned cancel func is called; the result arrives on the channel.
func startTestServer(t *testing.T, s *Server) (addr string, cancel context.CancelFunc, result <-chan error) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	res := make(chan error, 1)
	go func() { res <- s.serve(ctx, ln, nil) }()

	return ln.Addr().String(), cancel, res
}

func TestShutdownWaitsForRunningRenders(t *testing.T) {
	s := NewServer(&Config{ShutdownTimeout: 5 * time.Second})
	addr, cancel, result := startTestServer(t, s)

	resp, err := http.Get("http://" + addr + "/health")
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	done, err := s.beginRender()
	require.NoError(t, err)
	cancel()

	select {
	case err := <-result:
		t.Fatalf("shutdown returned before the render finished: %v", err)
	case <-time.After(200 * time.Millisecond):
	}

	// new renders are refused during shutdown
	require.Eventually(t, func() bool {
		release, err := s.beginRender()
		if err == nil {
			release() // shutdown not yet past the request drain, don't leak the registration
		}
		return err == errShuttingDown
	}, 2*time.Second, 10*time.Millisecond)

	done()
	select {
	case err := <-result:
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("shutdown did not finish after the render completed")
	}

	// listener is closed
	_, err = http.Get("http://" + addr + "/health")
	assert.Error(t, err)
}

func TestShutdownTimeoutAbandonsRenders(t *testing.T) {
	s := NewServer(&Config{ShutdownTimeout: 100 * time.Millisecond})
	_, cancel, result := startTestServer(t, s)

	done, err := s.beginRender()
	require.NoError(t, err)
	defer done()
	cancel()

	select {
	case err := <-result:
		require.Error(t, err)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
		assert.Contains(t, err.Error(), "abandoned 1 running render(s)")
	case <-time.After(2 * time.Second):
		t.Fatal("shutdown did not respect the timeout")
	}
}

func TestCleanupStaleTempDirs(t *testing.T) {
	dir := t.TempDir()
	stale := filepath.Join(dir, "cid-app-stale")
	fresh := filepath.Join(dir, "cid-app-fresh")
	other := filepath.Join(dir, "other-stale")
	for _, d := range []string{stale, fresh, other} {
		require.NoError(t, os.Mkdir(d, 0o755))
	}
	old := time.Now().Add(-2 * time.Hour)
	require.NoError(t, os.Chtimes(stale, old, old))
	require.NoError(t, os.Chtimes(other, old, old))

	cleanupStaleTempDirs(dir, time.Hour)

	assert.NoDirExists(t, stale)
	assert.DirExists(t, fresh)
	assert.DirExists(t, other)
}
