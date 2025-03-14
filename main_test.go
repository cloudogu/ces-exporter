package main

import (
	"context"
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"
)

func Test_configureLogger(t *testing.T) {
	t.Run("should configure logger with log-level from config", func(t *testing.T) {
		conf := core.Configuration{LogLevel: "DEBUG"}

		configureLogger(conf)

		textHandler, ok := slog.Default().Handler().(*slog.TextHandler)
		require.True(t, ok)
		assert.True(t, textHandler.Enabled(context.TODO(), slog.LevelDebug))
	})

	t.Run("should configure logger with log-level info if config not valid", func(t *testing.T) {
		conf := core.Configuration{LogLevel: "NO_NO_LOG"}

		configureLogger(conf)

		textHandler, ok := slog.Default().Handler().(*slog.TextHandler)
		require.True(t, ok)
		assert.True(t, textHandler.Enabled(context.TODO(), slog.LevelInfo))
	})
}

func Test_createServer(t *testing.T) {
	conf := core.Configuration{BasePath: "/ces-exporter"}
	router := createServer(conf)
	require.NotNil(t, router)

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/ces-exporter/health", nil)
	require.NoError(t, err)

	router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, "healthy", rr.Body.String())
}

func Test_main(t *testing.T) {

	t.Run("should start server", func(t *testing.T) {
		err := os.Setenv("API_KEY", "myApiKey")
		err = os.Setenv("NAMESPACE", "ecosystem")
		require.NoError(t, err)

		// Create a channel to receive signals
		sigChan := make(chan os.Signal, 1)
		// Register SIGINT to the signal channel
		signal.Notify(sigChan, syscall.SIGINT)

		go func() {
			assert.NotPanics(t, main)
		}()

		time.Sleep(100 * time.Millisecond)

		// assert correct start of the server
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			assert.True(t, checkPort(8080, false))
		}()

		wg.Wait()

		sendSignal(syscall.SIGINT)

		// assert graceful shutdown
		wg.Add(1)
		go func() {
			defer wg.Done()
			assert.True(t, checkPort(8080, true))
		}()

		wg.Wait()
	})
}

func checkPort(port int, available bool) bool {
	ctxWithCancel, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancel()

	resultChan := make(chan bool)
	go func() {
	loop:
		for {
			select {
			case <-ctxWithCancel.Done():
				break loop
			default:
				if result := isPortAvailable(port); result == available {
					resultChan <- result
				}
			}

		}
	}()

	select {
	case <-ctxWithCancel.Done():
		return false
	case <-resultChan:
		return true
	}
}

func isPortAvailable(port int) bool {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		// Port is not available
		return false
	}

	defer func() {
		_ = listener.Close()
	}()

	// Port is available
	return true
}

// Function to send signal to the process
func sendSignal(sig os.Signal) {
	pid := os.Getpid()
	process, err := os.FindProcess(pid)
	if err != nil {
		// Handle error
		return
	}
	process.Signal(sig)
}
