package main

import (
	"context"
	"fmt"
	"github.com/cloudogu/ces-exporter/configuration"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/cloudogu/ces-exporter/export"
	"github.com/cloudogu/ces-exporter/maintenance"
	"github.com/cloudogu/ces-exporter/systeminfo"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Printf("%s\n", err)
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

	srv := createServer(config)

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

func createServer(config core.Configuration) http.Handler {
	rootHandler := http.NewServeMux()
	rootHandler.HandleFunc("GET /health", core.Health)

	rootHandler.HandleFunc("GET /system-info", systeminfo.GetSystemInfo)

	rootHandler.HandleFunc("GET /configuration", configuration.GetConfig)

	rootHandler.HandleFunc("GET /export/{doguName}", export.GetExportDogu)
	rootHandler.HandleFunc("POST /export/{doguName}", export.SetExportDogu)
	rootHandler.HandleFunc("GET /export/mode", export.GetExportMode)

	rootHandler.HandleFunc("GET /maintenance/mode", maintenance.GetMaintenanceMode)
	rootHandler.HandleFunc("POST /maintenance/mode", maintenance.SetMaintenanceMode)

	router := http.NewServeMux()
	router.Handle(fmt.Sprintf("%s/", config.BasePath), http.StripPrefix(config.BasePath, rootHandler))

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
