package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/cloudogu/ces-exporter/configuration"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/cloudogu/ces-exporter/export"
	"github.com/cloudogu/ces-exporter/maintenance"
	"github.com/cloudogu/ces-exporter/systeminfo"
	componentEcoClient "github.com/cloudogu/k8s-component-lib/client"
	ecoSystemV2 "github.com/cloudogu/k8s-dogu-lib/v2/client"
	libdogu "github.com/cloudogu/k8s-registry-lib/dogu"
	"github.com/cloudogu/k8s-registry-lib/repository"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	rclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type multinodeControllerProvider struct {
	componentClient ecosystemComponentClient
	doguClient      ecosystemDogusClient
	client          kubernetesClient
	config          *core.Configuration
	bclient         *core.BackupScheduleRuntimeClient
	rtclient        rclient.Client
}

func newMultinodeControllerProvider(config core.Configuration) (*multinodeControllerProvider, error) {
	clusterConfig, err := ctrl.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get k8s cluster config: %w", err)
	}

	rtclient, err := rclient.New(clusterConfig, rclient.Options{})
	if err != nil {
		return nil, fmt.Errorf("error creating client for the multinode controller: %w", err)
	}

	bclient := core.NewBackupScheduleRuntimeClient(rtclient, config.Namespace)

	componentClient, err := componentEcoClient.NewForConfig(clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create component client: %w", err)
	}

	doguClient, err := ecoSystemV2.NewForConfig(clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create dogu client: %w", err)
	}

	cl, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client: %w", err)
	}

	return &multinodeControllerProvider{
		componentClient: componentClient,
		doguClient:      doguClient,
		client:          cl,
		config:          &config,
		bclient:         bclient,
		rtclient:        rtclient,
	}, nil
}

func (m *multinodeControllerProvider) createControllers(_ context.Context) (*systeminfo.Controller, *configuration.Controller, *maintenance.Controller, *export.Controller) {
	configMaps := m.client.CoreV1().ConfigMaps(m.config.Namespace)
	secrets := m.client.CoreV1().Secrets(m.config.Namespace)
	globalConfigRepo := repository.NewGlobalConfigRepository(configMaps)
	sensitiveRepo := repository.NewSensitiveDoguConfigRepository(secrets)
	doguRepo := repository.NewDoguConfigRepository(configMaps)
	doguVersionReg := libdogu.NewDoguVersionRegistry(configMaps)
	services := m.client.CoreV1().Services(m.config.Namespace)
	dogus := m.doguClient.Dogus(m.config.Namespace)

	systemInfoProvider := systeminfo.NewMultinodeSystemInfoProvider(
		configMaps,
		m.client.CoreV1().PersistentVolumeClaims(m.config.Namespace),
		m.config.Namespace,
		m.componentClient.Components(m.config.Namespace),
	)
	systemInfoController := systeminfo.NewController(systemInfoProvider)

	exportModeProvider := export.NewMultinodeExportModeProvider(m.config.Namespace, configMaps, dogus, services)
	exportModeController := export.NewController(exportModeProvider)

	// start cron job for setting export mode. See env variable "EXPORT_CRON" for timetable
	go m.startCronJob(dogus)

	configurationProvider := configuration.NewMultinodeConfigurationProvider(
		m.config.Namespace,
		sensitiveRepo,
		doguRepo,
		globalConfigRepo,
		doguVersionReg,
		m.bclient,
		secrets,
	)
	configController := configuration.NewController(configurationProvider)

	maintenanceModeAdapter := repository.NewMaintenanceModeAdapter("ces-exporter", m.rtclient, m.config.Namespace)
	maintenanceModeProvider := maintenance.NewMultinodeProvider(maintenanceModeAdapter)
	maintenanceModeController := maintenance.NewController(maintenanceModeProvider)

	return systemInfoController, configController, maintenanceModeController, exportModeController
}

// startCronJob starts an asynchronous task - it can not return anything but will log error if the job fails
func (m *multinodeControllerProvider) startCronJob(dogus ecoSystemV2.DoguInterface) {
	if m.config.CronExp == "" {
		return
	}
	cj := export.NewCronJob(m.config.CronExp, dogus, m.config.Namespace, m.config.VerboseCron)
	err := cj.Run()

	if err != nil {
		slog.Error(fmt.Sprintf("failed to start cronjob: %s", err.Error()))
	}
}
