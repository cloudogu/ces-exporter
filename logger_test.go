package main

import (
	"context"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"testing"
)

func Test_configureLogger(t *testing.T) {
	t.Run("should configure logger with log-level from config", func(t *testing.T) {
		conf := core.Configuration{LogLevel: "DEBUG"}

		configureLogger(conf)

		textHandler, ok := slog.Default().Handler().(*slog.TextHandler)
		require.True(t, ok)
		assert.True(t, textHandler.Enabled(context.TODO(), slog.LevelDebug))
	})

	t.Run("should configure logger with log-level info if config not valid", func(t *testing.T) {
		conf := core.Configuration{LogLevel: "NO_NO_LOG"}

		configureLogger(conf)

		textHandler, ok := slog.Default().Handler().(*slog.TextHandler)
		require.True(t, ok)
		assert.True(t, textHandler.Enabled(context.TODO(), slog.LevelInfo))
	})
}
