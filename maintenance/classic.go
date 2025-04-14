package maintenance

import (
	"context"
	"log/slog"
)

type ClassicMaintenanceModeProvider struct {
}

func NewClassicProvider() *ClassicMaintenanceModeProvider {
	return &ClassicMaintenanceModeProvider{}
}

// SetMaintenanceMode activates or deactivates the maintenance mode by adding/removing the key maintenance to the global-config
func (m ClassicMaintenanceModeProvider) SetMaintenanceMode(mReq maintenanceModeRequest, ctx context.Context) (*MaintenanceModeStatus, error) {
	status := &MaintenanceModeStatus{
		IsActive: mReq.Activate,
	}

	slog.Info("Maintenance-Mode activated")
	return status, nil
}

// GetMaintenanceMode Get maintenance mode by checking if key maintenance exists in global-config
func (m ClassicMaintenanceModeProvider) GetMaintenanceMode(ctx context.Context) (*MaintenanceModeStatus, error) {

	status := &MaintenanceModeStatus{
		IsActive: false,
	}

	return status, nil
}
