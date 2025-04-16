package configuration

import (
	"context"
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/cloudogu/ces-exporter/etcd"
	"log/slog"
)

var _ Provider = (*ClassicConfigurationProvider)(nil)

type ClassicConfigurationProvider struct {
}

func NewClassicConfigurationProvider() *ClassicConfigurationProvider {
	return &ClassicConfigurationProvider{}
}

func (c ClassicConfigurationProvider) getBackupSchedules(ctx context.Context) ([]core.BackupSchedule, error) {
	return []core.BackupSchedule{}, nil
}

func (c ClassicConfigurationProvider) getGlobalConfigs(_ context.Context) ([]core.KeyValue, error) {
	configs, err := etcd.GetGlobalConfig(make([]string, 0))
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return configs, nil
}

func (c ClassicConfigurationProvider) getDoguConfigs(ctx context.Context) ([]core.DoguConfig, error) {
	slog.Debug("get dogu configs...")
	return nil, nil
}
