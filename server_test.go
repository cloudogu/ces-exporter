package main

import (
	"context"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.etcd.io/etcd/client/v2"
	"k8s.io/client-go/rest"
	"net/http"
	"net/http/httptest"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"testing"
	"time"
)

func Test_createServer(t *testing.T) {
	t.Run("for multinode", func(t *testing.T) {
		conf := core.Configuration{BasePath: "/ces-exporter"}
		cl := newMockKubernetesClient(t)
		cv1 := newMockCorev1Interface(t)
		cl.EXPECT().CoreV1().Return(cv1)
		cv1.EXPECT().ConfigMaps(mock.Anything).Return(nil)
		cv1.EXPECT().PersistentVolumeClaims(mock.Anything).Return(nil)
		cv1.EXPECT().Secrets("").Return(nil)
		cv1.EXPECT().Services("").Return(nil)

		componentClient := newMockEcosystemComponentClient(t)
		componentClient.EXPECT().Components(mock.Anything).Return(nil)

		doguClient := newMockEcosystemDogusClient(t)
		doguClient.EXPECT().Dogus(mock.Anything).Return(nil)

		provider, err := newMultinodeControllerProvider(conf)
		require.NoError(t, err)

		provider.componentClient = componentClient
		provider.doguClient = doguClient
		provider.client = cl
		exCtx := server{
			config:             &conf,
			controllerProvider: provider,
		}
		router := exCtx.createEndpoints(context.Background())
		require.NotNil(t, router)

		rr := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/ces-exporter/health", nil)
		require.NoError(t, err)

		router.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "healthy", rr.Body.String())
	})

	t.Run("for classic", func(t *testing.T) {
		err := os.Setenv(fqdnEnv, "fqdn")
		require.NoError(t, err)
		conf := core.Configuration{BasePath: "/ces-exporter", IsClassic: true}
		provider, err := newClassicControllerProvider(&conf)
		require.NoError(t, err)
		exCtx := server{
			config:             &conf,
			controllerProvider: provider,
		}
		router := exCtx.createEndpoints(context.Background())
		require.NotNil(t, router)

		rr := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/ces-exporter/health", nil)
		require.NoError(t, err)

		router.ServeHTTP(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "healthy", rr.Body.String())

		err = os.Unsetenv(fqdnEnv)
		require.NoError(t, err)
	})

}

func TestNewServer(t *testing.T) {
	t.Run("will init for classic", func(t *testing.T) {
		err := os.Setenv(fqdnEnv, "fqdn")
		require.NoError(t, err)

		srv, err := newServer(core.Configuration{
			IsClassic: true,
		})
		require.NoError(t, err)
		require.NotNil(t, srv)

		err = os.Unsetenv(fqdnEnv)
		require.NoError(t, err)
	})

	t.Run("will init for multinode", func(t *testing.T) {
		// override default controller method to retrieve a kube config
		oldGetConfigDelegate := ctrl.GetConfig
		defer func() {
			ctrl.GetConfig = oldGetConfigDelegate
		}()
		ctrl.GetConfig = func() (*rest.Config, error) {
			return &rest.Config{}, nil
		}

		srv, err := newServer(core.Configuration{
			IsClassic: false,
		})
		require.NoError(t, err)
		require.NotNil(t, srv)
	})
}

func TestUpdateApiKeysAndSshKey(t *testing.T) {
	t.Run("will update apiKey", func(t *testing.T) {
		err := os.Setenv(fqdnEnv, "fqdn")
		require.NoError(t, err)
		confCtx := newMockWatchConfigurationContext(t)
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
