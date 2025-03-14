package systeminfo

import (
	"github.com/cloudogu/ces-exporter/core"
	"net/http"
)

func GetSystemInfo(w http.ResponseWriter, r *http.Request) {
	info := &systemInfo{}

	//TODO implement me

	core.JSON(w, http.StatusOK, info)
}
