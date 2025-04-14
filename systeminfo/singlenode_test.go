package systeminfo

import (
	"github.com/cloudogu/ces-exporter/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestSingleNodeSystemInfoProvider_getFqdn(t *testing.T) {
	t.Run("return fqdn from global config", func(t *testing.T) {
		expFqdn := "testFqdn"

		globalConfigMock := newMockGetGlobalConfigFunc(t)
		globalConfigMock.EXPECT().Execute(mock.Anything).Return([]core.KeyValue{
			{
				Key:   globalConfigKeyFqdn,
				Value: expFqdn,
			},
		}, nil)

		sp := SingleNodeSystemInfoProvider{
			getGlobalConfig: globalConfigMock.Execute,
		}

		fqdn, err := sp.getFqdn()
		assert.NoError(t, err)
		assert.Equal(t, expFqdn, fqdn)
	})

	t.Run("error while getting global config", func(t *testing.T) {
		globalConfigMock := newMockGetGlobalConfigFunc(t)
		globalConfigMock.EXPECT().Execute(mock.Anything).Return(nil, assert.AnError)

		sp := SingleNodeSystemInfoProvider{
			getGlobalConfig: globalConfigMock.Execute,
		}

		_, err := sp.getFqdn()
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("error while looking up key for fqdn", func(t *testing.T) {
		globalConfigMock := newMockGetGlobalConfigFunc(t)
		globalConfigMock.EXPECT().Execute(mock.Anything).Return([]core.KeyValue{
			{
				Key:   "unknownFqdn",
				Value: "wrongFqdn",
			},
		}, nil)

		sp := SingleNodeSystemInfoProvider{
			getGlobalConfig: globalConfigMock.Execute,
		}

		_, err := sp.getFqdn()
		assert.ErrorIs(t, err, errFqdnGlobalConfigKeyNotFound)
	})

}
