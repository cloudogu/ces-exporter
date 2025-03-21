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

func Unauthorized(w http.ResponseWriter, message string) {
	ErrorResponse(w, http.StatusUnauthorized, message)
}

func ErrorResponse(w http.ResponseWriter, status int, message string) {
	JSON(w, status, &ApiError{Code: status, Message: message})
}

func InternalServerErrorResponse(w http.ResponseWriter, err error) {
	ErrorResponse(w, http.StatusInternalServerError, err.Error())
}

func JSON[T any](w http.ResponseWriter, status int, v T) {
	w.Header().Set("Content-Type", "application/json")

	bytes, err := json.Marshal(v)
	if err != nil {
		slog.Error("error while encoding JSON response", "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	if _, err := fmt.Fprintln(w, string(bytes)); err != nil {
		slog.Error("error writing JSON response", "err", err)
		return
	}
}

func Decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}
