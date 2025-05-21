package etcd

import (
	"encoding/json"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/cloudogu/ces-exporter/etcd/mocks"
	cesLibCore "github.com/cloudogu/cesapp-lib/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.etcd.io/etcd/client/v2"
	"os"
	"path"
	"slices"
	"testing"
)

func TestGetDoguSpec(t *testing.T) {
	// fake once call
	clientOnce.Do(func() {})

	const usermgtDogu = "usermgt"

	t.Run("GetDoguSpec from Usermgt", func(t *testing.T) {
		doguJson, err := os.ReadFile("testdata/dogu.json")
		require.NoError(t, err)

		var expDoguSpec cesLibCore.Dogu
		err = json.Unmarshal(doguJson, &expDoguSpec)
		require.NoError(t, err)

		oldDoguGetKeyValues := doguGetKeyValues
		defer func() {
			doguGetKeyValues = oldDoguGetKeyValues
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

		doguGetKeyValues = getKeyValuesMock.Execute

		doguSpec, err := GetDoguSpec(usermgtDogu)
		assert.NoError(t, err)
		assert.Equal(t, expDoguSpec, doguSpec)
	})

	t.Run("current dogu not found", func(t *testing.T) {
		oldDoguGetKeyValues := doguGetKeyValues
		defer func() {
			doguGetKeyValues = oldDoguGetKeyValues
		}()

		getKeyValuesMock := mocks.NewGetKeyValuesClientFuncType(t)
		getKeyValuesMock.EXPECT().Execute(mock.Anything).Return(nil, client.Error{
			Code: client.ErrorCodeKeyNotFound,
		})

		doguGetKeyValues = getKeyValuesMock.Execute

		_, err := GetDoguSpec(usermgtDogu)
		assert.ErrorIs(t, err, ErrDoguNotFound)
	})

	t.Run("etcd client returns error", func(t *testing.T) {
		oldDoguGetKeyValues := doguGetKeyValues
		defer func() {
			doguGetKeyValues = oldDoguGetKeyValues
		}()

		getKeyValuesMock := mocks.NewGetKeyValuesClientFuncType(t)
		getKeyValuesMock.EXPECT().Execute(mock.Anything).Return(nil, assert.AnError)

		doguGetKeyValues = getKeyValuesMock.Execute

		_, err := GetDoguSpec(usermgtDogu)
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("error getting dogu in specific version", func(t *testing.T) {
		doguJson, err := os.ReadFile("testdata/dogu.json")
		require.NoError(t, err)

		var expDoguSpec cesLibCore.Dogu
		err = json.Unmarshal(doguJson, &expDoguSpec)
		require.NoError(t, err)

		oldDoguGetKeyValues := doguGetKeyValues
		defer func() {
			doguGetKeyValues = oldDoguGetKeyValues
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
		getKeyValuesMock.EXPECT().Execute(doguSpecPath).Return(nil, assert.AnError)

		doguGetKeyValues = getKeyValuesMock.Execute

		_, err = GetDoguSpec(usermgtDogu)
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("error unmarshalling dogu json", func(t *testing.T) {
		doguJson, err := os.ReadFile("testdata/dogu.json")
		require.NoError(t, err)

		var expDoguSpec cesLibCore.Dogu
		err = json.Unmarshal(doguJson, &expDoguSpec)
		require.NoError(t, err)

		oldDoguGetKeyValues := doguGetKeyValues
		defer func() {
			doguGetKeyValues = oldDoguGetKeyValues
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
				Value: "invalid",
			},
		}, nil)

		doguGetKeyValues = getKeyValuesMock.Execute

		_, err = GetDoguSpec(usermgtDogu)
		assert.Error(t, err)
	})
}

func TestGetAllDogus(t *testing.T) {
	// fake once call
	clientOnce.Do(func() {})

	t.Run("get all dogus", func(t *testing.T) {
		oldDoguGetKeyValues := doguGetKeyValues
		defer func() {
			doguGetKeyValues = oldDoguGetKeyValues
		}()

		keyValuesFuncMock := mocks.NewGetKeyValuesClientFuncType(t)
		keyValuesFuncMock.EXPECT().Execute(mock.Anything).Return([]core.KeyValue{
			{
				Key:   "/usermgt/current",
				Value: "1.1.1",
			},
			{
				Key:   "/usermgt/1.1.1",
				Value: "",
			},
			{
				Key:   "/admin/current",
				Value: "1.1.1",
			},
			{
				Key:   "/admin/1.1.1",
				Value: "",
			},
			{
				Key:   "/confluence/current",
				Value: "1.1.1",
			},
			{
				Key:   "/confluence/1.1.1",
				Value: "",
			},
			{
				Key:   "/backup/current",
				Value: "1.1.1",
			},
			{
				Key:   "/backup/1.1.1",
				Value: "",
			},
		}, nil)

		doguGetKeyValues = keyValuesFuncMock.Execute

		doguList, err := GetAllDogus()
		assert.NoError(t, err)
		assert.Len(t, doguList, 4)

		expDoguList := []string{"usermgt", "admin", "confluence", "backup"}

		for _, expDogu := range expDoguList {
			assert.True(t, slices.Contains(doguList, expDogu))
		}
	})

	t.Run("getKeyValues returns error", func(t *testing.T) {
		oldDoguGetKeyValues := doguGetKeyValues
		defer func() {
			doguGetKeyValues = oldDoguGetKeyValues
		}()

		keyValuesFuncMock := mocks.NewGetKeyValuesClientFuncType(t)
		keyValuesFuncMock.EXPECT().Execute(mock.Anything).Return(nil, assert.AnError)

		doguGetKeyValues = keyValuesFuncMock.Execute

		doguList, err := GetAllDogus()
		assert.ErrorIs(t, err, assert.AnError)
		assert.Len(t, doguList, 0)
	})
}

func Test_getDoguConfigKeySet(t *testing.T) {
	// fake once call
	clientOnce.Do(func() {})

	const usermgtDogu = "usermgt"

	t.Run("return config key set from dogu.json", func(t *testing.T) {
		doguJson, err := os.ReadFile("testdata/dogu.json")
		require.NoError(t, err)

		var expDoguSpec cesLibCore.Dogu
		err = json.Unmarshal(doguJson, &expDoguSpec)
		require.NoError(t, err)

		oldDoguGetKeyValues := doguGetKeyValues
		defer func() {
			doguGetKeyValues = oldDoguGetKeyValues
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

		doguGetKeyValues = getKeyValuesMock.Execute

		doguConfigKeySet, err := getDoguConfigKeySet(usermgtDogu)
		assert.NoError(t, err)

		assert.Len(t, doguConfigKeySet, len(expDoguSpec.Configuration))

		for _, config := range expDoguSpec.Configuration {
			included, ok := doguConfigKeySet[config.Name]
			assert.True(t, ok)
			assert.True(t, included)
		}
	})

	t.Run("exclude global config from dogu.json", func(t *testing.T) {
		doguJson, err := os.ReadFile("testdata/dogu.json")
		require.NoError(t, err)

		var expDoguSpec cesLibCore.Dogu
		err = json.Unmarshal(doguJson, &expDoguSpec)
		require.NoError(t, err)

		oldDoguGetKeyValues := doguGetKeyValues
		defer func() {
			doguGetKeyValues = oldDoguGetKeyValues
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

		doguGetKeyValues = getKeyValuesMock.Execute

		doguConfigKeySet, err := getDoguConfigKeySet(usermgtDogu, ExcludeGlobalConfig())
		assert.NoError(t, err)

		assert.Len(t, doguConfigKeySet, len(expDoguSpec.Configuration))

		for _, config := range expDoguSpec.Configuration {
			included, ok := doguConfigKeySet[config.Name]
			assert.True(t, ok)
			if config.Name == "mail/sender" {
				assert.False(t, included)
			} else {
				assert.True(t, included)
			}
		}
	})

	t.Run("GetDoguSpec returns error", func(t *testing.T) {
		oldDoguGetKeyValues := doguGetKeyValues
		defer func() {
			doguGetKeyValues = oldDoguGetKeyValues
		}()

		keyValuesFuncMock := mocks.NewGetKeyValuesClientFuncType(t)
		keyValuesFuncMock.EXPECT().Execute(mock.Anything).Return(nil, assert.AnError)

		doguGetKeyValues = keyValuesFuncMock.Execute

		doguList, err := getDoguConfigKeySet(usermgtDogu)
		assert.ErrorIs(t, err, assert.AnError)
		assert.Len(t, doguList, 0)
	})
}
