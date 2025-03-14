package export

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetExportDogu(t *testing.T) {
	t.Run("should return get export dogu", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/export/myDogu", nil)
		req.SetPathValue("doguName", "myDogu")
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetExportDogu)

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "{\"dogu\":\"myDogu\",\"volumePath\":\"\",\"exporterPort\":0}\n", rr.Body.String())
	})

	t.Run("should return bad request when doguName is empty", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/export/myDogu", nil)
		req.SetPathValue("doguName", "")
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetExportDogu)

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
		require.Equal(t, "{\"code\":400,\"message\":\"doguName must not be empty\"}\n", rr.Body.String())
	})
}

func TestSetExportDogu(t *testing.T) {
	t.Run("should return set export dogu", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/export/myDogu", nil)
		req.SetPathValue("doguName", "otherDogu")
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(SetExportDogu)

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "{\"dogu\":\"otherDogu\",\"volumePath\":\"\",\"exporterPort\":0}\n", rr.Body.String())
	})

	t.Run("should return bad request when doguName is empty", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/export/myDogu", nil)
		req.SetPathValue("doguName", "")
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(SetExportDogu)

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
		require.Equal(t, "{\"code\":400,\"message\":\"doguName must not be empty\"}\n", rr.Body.String())
	})
}

func TestGetExportMode(t *testing.T) {
	t.Run("should return get export mode", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/export/mode", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetExportMode)

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "{\"isActive\":false}\n", rr.Body.String())
	})
}
