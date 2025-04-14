package systeminfo

import (
	"errors"
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
)

var (
	errFqdnGlobalConfigKeyNotFound = errors.New("global config key for fqdn not found")
)

type getGlobalConfigFunc func(ignoreKeys []string) (core.GlobalConfig, error)

type SingleNodeSystemInfoProvider struct {
	getGlobalConfig getGlobalConfigFunc
}

func NewSingleNodeSystemInfoProvider() *SingleNodeSystemInfoProvider {
	return &SingleNodeSystemInfoProvider{}
}

func (sp *SingleNodeSystemInfoProvider) getFqdn() (string, error) {
	globalCfg, err := sp.getGlobalConfig(nil)
	if err != nil {
		return "", fmt.Errorf("error getting global config: %w", err)
	}

	for _, kvPair := range globalCfg {
		if kvPair.Key == globalConfigKeyFqdn {
			return kvPair.Value, nil
		}
	}

	return "", errFqdnGlobalConfigKeyNotFound
}
