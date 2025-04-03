package maintenance

import (
	"bytes"
	"context"
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

func TestGetMaintenanceMode(t *testing.T) {
	t.Run("should return maintenance mode", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/maintenance/mode", nil)
		require.NoError(t, err)

		mms := MaintenanceModeStatus{IsActive: false}

		maintenanceModeProvider := NewMockProvider(t)
		maintenanceModeProvider.EXPECT().GetMaintenanceMode(mock.Anything).Return(&mms, nil)
		mmc := NewMultinodeMaintenanceModeController(maintenanceModeProvider)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(mmc.GetMaintenanceMode)

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "{\"isActive\":false}\n", rr.Body.String())
	})

	t.Run("should return internal server error", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/maintenance/mode", nil)
		require.NoError(t, err)

		maintenanceModeProvider := NewMockProvider(t)
		maintenanceModeProvider.EXPECT().GetMaintenanceMode(mock.Anything).Return(nil, fmt.Errorf("testerror"))
		mmc := NewMultinodeMaintenanceModeController(maintenanceModeProvider)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(mmc.GetMaintenanceMode)

		handler.ServeHTTP(rr, req)

		requireHttpError(t, 500, "testerror", rr)
	})
}

func TestSetMaintenanceMode(t *testing.T) {
	t.Run("should activate maintenance mode", func(t *testing.T) {
		mmReq := &maintenanceModeRequest{
			Activate: true,
			Message: Message{
				Title: "Test Title",
				Text:  "Test Message!!!",
			},
		}
		body, err := json.Marshal(mmReq)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/maintenance/mode", bytes.NewBuffer(body))
		require.NoError(t, err)

		mms := MaintenanceModeStatus{IsActive: true}
		maintenanceModeProvider := NewMockProvider(t)
		maintenanceModeProvider.EXPECT().SetMaintenanceMode(*mmReq, context.Background()).Return(&mms, nil)
		mmc := NewMultinodeMaintenanceModeController(maintenanceModeProvider)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(mmc.SetMaintenanceMode)

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "{\"isActive\":true}\n", rr.Body.String())
	})

	t.Run("should return internal server error", func(t *testing.T) {
		mmReq := &maintenanceModeRequest{
			Activate: false,
			Message: Message{
				Title: "Test Title",
				Text:  "Test Message!!!",
			},
		}
		body, err := json.Marshal(mmReq)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/maintenance/mode", bytes.NewBuffer(body))
		require.NoError(t, err)

		maintenanceModeProvider := NewMockProvider(t)
		maintenanceModeProvider.EXPECT().SetMaintenanceMode(*mmReq, context.Background()).Return(nil, fmt.Errorf("testerror"))
		mmc := NewMultinodeMaintenanceModeController(maintenanceModeProvider)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(mmc.SetMaintenanceMode)

		handler.ServeHTTP(rr, req)

		requireHttpError(t, 500, "testerror", rr)
	})

	t.Run("should deactivate maintenance mode", func(t *testing.T) {
		mmReq := &maintenanceModeRequest{
			Activate: false,
			Message: Message{
				Title: "Test Title",
				Text:  "Test Message!!!",
			},
		}
		body, err := json.Marshal(mmReq)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/maintenance/mode", bytes.NewBuffer(body))
		require.NoError(t, err)

		mms := MaintenanceModeStatus{IsActive: false}
		maintenanceModeProvider := NewMockProvider(t)
		maintenanceModeProvider.EXPECT().SetMaintenanceMode(*mmReq, mock.Anything).Return(&mms, nil)
		mmc := NewMultinodeMaintenanceModeController(maintenanceModeProvider)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(mmc.SetMaintenanceMode)

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "{\"isActive\":false}\n", rr.Body.String())
	})

	t.Run("should return BadRequest when setting maintenance-mode with invalid body", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/maintenance/mode", strings.NewReader("not valid"))
		require.NoError(t, err)

		maintenanceModeProvider := NewMockProvider(t)
		mmc := NewMultinodeMaintenanceModeController(maintenanceModeProvider)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(mmc.SetMaintenanceMode)

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
		require.Contains(t, rr.Body.String(), "{\"code\":400,\"message\":\"error decoding maintenance-mode request: decode json: invalid character")
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
