package systeminfo

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetSystemInfo(t *testing.T) {
	t.Run("should return system-info", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/system-info", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetSystemInfo)

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "{\"fqdn\":\"\",\"isMultinode\":false,\"dogus\":null,\"components\":null}\n", rr.Body.String())
	})
}
