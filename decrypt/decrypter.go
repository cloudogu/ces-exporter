package decrypt

import (
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/cloudogu/cesapp-lib/keys"
	"log/slog"
	"path"
	"strings"
	"sync"
)

const (
	// keyProviderKey is the key used to identify the key provider entry in the global config.
	keyProviderKey = "key_provider"
)

var (
	// keyProvider is the singleton instance of the KeyProvider.
	keyProvider *keys.KeyProvider
	// errKeyProvider stores any error that occurred during KeyProvider initialization.
	errKeyProvider error
	// keyProviderOnce ensures the KeyProvider is only initialized once.
	keyProviderOnce sync.Once
)

var (
	doguVolumePath = "/data"
)

// GetGlobalConfigFunc defines a function type that retrieves the global config,
// optionally ignoring some keys.
type GetGlobalConfigFunc func(ignoreKeys []string) (core.GlobalConfig, error)

// getPrivateKeyPath constructs the file path to the private key PEM file
// for a given dogu name.
func getPrivateKeyPath(dogu string) string {
	doguPrivateKeyPath := fmt.Sprintf("/%s/volumes/_private/private.pem", dogu)
	return path.Join(doguVolumePath, doguPrivateKeyPath)
}

func createKeyProvider(getGCfg GetGlobalConfigFunc) (*keys.KeyProvider, error) {
	globalCfg, err := getGCfg([]string{})
	if err != nil {
		return nil, fmt.Errorf("failed to get global config: %w", err)
	}

	for _, kv := range globalCfg {
		if strings.Contains(kv.Key, keyProviderKey) {
			provider, pErr := keys.NewKeyProvider(kv.Value)
			if pErr != nil {
				return nil, fmt.Errorf("failed to create key provider: %w", pErr)
			}

			slog.Info("Use KeyProvider to decrypt sensitive configs", "provider", kv.Value)

			return provider, nil
		}
	}

	return nil, fmt.Errorf("no key provider found for global config")
}

// GetKeyProvider initializes and returns a singleton instance of KeyProvider.
// It reads the global config using the provided function and extracts the
// key provider configuration.
//
// The function ensures thread-safe lazy initialization and returns the same
// instance for subsequent calls. If an error occurs during initialization,
// it is returned alongside a nil provider.
func GetKeyProvider(getGCfg GetGlobalConfigFunc) (*keys.KeyProvider, error) {
	keyProviderOnce.Do(func() {
		keyProvider, errKeyProvider = createKeyProvider(getGCfg)
	})

	return keyProvider, errKeyProvider
}

// CreateDecrypter creates a Decrypter instance using the private key for the specified dogu.
// It retrieves the KeyProvider and loads the private key from the filesystem.
// Returns an error if the provider or private key could not be retrieved.
func CreateDecrypter(dogu string, getGCfg GetGlobalConfigFunc) (Decrypter, error) {
	provider, err := GetKeyProvider(getGCfg)
	if err != nil {
		return Decrypter{}, fmt.Errorf("failed to get key provider: %w", err)
	}

	keyPair, err := provider.FromPrivateKeyPath(getPrivateKeyPath(dogu))
	if err != nil {
		return Decrypter{}, fmt.Errorf("failed to create key pair: %w", err)
	}

	return Decrypter{privateKey: keyPair.Private()}, nil
}

// Decrypter is a wrapper around a private key that can decrypt encrypted input.
type Decrypter struct {
	privateKey *keys.PrivateKey
}

// Decrypt decrypts the given input string using the private key.
// Returns the decrypted string or an error if decryption fails.
func (d Decrypter) Decrypt(input string) (string, error) {
	return d.privateKey.Decrypt(input)
}
