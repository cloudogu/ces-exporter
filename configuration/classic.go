package configuration

import (
	"context"
	"fmt"
	"log/slog"
)

var _ Provider = (*ClassicConfigurationProvider)(nil)

type ClassicConfigurationProvider struct {
	namespace           string
	sensitiveRepo       doguConfigRepository
	doguConfigRepo      doguConfigRepository
	globalConfigRepo    globalConfigRepository
	doguVersionRegistry doguVersionRegistry
	client              backupScheduleRuntimeClient
}

func NewClassicConfigurationProvider(
	namespace string,
	sensitiveRepo doguConfigRepository,
	doguConfigRepo doguConfigRepository,
	globalConfigRepo globalConfigRepository,
	doguVersionRegistry doguVersionRegistry,
	client backupScheduleRuntimeClient,
) *ClassicConfigurationProvider {
	return &ClassicConfigurationProvider{
		namespace:           namespace,
		sensitiveRepo:       sensitiveRepo,
		doguConfigRepo:      doguConfigRepo,
		globalConfigRepo:    globalConfigRepo,
		client:              client,
		doguVersionRegistry: doguVersionRegistry,
	}
}

func (c ClassicConfigurationProvider) getGlobalConfigs(ctx context.Context) ([]keyValue, error) {
	repo, err := c.globalConfigRepo.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get global config: %w", err)
	}

	var globalConfigs []keyValue
	for k, v := range repo.GetAll() {
		globalConfigs = append(globalConfigs, keyValue{
			Key:   k.String(),
			Value: v.String(),
		})
	}

	return globalConfigs, nil
}

func (c ClassicConfigurationProvider) getDoguConfigs(ctx context.Context) ([]doguConfig, error) {
	slog.Debug("get dogu configs...")
	dogus, err := c.doguVersionRegistry.GetCurrentOfAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get installed dogus: %w", err)
	}

	var doguConfigs []doguConfig
	for _, d := range dogus {
		slog.Debug(fmt.Sprintf("get configs for dogu %s", d.Name))
		dConfig, err := c.doguConfigRepo.Get(ctx, d.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get dogu config for dogu %s: %w", d.Name, err)
		}
		dConfigKeys := []keyValue{}
		slog.Debug(fmt.Sprintf("found %d normal config keys for dogu %s", len(dConfig.GetAll()), d.Name.String()))
		for k, v := range dConfig.GetAll() {
			slog.Debug(fmt.Sprintf("found normal config key %s for dogu %s", k, d.Name.String()))
			dConfigKeys = append(dConfigKeys, keyValue{
				Key:   k.String(),
				Value: v.String(),
			})
		}

		sConfig, err := c.sensitiveRepo.Get(ctx, d.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to get sensitive dogu config for dogu %s: %w", d.Name, err)
		}
		dSecretKeys := []keyValue{}
		slog.Debug(fmt.Sprintf("found %d sensitive config keys for dogu %s", len(sConfig.GetAll()), d.Name.String()))
		for k, v := range sConfig.GetAll() {
			slog.Debug(fmt.Sprintf("found sensitive config key %s for dogu %s", k, d.Name.String()))
			dSecretKeys = append(dSecretKeys, keyValue{
				Key:   k.String(),
				Value: v.String(),
			})
		}

		doguConfigs = append(doguConfigs, doguConfig{
			Name:            d.Name.String(),
			NormalConfig:    dConfigKeys,
			SensitiveConfig: dSecretKeys,
			LocalConfig:     []keyValue{},
		})
	}

	return doguConfigs, nil
}

func (c ClassicConfigurationProvider) getBackupSchedules(ctx context.Context) ([]backupSchedule, error) {
	var schedulesResult []backupSchedule

	schedules, err := c.client.ListBackupSchedules(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list backup schedules: %w", err)
	}

	for _, schedule := range schedules.Items {
		schedulesResult = append(schedulesResult, backupSchedule{
			Name:     schedule.Name,
			Schedule: schedule.Spec.Schedule,
		})
	}

	return schedulesResult, nil
}
