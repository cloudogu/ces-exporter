package configuration

import (
	"github.com/cloudogu/ces-exporter/core"
	"net/http"
)

func GetConfig(w http.ResponseWriter, r *http.Request) {
	config := &configuration{}

	//TODO implement me

	core.JSON(w, http.StatusOK, config)
}
