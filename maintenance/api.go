package maintenance

import (
	"context"
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	"log/slog"
	"net/http"
)

type Provider interface {
	ActivateMaintenanceMode(mReq maintenanceModeRequest, ctx context.Context) (*MaintenanceModeStatus, error)
	DeactivateMaintenanceMode(ctx context.Context) (*MaintenanceModeStatus, error)
	GetMaintenanceMode(ctx context.Context) (*MaintenanceModeStatus, error)
}

type MultinodeMaintenanceModeController struct {
	provider Provider
}

func NewMultinodeMaintenanceModeController(provider Provider) *MultinodeMaintenanceModeController {
	return &MultinodeMaintenanceModeController{provider: provider}
}

/*
GetMaintenanceMode
checks if the maintenance mode is currently active
*/
func (m *MultinodeMaintenanceModeController) GetMaintenanceMode(w http.ResponseWriter, r *http.Request) {
	status, err := m.provider.GetMaintenanceMode(r.Context())
	if err != nil {
		core.InternalServerErrorResponse(w, err)
		return
	}

	core.JSON(w, http.StatusOK, status)
}

/*
SetMaintenanceMode
activates or deactivates the maintenance mode bases on the given activate parameter
*/
func (m *MultinodeMaintenanceModeController) SetMaintenanceMode(w http.ResponseWriter, r *http.Request) {
	mReq, err := core.Decode[maintenanceModeRequest](r)
	if err != nil {
		core.BadRequest(w, fmt.Sprintf("error decoding maintenance-mode request: %v", err))
		return
	}

	status := &MaintenanceModeStatus{}

	if mReq.Activate {
		status, err = m.provider.ActivateMaintenanceMode(mReq, r.Context())
	} else {
		status, err = m.provider.DeactivateMaintenanceMode(r.Context())
	}

	if err != nil {
		core.InternalServerErrorResponse(w, err)
		return
	}

	core.JSON(w, http.StatusOK, status)
}
