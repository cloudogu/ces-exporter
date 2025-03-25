package main

import (
	"context"
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	ctrl "sigs.k8s.io/controller-runtime"
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
	client := newMockKubernetesClient(t)
	ecosystemClient := newMockV1AlphaClientInterface(t)
	cv1 := newMockCorev1Interface(t)
	client.EXPECT().CoreV1().Return(cv1)
	cv1.EXPECT().ConfigMaps(mock.Anything).Return(nil)
	cv1.EXPECT().PersistentVolumeClaims(mock.Anything).Return(nil)
	cv1.EXPECT().Secrets("").Return(nil)
	ecosystemClient.EXPECT().Components(mock.Anything).Return(nil)
	exCtx := exporterContext{
		ecosystemClient: ecosystemClient,
		client:          client,
		config:          conf,
	}
	router := exCtx.createServer()
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
		// override default controller method to retrieve a kube config
		oldGetConfigDelegate := ctrl.GetConfig
		defer func() {
			ctrl.GetConfig = oldGetConfigDelegate
		}()
		ctrl.GetConfig = func() (*rest.Config, error) {
			return &rest.Config{}, nil
		}

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
