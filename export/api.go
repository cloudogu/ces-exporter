package export

import (
	"github.com/cloudogu/ces-exporter/core"
	"net/http"
)

func GetExportDogu(w http.ResponseWriter, r *http.Request) {
	doguName := r.PathValue("doguName")
	if doguName == "" {
		core.BadRequest(w, "doguName must not be empty")
		return
	}

	//TODO implement me

	doguExp := &doguExport{
		Dogu: doguName,
	}

	core.JSON(w, http.StatusOK, doguExp)
}

func SetExportDogu(w http.ResponseWriter, r *http.Request) {
	doguName := r.PathValue("doguName")
	if doguName == "" {
		core.BadRequest(w, "doguName must not be empty")
		return
	}

	//TODO implement me

	doguExp := &doguExport{
		Dogu: doguName,
	}

	core.JSON(w, http.StatusOK, doguExp)
}

func GetExportMode(w http.ResponseWriter, r *http.Request) {
	status := &exportModeStatus{
		IsActive: false,
	}

	//TODO implement me

	core.JSON(w, http.StatusOK, status)
}
