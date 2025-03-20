package util

import (
	"log/slog"
	"net/http"
)

func HandleUnexpectedError(w http.ResponseWriter, err error) {
	slog.Error(err.Error())
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
