package etcd

import (
	"encoding/json"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/cloudogu/ces-exporter/decrypt"
	"github.com/cloudogu/ces-exporter/etcd/mocks"
	cesLibCore "github.com/cloudogu/cesapp-lib/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"slices"
	"testing"
)

func TestConfigFilterOptions(t *testing.T) {
	initialKeyValues := []core.KeyValue{
		{Key: "/public.pem"},
		{Key: "/sa-cas/cas_client_id"},
		{Key: "/sa-ldap/password", Value: "encrypted"},
		{Key: "/sa-ldap/username", Value: "encrypted"},
		{Key: "/logging/root"},
	}

	t.Run("filterKeys", func(t *testing.T) {
		t.Run("include keys from filterMap", func(t *testing.T) {
			option := filterKeys(map[string]bool{
				"sa-cas/cas_client_id": true,
				"logging/root":         true,
			}, false, []string{})

			filteredKeyValues := option(initialKeyValues)
			assert.Len(t, filteredKeyValues, 2)

			filteredKeys := make([]string, len(filteredKeyValues))
			for i := range filteredKeyValues {
				filteredKeys[i] = filteredKeyValues[i].Key
			}

			assert.True(t, slices.Contains(filteredKeys, "/sa-cas/cas_client_id"))
			assert.True(t, slices.Contains(filteredKeys, "/logging/root"))
		})

		t.Run("exclude keys from filterMap", func(t *testing.T) {
			option := filterKeys(map[string]bool{
				"sa-cas/cas_client_id": true,
				"logging/root":         true,
			}, true, []string{})

			filteredKeyValues := option(initialKeyValues)
			assert.Len(t, filteredKeyValues, 3)

			filteredKeys := make([]string, len(filteredKeyValues))
			for i := range filteredKeyValues {
				filteredKeys[i] = filteredKeyValues[i].Key
			}

			assert.True(t, slices.Contains(filteredKeys, "/public.pem"))
			assert.True(t, slices.Contains(filteredKeys, "/sa-ldap/password"))
			assert.True(t, slices.Contains(filteredKeys, "/sa-ldap/username"))
		})

		t.Run("exception to include all service accounts", func(t *testing.T) {
			option := filterKeys(
				map[string]bool{
					"logging/root": true,
				},
				false,
				[]string{"sa-"})

			filteredKeyValues := option(initialKeyValues)
			assert.Len(t, filteredKeyValues, 4)

			filteredKeys := make([]string, len(filteredKeyValues))
			for i := range filteredKeyValues {
				filteredKeys[i] = filteredKeyValues[i].Key
			}

			assert.True(t, slices.Contains(filteredKeys, "/logging/root"))
			assert.True(t, slices.Contains(filteredKeys, "/sa-cas/cas_client_id"))
			assert.True(t, slices.Contains(filteredKeys, "/sa-ldap/password"))
			assert.True(t, slices.Contains(filteredKeys, "/sa-ldap/username"))
		})
	})

	t.Run("ignoreSubKeys", func(t *testing.T) {
		tests := []struct {
			name      string
			regexes   []string
			expLength int
		}{
			{
				name:      "empty regexes",
				regexes:   []string{},
				expLength: 5,
			},
			{
				name: "ignore pem",
				regexes: []string{
					`public\.pem$`,
				},
				expLength: 4,
			},
			{
				name: "ignore service accounts",
				regexes: []string{
					`sa-[^/]+`,
				},
				expLength: 2,
			},
			{
				name: "ignore service accounts and pem",
				regexes: []string{
					`public\.pem$`,
					`sa-[^/]+`,
				},
				expLength: 1,
			},
			{
				name: "invalid regex",
				regexes: []string{
					`[a-z`,
				},
				expLength: 5,
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				option := ignoreSubKeys(test.regexes)
				filteredKeyValues := option(initialKeyValues)

				assert.Len(t, filteredKeyValues, test.expLength)
			})
		}
	})

	t.Run("filterEncryptedKeys", func(t *testing.T) {
		t.Run("include encrypted keys", func(t *testing.T) {
			decrypterMock := mocks.NewDecrypter(t)
			decrypterMock.EXPECT().Decrypt("encrypted").Return("", nil)
			decrypterMock.EXPECT().Decrypt("").Return("", assert.AnError)

			option := filterEncryptedKeys(decrypterMock, false)

			filteredKeyValues := option(initialKeyValues)
			assert.Len(t, filteredKeyValues, 2)

			filteredKeys := make([]string, len(filteredKeyValues))
			for i := range filteredKeyValues {
				filteredKeys[i] = filteredKeyValues[i].Key
			}

			assert.True(t, slices.Contains(filteredKeys, "/sa-ldap/password"))
			assert.True(t, slices.Contains(filteredKeys, "/sa-ldap/username"))
		})

		t.Run("exclude encrypted keys", func(t *testing.T) {
			decrypterMock := mocks.NewDecrypter(t)
			decrypterMock.EXPECT().Decrypt("encrypted").Return("", nil)
			decrypterMock.EXPECT().Decrypt("").Return("", assert.AnError)

			option := filterEncryptedKeys(decrypterMock, true)

			filteredKeyValues := option(initialKeyValues)
			assert.Len(t, filteredKeyValues, 3)

			filteredKeys := make([]string, len(filteredKeyValues))
			for i := range filteredKeyValues {
				filteredKeys[i] = filteredKeyValues[i].Key
			}

			assert.True(t, slices.Contains(filteredKeys, "/public.pem"))
			assert.True(t, slices.Contains(filteredKeys, "/sa-cas/cas_client_id"))
			assert.True(t, slices.Contains(filteredKeys, "/logging/root"))
		})
	})

	t.Run("decryptEncryptedKeys", func(t *testing.T) {
		decrypterMock := mocks.NewDecrypter(t)
		decrypterMock.EXPECT().Decrypt("encrypted").Return("decrypted", nil)
		decrypterMock.EXPECT().Decrypt("").Return("", assert.AnError)

		option := decryptEncryptedKeys(decrypterMock)

		filteredKeyValues := option(initialKeyValues)
		assert.Len(t, filteredKeyValues, 5)

		for _, kv := range filteredKeyValues {
			switch kv.Key {
			case "/sa-ldap/password", "/sa-ldap/username":
				assert.True(t, kv.Value == "decrypted")
			default:
				assert.True(t, kv.Value == "")
			}
		}
	})
}

