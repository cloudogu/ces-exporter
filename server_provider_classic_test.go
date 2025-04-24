package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.etcd.io/etcd/client/v2"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestWriteAuthorizedKey(t *testing.T) {
	t.Run("print slog error if write function fails", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, nil))
		slog.SetDefault(logger)

		writeAuthorizedKey("", func(name string, data []byte, perm os.FileMode) error {
			return fmt.Errorf("testerror")
		})
		assert.Contains(t, buf.String(), "level=ERROR msg=\"Could not write changed ssh key to file: testerror")
	})

	t.Run("print slog info if write function succeeds", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, nil))
		slog.SetDefault(logger)

		writeAuthorizedKey("", func(name string, data []byte, perm os.FileMode) error {
			return nil
		})
		assert.Contains(t, buf.String(), "level=INFO msg=\"Successfully wrote new ssh key to authorized_keys file...")
	})
}

func TestUpdateApiKeysAndSshKey(t *testing.T) {
	t.Run("will update apiKey", func(t *testing.T) {
		err := os.Setenv(fqdnEnv, "fqdn")
		require.NoError(t, err)
		confCtx := newMockWatchConfigurationContext(t)
		confCtx.EXPECT().Get(regKeyVolumeIncreaseFactor).Return("0.3", nil)
		confCtx.EXPECT().Get(regKeyApi).Return("oldval", nil).Once()
		confCtx.EXPECT().Get(regKeySsh).Return("oldssh", nil).Once()
		confCtx.EXPECT().
			Watch(mock.Anything, regKeyApi, mock.Anything, mock.Anything).
			Run(
				func(
					ctx context.Context,
					key string,
					recursive bool,
					eventChannel chan *client.Response) {
					time.Sleep(500 * time.Millisecond)
					response := &client.Response{
						Node: &client.Node{
							Value: "newval",
						},
					}
					eventChannel <- response
				},
			).Once()
		confCtx.EXPECT().
			Watch(mock.Anything, regKeySsh, mock.Anything, mock.Anything).
			Run(
				func(
					ctx context.Context,
					key string,
					recursive bool,
					eventChannel chan *client.Response) {
					time.Sleep(500 * time.Millisecond)
					response := &client.Response{
						Node: &client.Node{
							Value: "newssh",
						},
					}
					eventChannel <- response
				},
			).Once()
		conf := core.Configuration{
			ApiKey:    "original",
			IsClassic: true,
		}
		srv, err := newServer(conf)
		require.NoError(t, err)
		provider, err := newClassicControllerProvider(srv.config)
		require.NoError(t, err)

		provider.reg = confCtx
		counter := 0
		provider.write = func(name string, data []byte, perm os.FileMode) error {
			assert.Equal(t, "/root/.ssh/authorized_keys", name)
			if counter == 0 {
				assert.Equal(t, "oldssh", string(data))
				counter++
			} else {
				assert.Equal(t, "newssh", string(data))
			}
			return nil
		}

		srv.controllerProvider = provider

		_, _, _, _ = provider.createControllers(context.Background())

		assert.Equal(t, "original", srv.config.ApiKey)
		time.Sleep(100 * time.Millisecond)
		assert.Equal(t, "oldval", srv.config.ApiKey)
		time.Sleep(500 * time.Millisecond)
		assert.Equal(t, "newval", srv.config.ApiKey)
	})
}

func Test_getVolumeIncreaseFactor(t *testing.T) {
	regMock := newMockWatchConfigurationContext(t)

	var mockedReturnValue string
	var mockedError error

	regMock.EXPECT().Get(regKeyVolumeIncreaseFactor).RunAndReturn(func(_ string) (string, error) {
		return mockedReturnValue, mockedError
	})

	tests := []struct {
		name                 string
		volumeIncreaseFactor string
		mErr                 error
		expected             float32
	}{
		{name: "30%", volumeIncreaseFactor: "0.3", expected: 0.3},
		{name: "110%", volumeIncreaseFactor: "1.1", expected: 1.1},
		{name: "50,55%", volumeIncreaseFactor: "0.555", expected: 0.555},
		{name: "invalid factor", volumeIncreaseFactor: "invalid", expected: defaultVolumeIncreaseFactor},
		{name: "registry error", mErr: assert.AnError, expected: defaultVolumeIncreaseFactor},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockedReturnValue = tt.volumeIncreaseFactor
			mockedError = tt.mErr

			actual := getVolumeIncreaseFactor(regMock)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
