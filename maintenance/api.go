package maintenance

import (
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	"net/http"
)

func GetMaintenanceMode(w http.ResponseWriter, r *http.Request) {
	status := &maintenanceModeStatus{
		IsActive: false,
	}

	//TODO implement me

	core.JSON(w, http.StatusOK, status)
}

func SetMaintenanceMode(w http.ResponseWriter, r *http.Request) {
	mReq, err := core.Decode[maintenanceModeRequest](r)
	if err != nil {
		core.BadRequest(w, fmt.Sprintf("error decoding maintenance-mode request: %v", err))
		return
	}

	status := &maintenanceModeStatus{
		IsActive: mReq.Activate,
	}

	//TODO implement me

	core.JSON(w, http.StatusOK, status)
}
