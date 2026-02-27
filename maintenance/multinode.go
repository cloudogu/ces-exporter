package maintenance

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/cloudogu/k8s-registry-lib/repository"
)

type maintenanceAdapter interface {
	Activate(ctx context.Context, content repository.MaintenanceModeDescription, force bool) error
	Deactivate(ctx context.Context, force bool) error
	GetStatus(ctx context.Context) (repository.MaintenanceModeDescription, bool, error)
}

type MultinodeMaintenanceModeProvider struct {
	adapter maintenanceAdapter
}

func NewMultinodeProvider(adapter maintenanceAdapter) *MultinodeMaintenanceModeProvider {
	return &MultinodeMaintenanceModeProvider{adapter: adapter}
}

// SetMaintenanceMode activates or deactivates the maintenance mode by adding/removing the key maintenance to the global-config
func (m MultinodeMaintenanceModeProvider) SetMaintenanceMode(mReq maintenanceModeRequest, ctx context.Context) (*MaintenanceModeStatus, error) {
	if mReq.Activate {
		err := m.adapter.Activate(ctx, repository.MaintenanceModeDescription{
			Title: mReq.Message.Title,
			Text:  mReq.Message.Text,
		}, true)
		if err != nil {
			return nil, fmt.Errorf("activate maintenance mode: %w", err)
		}
	} else {
		err := m.adapter.Deactivate(ctx, true)
		if err != nil {
			return nil, fmt.Errorf("deactivate maintenance mode: %w", err)
		}
	}

	status := &MaintenanceModeStatus{
		IsActive: mReq.Activate,
	}

	slog.Info("Maintenance-Mode set")
	return status, nil
}

// GetMaintenanceMode Get maintenance mode by checking if key maintenance exists in global-config
func (m MultinodeMaintenanceModeProvider) GetMaintenanceMode(ctx context.Context) (*MaintenanceModeStatus, error) {
	_, active, err := m.adapter.GetStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get maintenance status: %w", err)
	}

	status := &MaintenanceModeStatus{
		IsActive: active,
	}
	return status, nil
}
