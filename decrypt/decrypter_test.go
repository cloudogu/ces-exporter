package decrypt

import (
	"github.com/cloudogu/ces-exporter/core"
	"github.com/cloudogu/ces-exporter/decrypt/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func Test_createKeyProvider(t *testing.T) {
	t.Run("create key provider", func(t *testing.T) {
		getGlobalCfgMock := mocks.NewGetGlobalConfigFunc(t)
		getGlobalCfgMock.EXPECT().Execute(mock.Anything).Return([]core.KeyValue{
			{
				Key:   keyProviderKey,
				Value: "pkcs1v15",
			},
		}, nil)

		provider, err := createKeyProvider(getGlobalCfgMock.Execute)

		assert.NoError(t, err)
		assert.NotNil(t, provider)
	})

	t.Run("error receiving global config", func(t *testing.T) {
		getGlobalCfgMock := mocks.NewGetGlobalConfigFunc(t)
		getGlobalCfgMock.EXPECT().Execute(mock.Anything).Return(core.GlobalConfig{}, assert.AnError)

		provider, err := createKeyProvider(getGlobalCfgMock.Execute)

		assert.ErrorIs(t, err, assert.AnError)
		assert.Nil(t, provider)
	})

	t.Run("global config does not contain key provider key", func(t *testing.T) {
		getGlobalCfgMock := mocks.NewGetGlobalConfigFunc(t)
		getGlobalCfgMock.EXPECT().Execute(mock.Anything).Return([]core.KeyValue{
			{
				Key:   "invalid",
				Value: "pkcs1v15",
			},
		}, nil)

		provider, err := createKeyProvider(getGlobalCfgMock.Execute)

		assert.Error(t, err)
		assert.Nil(t, provider)
	})

	t.Run("invalid keyProvider", func(t *testing.T) {
		getGlobalCfgMock := mocks.NewGetGlobalConfigFunc(t)
		getGlobalCfgMock.EXPECT().Execute(mock.Anything).Return([]core.KeyValue{
			{
				Key:   keyProviderKey,
				Value: "invalid",
			},
		}, nil)

		provider, err := createKeyProvider(getGlobalCfgMock.Execute)

		assert.Error(t, err)
		assert.Nil(t, provider)
	})
}

func TestGetKeyProvider(t *testing.T) {
	getGlobalCfgMock := mocks.NewGetGlobalConfigFunc(t)
	getGlobalCfgMock.EXPECT().Execute(mock.Anything).Return([]core.KeyValue{
		{
			Key:   keyProviderKey,
			Value: "pkcs1v15",
		},
	}, nil)

	provider, err := GetKeyProvider(getGlobalCfgMock.Execute)

	assert.NoError(t, err)
	assert.NotNil(t, provider)

	// ensure clientOnce is only called once
	keyProviderOnce.Do(func() {
		keyProvider = nil
		errKeyProvider = assert.AnError
	})

	assert.NoError(t, err)
	assert.NotNil(t, keyProvider)
}
