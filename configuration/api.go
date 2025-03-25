package configuration

import (
	"context"
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	"net/http"
)

type Controller struct {
	systemInfoProvider ConfigurationProvider
}

type ConfigurationProvider interface {
	getGlobalConfigs(ctx context.Context) ([]keyValue, error)
	getDoguConfigs(ctx context.Context) ([]doguConfig, error)
	getBackupSchedules(ctx context.Context) ([]backupSchedule, error)
}

func NewController(provider ConfigurationProvider) *Controller {
	return &Controller{
		systemInfoProvider: provider,
	}
}

func (c Controller) GetConfig(w http.ResponseWriter, r *http.Request) {
	globalConfigs, err := c.systemInfoProvider.getGlobalConfigs(r.Context())
	if err != nil {
		core.InternalServerErrorResponse(w, fmt.Errorf("failed to get global configs: %w", err))
		return
	}

	doguConfigs, err := c.systemInfoProvider.getDoguConfigs(r.Context())
	if err != nil {
		core.InternalServerErrorResponse(w, fmt.Errorf("failed to get dogu configs: %w", err))
		return
	}

	schedulesResult, err := c.systemInfoProvider.getBackupSchedules(r.Context())
	if err != nil {
		core.InternalServerErrorResponse(w, fmt.Errorf("failed to get backup schedules: %w", err))
		return
	}

	response := &configuration{
		GlobalConfig:    globalConfigs,
		DoguConfigs:     doguConfigs,
		BackupSchedules: schedulesResult,
	}

	core.JSON(w, http.StatusOK, response)
}
