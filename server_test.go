package main

import (
	"github.com/cloudogu/ces-exporter/core"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
	exCtx := server{
		ecosystemClient: ecosystemClient,
		client:          client,
		config:          &conf,
	}
	router := exCtx.createEndpoints()
	require.NotNil(t, router)

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/ces-exporter/health", nil)
	require.NoError(t, err)

	router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, "healthy", rr.Body.String())
}
