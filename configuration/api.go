package configuration

import (
	"context"
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	"net/http"
)

type Controller struct {
	systemInfoProvider Provider
}

type Provider interface {
	getGlobalConfigs(ctx context.Context) ([]core.KeyValue, error)
	getDoguConfigs(ctx context.Context) ([]core.DoguConfig, error)
	getBackupSchedules(ctx context.Context) ([]core.BackupSchedule, error)
}

func NewController(provider Provider) *Controller {
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

	response := &core.ExportResponse{
		GlobalConfig:    globalConfigs,
		DoguConfigs:     doguConfigs,
		BackupSchedules: schedulesResult,
	}

	core.JSON(w, http.StatusOK, response)
}
