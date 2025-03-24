package configuration

import (
	"context"
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/cloudogu/k8s-registry-lib/repository"
	"k8s.io/client-go/kubernetes/typed/core/v1"
	"log/slog"
	"net/http"
)

type Controller struct {
	systemInfoProvider ConfigurationProvider
	configMaps         v1.ConfigMapInterface
	secrets            v1.SecretInterface
	client             *core.BackupScheduleRuntimeClient
}

type ConfigurationProvider interface {
}

func NewController(provider ConfigurationProvider, configMaps v1.ConfigMapInterface, secrets v1.SecretInterface, client *core.BackupScheduleRuntimeClient) *Controller {
	return &Controller{
		systemInfoProvider: provider,
		configMaps:         configMaps,
		secrets:            secrets,
		client:             client,
	}
}

func (c Controller) GetConfig(w http.ResponseWriter, r *http.Request) {
	globalConfigs, err := c.getGlobalConfigs(r.Context())
	if err != nil {
		core.InternalServerErrorResponse(w, err)
		return
	}

	doguConfigs, err := c.getDoguConfigs(r.Context())
	if err != nil {
		core.InternalServerErrorResponse(w, err)
		return
	}

	schedulesResult, err := c.getBackupSchedules(r.Context())
	if err != nil {
		core.InternalServerErrorResponse(w, err)
		return
	}

	response := &configuration{
		GlobalConfig:    globalConfigs,
		DoguConfigs:     doguConfigs,
		BackupSchedules: schedulesResult,
	}

	core.JSON(w, http.StatusOK, response)
}

func (c Controller) getBackupSchedules(ctx context.Context) ([]backupSchedule, error) {
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

func (c Controller) getDoguConfigs(ctx context.Context) ([]doguConfig, error) {
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

func (c Controller) getGlobalConfigs(ctx context.Context) ([]keyValue, error) {
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