func TestGetGlobalConfig(t *testing.T) {
	// fake once call
	clientOnce.Do(func() {})

	oldConfigGetKeyValues := configGetKeyValues
	defer func() {
		configGetKeyValues = oldConfigGetKeyValues
	}()

	configGetKeyValuesMock := mocks.NewGetKeyValuesClientFuncType(t)
	configGetKeyValuesMock.EXPECT().Execute(globalConfigPath, mock.Anything).Return(core.GlobalConfig{}, nil)

	configGetKeyValues = configGetKeyValuesMock.Execute

	gCfg, err := GetGlobalConfig([]string{})
	assert.NoError(t, err)
	assert.NotNil(t, gCfg)
}

func TestGetNormalConfig(t *testing.T) {
	// fake once call
	clientOnce.Do(func() {})

	const usermgtDogu = "usermgt"

	t.Run("GetNormalConfig", func(t *testing.T) {
		doguJson, err := os.ReadFile("testdata/dogu.json")
		require.NoError(t, err)

		var expDoguSpec cesLibCore.Dogu
		err = json.Unmarshal(doguJson, &expDoguSpec)
		require.NoError(t, err)

		oldConfigGetKeyValues := configGetKeyValues
		defer func() {
			configGetKeyValues = oldConfigGetKeyValues
			doguGetKeyValues = oldConfigGetKeyValues
		}()

		oldCreateDecryptFunc := createDecryptFunc
		defer func() {
			createDecryptFunc = oldCreateDecryptFunc
		}()

		doguCurrentPath := path.Join(doguPath, usermgtDogu, "current")
		doguSpecPath := path.Join(doguPath, usermgtDogu, expDoguSpec.Version)

		getKeyValuesMock := mocks.NewGetKeyValuesClientFuncType(t)
		getKeyValuesMock.EXPECT().Execute(doguCurrentPath).Return([]core.KeyValue{
			{
				Key:   doguCurrentPath,
				Value: expDoguSpec.Version,
			},
		}, nil)
		getKeyValuesMock.EXPECT().Execute(doguSpecPath).Return([]core.KeyValue{
			{
				Key:   doguSpecPath,
				Value: string(doguJson),
			},
		}, nil)

		getKeyValuesMock.EXPECT().Execute(
			path.Join("/config", usermgtDogu),
			mock.AnythingOfType("core.FilterOption"),
			mock.AnythingOfType("core.FilterOption"),
			mock.AnythingOfType("core.FilterOption"),
		).Return([]core.KeyValue{}, nil)

		configGetKeyValues = getKeyValuesMock.Execute
		doguGetKeyValues = getKeyValuesMock.Execute

		createDecryptFunc = func(dogu string, getGCfg decrypt.GetGlobalConfigFunc) (decrypt.Decrypter, error) {
			return decrypt.Decrypter{}, nil
		}

		_, err = GetNormalConfig(usermgtDogu, []string{})
		assert.NoError(t, err)
	})

	t.Run("getDoguConfigKeySet returns error", func(t *testing.T) {
		oldConfigGetKeyValues := configGetKeyValues
		defer func() {
			configGetKeyValues = oldConfigGetKeyValues
			doguGetKeyValues = oldConfigGetKeyValues
		}()

		oldCreateDecryptFunc := createDecryptFunc
		defer func() {
			createDecryptFunc = oldCreateDecryptFunc
		}()

		doguCurrentPath := path.Join(doguPath, usermgtDogu, "current")

		getKeyValuesMock := mocks.NewGetKeyValuesClientFuncType(t)
		getKeyValuesMock.EXPECT().Execute(doguCurrentPath).Return(nil, assert.AnError)

		configGetKeyValues = getKeyValuesMock.Execute
		doguGetKeyValues = getKeyValuesMock.Execute

		createDecryptFunc = func(dogu string, getGCfg decrypt.GetGlobalConfigFunc) (decrypt.Decrypter, error) {
			return decrypt.Decrypter{}, nil
		}

		_, err := GetNormalConfig(usermgtDogu, []string{})
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("createDecryptFunc returns error", func(t *testing.T) {
		doguJson, err := os.ReadFile("testdata/dogu.json")
		require.NoError(t, err)

		var expDoguSpec cesLibCore.Dogu
		err = json.Unmarshal(doguJson, &expDoguSpec)
		require.NoError(t, err)

		oldConfigGetKeyValues := configGetKeyValues
		defer func() {
			configGetKeyValues = oldConfigGetKeyValues
			doguGetKeyValues = oldConfigGetKeyValues
		}()

		oldCreateDecryptFunc := createDecryptFunc
		defer func() {
			createDecryptFunc = oldCreateDecryptFunc
		}()

		doguCurrentPath := path.Join(doguPath, usermgtDogu, "current")
		doguSpecPath := path.Join(doguPath, usermgtDogu, expDoguSpec.Version)

		getKeyValuesMock := mocks.NewGetKeyValuesClientFuncType(t)
		getKeyValuesMock.EXPECT().Execute(doguCurrentPath).Return([]core.KeyValue{
			{
				Key:   doguCurrentPath,
				Value: expDoguSpec.Version,
			},
		}, nil)
		getKeyValuesMock.EXPECT().Execute(doguSpecPath).Return([]core.KeyValue{
			{
				Key:   doguSpecPath,
				Value: string(doguJson),
			},
		}, nil)

		configGetKeyValues = getKeyValuesMock.Execute
		doguGetKeyValues = getKeyValuesMock.Execute

		createDecryptFunc = func(dogu string, getGCfg decrypt.GetGlobalConfigFunc) (decrypt.Decrypter, error) {
			return decrypt.Decrypter{}, assert.AnError
		}

		_, err = GetNormalConfig(usermgtDogu, []string{})
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("configGetKeyValues returns error", func(t *testing.T) {
		doguJson, err := os.ReadFile("testdata/dogu.json")
		require.NoError(t, err)

		var expDoguSpec cesLibCore.Dogu
		err = json.Unmarshal(doguJson, &expDoguSpec)
		require.NoError(t, err)

		oldConfigGetKeyValues := configGetKeyValues
		defer func() {
			configGetKeyValues = oldConfigGetKeyValues
			doguGetKeyValues = oldConfigGetKeyValues
		}()

		oldCreateDecryptFunc := createDecryptFunc
		defer func() {
			createDecryptFunc = oldCreateDecryptFunc
		}()

		doguCurrentPath := path.Join(doguPath, usermgtDogu, "current")
		doguSpecPath := path.Join(doguPath, usermgtDogu, expDoguSpec.Version)

		getKeyValuesMock := mocks.NewGetKeyValuesClientFuncType(t)
		getKeyValuesMock.EXPECT().Execute(doguCurrentPath).Return([]core.KeyValue{
			{
				Key:   doguCurrentPath,
				Value: expDoguSpec.Version,
			},
		}, nil)
		getKeyValuesMock.EXPECT().Execute(doguSpecPath).Return([]core.KeyValue{
			{
				Key:   doguSpecPath,
				Value: string(doguJson),
			},
		}, nil)

		getKeyValuesMock.EXPECT().Execute(
			path.Join("/config", usermgtDogu),
			mock.AnythingOfType("core.FilterOption"),
			mock.AnythingOfType("core.FilterOption"),
			mock.AnythingOfType("core.FilterOption"),
		).Return(nil, assert.AnError)

		configGetKeyValues = getKeyValuesMock.Execute
		doguGetKeyValues = getKeyValuesMock.Execute

		createDecryptFunc = func(dogu string, getGCfg decrypt.GetGlobalConfigFunc) (decrypt.Decrypter, error) {
			return decrypt.Decrypter{}, nil
		}

		_, err = GetNormalConfig(usermgtDogu, []string{})
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestGetSensitiveConfig(t *testing.T) {
	// fake once call
	clientOnce.Do(func() {})

	const usermgtDogu = "usermgt"

	t.Run("GetSensitiveConfig", func(t *testing.T) {
		doguJson, err := os.ReadFile("testdata/dogu.json")
		require.NoError(t, err)

		var expDoguSpec cesLibCore.Dogu
		err = json.Unmarshal(doguJson, &expDoguSpec)
		require.NoError(t, err)

		oldConfigGetKeyValues := configGetKeyValues
		defer func() {
			configGetKeyValues = oldConfigGetKeyValues
			doguGetKeyValues = oldConfigGetKeyValues
		}()

		oldCreateDecryptFunc := createDecryptFunc
		defer func() {
			createDecryptFunc = oldCreateDecryptFunc
		}()

		doguCurrentPath := path.Join(doguPath, usermgtDogu, "current")
		doguSpecPath := path.Join(doguPath, usermgtDogu, expDoguSpec.Version)

		getKeyValuesMock := mocks.NewGetKeyValuesClientFuncType(t)
		getKeyValuesMock.EXPECT().Execute(doguCurrentPath).Return([]core.KeyValue{
			{
				Key:   doguCurrentPath,
				Value: expDoguSpec.Version,
			},
		}, nil)
		getKeyValuesMock.EXPECT().Execute(doguSpecPath).Return([]core.KeyValue{
			{
				Key:   doguSpecPath,
				Value: string(doguJson),
			},
		}, nil)

		getKeyValuesMock.EXPECT().Execute(
			path.Join("/config", usermgtDogu),
			mock.AnythingOfType("core.FilterOption"),
			mock.AnythingOfType("core.FilterOption"),
			mock.AnythingOfType("core.FilterOption"),
		).Return([]core.KeyValue{}, nil)

		configGetKeyValues = getKeyValuesMock.Execute
		doguGetKeyValues = getKeyValuesMock.Execute

		createDecryptFunc = func(dogu string, getGCfg decrypt.GetGlobalConfigFunc) (decrypt.Decrypter, error) {
			return decrypt.Decrypter{}, nil
		}

		_, err = GetSensitiveConfig(usermgtDogu, []string{})
		assert.NoError(t, err)
	})

	t.Run("getDoguConfigKeySet returns error", func(t *testing.T) {
		oldConfigGetKeyValues := configGetKeyValues
		defer func() {
			configGetKeyValues = oldConfigGetKeyValues
			doguGetKeyValues = oldConfigGetKeyValues
		}()

		oldCreateDecryptFunc := createDecryptFunc
		defer func() {
			createDecryptFunc = oldCreateDecryptFunc
		}()

		doguCurrentPath := path.Join(doguPath, usermgtDogu, "current")

		getKeyValuesMock := mocks.NewGetKeyValuesClientFuncType(t)
		getKeyValuesMock.EXPECT().Execute(doguCurrentPath).Return(nil, assert.AnError)

		configGetKeyValues = getKeyValuesMock.Execute
		doguGetKeyValues = getKeyValuesMock.Execute

		createDecryptFunc = func(dogu string, getGCfg decrypt.GetGlobalConfigFunc) (decrypt.Decrypter, error) {
			return decrypt.Decrypter{}, nil
		}

		_, err := GetSensitiveConfig(usermgtDogu, []string{})
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("createDecryptFunc returns error", func(t *testing.T) {
		doguJson, err := os.ReadFile("testdata/dogu.json")
		require.NoError(t, err)

		var expDoguSpec cesLibCore.Dogu
		err = json.Unmarshal(doguJson, &expDoguSpec)
		require.NoError(t, err)

		oldConfigGetKeyValues := configGetKeyValues
		defer func() {
			configGetKeyValues = oldConfigGetKeyValues
			doguGetKeyValues = oldConfigGetKeyValues
		}()

		oldCreateDecryptFunc := createDecryptFunc
		defer func() {
			createDecryptFunc = oldCreateDecryptFunc
		}()

		doguCurrentPath := path.Join(doguPath, usermgtDogu, "current")
		doguSpecPath := path.Join(doguPath, usermgtDogu, expDoguSpec.Version)

		getKeyValuesMock := mocks.NewGetKeyValuesClientFuncType(t)
		getKeyValuesMock.EXPECT().Execute(doguCurrentPath).Return([]core.KeyValue{
			{
				Key:   doguCurrentPath,
				Value: expDoguSpec.Version,
			},
		}, nil)
		getKeyValuesMock.EXPECT().Execute(doguSpecPath).Return([]core.KeyValue{
			{
				Key:   doguSpecPath,
				Value: string(doguJson),
			},
		}, nil)

		configGetKeyValues = getKeyValuesMock.Execute
		doguGetKeyValues = getKeyValuesMock.Execute

		createDecryptFunc = func(dogu string, getGCfg decrypt.GetGlobalConfigFunc) (decrypt.Decrypter, error) {
			return decrypt.Decrypter{}, assert.AnError
		}

		_, err = GetSensitiveConfig(usermgtDogu, []string{})
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("configGetKeyValues returns error", func(t *testing.T) {
		doguJson, err := os.ReadFile("testdata/dogu.json")
		require.NoError(t, err)

		var expDoguSpec cesLibCore.Dogu
		err = json.Unmarshal(doguJson, &expDoguSpec)
		require.NoError(t, err)

		oldConfigGetKeyValues := configGetKeyValues
		defer func() {
			configGetKeyValues = oldConfigGetKeyValues
			doguGetKeyValues = oldConfigGetKeyValues
		}()

		oldCreateDecryptFunc := createDecryptFunc
		defer func() {
			createDecryptFunc = oldCreateDecryptFunc
		}()

		doguCurrentPath := path.Join(doguPath, usermgtDogu, "current")
		doguSpecPath := path.Join(doguPath, usermgtDogu, expDoguSpec.Version)

		getKeyValuesMock := mocks.NewGetKeyValuesClientFuncType(t)
		getKeyValuesMock.EXPECT().Execute(doguCurrentPath).Return([]core.KeyValue{
			{
				Key:   doguCurrentPath,
				Value: expDoguSpec.Version,
			},
		}, nil)
		getKeyValuesMock.EXPECT().Execute(doguSpecPath).Return([]core.KeyValue{
			{
				Key:   doguSpecPath,
				Value: string(doguJson),
			},
		}, nil)

		getKeyValuesMock.EXPECT().Execute(
			path.Join("/config", usermgtDogu),
			mock.AnythingOfType("core.FilterOption"),
			mock.AnythingOfType("core.FilterOption"),
			mock.AnythingOfType("core.FilterOption"),
		).Return(nil, assert.AnError)

		configGetKeyValues = getKeyValuesMock.Execute
		doguGetKeyValues = getKeyValuesMock.Execute

		createDecryptFunc = func(dogu string, getGCfg decrypt.GetGlobalConfigFunc) (decrypt.Decrypter, error) {
			return decrypt.Decrypter{}, nil
		}

		_, err = GetSensitiveConfig(usermgtDogu, []string{})
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestGetLocalConfig(t *testing.T) {
	// fake once call
	clientOnce.Do(func() {})

	const usermgtDogu = "usermgt"

	t.Run("GetLocalConfig", func(t *testing.T) {
		doguJson, err := os.ReadFile("testdata/dogu.json")
		require.NoError(t, err)

		var expDoguSpec cesLibCore.Dogu
		err = json.Unmarshal(doguJson, &expDoguSpec)
		require.NoError(t, err)

		oldConfigGetKeyValues := configGetKeyValues
		defer func() {
			configGetKeyValues = oldConfigGetKeyValues
			doguGetKeyValues = oldConfigGetKeyValues
		}()

		oldCreateDecryptFunc := createDecryptFunc
		defer func() {
			createDecryptFunc = oldCreateDecryptFunc
		}()

		doguCurrentPath := path.Join(doguPath, usermgtDogu, "current")
		doguSpecPath := path.Join(doguPath, usermgtDogu, expDoguSpec.Version)

		getKeyValuesMock := mocks.NewGetKeyValuesClientFuncType(t)
		getKeyValuesMock.EXPECT().Execute(doguCurrentPath).Return([]core.KeyValue{
			{
				Key:   doguCurrentPath,
				Value: expDoguSpec.Version,
			},
		}, nil)
		getKeyValuesMock.EXPECT().Execute(doguSpecPath).Return([]core.KeyValue{
			{
				Key:   doguSpecPath,
				Value: string(doguJson),
			},
		}, nil)

		getKeyValuesMock.EXPECT().Execute(
			path.Join("/config", usermgtDogu),
			mock.AnythingOfType("core.FilterOption"),
			mock.AnythingOfType("core.FilterOption"),
			mock.AnythingOfType("core.FilterOption"),
		).Return([]core.KeyValue{}, nil)

		configGetKeyValues = getKeyValuesMock.Execute
		doguGetKeyValues = getKeyValuesMock.Execute

		createDecryptFunc = func(dogu string, getGCfg decrypt.GetGlobalConfigFunc) (decrypt.Decrypter, error) {
			return decrypt.Decrypter{}, nil
		}

		_, err = GetLocalConfig(usermgtDogu, []string{})
		assert.NoError(t, err)
	})

	t.Run("getDoguConfigKeySet returns error", func(t *testing.T) {
		oldConfigGetKeyValues := configGetKeyValues
		defer func() {
			configGetKeyValues = oldConfigGetKeyValues
			doguGetKeyValues = oldConfigGetKeyValues
		}()

		oldCreateDecryptFunc := createDecryptFunc
		defer func() {
			createDecryptFunc = oldCreateDecryptFunc
		}()

		doguCurrentPath := path.Join(doguPath, usermgtDogu, "current")

		getKeyValuesMock := mocks.NewGetKeyValuesClientFuncType(t)
		getKeyValuesMock.EXPECT().Execute(doguCurrentPath).Return(nil, assert.AnError)

		configGetKeyValues = getKeyValuesMock.Execute
		doguGetKeyValues = getKeyValuesMock.Execute

		createDecryptFunc = func(dogu string, getGCfg decrypt.GetGlobalConfigFunc) (decrypt.Decrypter, error) {
			return decrypt.Decrypter{}, nil
		}

		_, err := GetLocalConfig(usermgtDogu, []string{})
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("createDecryptFunc returns error", func(t *testing.T) {
		doguJson, err := os.ReadFile("testdata/dogu.json")
		require.NoError(t, err)

		var expDoguSpec cesLibCore.Dogu
		err = json.Unmarshal(doguJson, &expDoguSpec)
		require.NoError(t, err)

		oldConfigGetKeyValues := configGetKeyValues
		defer func() {
			configGetKeyValues = oldConfigGetKeyValues
			doguGetKeyValues = oldConfigGetKeyValues
		}()

		oldCreateDecryptFunc := createDecryptFunc
		defer func() {
			createDecryptFunc = oldCreateDecryptFunc
		}()

		doguCurrentPath := path.Join(doguPath, usermgtDogu, "current")
		doguSpecPath := path.Join(doguPath, usermgtDogu, expDoguSpec.Version)

		getKeyValuesMock := mocks.NewGetKeyValuesClientFuncType(t)
		getKeyValuesMock.EXPECT().Execute(doguCurrentPath).Return([]core.KeyValue{
			{
				Key:   doguCurrentPath,
				Value: expDoguSpec.Version,
			},
		}, nil)
		getKeyValuesMock.EXPECT().Execute(doguSpecPath).Return([]core.KeyValue{
			{
				Key:   doguSpecPath,
				Value: string(doguJson),
			},
		}, nil)

		configGetKeyValues = getKeyValuesMock.Execute
		doguGetKeyValues = getKeyValuesMock.Execute

		createDecryptFunc = func(dogu string, getGCfg decrypt.GetGlobalConfigFunc) (decrypt.Decrypter, error) {
			return decrypt.Decrypter{}, assert.AnError
		}

		_, err = GetLocalConfig(usermgtDogu, []string{})
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("configGetKeyValues returns error", func(t *testing.T) {
		doguJson, err := os.ReadFile("testdata/dogu.json")
		require.NoError(t, err)

		var expDoguSpec cesLibCore.Dogu
		err = json.Unmarshal(doguJson, &expDoguSpec)
		require.NoError(t, err)

		oldConfigGetKeyValues := configGetKeyValues
		defer func() {
			configGetKeyValues = oldConfigGetKeyValues
			doguGetKeyValues = oldConfigGetKeyValues
		}()

		oldCreateDecryptFunc := createDecryptFunc
		defer func() {
			createDecryptFunc = oldCreateDecryptFunc
		}()

		doguCurrentPath := path.Join(doguPath, usermgtDogu, "current")
		doguSpecPath := path.Join(doguPath, usermgtDogu, expDoguSpec.Version)

		getKeyValuesMock := mocks.NewGetKeyValuesClientFuncType(t)
		getKeyValuesMock.EXPECT().Execute(doguCurrentPath).Return([]core.KeyValue{
			{
				Key:   doguCurrentPath,
				Value: expDoguSpec.Version,
			},
		}, nil)
		getKeyValuesMock.EXPECT().Execute(doguSpecPath).Return([]core.KeyValue{
			{
				Key:   doguSpecPath,
				Value: string(doguJson),
			},
		}, nil)

		getKeyValuesMock.EXPECT().Execute(
			path.Join("/config", usermgtDogu),
			mock.AnythingOfType("core.FilterOption"),
			mock.AnythingOfType("core.FilterOption"),
			mock.AnythingOfType("core.FilterOption"),
		).Return(nil, assert.AnError)

		configGetKeyValues = getKeyValuesMock.Execute
		doguGetKeyValues = getKeyValuesMock.Execute

		createDecryptFunc = func(dogu string, getGCfg decrypt.GetGlobalConfigFunc) (decrypt.Decrypter, error) {
			return decrypt.Decrypter{}, nil
		}

		_, err = GetLocalConfig(usermgtDogu, []string{})
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestGetConfig(t *testing.T) {
	// fake once call
	clientOnce.Do(func() {})

	const usermgtDogu = "usermgt"

	t.Run("GetConfig", func(t *testing.T) {
		doguJson, err := os.ReadFile("testdata/dogu.json")
		require.NoError(t, err)

		var expDoguSpec cesLibCore.Dogu
		err = json.Unmarshal(doguJson, &expDoguSpec)
		require.NoError(t, err)

		oldConfigGetKeyValues := configGetKeyValues
		defer func() {
			configGetKeyValues = oldConfigGetKeyValues
			doguGetKeyValues = oldConfigGetKeyValues
		}()

		oldCreateDecryptFunc := createDecryptFunc
		defer func() {
			createDecryptFunc = oldCreateDecryptFunc
		}()

		doguCurrentPath := path.Join(doguPath, usermgtDogu, "current")
		doguSpecPath := path.Join(doguPath, usermgtDogu, expDoguSpec.Version)

		getKeyValuesMock := mocks.NewGetKeyValuesClientFuncType(t)
		getKeyValuesMock.EXPECT().Execute(doguCurrentPath).Return([]core.KeyValue{
			{
				Key:   doguCurrentPath,
				Value: expDoguSpec.Version,
			},
		}, nil)
		getKeyValuesMock.EXPECT().Execute(doguSpecPath).Return([]core.KeyValue{
			{
				Key:   doguSpecPath,
				Value: string(doguJson),
			},
		}, nil)

		getKeyValuesMock.EXPECT().Execute(
			path.Join("/config", usermgtDogu),
			mock.AnythingOfType("core.FilterOption"),
			mock.AnythingOfType("core.FilterOption"),
			mock.AnythingOfType("core.FilterOption"),
		).Return([]core.KeyValue{}, nil)

		configGetKeyValues = getKeyValuesMock.Execute
		doguGetKeyValues = getKeyValuesMock.Execute

		createDecryptFunc = func(dogu string, getGCfg decrypt.GetGlobalConfigFunc) (decrypt.Decrypter, error) {
			return decrypt.Decrypter{}, nil
		}

		_, err = GetConfig(usermgtDogu, []string{})
		assert.NoError(t, err)
	})

	t.Run("GetNormalConfig returns error", func(t *testing.T) {
		oldConfigGetKeyValues := configGetKeyValues
		defer func() {
			configGetKeyValues = oldConfigGetKeyValues
			doguGetKeyValues = oldConfigGetKeyValues
		}()

		oldCreateDecryptFunc := createDecryptFunc
		defer func() {
			createDecryptFunc = oldCreateDecryptFunc
		}()

		doguCurrentPath := path.Join(doguPath, usermgtDogu, "current")

		getKeyValuesMock := mocks.NewGetKeyValuesClientFuncType(t)
		getKeyValuesMock.EXPECT().Execute(doguCurrentPath).Return(nil, assert.AnError)

		configGetKeyValues = getKeyValuesMock.Execute
		doguGetKeyValues = getKeyValuesMock.Execute

		createDecryptFunc = func(dogu string, getGCfg decrypt.GetGlobalConfigFunc) (decrypt.Decrypter, error) {
			return decrypt.Decrypter{}, nil
		}

		_, err := GetConfig(usermgtDogu, []string{})
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("GetLocalConfig returns error", func(t *testing.T) {
		doguJson, err := os.ReadFile("testdata/dogu.json")
		require.NoError(t, err)

		var expDoguSpec cesLibCore.Dogu
		err = json.Unmarshal(doguJson, &expDoguSpec)
		require.NoError(t, err)

		oldConfigGetKeyValues := configGetKeyValues
		defer func() {
			configGetKeyValues = oldConfigGetKeyValues
			doguGetKeyValues = oldConfigGetKeyValues
		}()

		oldCreateDecryptFunc := createDecryptFunc
		defer func() {
			createDecryptFunc = oldCreateDecryptFunc
		}()

		doguCurrentPath := path.Join(doguPath, usermgtDogu, "current")
		doguSpecPath := path.Join(doguPath, usermgtDogu, expDoguSpec.Version)

		getKeyValuesMock := mocks.NewGetKeyValuesClientFuncType(t)

		getKeyValuesMock.EXPECT().Execute(doguCurrentPath).
			RunAndReturn(
				func(s string, option ...core.FilterOption) ([]core.KeyValue, error) {
					if len(getKeyValuesMock.Calls) == 4 {
						return nil, assert.AnError
					}

					return []core.KeyValue{
						{
							Key:   doguCurrentPath,
							Value: expDoguSpec.Version,
						},
					}, nil
				})

		getKeyValuesMock.EXPECT().Execute(doguSpecPath).Return([]core.KeyValue{
			{
				Key:   doguSpecPath,
				Value: string(doguJson),
			},
		}, nil)

		getKeyValuesMock.EXPECT().Execute(
			path.Join("/config", usermgtDogu),
			mock.AnythingOfType("core.FilterOption"),
			mock.AnythingOfType("core.FilterOption"),
			mock.AnythingOfType("core.FilterOption"),
		).Return([]core.KeyValue{}, nil)

		configGetKeyValues = getKeyValuesMock.Execute
		doguGetKeyValues = getKeyValuesMock.Execute

		createDecryptFunc = func(dogu string, getGCfg decrypt.GetGlobalConfigFunc) (decrypt.Decrypter, error) {
			return decrypt.Decrypter{}, nil
		}

		_, err = GetConfig(usermgtDogu, []string{})
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("GetSensitiveConfig returns error", func(t *testing.T) {
		doguJson, err := os.ReadFile("testdata/dogu.json")
		require.NoError(t, err)

		var expDoguSpec cesLibCore.Dogu
		err = json.Unmarshal(doguJson, &expDoguSpec)
		require.NoError(t, err)

		oldConfigGetKeyValues := configGetKeyValues
		defer func() {
			configGetKeyValues = oldConfigGetKeyValues
			doguGetKeyValues = oldConfigGetKeyValues
		}()

		oldCreateDecryptFunc := createDecryptFunc
		defer func() {
			createDecryptFunc = oldCreateDecryptFunc
		}()

		doguCurrentPath := path.Join(doguPath, usermgtDogu, "current")
		doguSpecPath := path.Join(doguPath, usermgtDogu, expDoguSpec.Version)

		getKeyValuesMock := mocks.NewGetKeyValuesClientFuncType(t)

		getKeyValuesMock.EXPECT().Execute(doguCurrentPath).
			RunAndReturn(
				func(s string, option ...core.FilterOption) ([]core.KeyValue, error) {
					if len(getKeyValuesMock.Calls) == 7 {
						return nil, assert.AnError
					}

					return []core.KeyValue{
						{
							Key:   doguCurrentPath,
							Value: expDoguSpec.Version,
						},
					}, nil
				})

		getKeyValuesMock.EXPECT().Execute(doguSpecPath).Return([]core.KeyValue{
			{
				Key:   doguSpecPath,
				Value: string(doguJson),
			},
		}, nil)

		getKeyValuesMock.EXPECT().Execute(
			path.Join("/config", usermgtDogu),
			mock.AnythingOfType("core.FilterOption"),
			mock.AnythingOfType("core.FilterOption"),
			mock.AnythingOfType("core.FilterOption"),
		).Return([]core.KeyValue{}, nil)

		configGetKeyValues = getKeyValuesMock.Execute
		doguGetKeyValues = getKeyValuesMock.Execute

		createDecryptFunc = func(dogu string, getGCfg decrypt.GetGlobalConfigFunc) (decrypt.Decrypter, error) {
			return decrypt.Decrypter{}, nil
		}

		_, err = GetConfig(usermgtDogu, []string{})
		assert.ErrorIs(t, err, assert.AnError)
	})
}
