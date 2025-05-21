package etcd

import (
	"encoding/json"
	"fmt"
	cesLibCore "github.com/cloudogu/cesapp-lib/core"
	"log/slog"
	"path"
	"strings"
)

const (
	// doguPath is the base etcd path where Dogu specifications are stored.
	doguPath = "/dogu_v2"
)

var (
	doguGetKeyValues getKeyValuesClientFuncType = getKeyValues
)

// ExcludeOption is a function type used to filter configuration fields from a Dogu spec.
type ExcludeOption func(cfg cesLibCore.ConfigurationField) bool

// ExcludeGlobalConfig is an ExcludeOption that excludes configuration fields marked as global.
func ExcludeGlobalConfig() ExcludeOption {
	return func(cfg cesLibCore.ConfigurationField) bool {
		return cfg.Global
	}
}

// getDoguConfigKeySet returns a set of configuration key names defined in the Dogu spec
// for the given Dogu. Optionally excludes certain fields using provided ExcludeOptions.
//
// Each returned key has a value of true in the map if included, or false if excluded.
func getDoguConfigKeySet(dogu string, exclude ...ExcludeOption) (map[string]bool, error) {
	doguConfigKeySet := make(map[string]bool)

	doguSpec, err := GetDoguSpec(dogu)
	if err != nil {
		return nil, fmt.Errorf("failed to get dogu spec: %w", err)
	}

	for _, config := range doguSpec.Configuration {
		doguConfigKeySet[config.Name] = true

		for _, e := range exclude {
			if e(config) {
				doguConfigKeySet[config.Name] = false
			}
		}
	}

	return doguConfigKeySet, nil
}

// GetAllDogus returns a list of all Dogu names present in the etcd registry.
//
// It scans for keys ending with `/current` under the dogu path and extracts the
// Dogu name from the path components.
func GetAllDogus() ([]string, error) {
	kvSlice, err := doguGetKeyValues(doguPath)
	if err != nil {
		return nil, fmt.Errorf("could not list installed dogus %s: %w", doguPath, err)
	}

	dogus := make([]string, 0, len(kvSlice))

	for _, kv := range kvSlice {
		if strings.HasSuffix(kv.Key, "/current") {
			keyComponents := strings.Split(kv.Key, "/")

			dogus = append(dogus, keyComponents[1])
		}
	}

	return dogus, nil
}

// GetDoguSpec retrieves the current Dogu specification for the given Dogu name.
//
// It first determines the active version of the Dogu by reading the `/current` key,
// then retrieves and unmarshals the corresponding JSON specification.
func GetDoguSpec(dogu string) (cesLibCore.Dogu, error) {
	currentDoguList, err := doguGetKeyValues(path.Join(doguPath, dogu, "current"))
	if err != nil {
		if isKeyNotFoundError(err) {
			err = ErrDoguNotFound
		}

		return cesLibCore.Dogu{}, fmt.Errorf("could not read key 'current' for dogu %s: %w", dogu, err)
	}

	doguVersion := currentDoguList[0].Value

	slog.Debug("Found current version for dogu", "dogu", dogu, "version", doguVersion)

	doguJsonSlice, err := doguGetKeyValues(path.Join(doguPath, dogu, doguVersion))
	if err != nil {
		return cesLibCore.Dogu{}, fmt.Errorf("could not read dogu json for dogu %s and version %s: %w", dogu, doguVersion, err)
	}

	doguJson := doguJsonSlice[0].Value

	var doguSpec cesLibCore.Dogu

	err = json.Unmarshal([]byte(doguJson), &doguSpec)
	if err != nil {
		return cesLibCore.Dogu{}, fmt.Errorf("could not unmarshal dogu json for %s and version %s: %w", dogu, doguVersion, err)
	}

	return doguSpec, nil
}
