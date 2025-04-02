package main

import (
	"context"
	"fmt"
	"github.com/cloudogu/ces-exporter/configuration"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/cloudogu/ces-exporter/export"
	"github.com/cloudogu/ces-exporter/maintenance"
	"github.com/cloudogu/ces-exporter/systeminfo"
	componentEcoClient "github.com/cloudogu/k8s-component-operator/pkg/api/ecosystem"
	ecoSystemV2 "github.com/cloudogu/k8s-dogu-operator/v3/api/ecoSystem"
	librepo "github.com/cloudogu/k8s-registry-lib/repository"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	ctrl "sigs.k8s.io/controller-runtime"
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
	doguClient      ecoSystemV2.EcoSystemV2Interface
	client          kubernetesClient
	config          core.Configuration
}

func newExporterContext(config core.Configuration) (*exporterContext, error) {
	clusterConfig, err := ctrl.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get k8s cluster config: %w", err)
	}

	ecosystemClient, err := componentEcoClient.NewForConfig(clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create config client: %w", err)
	}

	client, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client: %w", err)
	}

	doguClient, err := ecoSystemV2.NewForConfig(clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create dogu client: %w", err)
	}

	return &exporterContext{
		ecosystemClient,
		doguClient,
		client,
		config,
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

	dogus, _ := ec.doguClient.Dogus(ec.config.Namespace).List(context.Background(), metav1.ListOptions{})
	slog.Info("dogus list-:")
	for _, d := range dogus.Items {
		slog.Info(d.Name)
		slog.Info(fmt.Sprintf("%v", d.Spec.ExportMode))
	}
	slog.Info(fmt.Sprintf("%d", len(dogus.Items)))

	//go ec.startCronJob()

	systemInfoProvider := systeminfo.NewMultinodeSystemInfoProvider(
		ec.client.CoreV1().ConfigMaps(ec.config.Namespace),
		ec.client.CoreV1().PersistentVolumeClaims(ec.config.Namespace),
		ec.config.Namespace,
		ec.ecosystemClient.Components(ec.config.Namespace),
	)
	systemInfoController := systeminfo.NewController(systemInfoProvider)

	exportModeProvider := export.NewMultinodeExportModeProvider(ec.client, ec.config.Namespace)
	exportModeController := export.NewMultinodeExportModeController(exportModeProvider)

	authMiddleware := core.NewAuthMiddleware(ec.config)

	rootHandler := http.NewServeMux()
	rootHandler.HandleFunc("GET /health", core.Health)

	rootHandler.HandleFunc("GET /system-info", authMiddleware(systemInfoController.GetSystemInfo))

	rootHandler.HandleFunc("GET /configuration", authMiddleware(configuration.GetConfig))

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

// this starts a asynchronous task - it can not return anything but will log error if the job fails
func (ec exporterContext) startCronJob() {
	cron_expr := os.Getenv(export.ExportCronJobEnv)
	if cron_expr == "" {
		// step out if no expression is configured
		return
	}

	cj := export.NewCronJob(cron_expr)
	err := cj.Run(ec.callCronJob)

	if err != nil {
		slog.Error("Failed to start cronjob:", "err", err)
	}
}

// this handles the actuall exporter
func (ec exporterContext) callCronJob() (int, error) {
	slog.Info("Hallo Welt")
	configMaps := ec.client.CoreV1().ConfigMaps(ec.config.Namespace)
	slog.Info("Hallo Welt2")
	doguRepo := librepo.NewDoguConfigRepository(configMaps)
	slog.Info("Hallo Welt3")
	conf, err := doguRepo.Get(context.Background(), "usermgt")
	slog.Info("Hallo Wel4")
	if err != nil {
		return 1, fmt.Errorf("Error getting dogu config for", "err", err)
	}
	slog.Info("Hallo Welt5")
	for _, e := range conf.GetAll() {
		slog.Info("Hallo Welt6")
		slog.Info(fmt.Sprintf("%v", e))
	}
	slog.Info("Hallo Welt7")
	return 0, nil
}
