package main

import (
	"context"
	"fmt"
	"github.com/cloudogu/ces-exporter/configuration"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/cloudogu/ces-exporter/export"
	"github.com/cloudogu/ces-exporter/maintenance"
	"github.com/cloudogu/ces-exporter/systeminfo"
	bup "github.com/cloudogu/k8s-backup-operator/pkg/api/v1"
	componentEcoClient "github.com/cloudogu/k8s-component-operator/pkg/api/ecosystem"
	ecoSystemV2 "github.com/cloudogu/k8s-dogu-operator/v3/api/ecoSystem"
	libdogu "github.com/cloudogu/k8s-registry-lib/dogu"
	"github.com/cloudogu/k8s-registry-lib/repository"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	ctrl "sigs.k8s.io/controller-runtime"
	rclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
	"time"
)

type v1AlphaClientInterface interface {
	componentEcoClient.ComponentV1Alpha1Interface
}

type kubernetesClient interface {
	kubernetes.Interface
}

type exporterContext struct {
	ecosystemClient v1AlphaClientInterface
	doguClient      *export.EcosystemDoguClient
	serviceClient   *export.EcosystemServiceClient
	client          kubernetesClient
	config          core.Configuration
	bclient         *core.BackupScheduleRuntimeClient
}

func newExporterContext(config core.Configuration) (*exporterContext, error) {
	clusterConfig, err := ctrl.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get k8s cluster config: %w", err)
	}

	rtclient, err := rclient.New(clusterConfig, rclient.Options{})
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}

	bclient := core.NewBackupScheduleRuntimeClient(rtclient, config.Namespace)

	ecosystemClient, err := componentEcoClient.NewForConfig(clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create config client: %w", err)
	}

	client, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client: %w", err)
	}

	dc, err := ecoSystemV2.NewForConfig(clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create dogu client: %w", err)
	}

	doguClient := export.NewEcosystemDoguClient(config.Namespace, dc)

	serviceClient := export.NewServiceClient(client.CoreV1().Services(config.Namespace))

	return &exporterContext{
		ecosystemClient,
		doguClient,
		serviceClient,
		client,
		config,
		bclient,
	}, nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		slog.Error("error starting ces-exporter", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	err := bup.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatalf("Failed to register BackupSchedule scheme: %v", err)
	}

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	config, err := core.ReadConfigFromEnv()
	if err != nil {
		return err
	}

	configureLogger(config)

	eCtx, err := newExporterContext(config)
	if err != nil {
		return fmt.Errorf("failed to create exporter context: %w", err)
	}

	srv := eCtx.createServer()

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: srv,
	}
	go func() {
		slog.Info(fmt.Sprintf("listening on %s", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("error start serving.", "err", err)
		}
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		slog.Info("shutting down...")
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil { //nolint:contextcheck
			slog.Error("error shutting down http server.", "err", err)
		}
	}()
	wg.Wait()
	return nil
}

func (ec exporterContext) createServer() http.Handler {
	configMaps := ec.client.CoreV1().ConfigMaps(ec.config.Namespace)
	secrets := ec.client.CoreV1().Secrets(ec.config.Namespace)
	globalConfigRepo := repository.NewGlobalConfigRepository(configMaps)
	sensitiveRepo := repository.NewSensitiveDoguConfigRepository(secrets)
	doguRepo := repository.NewDoguConfigRepository(configMaps)
	doguVersionReg := libdogu.NewDoguVersionRegistry(configMaps)

	// start cron job for setting export mode. See env variable "EXPORT_CRON" for timetable
	go ec.startCronJob()

	systemInfoProvider := systeminfo.NewMultinodeSystemInfoProvider(
		configMaps,
		ec.client.CoreV1().PersistentVolumeClaims(ec.config.Namespace),
		ec.config.Namespace,
		ec.ecosystemClient.Components(ec.config.Namespace),
	)
	systemInfoController := systeminfo.NewController(systemInfoProvider)

	exportModeProvider := export.NewMultinodeExportModeProvider(ec.config.Namespace, configMaps, ec.doguClient, ec.serviceClient)
	exportModeController := export.NewMultinodeExportModeController(exportModeProvider)

	configurationProvider := configuration.NewMultinodeConfigurationProvider(ec.config.Namespace, sensitiveRepo, doguRepo, globalConfigRepo, doguVersionReg, ec.bclient)
	configController := configuration.NewController(configurationProvider)

	authMiddleware := core.NewAuthMiddleware(ec.config)

	rootHandler := http.NewServeMux()
	rootHandler.HandleFunc("GET /health", core.Health)

	rootHandler.HandleFunc("GET /system-info", authMiddleware(systemInfoController.GetSystemInfo))

	rootHandler.HandleFunc("GET /configuration", authMiddleware(configController.GetConfig))

	rootHandler.HandleFunc("GET /export/dogu", authMiddleware(exportModeController.GetExportDogu))
	rootHandler.HandleFunc("POST /export/dogu/{doguName}", authMiddleware(exportModeController.SetExportDogu))
	rootHandler.HandleFunc("GET /export/mode", authMiddleware(exportModeController.GetExportMode))

	rootHandler.HandleFunc("GET /maintenance/mode", authMiddleware(maintenance.GetMaintenanceMode))
	rootHandler.HandleFunc("POST /maintenance/mode", authMiddleware(maintenance.SetMaintenanceMode))

	router := http.NewServeMux()
	router.Handle(fmt.Sprintf("%s/", ec.config.BasePath), http.StripPrefix(ec.config.BasePath, rootHandler))

	return router
}

func configureLogger(conf core.Configuration) {
	var level slog.Level
	var err = level.UnmarshalText([]byte(conf.LogLevel))
	if err != nil {
		slog.Error("error parsing log level. Setting log level to INFO.", "err", err)
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: false,
		Level:     level,
	}))
	slog.SetDefault(logger)

	slog.Info("configured logger", "level", level.String())
}

/*
this starts an asynchronous task - it can not return anything but will log error if the job fails
*/
func (ec exporterContext) startCronJob() {
	if ec.config.CronExp == "" {
		return
	}
	cj := export.NewCronJob(ec.config.CronExp, ec.doguClient, ec.config.Namespace, ec.config.VerboseCron)
	err := cj.Run()

	if err != nil {
		slog.Error("Failed to start cronjob:", "err", err)
	}
}
