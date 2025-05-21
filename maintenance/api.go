package maintenance

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/cloudogu/k8s-registry-lib/config"
	"net/http"
)

type Provider interface {
	SetMaintenanceMode(mReq maintenanceModeRequest, ctx context.Context) (*MaintenanceModeStatus, error)
	GetMaintenanceMode(ctx context.Context) (*MaintenanceModeStatus, error)
}

type Controller struct {
	provider Provider
}

func NewController(provider Provider) *Controller {
	return &Controller{provider: provider}
}

// GetMaintenanceMode checks if the maintenance mode is currently active
func (m *Controller) GetMaintenanceMode(w http.ResponseWriter, r *http.Request) {
	status, err := m.provider.GetMaintenanceMode(r.Context())
	if err != nil {
		core.InternalServerErrorResponse(w, err)
		return
	}

	core.JSON(w, http.StatusOK, status)
}

// SetMaintenanceMode activates or deactivates the maintenance mode bases on the given activate parameter
func (m *Controller) SetMaintenanceMode(w http.ResponseWriter, r *http.Request) {
	mReq, err := core.Decode[maintenanceModeRequest](r)
	if err != nil {
		core.BadRequest(w, fmt.Sprintf("error decoding maintenance-mode request: %v", err))
		return
	}

	status, err := m.provider.SetMaintenanceMode(mReq, r.Context())

	if err != nil {
		core.InternalServerErrorResponse(w, err)
		return
	}

	core.JSON(w, http.StatusOK, status)
}

// BuildMaintenanceJSON results in this json format: {"title": "some title", "text": "some text"}
func BuildMaintenanceJSON(mReq maintenanceModeRequest) (config.Value, error) {
	jsonConfig, err := json.Marshal(mReq.Message)
	if err != nil {
		return "", fmt.Errorf("unable to create maintenance mode config json: %s", err)
	}

	return config.Value(jsonConfig), nil
}
