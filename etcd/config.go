package etcd

import (
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/cloudogu/ces-exporter/decrypt"
	"log/slog"
	"path"
	"regexp"
	"strings"
)

const (
	// globalConfigPath is the path where the global configuration is stored.
	globalConfigPath = "/config/_global"
	// doguConfigPath is the base path for all Dogu-specific configurations.
	doguConfigPath = "/config"
)

var (
	// createDecryptFunc references the decrypter constructor used throughout this module.
	createDecryptFunc = decrypt.CreateDecrypter
	// configGetKeyValues references the getKeyValues from client.go used throughout this module.
	configGetKeyValues getKeyValuesClientFuncType = getKeyValues
)

// decrypter defines an interface for decrypting encrypted strings.
type decrypter interface {
	Decrypt(input string) (string, error)
}

// filterKeys creates a FilterOption that includes or excludes key-value pairs
// based on a predefined key set and a list of exceptions.
func filterKeys(filterSet map[string]bool, exclude bool, exceptions []string) core.FilterOption {
	return func(kvs []core.KeyValue) []core.KeyValue {
		filtered := make([]core.KeyValue, 0, len(kvs))

		for _, kv := range kvs {
			key := strings.TrimPrefix(kv.Key, "/")
			if exists := filterSet[key]; exists == exclude {
				keep := false
				for _, exception := range exceptions {
					if strings.Contains(key, exception) {
						keep = true
						break
					}
				}

				if !keep {
					continue
				}
			}

			filtered = append(filtered, kv)
		}

		return filtered
	}
}

// ignoreSubKeys returns a FilterOption that removes key-value pairs
// whose keys matches the specified regexes.
func ignoreSubKeys(regexes []string) core.FilterOption {
	regexList := make([]*regexp.Regexp, 0, len(regexes))
	for _, regexString := range regexes {
		regex, err := regexp.Compile(regexString)
		if err != nil {
			slog.Warn("failed to compile regex, regex will be ignored", "regex", regexString, "error", err)
			continue
		}

		regexList = append(regexList, regex)
	}

	return func(kvs []core.KeyValue) []core.KeyValue {
		filtered := make([]core.KeyValue, 0, len(kvs))
		for _, kv := range kvs {
			skip := false

			for _, regex := range regexList {
				if regex.MatchString(kv.Key) {
					skip = true
					break
				}
			}

			if skip {
				continue
			}

			filtered = append(filtered, kv)
		}

		return filtered
	}
}

// filterEncryptedKeys returns a FilterOption that includes or excludes encrypted values,
// based on the `exclude` flag. When exclude is true, encrypted values are removed;
// otherwise, they are decrypted and included.
func filterEncryptedKeys(d decrypter, exclude bool) core.FilterOption {
	return func(kvs []core.KeyValue) []core.KeyValue {
		filtered := make([]core.KeyValue, 0, len(kvs))

		for _, kv := range kvs {
			dValue, err := d.Decrypt(kv.Value)

			if (err == nil && exclude) || (err != nil && !exclude) {
				continue
			}

			if !exclude {
				kv = core.KeyValue{
					Key:   kv.Key,
					Value: dValue,
				}
			}

			filtered = append(filtered, kv)
		}

		return filtered
	}
}

// decryptEncryptedKeys returns a FilterOption that decrypts all decryptable key-values.
// If decryption fails, the original value is preserved.
func decryptEncryptedKeys(d decrypter) core.FilterOption {
	return func(kvs []core.KeyValue) []core.KeyValue {
		filtered := make([]core.KeyValue, 0, len(kvs))

		for _, kv := range kvs {
			dValue, err := d.Decrypt(kv.Value)
			if err == nil {
				filtered = append(filtered, core.KeyValue{
					Key:   kv.Key,
					Value: dValue,
				})
			} else {
				filtered = append(filtered, kv)
			}
		}

		return filtered
	}
}

// GetGlobalConfig returns the global configuration, optionally excluding specified keys.
func GetGlobalConfig(ignoreKeys []string) (core.GlobalConfig, error) {
	return configGetKeyValues(
		globalConfigPath,
		ignoreSubKeys(ignoreKeys),
	)
}

