package maintenance

import (
	"bytes"
	"encoding/json"
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

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetMaintenanceMode)

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "{\"isActive\":false}\n", rr.Body.String())
	})
}

func TestSetMaintenanceMode(t *testing.T) {
	t.Run("should set maintenance mode", func(t *testing.T) {
		body, err := json.Marshal(&maintenanceModeRequest{
			Activate: true,
			Title:    "Test Title",
			Message:  "Test Message!!!",
		})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/maintenance/mode", bytes.NewBuffer(body))
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(SetMaintenanceMode)

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "{\"isActive\":true}\n", rr.Body.String())
	})

	t.Run("should return BadRequest when setting maintenance-mode with invalid body", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/maintenance/mode", strings.NewReader("not valid"))
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(SetMaintenanceMode)

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
		require.Contains(t, rr.Body.String(), "{\"code\":400,\"message\":\"error decoding maintenance-mode request: decode json: invalid character")
	})
}
