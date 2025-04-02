package core

import (
	"net/http"
)

const apiKeyHeaderName = "X-CES-EXPORTER-API-KEY"

func NewAuthMiddleware(config Configuration) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get(apiKeyHeaderName)

			w.Header().Add("Access-Control-Allow-Origin", "*")

			if apiKey != config.ApiKey {
				Unauthorized(w, http.StatusText(http.StatusUnauthorized))
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}
