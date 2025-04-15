package main

import (
	"context"
	"github.com/cloudogu/ces-exporter/core"
	"log/slog"
	"os"
)

func exitWithError(err error) {
	slog.Error("error starting ces-exporter", "err", err)
	os.Exit(1)
}

func main() {
	ctx := context.Background()

	config, err := core.ReadConfigFromEnv()
	if err != nil {
		exitWithError(err)
	}

	configureLogger(config)

	srv, err := newServer(ctx, config)
	if err != nil {
		exitWithError(err)
	}

	if err := srv.run(ctx); err != nil {
		exitWithError(err)
	}
}
