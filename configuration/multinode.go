package configuration

import (
	"context"
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	"log/slog"
)

var _ Provider = (*MultinodeConfigurationProvider)(nil)

type MultinodeConfigurationProvider struct {
	namespace           string
	sensitiveRepo       doguConfigRepository
	doguConfigRepo      doguConfigRepository
	globalConfigRepo    globalConfigRepository
	doguVersionRegistry doguVersionRegistry
	client              backupScheduleRuntimeClient
}

func NewMultinodeConfigurationProvider(
	namespace string,
	sensitiveRepo doguConfigRepository,
	doguConfigRepo doguConfigRepository,
	globalConfigRepo globalConfigRepository,
	doguVersionRegistry doguVersionRegistry,
	client backupScheduleRuntimeClient,
) *MultinodeConfigurationProvider {
	return &MultinodeConfigurationProvider{
		namespace:           namespace,
		sensitiveRepo:       sensitiveRepo,
		doguConfigRepo:      doguConfigRepo,
		globalConfigRepo:    globalConfigRepo,
		client:              client,
		doguVersionRegistry: doguVersionRegistry,
	}
}

func (c MultinodeConfigurationProvider) getGlobalConfigs(ctx context.Context) ([]core.KeyValue, error) {
	repo, err := c.globalConfigRepo.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get global config: %w", err)
	}

	var globalConfigs []core.KeyValue
	for k, v := range repo.GetAll() {
		globalConfigs = append(globalConfigs, core.KeyValue{
			Key:   k.String(),
			Value: v.String(),
		})
	}

	return globalConfigs, nil
}

func (c MultinodeConfigurationProvider) getDoguConfigs(ctx context.Context) ([]core.DoguConfig, error) {
	slog.Debug("get dogu configs...")
	dogus, err := c.doguVersionRegistry.GetCurrentOfAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get installed dogus: %w", err)
	}

	var doguConfigs []core.DoguConfig
	for _, d := range dogus {
		slog.Debug(fmt.Sprintf("get configs for dogu %s", d.Name))
		dConfig, err := c.doguConfigRepo.Get(ctx, d.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get dogu config for dogu %s: %w", d.Name, err)
		}
		dConfigKeys := []core.KeyValue{}
		slog.Debug(fmt.Sprintf("found %d normal config keys for dogu %s", len(dConfig.GetAll()), d.Name.String()))
		for k, v := range dConfig.GetAll() {
			slog.Debug(fmt.Sprintf("found normal config key %s for dogu %s", k, d.Name.String()))
			dConfigKeys = append(dConfigKeys, core.KeyValue{
				Key:   k.String(),
				Value: v.String(),
			})
		}

		sConfig, err := c.sensitiveRepo.Get(ctx, d.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get sensitive dogu config for dogu %s: %w", d.Name, err)
		}
		dSecretKeys := []core.KeyValue{}
		slog.Debug(fmt.Sprintf("found %d sensitive config keys for dogu %s", len(sConfig.GetAll()), d.Name.String()))
		for k, v := range sConfig.GetAll() {
			slog.Debug(fmt.Sprintf("found sensitive config key %s for dogu %s", k, d.Name.String()))
			dSecretKeys = append(dSecretKeys, core.KeyValue{
				Key:   k.String(),
				Value: v.String(),
			})
		}

		doguConfigs = append(doguConfigs, core.DoguConfig{
			Name:            d.Name.String(),
			NormalConfig:    dConfigKeys,
			SensitiveConfig: dSecretKeys,
			LocalConfig:     []core.KeyValue{},
		})
	}

	return doguConfigs, nil
}

func (c MultinodeConfigurationProvider) getBackupSchedules(ctx context.Context) ([]core.BackupSchedule, error) {
	var schedulesResult []core.BackupSchedule

	schedules, err := c.client.ListBackupSchedules(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list backup schedules: %w", err)
	}

	for _, schedule := range schedules.Items {
		schedulesResult = append(schedulesResult, core.BackupSchedule{
			Name:     schedule.Name,
			Schedule: schedule.Spec.Schedule,
		})
	}

	return schedulesResult, nil
}
