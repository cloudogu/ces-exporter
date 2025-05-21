package systeminfo

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

func TestGetSystemInfo(t *testing.T) {
	t.Run("should return system-info", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/system-info", nil)
		require.NoError(t, err)

		components := []component{
			{Name: "c1", Version: "v1"},
			{Name: "c2", Version: "v2"},
		}
		dogus := []dogu{
			{Name: "d1", Version: "v1", Volume: volume{SizeInBytes: 1}},
			{Name: "d2", Version: "v2", Volume: volume{SizeInBytes: 2}},
		}

		fqdn := "random-fqdn"

		rr := httptest.NewRecorder()
		provider := NewMockSystemInfoProvider(t)
		provider.EXPECT().getFqdn(mock.Anything).Return(fqdn, nil)
		provider.EXPECT().getComponents(mock.Anything).Return(components, nil)
		provider.EXPECT().getDogus(mock.Anything).Return(dogus, nil)
		provider.EXPECT().isMultinode().Return(true)
		controller := NewController(provider)
		handler := http.HandlerFunc(controller.GetSystemInfo)

		r := &systemInfo{
			FQDN:        fqdn,
			IsMultinode: true,
			Dogus:       dogus,
			Components:  components,
		}

		expectedResult, err := json.Marshal(r)
		require.NoError(t, err)

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, string(expectedResult), strings.Replace(rr.Body.String(), "\n", "", -1))
	})

	t.Run("fails on get fqdn", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/system-info", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		provider := NewMockSystemInfoProvider(t)
		provider.EXPECT().getFqdn(mock.Anything).Return("", fmt.Errorf("testerror"))

		controller := NewController(provider)
		handler := http.HandlerFunc(controller.GetSystemInfo)

		handler.ServeHTTP(rr, req)

		requireHttpError(t, http.StatusInternalServerError, "failed to get fqdn: testerror", rr)
	})

	t.Run("fails on get dogus", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/system-info", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		provider := NewMockSystemInfoProvider(t)
		provider.EXPECT().getFqdn(mock.Anything).Return("fqdn", nil)
		provider.EXPECT().getDogus(mock.Anything).Return(nil, fmt.Errorf("testerror"))

		controller := NewController(provider)
		handler := http.HandlerFunc(controller.GetSystemInfo)

		handler.ServeHTTP(rr, req)

		requireHttpError(t, http.StatusInternalServerError, "failed to get dogus: testerror", rr)
		provider.AssertExpectations(t)
	})

	t.Run("fails on get components", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/system-info", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		provider := NewMockSystemInfoProvider(t)
		provider.EXPECT().getFqdn(mock.Anything).Return("fqdn", nil)
		provider.EXPECT().getDogus(mock.Anything).Return(nil, nil)
		provider.EXPECT().getComponents(mock.Anything).Return(nil, fmt.Errorf("testerror"))

		controller := NewController(provider)
		handler := http.HandlerFunc(controller.GetSystemInfo)

		handler.ServeHTTP(rr, req)

		requireHttpError(t, http.StatusInternalServerError, "failed to get components: testerror", rr)
		provider.AssertExpectations(t)
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
