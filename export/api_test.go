package export

import (
	"encoding/json"
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetExportDogu(t *testing.T) {
	t.Run("should return get export dogu", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/export/myDogu", nil)
		req.SetPathValue("doguName", "myDogu")
		require.NoError(t, err)

		dogu := DoguExport{
			Dogu:         "myDogu",
			VolumePath:   "",
			ExporterPort: 0,
		}

		ep := NewMockProvider(t)
		ep.EXPECT().GetExportDogu(mock.Anything).Return(&dogu, nil)
		mec := NewMultinodeExportModeController(ep)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(mec.GetExportDogu)

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "{\"dogu\":\"myDogu\",\"volumePath\":\"\",\"exporterPort\":0}\n", rr.Body.String())
	})

	t.Run("should throw internal server error", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/export/myDogu", nil)
		req.SetPathValue("doguName", "myDogu")
		require.NoError(t, err)

		ep := NewMockProvider(t)
		ep.EXPECT().GetExportDogu(mock.Anything).Return(nil, fmt.Errorf("testerror"))
		mec := NewMultinodeExportModeController(ep)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(mec.GetExportDogu)

		handler.ServeHTTP(rr, req)

		requireHttpError(t, 500, "testerror", rr)
	})

	t.Run("should return bad request when doguName is empty", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/export/myDogu", nil)
		req.SetPathValue("doguName", "")
		require.NoError(t, err)

		ep := NewMockProvider(t)
		mec := NewMultinodeExportModeController(ep)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(mec.GetExportDogu)

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

		dogu := DoguExport{
			Dogu:         "otherDogu",
			VolumePath:   "",
			ExporterPort: 0,
		}

		ep := NewMockProvider(t)
		ep.EXPECT().GetExportDogu(mock.Anything).Return(&dogu, nil)
		mec := NewMultinodeExportModeController(ep)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(mec.SetExportDogu)

		handler.ServeHTTP(rr, req)

		requireHttpError(t, 500, "testerror", rr)
	})

	t.Run("should throw internal server error", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/export/myDogu", nil)
		req.SetPathValue("doguName", "otherDogu")
		require.NoError(t, err)

		ep := NewMockProvider(t)
		ep.EXPECT().GetExportDogu(mock.Anything).Return(nil, fmt.Errorf("testerror"))
		mec := NewMultinodeExportModeController(ep)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(mec.SetExportDogu)

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "{\"dogu\":\"otherDogu\",\"volumePath\":\"\",\"exporterPort\":0}\n", rr.Body.String())
	})

	t.Run("should return bad request when doguName is empty", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/export/myDogu", nil)
		req.SetPathValue("doguName", "")
		require.NoError(t, err)

		ep := NewMockProvider(t)
		mec := NewMultinodeExportModeController(ep)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(mec.SetExportDogu)

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
		require.Equal(t, "{\"code\":400,\"message\":\"doguName must not be empty\"}\n", rr.Body.String())
	})
}

// func TestGetExportMode(t *testing.T) {
// 	t.Run("should return get export mode", func(t *testing.T) {
// 		req, err := http.NewRequest("GET", "/export/mode", nil)
// 		require.NoError(t, err)
//
// 		rr := httptest.NewRecorder()
// 		handler := http.HandlerFunc(GetExportMode)
//
// 		handler.ServeHTTP(rr, req)
//
// 		require.Equal(t, http.StatusOK, rr.Code)
// 		require.Equal(t, "{\"isActive\":false}\n", rr.Body.String())
// 	})
// }

func requireHttpError(t *testing.T, code int, msg string, recorder *httptest.ResponseRecorder) {
	t.Helper()
	assert.Equal(t, code, recorder.Code)
	body := strings.Replace(recorder.Body.String(), "\n", "", -1)
	marshalled, err := json.Marshal(&core.ApiError{
		Code:    code,
		Message: msg,
	})
	require.NoError(t, err)

	assert.Equal(t, string(marshalled), body)
}
