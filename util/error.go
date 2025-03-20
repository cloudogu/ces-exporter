package util

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func HandleUnexpectedError(w http.ResponseWriter, err error) {
	HandleErrorWithCode(w, err, http.StatusInternalServerError)
}

func HandleErrorWithCode(w http.ResponseWriter, err error, code int) {
	slog.Error(err.Error())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	errorResponse := ErrorResponse{
		Code:    code,
		Message: err.Error(),
	}

	err = json.NewEncoder(w).Encode(errorResponse)
	if err != nil {
		slog.Error(err.Error())
	}
}
