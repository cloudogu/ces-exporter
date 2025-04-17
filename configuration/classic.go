package configuration

import (
	"context"
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/cloudogu/ces-exporter/etcd"
)

var skippedKeys = map[string][]string{
	"backup": {
		"time",
		"active",
	},
	"global": {
		"certificate",
		"excluded_etcd_keys",
		"debug",
		"key_provider",
		"support_mode_etcd_exclude",
	},
	"*": {
		"public.pem",
	},
}

var _ Provider = (*ClassicProvider)(nil)

type ClassicProvider struct {
}

func NewClassicProvider() Provider {
	return &ClassicProvider{}
}

func (c ClassicProvider) getBackupSchedules(_ context.Context) ([]core.BackupSchedule, error) {
	return getBackupSchedules()
}

func (c ClassicProvider) getGlobalConfigs(_ context.Context) ([]core.KeyValue, error) {
	return etcd.GetGlobalConfig(skippedKeys["global"])
}

func (c ClassicProvider) getDoguConfigs(_ context.Context) ([]core.DoguConfig, error) {
	dogus, err := etcd.GetAllDogus()
	if err != nil {
		return nil, fmt.Errorf("failed to get all dogus: %w", err)
	}

	toSkipForAllDogus := skippedKeys["*"]

	var result []core.DoguConfig
	for _, d := range dogus {
		toSkipForDogu := append(skippedKeys[d], toSkipForAllDogus...)

		normalConfig, err := etcd.GetNormalConfig(d, toSkipForDogu)
		if err != nil {
			return nil, fmt.Errorf("failed to get normal config of dogu %s: %w", d, err)
		}

		senstivieConfig, err := etcd.GetSensitiveConfig(d, toSkipForDogu)
		if err != nil {
			return nil, fmt.Errorf("failed to get normal config of dogu %s: %w", d, err)
		}

		localConfig, err := etcd.GetLocalConfig(d, toSkipForDogu)
		if err != nil {
			return nil, fmt.Errorf("failed to get normal config of dogu %s: %w", d, err)
		}

		result = append(result, core.DoguConfig{
			Name:            d,
			NormalConfig:    normalConfig,
			LocalConfig:     localConfig,
			SensitiveConfig: senstivieConfig,
		})
	}

	return result, nil
}
