package core

import (
	"fmt"
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

type Configuration struct {
	LogLevel                 string
	BasePath                 string
	ApiKey                   string
	Namespace                string
	IsClassic                bool
	ClassicOnlyConfiguration ClassicOnlyConfiguration
}

type ClassicOnlyConfiguration struct {
	Fqdn string
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
		conf.ApiKey = "1"
		//return conf, fmt.Errorf(errorFormat, apiKeyEnv)
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
