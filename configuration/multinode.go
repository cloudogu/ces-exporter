package configuration

import (
	"context"
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/cloudogu/k8s-registry-lib/repository"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"log/slog"
)

type MultinodeConfigurationProvider struct {
	namespace  string
	configMaps v1.ConfigMapInterface
	secrets    v1.SecretInterface
	client     *core.BackupScheduleRuntimeClient
}

func NewMultinodeConfigurationProvider(namespace string, configMaps v1.ConfigMapInterface, secrets v1.SecretInterface) *MultinodeConfigurationProvider {
	return &MultinodeConfigurationProvider{
		namespace:  namespace,
		configMaps: configMaps,
		secrets:    secrets,
	}
}

func (c MultinodeConfigurationProvider) getGlobalConfigs(ctx context.Context) ([]keyValue, error) {
	globalConfigRepo := repository.NewGlobalConfigRepository(c.configMaps)

	repo, err := globalConfigRepo.Get(ctx)
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

func (c MultinodeConfigurationProvider) getDoguConfigs(ctx context.Context) ([]doguConfig, error) {
	slog.Debug("get dogu configs...")
	dogus, err := core.GetInstalledDogus(ctx, c.configMaps)
	if err != nil {
		return nil, fmt.Errorf("failed to get installed dogus: %w", err)
	}

	var doguConfigs []doguConfig
	for _, d := range dogus {
		slog.Debug(fmt.Sprintf("get configs for dogu %s", d.Name))
		doguConfigRepo := repository.NewDoguConfigRepository(c.configMaps)
		dConfig, err := doguConfigRepo.Get(ctx, d.Name)
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

		sensitiveConfigRepo := repository.NewSensitiveDoguConfigRepository(c.secrets)
		sConfig, err := sensitiveConfigRepo.Get(ctx, d.Name)
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

func (c MultinodeConfigurationProvider) getBackupSchedules(ctx context.Context) ([]backupSchedule, error) {
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
