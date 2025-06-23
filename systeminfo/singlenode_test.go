package systeminfo

import (
	"context"
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	cesLibCore "github.com/cloudogu/cesapp-lib/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestNewSingleNodeSystemInfoProvider(t *testing.T) {
	sp := NewSingleNodeSystemInfoProvider("testDir", 0)
	assert.NotNil(t, sp)
}

func TestSingleNodeSystemInfoProvider_getFqdn(t *testing.T) {
	t.Run("return fqdn from global config", func(t *testing.T) {
		expFqdn := "testFqdn"

		globalConfigMock := newMockGetGlobalConfig(t)
		globalConfigMock.EXPECT().Execute(mock.Anything).Return([]core.KeyValue{
			{
				Key:   globalConfigKeyFqdn,
				Value: expFqdn,
			},
		}, nil)

		sp := SingleNodeSystemInfoProvider{
			getGlobalConfig: globalConfigMock.Execute,
		}

		fqdn, err := sp.getFqdn(context.TODO())
		assert.NoError(t, err)
		assert.Equal(t, expFqdn, fqdn)
	})

	t.Run("error while getting global config", func(t *testing.T) {
		globalConfigMock := newMockGetGlobalConfig(t)
		globalConfigMock.EXPECT().Execute(mock.Anything).Return(nil, assert.AnError)

		sp := SingleNodeSystemInfoProvider{
			getGlobalConfig: globalConfigMock.Execute,
		}

		_, err := sp.getFqdn(context.TODO())
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("error while looking up key for fqdn", func(t *testing.T) {
		globalConfigMock := newMockGetGlobalConfig(t)
		globalConfigMock.EXPECT().Execute(mock.Anything).Return([]core.KeyValue{
			{
				Key:   "unknownFqdn",
				Value: "wrongFqdn",
			},
		}, nil)

		sp := SingleNodeSystemInfoProvider{
			getGlobalConfig: globalConfigMock.Execute,
		}

		_, err := sp.getFqdn(context.TODO())
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

		const staticVolumeSize = int64(100)
		const expStaticVolumeSize = int64(1073741824)

		doguGetterMock := newMockDoguGetter(t)

		doguGetterMock.EXPECT().GetAllDogus().Return([]string{"admin", "usermgt", "confluence"}, nil)
		doguGetterMock.EXPECT().GetDoguSpec("admin").Return(adminDoguSpec, nil)
		doguGetterMock.EXPECT().GetDoguSpec("usermgt").Return(usermgtDoguSpec, nil)
		doguGetterMock.EXPECT().GetDoguSpec("confluence").Return(confluenceDoguSpec, nil)

		volumeSizeGetterMock := newMockVolumeSizeGetter(t)
		volumeSizeGetterMock.EXPECT().GetVolumeSizeInBytes(mock.Anything).Return(staticVolumeSize, nil)

		sp := SingleNodeSystemInfoProvider{
			getGlobalConfig:  nil,
			doguGetter:       doguGetterMock,
			volumeSizeGetter: volumeSizeGetterMock,
		}

		dogus, err := sp.getDogus(context.TODO())
		assert.NoError(t, err)
		assert.Equal(t, len(expDoguMap), len(dogus))

		for _, d := range dogus {
			expDoguSpec, ok := expDoguMap[d.Name]
			assert.True(t, ok)
			assert.Equal(t, expDoguSpec.Version, d.Version)
			assert.Equal(t, expStaticVolumeSize, d.Volume.SizeInBytes)
		}
	})

	t.Run("error getting all dogus currently installed", func(t *testing.T) {
		doguGetterMock := newMockDoguGetter(t)

		doguGetterMock.EXPECT().GetAllDogus().Return(nil, assert.AnError)

		sp := SingleNodeSystemInfoProvider{
			getGlobalConfig: nil,
			doguGetter:      doguGetterMock,
		}

		_, err := sp.getDogus(context.TODO())
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

		const staticVolumeSize = int64(100)

		doguGetterMock := newMockDoguGetter(t)

		doguGetterMock.EXPECT().GetAllDogus().Return([]string{"admin", "usermgt", "confluence"}, nil)
		doguGetterMock.EXPECT().GetDoguSpec("admin").Return(adminDoguSpec, nil)
		doguGetterMock.EXPECT().GetDoguSpec("usermgt").Return(usermgtDoguSpec, nil)
		doguGetterMock.EXPECT().GetDoguSpec("confluence").Return(cesLibCore.Dogu{}, assert.AnError)

		volumeSizeGetterMock := newMockVolumeSizeGetter(t)
		volumeSizeGetterMock.EXPECT().GetVolumeSizeInBytes(mock.Anything).Return(staticVolumeSize, nil)

		sp := SingleNodeSystemInfoProvider{
			getGlobalConfig:  nil,
			doguGetter:       doguGetterMock,
			volumeSizeGetter: volumeSizeGetterMock,
		}

		_, err := sp.getDogus(context.TODO())
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestSingleNodeSystemInfoProvider_getComponents(t *testing.T) {
	sp := SingleNodeSystemInfoProvider{}

	components, err := sp.getComponents(context.TODO())
	assert.NoError(t, err)
	assert.NotNil(t, components)
	assert.Len(t, components, 0)
}

func TestSingleNodeSystemInfoProvider_isMultinode(t *testing.T) {
	sp := SingleNodeSystemInfoProvider{}
	assert.False(t, sp.isMultinode())
}

func TestSingleNodeVolumeSizeGetter_GetVolumeSizeInBytes(t *testing.T) {
	t.Run("return volume size in bytes", func(t *testing.T) {
		mockOutput := `Total   Exclusive  Set shared  Filename
158193729536  158193729536           0  /var/lib/ces/nexus`

		cmdExecutorMock := newMockCommandExecutor(t)
		cmdExecutorMock.EXPECT().Output().Return([]byte(mockOutput), nil)

		sp := singleNodeVolumeSizeGetter{
			volumeBasePath: "tesPath",
			execCommand: func(command string, args ...string) commandExecutor {
				return cmdExecutorMock
			},
		}

		sizeInBytes, err := sp.GetVolumeSizeInBytes("testDogu")
		assert.NoError(t, err)
		assert.Equal(t, int64(158193729536), sizeInBytes)
	})

	t.Run("error due to missing header line", func(t *testing.T) {
		mockOutput := `158193729536  158193729536  0  /var/lib/ces/nexus`

		cmdExecutorMock := newMockCommandExecutor(t)
		cmdExecutorMock.EXPECT().Output().Return([]byte(mockOutput), nil)

		sp := singleNodeVolumeSizeGetter{
			volumeBasePath: "tesPath",
			execCommand: func(command string, args ...string) commandExecutor {
				return cmdExecutorMock
			},
		}

		_, err := sp.GetVolumeSizeInBytes("testDogu")
		assert.Error(t, err)
	})

	t.Run("exec command returns error", func(t *testing.T) {
		cmdExecutorMock := newMockCommandExecutor(t)
		cmdExecutorMock.EXPECT().Output().Return(nil, assert.AnError)

		sp := singleNodeVolumeSizeGetter{
			volumeBasePath: "tesPath",
			execCommand: func(command string, args ...string) commandExecutor {
				return cmdExecutorMock
			},
		}

		_, err := sp.GetVolumeSizeInBytes("testDogu")
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("error due to empty line", func(t *testing.T) {
		mockOutput := `Total   Exclusive  Set shared  Filename
             
158193729536  158193729536           0  /var/lib/ces/nexus`

		cmdExecutorMock := newMockCommandExecutor(t)
		cmdExecutorMock.EXPECT().Output().Return([]byte(mockOutput), nil)

		sp := singleNodeVolumeSizeGetter{
			volumeBasePath: "tesPath",
			execCommand: func(command string, args ...string) commandExecutor {
				return cmdExecutorMock
			},
		}

		_, err := sp.GetVolumeSizeInBytes("testDogu")
		assert.Error(t, err)
	})

	t.Run("error parsing total size to int64", func(t *testing.T) {
		mockOutput := `Total   Exclusive  Set shared  Filename
badFormat  158193729536           0  /var/lib/ces/nexus`

		cmdExecutorMock := newMockCommandExecutor(t)
		cmdExecutorMock.EXPECT().Output().Return([]byte(mockOutput), nil)

		sp := singleNodeVolumeSizeGetter{
			volumeBasePath: "tesPath",
			execCommand: func(command string, args ...string) commandExecutor {
				return cmdExecutorMock
			},
		}

		_, err := sp.GetVolumeSizeInBytes("testDogu")
		assert.Error(t, err)
	})
}

func TestCalculateTargetVolumeSize(t *testing.T) {
	const GiB = 1024 * 1024 * 1024

	tests := []struct {
		inActualSize     int64
		inIncreateFactor float32
		expectedSize     int64
	}{
		{inActualSize: 0, inIncreateFactor: 1.3, expectedSize: 1 * GiB},
		{inActualSize: 100, inIncreateFactor: 1, expectedSize: 1 * GiB},
		{inActualSize: 100, inIncreateFactor: 1.3, expectedSize: 1 * GiB},
		{inActualSize: 1*GiB + 500, inIncreateFactor: 1, expectedSize: 2 * GiB},
		{inActualSize: 176 * GiB, inIncreateFactor: 1.8, expectedSize: 317 * GiB},
		{inActualSize: 2 * GiB, inIncreateFactor: 3.2, expectedSize: 7 * GiB},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%+v", tc), func(t *testing.T) {
			assert.Equal(t, tc.expectedSize, calculateTargetVolumeSize(tc.inIncreateFactor, tc.inActualSize))
		})
	}
}
