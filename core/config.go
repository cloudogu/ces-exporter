package core

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/client/v2"
	"os"
	"strings"
)

const (
	logLevelEnv  = "LOG_LEVEL"
	basePathEnv  = "BASE_PATH"
	apiKeyEnv    = "API_KEY"
	namespaceEnv = "NAMESPACE"
	fqdnEnv      = "FQDN"
	errorFormat  = "environment variable %s is not set"
	modeClassic  = "classic"
)

type watchConfigurationContext interface {
	Watch(ctx context.Context, key string, recursive bool, eventChannel chan *client.Response)
	Get(key string) (string, error)
}

type Configuration struct {
	LogLevel                 string
	BasePath                 string
	ApiKey                   string
	Namespace                string
	IsClassic                bool
	ClassicOnlyConfiguration ClassicOnlyConfiguration
}

type WriteFileFunc func(name string, data []byte, perm os.FileMode) error

type ClassicOnlyConfiguration struct {
	Fqdn      string
	Registry  watchConfigurationContext
	WriteFile WriteFileFunc
}

func ReadConfigFromEnv() (Configuration, error) {
	conf := Configuration{}

	mode := os.Getenv("MODE")
	conf.IsClassic = mode == modeClassic

	conf.LogLevel = os.Getenv(logLevelEnv)
	if conf.LogLevel == "" {
		conf.LogLevel = "INFO"
	}

	conf.BasePath = os.Getenv(basePathEnv)
	if conf.BasePath == "" {
		conf.BasePath = "/ces-exporter"
	}
	conf.BasePath = strings.TrimSuffix(conf.BasePath, "/")

	conf.ApiKey = os.Getenv(apiKeyEnv)
	if conf.ApiKey == "" {
		return conf, fmt.Errorf(errorFormat, apiKeyEnv)
	}

	conf.Namespace = os.Getenv(namespaceEnv)
	if conf.Namespace == "" && !conf.IsClassic {
		return conf, fmt.Errorf(errorFormat, namespaceEnv)
	}

	if conf.IsClassic {
		conf.ClassicOnlyConfiguration = ClassicOnlyConfiguration{
			Fqdn: os.Getenv(fqdnEnv),
		}
	}

	return conf, nil
}
