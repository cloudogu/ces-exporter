package systeminfo

import (
	"github.com/cloudogu/ces-exporter/core"
	cesLibCore "github.com/cloudogu/cesapp-lib/core"
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

func TestSingleNodeSystemInfoProvider_getDogus(t *testing.T) {
	t.Run("return current dogus", func(t *testing.T) {
		adminDoguSpec := cesLibCore.Dogu{
			Name:    "premium/admin",
			Version: "2.13.2-1",
		}

		usermgtDoguSpec := cesLibCore.Dogu{
			Name:    "usermgt/admin",
			Version: "1.20.0-4",
		}

		confluenceDoguSpec := cesLibCore.Dogu{
			Name:    "premium/confluence",
			Version: "8.5.19-1",
		}

		expDoguMap := map[string]cesLibCore.Dogu{
			adminDoguSpec.Name:      adminDoguSpec,
			usermgtDoguSpec.Name:    usermgtDoguSpec,
			confluenceDoguSpec.Name: confluenceDoguSpec,
		}

		doguGetterMock := newMockDoguGetter(t)

		doguGetterMock.EXPECT().GetAllDogus().Return([]string{"admin", "usermgt", "confluence"}, nil)
		doguGetterMock.EXPECT().GetDoguSpec("admin").Return(adminDoguSpec, nil)
		doguGetterMock.EXPECT().GetDoguSpec("usermgt").Return(usermgtDoguSpec, nil)
		doguGetterMock.EXPECT().GetDoguSpec("confluence").Return(confluenceDoguSpec, nil)

		sp := SingleNodeSystemInfoProvider{
			getGlobalConfig: nil,
			doguGetter:      doguGetterMock,
		}

		dogus, err := sp.getDogus()
		assert.NoError(t, err)
		assert.Equal(t, len(expDoguMap), len(dogus))

		for _, d := range dogus {
			expDoguSpec, ok := expDoguMap[d.Name]
			assert.True(t, ok)
			assert.Equal(t, expDoguSpec.Version, d.Version)
		}
	})

	t.Run("error getting all dogus currently installed", func(t *testing.T) {
		doguGetterMock := newMockDoguGetter(t)

		doguGetterMock.EXPECT().GetAllDogus().Return(nil, assert.AnError)

		sp := SingleNodeSystemInfoProvider{
			getGlobalConfig: nil,
			doguGetter:      doguGetterMock,
		}

		_, err := sp.getDogus()
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("error receiving dogu spec", func(t *testing.T) {
		adminDoguSpec := cesLibCore.Dogu{
			Name:    "premium/admin",
			Version: "2.13.2-1",
		}

		usermgtDoguSpec := cesLibCore.Dogu{
			Name:    "usermgt/admin",
			Version: "1.20.0-4",
		}

		doguGetterMock := newMockDoguGetter(t)

		doguGetterMock.EXPECT().GetAllDogus().Return([]string{"admin", "usermgt", "confluence"}, nil)
		doguGetterMock.EXPECT().GetDoguSpec("admin").Return(adminDoguSpec, nil)
		doguGetterMock.EXPECT().GetDoguSpec("usermgt").Return(usermgtDoguSpec, nil)
		doguGetterMock.EXPECT().GetDoguSpec("confluence").Return(cesLibCore.Dogu{}, assert.AnError)

		sp := SingleNodeSystemInfoProvider{
			getGlobalConfig: nil,
			doguGetter:      doguGetterMock,
		}

		_, err := sp.getDogus()
		assert.ErrorIs(t, err, assert.AnError)
	})
}
