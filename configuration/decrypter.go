package configuration

import (
	"fmt"
	"github.com/cloudogu/cesapp-lib/keys"
	"log/slog"
	"strings"
	"sync"
)

const (
	_KeyProviderKey = "key_provider"
)

var (
	keyProvider     *keys.KeyProvider
	keyProviderErr  error
	keyProviderOnce sync.Once
)

func getPrivateKeyPath(dogu string) string {
	return fmt.Sprintf("/var/lib/ces/%s/volumes/_private/private.pem", dogu)
}

func GetKeyProvider(getGCfg getGlobalConfigFunc) (*keys.KeyProvider, error) {
	keyProviderOnce.Do(func() {
		globalCfg, err := getGCfg([]string{})
		if err != nil {
			keyProviderErr = fmt.Errorf("failed to get global config: %w", err)
			return
		}

		for _, kv := range globalCfg {
			if strings.Contains(kv.Key, _KeyProviderKey) {
				provider, pErr := keys.NewKeyProvider(kv.Value)
				if pErr != nil {
					keyProviderErr = fmt.Errorf("failed to create key provider: %w", pErr)
					return
				}

				slog.Debug("KeyProvided extracted", "provider", kv.Value)

				keyProvider = provider
				return
			}
		}

		keyProviderErr = fmt.Errorf("no key provider found for global config")
	})

	return keyProvider, keyProviderErr
}

func CreateDecrypter(dogu string, getGCfg getGlobalConfigFunc) (Decrypter, error) {
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

type Decrypter struct {
	privateKey *keys.PrivateKey
}

func (d Decrypter) Decrypt(input string) (string, error) {
	return d.privateKey.Decrypt(input)
}
