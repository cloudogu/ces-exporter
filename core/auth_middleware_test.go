package core

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewAuthMiddleware(t *testing.T) {
	t.Run("should not allow unauthorized requests without header", func(t *testing.T) {
		authMiddleware := NewAuthMiddleware(&Configuration{ApiKey: "test123"})

		handler := authMiddleware(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("OK"))
			require.NoError(t, err)
		})

		req, err := http.NewRequest("GET", "/test", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Equal(t, "{\"code\":401,\"message\":\"Unauthorized\"}\n", rr.Body.String())
	})

	t.Run("should not allow unauthorized requests with empty header", func(t *testing.T) {
		authMiddleware := NewAuthMiddleware(&Configuration{ApiKey: "test123"})

		handler := authMiddleware(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("OK"))
			require.NoError(t, err)
		})

		req, err := http.NewRequest("GET", "/test", nil)
		req.Header.Add(apiKeyHeaderName, "")
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Equal(t, "{\"code\":401,\"message\":\"Unauthorized\"}\n", rr.Body.String())
	})

	t.Run("should not allow unauthorized requests with wrong header", func(t *testing.T) {
		authMiddleware := NewAuthMiddleware(&Configuration{ApiKey: "test123"})

		handler := authMiddleware(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("OK"))
			require.NoError(t, err)
		})

		req, err := http.NewRequest("GET", "/test", nil)
		req.Header.Add(apiKeyHeaderName, "notValid")
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Equal(t, "{\"code\":401,\"message\":\"Unauthorized\"}\n", rr.Body.String())
	})

	t.Run("should allow requests with correct header", func(t *testing.T) {
		authMiddleware := NewAuthMiddleware(&Configuration{ApiKey: "test123"})

		handler := authMiddleware(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("OK"))
			require.NoError(t, err)
		})

		req, err := http.NewRequest("GET", "/test", nil)
		req.Header.Add(apiKeyHeaderName, "test123")
		require.NoError(t, err)

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "OK", rr.Body.String())
	})
}
