package maintenance

import (
	"context"
	"fmt"
	"github.com/cloudogu/k8s-registry-lib/config"
	"github.com/cloudogu/k8s-registry-lib/repository"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"log/slog"
	"strings"
)

const (
	maintenanceModeKey = "maintenance"
)

type configMaps interface {
	corev1.ConfigMapInterface
}

type MultinodeMaintenanceModeProvider struct {
	configMaps configMaps
}

func NewMultinodeMaintenanceModeProvider(maps configMaps) *MultinodeMaintenanceModeProvider {
	return &MultinodeMaintenanceModeProvider{configMaps: maps}
}

/*
ActivateMaintenanceMode
activates the maintenance mode by adding the key maintenance to the global-config
*/
func (m MultinodeMaintenanceModeProvider) ActivateMaintenanceMode(mReq maintenanceModeRequest, ctx context.Context) (*MaintenanceModeStatus, error) {
	globalConfigRepo := repository.NewGlobalConfigRepository(m.configMaps)
	globalConfig, err := globalConfigRepo.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get global config: %w", err)
	}

	_, exists := globalConfig.Get(maintenanceModeKey)
	if exists {
		// remove the key if it already exists to update title and message afterward
		globalConfig.Delete(maintenanceModeKey)
	}

	_, err = globalConfig.Set(maintenanceModeKey, BuildMaintenanceJSON(mReq))
	if err != nil {
		return nil, fmt.Errorf("could not set maintenance mode: %w", err)
	}

	_, err = globalConfigRepo.Update(ctx, globalConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to update global config: %w", err)
	}

	status := &MaintenanceModeStatus{
		IsActive: true,
	}

	slog.Info("Maintenance-Mode activated")

	return status, nil
}

/*
DeactivateMaintenanceMode
deactivates the maintenance mode by removing the key maintenance from the global-config
*/
func (m MultinodeMaintenanceModeProvider) DeactivateMaintenanceMode(ctx context.Context) (*MaintenanceModeStatus, error) {
	globalConfigRepo := repository.NewGlobalConfigRepository(m.configMaps)
	globalConfig, err := globalConfigRepo.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get global config: %w", err)
	}

	// this just returns the new configmap, which can be ignored
	globalConfig.Delete(maintenanceModeKey)

	_, err = globalConfigRepo.Update(ctx, globalConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to update global config: %w", err)
	}

	status := &MaintenanceModeStatus{
		IsActive: false,
	}

	slog.Info("Maintenance-Mode deactivated")

	return status, nil
}

/*
GetMaintenanceMode
Get maintenance mode by checking if key maintenance exists in global-config
*/
func (m MultinodeMaintenanceModeProvider) GetMaintenanceMode(ctx context.Context) (*MaintenanceModeStatus, error) {
	globalConfigRepo := repository.NewGlobalConfigRepository(m.configMaps)
	globalConfig, err := globalConfigRepo.Get(ctx)
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
results in this json format: {"title": "some title", "message": "some message"}
*/
func BuildMaintenanceJSON(mReq maintenanceModeRequest) config.Value {
	var sb strings.Builder
	sb.WriteString("{\"title\": \"")
	sb.WriteString(mReq.Title)
	sb.WriteString("\", \"message\": \"")
	sb.WriteString(mReq.Message)
	sb.WriteString("\"}")
	return config.Value(sb.String())
}
