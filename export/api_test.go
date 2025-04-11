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
		req, err := http.NewRequest("GET", "/export/dogu", nil)
		require.NoError(t, err)

		dogu := doguExport{
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

	t.Run("should throw 404 server error", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/export/dogu", nil)
		require.NoError(t, err)

		ep := NewMockProvider(t)
		ep.EXPECT().GetExportDogu(mock.Anything).Return(nil, fmt.Errorf("testerror"))
		mec := NewMultinodeExportModeController(ep)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(mec.GetExportDogu)

		handler.ServeHTTP(rr, req)

		requireHttpError(t, http.StatusNotFound, "failed to get export dogu: testerror", rr)
	})
}

func TestSetExportDogu(t *testing.T) {
	t.Run("should return set export dogu", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/export/dogu", nil)
		req.SetPathValue("doguName", "otherDogu")
		require.NoError(t, err)

		dogu := doguExport{
			Dogu:         "otherDogu",
			VolumePath:   "otherDogu-data",
			ExporterPort: 7022,
		}

		ep := NewMockProvider(t)
		ep.EXPECT().SetExportDogu(mock.Anything, mock.Anything).Return(&dogu, nil)
		mec := NewMultinodeExportModeController(ep)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(mec.SetExportDogu)

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "{\"dogu\":\"otherDogu\",\"volumePath\":\"otherDogu-data\",\"exporterPort\":7022}\n", rr.Body.String())

	})

	t.Run("should throw internal server error", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/export/dogu", nil)
		req.SetPathValue("doguName", "otherDogu")
		require.NoError(t, err)

		ep := NewMockProvider(t)
		ep.EXPECT().SetExportDogu(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("testerror"))
		mec := NewMultinodeExportModeController(ep)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(mec.SetExportDogu)

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
		requireHttpError(t, 500, "failed to set export dogu: testerror", rr)
	})

	t.Run("should return bad request when doguName is empty", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/export/dogu", nil)
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

func TestGetExportMode(t *testing.T) {
	t.Run("should return get export mode", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/export/mode", nil)
		require.NoError(t, err)

		exportMode := exportModeStatus{
			IsActive: true,
		}

		ep := NewMockProvider(t)
		ep.EXPECT().GetExportMode(mock.Anything).Return(&exportMode, nil)

		mec := NewMultinodeExportModeController(ep)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(mec.GetExportMode)

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "{\"isActive\":true}\n", rr.Body.String())
	})
	t.Run("should return get error on export mode", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/export/mode", nil)
		require.NoError(t, err)

		ep := NewMockProvider(t)
		ep.EXPECT().GetExportMode(mock.Anything).Return(nil, fmt.Errorf("testerror"))

		mec := NewMultinodeExportModeController(ep)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(mec.GetExportMode)

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
		require.Equal(t, "{\"code\":500,\"message\":\"testerror\"}\n", rr.Body.String())
	})
}

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
