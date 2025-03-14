package configuration

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetConfig(t *testing.T) {
	t.Run("should return config", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/configuration", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetConfig)

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "{\"global\":null,\"dogus\":null,\"backupSchedules\":null}\n", rr.Body.String())
	})
}
