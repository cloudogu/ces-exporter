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
	getGlobalConfig        getGlobalConfigFunc
	getDogus               getAllDogusFunc
	getNormalConfig        getNormalConfigFunc
	getSensitiveConfig     getSensitiveConfigFunc
	getLocalConfig         getLocalConfigFunc
	backupScheduleProvider backupScheduleProvider
}

func NewClassicProvider() *ClassicProvider {
	return &ClassicProvider{
		backupScheduleProvider: newBackupScheduleProvider(etcd.GetConfig),
		getGlobalConfig:        etcd.GetGlobalConfig,
		getDogus:               etcd.GetAllDogus,
		getNormalConfig:        etcd.GetNormalConfig,
		getSensitiveConfig:     etcd.GetSensitiveConfig,
		getLocalConfig:         etcd.GetLocalConfig,
	}
}

func (c ClassicProvider) getBackupSchedules(_ context.Context) ([]core.BackupSchedule, error) {
	return c.backupScheduleProvider.getBackupSchedules()
}

func (c ClassicProvider) getGlobalConfigs(_ context.Context) ([]core.KeyValue, error) {
	return c.getGlobalConfig(skippedKeys["global"])
}

func (c ClassicProvider) getDoguConfigs(_ context.Context) ([]core.DoguConfig, error) {
	dogus, err := c.getDogus()
	if err != nil {
		return nil, fmt.Errorf("failed to get all dogus: %w", err)
	}

	toSkipForAllDogus := skippedKeys["*"]

	var result []core.DoguConfig
	for _, d := range dogus {
		toSkipForDogu := append(skippedKeys[d], toSkipForAllDogus...)

		normalConfig, err := c.getNormalConfig(d, toSkipForDogu)
		if err != nil {
			return nil, fmt.Errorf("failed to get normal config of dogu %s: %w", d, err)
		}

		senstivieConfig, err := c.getSensitiveConfig(d, toSkipForDogu)
		if err != nil {
			return nil, fmt.Errorf("failed to get sensitive config of dogu %s: %w", d, err)
		}

		localConfig, err := c.getLocalConfig(d, toSkipForDogu)
		if err != nil {
			return nil, fmt.Errorf("failed to get local config of dogu %s: %w", d, err)
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