// GetConfig returns the combined normal, local, and sensitive Dogu configuration
// for the specified Dogu name, excluding specified keys.
func GetConfig(dogu string, ignoreKeys []string) (core.DoguConfig, error) {
	normalConfig, err := GetNormalConfig(dogu, ignoreKeys)
	if err != nil {
		return core.DoguConfig{}, fmt.Errorf("error getting normal dogu config: %w", err)
	}

	localConfig, err := GetLocalConfig(dogu, ignoreKeys)
	if err != nil {
		return core.DoguConfig{}, fmt.Errorf("error getting local dogu config: %w", err)
	}

	sensitiveConfig, err := GetSensitiveConfig(dogu, ignoreKeys)
	if err != nil {
		return core.DoguConfig{}, fmt.Errorf("error getting sensitive dogu config: %w", err)
	}

	return core.DoguConfig{
		Name:            dogu,
		NormalConfig:    normalConfig,
		LocalConfig:     localConfig,
		SensitiveConfig: sensitiveConfig,
	}, nil
}

// GetNormalConfig returns non-sensitive Dogu configuration keys that are not encrypted.
// It includes only those keys defined in the Dogu's JSON definition.
func GetNormalConfig(dogu string, ignoreKeys []string) ([]core.KeyValue, error) {
	doguConfigKeys, err := getDoguConfigKeySet(dogu, ExcludeGlobalConfig())
	if err != nil {
		return nil, fmt.Errorf("could not get dogu config keys: %w", err)
	}

	d, err := createDecryptFunc(dogu, GetGlobalConfig)
	if err != nil {
		return nil, fmt.Errorf("could not create decrypter: %w", err)
	}

	normalConfig, err := configGetKeyValues(
		path.Join(doguConfigPath, dogu),
		filterKeys(doguConfigKeys, false, []string{}), // only include keys from dogu.json
		ignoreSubKeys(ignoreKeys),                     // ignore keys provided by user
		filterEncryptedKeys(d, true),                  // exclude encrypted keys
	)
	if err != nil {
		return nil, fmt.Errorf("could not get dogu config: %w", err)
	}

	return normalConfig, nil
}

// GetLocalConfig returns Dogu configuration keys that are NOT defined in the Dogu's JSON
// and are decrypted if possible. Service account keys are ignored.
func GetLocalConfig(dogu string, ignoreKeys []string) ([]core.KeyValue, error) {
	doguConfigKeys, err := getDoguConfigKeySet(dogu)
	if err != nil {
		return nil, fmt.Errorf("could not get dogu config keys: %w", err)
	}

	d, err := createDecryptFunc(dogu, GetGlobalConfig)
	if err != nil {
		return nil, fmt.Errorf("could not create decrypter: %w", err)
	}

	// add service-account-keys to ignore-list for localConfig
	ignoreKeysWithServiceAccount := append(ignoreKeys, "/sa-")

	localConfig, err := configGetKeyValues(
		path.Join(doguConfigPath, dogu),
		filterKeys(doguConfigKeys, true, []string{}), // exclude keys from dogu.json
		decryptEncryptedKeys(d),                      // include all encrypted keys
		ignoreSubKeys(ignoreKeysWithServiceAccount),  // ignore keys provided by user
	)
	if err != nil {
		return nil, fmt.Errorf("could not get dogu config: %w", err)
	}

	return localConfig, nil
}

// GetSensitiveConfig returns only encrypted Dogu configuration keys
// that are listed in the Dogu's JSON definition.
func GetSensitiveConfig(dogu string, ignoreKeys []string) ([]core.KeyValue, error) {
	doguConfigKeys, err := getDoguConfigKeySet(dogu, ExcludeGlobalConfig())
	if err != nil {
		return nil, fmt.Errorf("could not get dogu config keys: %w", err)
	}

	d, err := createDecryptFunc(dogu, GetGlobalConfig)
	if err != nil {
		return nil, fmt.Errorf("could not create decrypter: %w", err)
	}

	sensitiveConfig, err := configGetKeyValues(
		path.Join(doguConfigPath, dogu),
		filterKeys(doguConfigKeys, false, []string{"sa-"}), // only include keys from dogu.json
		ignoreSubKeys(ignoreKeys),                          // ignore keys provided by user
		filterEncryptedKeys(d, false),                      // only include encrypted keys
	)
	if err != nil {
		return nil, fmt.Errorf("could not get sensitive dogu config: %w", err)
	}

	return sensitiveConfig, nil
}
