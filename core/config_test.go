package core

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadConfigFromEnv(t *testing.T) {
	t.Run("should read config from env for classic", func(t *testing.T) {
		err := os.Setenv("LOG_LEVEL", "DEBUG")
		require.NoError(t, err)
		err = os.Setenv("BASE_PATH", "/base-path")
		require.NoError(t, err)
		err = os.Setenv("API_KEY", "myApiKey")
		require.NoError(t, err)
		err = os.Setenv("NAMESPACE", "ecosystem")
		require.NoError(t, err)
		err = os.Setenv("MODE", "classic")
		require.NoError(t, err)

		conf, err := ReadConfigFromEnv()

		require.NoError(t, err)
		assert.Equal(t, "DEBUG", conf.LogLevel)
		assert.Equal(t, "/base-path", conf.BasePath)
		assert.Equal(t, "myApiKey", conf.ApiKey)
		assert.Equal(t, "ecosystem", conf.Namespace)
		assert.Equal(t, true, conf.IsClassic)
	})

	t.Run("should read config from env for multinode", func(t *testing.T) {
		err := os.Setenv("LOG_LEVEL", "DEBUG")
		require.NoError(t, err)
		err = os.Setenv("BASE_PATH", "/base-path")
		require.NoError(t, err)
		err = os.Setenv("API_KEY", "myApiKey")
		require.NoError(t, err)
		err = os.Setenv("NAMESPACE", "ecosystem")
		require.NoError(t, err)
		err = os.Unsetenv("MODE")
		require.NoError(t, err)

		conf, err := ReadConfigFromEnv()

		require.NoError(t, err)
		assert.Equal(t, "DEBUG", conf.LogLevel)
		assert.Equal(t, "/base-path", conf.BasePath)
		assert.Equal(t, "myApiKey", conf.ApiKey)
		assert.Equal(t, "ecosystem", conf.Namespace)
		assert.Equal(t, false, conf.IsClassic)
	})

	t.Run("should fail for missing api-key", func(t *testing.T) {
		err := os.Unsetenv("API_KEY")
		require.NoError(t, err)
		err = os.Unsetenv("MODE")
		require.NoError(t, err)
		_, err = ReadConfigFromEnv()

		require.Error(t, err)
		assert.ErrorContains(t, err, "environment variable API_KEY is not set")
	})

	t.Run("should fail for missing namespace", func(t *testing.T) {
		err := os.Setenv("API_KEY", "apiKey")
		require.NoError(t, err)
		err = os.Unsetenv("NAMESPACE")
		require.NoError(t, err)
		err = os.Unsetenv("MODE")
		require.NoError(t, err)
		_, err = ReadConfigFromEnv()

		require.Error(t, err)
		assert.ErrorContains(t, err, "environment variable NAMESPACE is not set")
	})

	t.Run("should read config from env and set default basePath if missing env", func(t *testing.T) {
		err := os.Setenv("API_KEY", "myApiKey")
		require.NoError(t, err)
		err = os.Setenv("NAMESPACE", "ecosystem")
		require.NoError(t, err)
		err = os.Unsetenv("BASE_PATH")
		require.NoError(t, err)
		err = os.Unsetenv("MODE")
		require.NoError(t, err)

		conf, err := ReadConfigFromEnv()

		require.NoError(t, err)
		assert.Equal(t, "myApiKey", conf.ApiKey)
		assert.Equal(t, "/ces-exporter", conf.BasePath)
	})

	t.Run("should read config from env and set default log level if missing env", func(t *testing.T) {
		err := os.Setenv("API_KEY", "myApiKey")
		require.NoError(t, err)
		err = os.Setenv("NAMESPACE", "ecosystem")
		require.NoError(t, err)
		err = os.Unsetenv("LOG_LEVEL")
		require.NoError(t, err)
		err = os.Unsetenv("MODE")
		require.NoError(t, err)

		conf, err := ReadConfigFromEnv()

		require.NoError(t, err)
		assert.Equal(t, "myApiKey", conf.ApiKey)
		assert.Equal(t, "INFO", conf.LogLevel)
	})
}
