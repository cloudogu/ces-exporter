package configuration

import (
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
	dogus, err := core.GetInstalledDogus(r.Context(), c.configMaps)
	if err != nil {
		core.InternalServerErrorResponse(w, err)
	}
	var globalConfigs []keyValue
	var doguConfigs []doguConfig

	globalConfigRepo := repository.NewGlobalConfigRepository(c.configMaps)
	repo, err := globalConfigRepo.Get(r.Context())
	if err != nil {
		core.InternalServerErrorResponse(w, err)
		return
	}

	for k, v := range repo.GetAll() {
		globalConfigs = append(globalConfigs, keyValue{
			Key:   k.String(),
			Value: v.String(),
		})
	}

	for _, d := range dogus {
		doguConfigRepo := repository.NewDoguConfigRepository(c.configMaps)
		dConfig, err := doguConfigRepo.Get(r.Context(), d.Name)
		if err != nil {
			core.InternalServerErrorResponse(w, err)
			return
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
		sConfig, err := sensitiveConfigRepo.Get(r.Context(), d.Name)
		if err != nil {
			core.InternalServerErrorResponse(w, err)
			return
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

	var schedulesResult []backupSchedule

	schedules, err := c.client.ListBackupSchedules()
	for _, i := range schedules.Items {
		slog.Info(fmt.Sprintf("found item: %v", i))
	}

	for _, schedule := range schedules.Items {
		schedulesResult = append(schedulesResult, backupSchedule{
			Name:     schedule.Name,
			Schedule: schedule.Spec.Schedule,
		})
	}

	response := &configuration{
		GlobalConfig:    globalConfigs,
		DoguConfigs:     doguConfigs,
		BackupSchedules: schedulesResult,
	}

	core.JSON(w, http.StatusOK, response)
}
