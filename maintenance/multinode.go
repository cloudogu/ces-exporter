package maintenance

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudogu/k8s-registry-lib/config"
	"log/slog"
)

const (
	maintenanceModeKey = "maintenance"
)

type globalConfigRepo interface {
	Get(ctx context.Context) (config.GlobalConfig, error)
	Update(ctx context.Context, globalConfig config.GlobalConfig) (config.GlobalConfig, error)
	Delete(ctx context.Context) error
}

type MultinodeMaintenanceModeProvider struct {
	globalConfigRepo globalConfigRepo
}

func NewMultinodeMaintenanceModeProvider(repo globalConfigRepo) *MultinodeMaintenanceModeProvider {
	return &MultinodeMaintenanceModeProvider{globalConfigRepo: repo}
}

/*
SetMaintenanceMode
activates or deactivates the maintenance mode by adding/removing the key maintenance to the global-config
*/
func (m MultinodeMaintenanceModeProvider) SetMaintenanceMode(mReq maintenanceModeRequest, ctx context.Context) (*MaintenanceModeStatus, error) {
	globalConfig, err := m.globalConfigRepo.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get global config: %w", err)
	}

	globalConfig.Delete(maintenanceModeKey)
	if mReq.Activate {
		jsonReq, err := BuildMaintenanceJSON(mReq)
		if err != nil {
			return nil, err
		}

		_, err = globalConfig.Set(maintenanceModeKey, jsonReq)
		if err != nil {
			return nil, fmt.Errorf("could not set maintenance mode: %w", err)
		}
	}

	_, err = m.globalConfigRepo.Update(ctx, globalConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to update global config: %w", err)
	}

	status := &MaintenanceModeStatus{
		IsActive: mReq.Activate,
	}

	slog.Info("Maintenance-Mode activated")
	return status, nil
}

/*
GetMaintenanceMode
Get maintenance mode by checking if key maintenance exists in global-config
*/
func (m MultinodeMaintenanceModeProvider) GetMaintenanceMode(ctx context.Context) (*MaintenanceModeStatus, error) {
	globalConfig, err := m.globalConfigRepo.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get global config: %w", err)
	}

	_, exists := globalConfig.Get(maintenanceModeKey)

	status := &MaintenanceModeStatus{
		IsActive: exists,
	}

	return status, nil
}

/*
BuildMaintenanceJSON
results in this json format: {"title": "some title", "text": "some text"}
*/
func BuildMaintenanceJSON(mReq maintenanceModeRequest) (config.Value, error) {
	jsonConfig, err := json.Marshal(mReq.Message)
	if err != nil {
		return "", fmt.Errorf("unable to create maintenance mode config json: %s", err)
	}

	return config.Value(jsonConfig), nil
}
