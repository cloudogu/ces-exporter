package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecode(t *testing.T) {
	type testStruct struct {
		Name   string `json:"name"`
		Active bool   `json:"active"`
	}

	t.Run("should decode json", func(t *testing.T) {
		body, err := json.Marshal(&testStruct{
			Name:   "TestName",
			Active: true,
		})
		require.NoError(t, err)

		req, err := http.NewRequest("POST", "/maintenance/mode", bytes.NewBuffer(body))

		decode, err := Decode[testStruct](req)

		require.NoError(t, err)
		assert.Equal(t, "TestName", decode.Name)
		assert.True(t, decode.Active)
	})

	t.Run("should return error decoding invalid json", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/maintenance/mode", strings.NewReader("invalid"))
		require.NoError(t, err)

		_, err = Decode[testStruct](req)

		require.Error(t, err)
		assert.ErrorContains(t, err, "decode json: invalid character 'i' looking for beginning of value")
	})
}

func TestJSON(t *testing.T) {
	type testStruct struct {
		Name   string `json:"name"`
		Active bool   `json:"active"`
	}

	t.Run("should write json response", func(t *testing.T) {
		rr := httptest.NewRecorder()

		JSON(rr, http.StatusOK, &testStruct{
			Name:   "TestName",
			Active: true,
		})

		require.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
		assert.Equal(t, "{\"name\":\"TestName\",\"active\":true}\n", rr.Body.String())
	})

	t.Run("should write error response on error encoding json", func(t *testing.T) {
		rr := httptest.NewRecorder()

		JSON(rr, http.StatusOK, make(chan int))

		require.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Equal(t, "text/plain; charset=utf-8", rr.Header().Get("Content-Type"))
		assert.Equal(t, "json: unsupported type: chan int\n", rr.Body.String())
	})
}

func TestBadRequest(t *testing.T) {
	t.Run("should write bad-request response", func(t *testing.T) {
		rr := httptest.NewRecorder()

		BadRequest(rr, "missing property")

		require.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
		assert.Equal(t, "{\"code\":400,\"message\":\"missing property\"}\n", rr.Body.String())
	})
}

func TestInternalServerError(t *testing.T) {
	t.Run("should write bad-request response", func(t *testing.T) {
		rr := httptest.NewRecorder()

		InternalServerErrorResponse(rr, fmt.Errorf("testerror"))

		require.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
		assert.Equal(t, "{\"code\":500,\"message\":\"testerror\"}\n", rr.Body.String())
	})
}
