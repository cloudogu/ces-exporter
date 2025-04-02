package configuration

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

func TestGetConfig(t *testing.T) {
	t.Run("should return config", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/configuration", nil)
		require.NoError(t, err)

		provider := NewMockProvider(t)
		globalConfigs := []keyValue{
			{
				Key:   "a",
				Value: "b",
			},
		}
		doguConfigs := []doguConfig{
			{
				Name: "d1",
				NormalConfig: []keyValue{
					{
						Key:   "c",
						Value: "d",
					},
				},
				LocalConfig: []keyValue{
					{
						Key:   "e",
						Value: "f",
					},
				},
				SensitiveConfig: []keyValue{
					{
						Key:   "g",
						Value: "h",
					},
				},
			},
		}
		backupSchedules := []backupSchedule{
			{
				Name:     "b1",
				Schedule: "12345",
			},
		}

		provider.EXPECT().getGlobalConfigs(mock.Anything).Return(globalConfigs, nil)
		provider.EXPECT().getDoguConfigs(mock.Anything).Return(doguConfigs, nil)
		provider.EXPECT().getBackupSchedules(mock.Anything).Return(backupSchedules, nil)
		controller := NewController(provider)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(controller.GetConfig)

		handler.ServeHTTP(rr, req)
		expectedResponse := &configuration{
			GlobalConfig:    globalConfigs,
			DoguConfigs:     doguConfigs,
			BackupSchedules: backupSchedules,
		}
		expectedResponseString, err := json.Marshal(expectedResponse)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, string(expectedResponseString), strings.Replace(rr.Body.String(), "\n", "", -1))
	})

	t.Run("fail on get backup schedules", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/configuration", nil)
		require.NoError(t, err)

		provider := NewMockProvider(t)
		provider.EXPECT().getGlobalConfigs(mock.Anything).Return(nil, nil)
		provider.EXPECT().getDoguConfigs(mock.Anything).Return(nil, nil)
		provider.EXPECT().getBackupSchedules(mock.Anything).Return(nil, fmt.Errorf("testerror"))
		controller := NewController(provider)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(controller.GetConfig)

		handler.ServeHTTP(rr, req)
		requireHttpError(t, 500, "failed to get backup schedules: testerror", rr)
	})

	t.Run("fail on get dogu configs", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/configuration", nil)
		require.NoError(t, err)

		provider := NewMockProvider(t)
		provider.EXPECT().getGlobalConfigs(mock.Anything).Return(nil, nil)
		provider.EXPECT().getDoguConfigs(mock.Anything).Return(nil, fmt.Errorf("testerror"))
		controller := NewController(provider)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(controller.GetConfig)

		handler.ServeHTTP(rr, req)
		requireHttpError(t, 500, "failed to get dogu configs: testerror", rr)
	})

	t.Run("fail on get global configs", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/configuration", nil)
		require.NoError(t, err)

		provider := NewMockProvider(t)
		provider.EXPECT().getGlobalConfigs(mock.Anything).Return(nil, fmt.Errorf("testerror"))
		controller := NewController(provider)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(controller.GetConfig)

		handler.ServeHTTP(rr, req)
		requireHttpError(t, 500, "failed to get global configs: testerror", rr)
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
