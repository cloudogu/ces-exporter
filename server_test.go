package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cloudogu/ces-exporter/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

func Test_createServer(t *testing.T) {
	t.Run("for multinode", func(t *testing.T) {
		// override default controller method to retrieve a kube config
		oldGetConfigDelegate := ctrl.GetConfig
		defer func() {
			ctrl.GetConfig = oldGetConfigDelegate
		}()
		ctrl.GetConfig = func() (*rest.Config, error) {
			return &rest.Config{}, nil
		}

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

	t.Run("will fail if fqdn is unset", func(t *testing.T) {
		err := os.Unsetenv(fqdnEnv)
		require.NoError(t, err)

		srv, err := newServer(core.Configuration{
			IsClassic: true,
		})
		assert.Error(t, err)
		assert.Nil(t, srv)
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
