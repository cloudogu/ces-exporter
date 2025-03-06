package core

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type ApiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func BadRequest(w http.ResponseWriter, message string) {
	ErrorResponse(w, http.StatusBadRequest, message)
}

func ErrorResponse(w http.ResponseWriter, status int, message string) {
	JSON(w, status, &ApiError{Code: status, Message: message})
}

func JSON[T any](w http.ResponseWriter, status int, v T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("error while encoding JSON-Response", "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}
