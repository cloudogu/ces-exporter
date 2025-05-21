package configuration

import (
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetBackupSchedulesClassic(t *testing.T) {
	t.Run("will work for valid time", func(t *testing.T) {
		provider := NewClassicProvider()
		provider.backupScheduleProvider = newBackupScheduleProvider(func(dogu string, ignoreKeys []string) (core.DoguConfig, error) {
			return core.DoguConfig{
				NormalConfig: []core.KeyValue{
					{
						Key:   "/time",
						Value: "13:30",
					},
				},
			}, nil
		})

		schedules, err := provider.getBackupSchedules(nil)
		require.NoError(t, err)

		require.Equal(t, 1, len(schedules))
		assert.Equal(t, "30 13 * * *", schedules[0].Schedule)
		assert.Equal(t, scheduledBackupName, schedules[0].Name)
	})
	t.Run("will return empty array on invalid time", func(t *testing.T) {
		provider := NewClassicProvider()
		provider.backupScheduleProvider = newBackupScheduleProvider(func(dogu string, ignoreKeys []string) (core.DoguConfig, error) {
			return core.DoguConfig{
				NormalConfig: []core.KeyValue{
					{
						Key:   "/time",
						Value: "1330",
					},
				},
			}, nil
		})

		schedules, err := provider.getBackupSchedules(nil)
		require.NoError(t, err)

		require.Equal(t, 0, len(schedules))
	})
	t.Run("will return error if error occur on get config", func(t *testing.T) {
		provider := NewClassicProvider()
		provider.backupScheduleProvider = newBackupScheduleProvider(func(dogu string, ignoreKeys []string) (core.DoguConfig, error) {
			return core.DoguConfig{}, fmt.Errorf("testerror")
		})

		schedules, err := provider.getBackupSchedules(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "testerror")
		assert.Equal(t, 0, len(schedules))
	})
}

func TestGetGlobalConfigsClassic(t *testing.T) {
	t.Run("will return things", func(t *testing.T) {
		returnValue := []core.KeyValue{
			{
				Key:   "a",
				Value: "b",
			},
		}
		provider := NewClassicProvider()
		provider.getGlobalConfig = func(ignoreKeys []string) (core.GlobalConfig, error) {
			assert.Equal(t, skippedKeys["global"], ignoreKeys)
			return returnValue, nil
		}

		configs, err := provider.getGlobalConfigs(nil)
		assert.Nil(t, err)
		assert.Equal(t, returnValue, configs)
	})
}

func TestGetDoguConfigsClassic(t *testing.T) {
	t.Run("runs successful", func(t *testing.T) {
		mockGetAllFunc := newMockGetAllDogusFunc(t)
		mockGetAllFunc.EXPECT().Execute().Return([]string{"d1", "d2"}, nil)

		mockNormalConfigFunc := newMockGetNormalConfigFunc(t)
		mockNormalConfigFunc.EXPECT().Execute("d1", skippedKeys["*"]).Return([]core.KeyValue{{Key: "a1", Value: "b1"}}, nil)
		mockNormalConfigFunc.EXPECT().Execute("d2", skippedKeys["*"]).Return([]core.KeyValue{{Key: "a2", Value: "b2"}}, nil)

		mockGetSensitiveFunc := newMockGetSensitiveConfigFunc(t)
		mockGetSensitiveFunc.EXPECT().Execute("d1", skippedKeys["*"]).Return([]core.KeyValue{{Key: "a1", Value: "b1"}}, nil)
		mockGetSensitiveFunc.EXPECT().Execute("d2", skippedKeys["*"]).Return([]core.KeyValue{{Key: "a2", Value: "b2"}}, nil)

		mockLocalConfigFunc := newMockGetLocalConfigFunc(t)
		mockLocalConfigFunc.EXPECT().Execute("d1", skippedKeys["*"]).Return([]core.KeyValue{{Key: "a1", Value: "b1"}}, nil)
		mockLocalConfigFunc.EXPECT().Execute("d2", skippedKeys["*"]).Return([]core.KeyValue{{Key: "a2", Value: "b2"}}, nil)

		provider := NewClassicProvider()
		provider.getDogus = func() ([]string, error) {
			return mockGetAllFunc.Execute()
		}

		provider.getNormalConfig = func(dogu string, ignoreKeys []string) ([]core.KeyValue, error) {
			return mockNormalConfigFunc.Execute(dogu, ignoreKeys)
		}

		provider.getSensitiveConfig = func(dogu string, ignoreKeys []string) ([]core.KeyValue, error) {
			return mockGetSensitiveFunc.Execute(dogu, ignoreKeys)
		}

		provider.getLocalConfig = func(dogu string, ignoreKeys []string) ([]core.KeyValue, error) {
			return mockLocalConfigFunc.Execute(dogu, ignoreKeys)
		}

		r, err := provider.getDoguConfigs(nil)
		require.NoError(t, err)
		require.Len(t, r, 2)

		assert.Equal(t, "d1", r[0].Name)
		assert.Len(t, r[0].NormalConfig, 1)
		assert.Equal(t, "a1", r[0].NormalConfig[0].Key)
		assert.Equal(t, "b1", r[0].NormalConfig[0].Value)
		assert.Len(t, r[0].LocalConfig, 1)
		assert.Equal(t, "a1", r[0].LocalConfig[0].Key)
		assert.Equal(t, "b1", r[0].LocalConfig[0].Value)
		assert.Len(t, r[0].SensitiveConfig, 1)
		assert.Equal(t, "a1", r[0].SensitiveConfig[0].Key)
		assert.Equal(t, "b1", r[0].SensitiveConfig[0].Value)

		assert.Equal(t, "d2", r[1].Name)
		assert.Len(t, r[1].NormalConfig, 1)
		assert.Equal(t, "a2", r[1].NormalConfig[0].Key)
		assert.Equal(t, "b2", r[1].NormalConfig[0].Value)
		assert.Len(t, r[1].LocalConfig, 1)
		assert.Equal(t, "a2", r[1].LocalConfig[0].Key)
		assert.Equal(t, "b2", r[1].LocalConfig[0].Value)
		assert.Len(t, r[1].SensitiveConfig, 1)
		assert.Equal(t, "a2", r[1].SensitiveConfig[0].Key)
		assert.Equal(t, "b2", r[1].SensitiveConfig[0].Value)
	})

	t.Run("fails on local config", func(t *testing.T) {
		mockGetAllFunc := newMockGetAllDogusFunc(t)
		mockGetAllFunc.EXPECT().Execute().Return([]string{"d1", "d2"}, nil)

		mockNormalConfigFunc := newMockGetNormalConfigFunc(t)
		mockNormalConfigFunc.EXPECT().Execute("d1", skippedKeys["*"]).Return([]core.KeyValue{{Key: "a1", Value: "b1"}}, nil)

		mockGetSensitiveFunc := newMockGetSensitiveConfigFunc(t)
		mockGetSensitiveFunc.EXPECT().Execute("d1", skippedKeys["*"]).Return([]core.KeyValue{{Key: "a1", Value: "b1"}}, nil)

		mockLocalConfigFunc := newMockGetLocalConfigFunc(t)
		mockLocalConfigFunc.EXPECT().Execute("d1", skippedKeys["*"]).Return([]core.KeyValue{}, fmt.Errorf("testerror"))

		provider := NewClassicProvider()
		provider.getDogus = func() ([]string, error) {
			return mockGetAllFunc.Execute()
		}

		provider.getNormalConfig = func(dogu string, ignoreKeys []string) ([]core.KeyValue, error) {
			return mockNormalConfigFunc.Execute(dogu, ignoreKeys)
		}

		provider.getSensitiveConfig = func(dogu string, ignoreKeys []string) ([]core.KeyValue, error) {
			return mockGetSensitiveFunc.Execute(dogu, ignoreKeys)
		}

		provider.getLocalConfig = func(dogu string, ignoreKeys []string) ([]core.KeyValue, error) {
			return mockLocalConfigFunc.Execute(dogu, ignoreKeys)
		}

		r, err := provider.getDoguConfigs(nil)
		require.Error(t, err)
		assert.Len(t, r, 0)
		assert.Contains(t, err.Error(), "failed to get local config of dogu d1")
		assert.Contains(t, err.Error(), "testerror")
	})

	t.Run("fails on sensitive config", func(t *testing.T) {
		mockGetAllFunc := newMockGetAllDogusFunc(t)
		mockGetAllFunc.EXPECT().Execute().Return([]string{"d1", "d2"}, nil)

		mockNormalConfigFunc := newMockGetNormalConfigFunc(t)
		mockNormalConfigFunc.EXPECT().Execute("d1", skippedKeys["*"]).Return([]core.KeyValue{{Key: "a1", Value: "b1"}}, nil)

		mockGetSensitiveFunc := newMockGetSensitiveConfigFunc(t)
		mockGetSensitiveFunc.EXPECT().Execute("d1", skippedKeys["*"]).Return([]core.KeyValue{}, fmt.Errorf("testerror"))

		mockLocalConfigFunc := newMockGetLocalConfigFunc(t)

		provider := NewClassicProvider()
		provider.getDogus = func() ([]string, error) {
			return mockGetAllFunc.Execute()
		}

		provider.getNormalConfig = func(dogu string, ignoreKeys []string) ([]core.KeyValue, error) {
			return mockNormalConfigFunc.Execute(dogu, ignoreKeys)
		}

		provider.getSensitiveConfig = func(dogu string, ignoreKeys []string) ([]core.KeyValue, error) {
			return mockGetSensitiveFunc.Execute(dogu, ignoreKeys)
		}

		provider.getLocalConfig = func(dogu string, ignoreKeys []string) ([]core.KeyValue, error) {
			return mockLocalConfigFunc.Execute(dogu, ignoreKeys)
		}

		r, err := provider.getDoguConfigs(nil)
		require.Error(t, err)
		assert.Len(t, r, 0)
		assert.Contains(t, err.Error(), "failed to get sensitive config of dogu d1")
		assert.Contains(t, err.Error(), "testerror")
	})

	t.Run("fails on sensitive config", func(t *testing.T) {
		mockGetAllFunc := newMockGetAllDogusFunc(t)
		mockGetAllFunc.EXPECT().Execute().Return([]string{"d1", "d2"}, nil)

		mockNormalConfigFunc := newMockGetNormalConfigFunc(t)
		mockNormalConfigFunc.EXPECT().Execute("d1", skippedKeys["*"]).Return([]core.KeyValue{}, fmt.Errorf("testerror"))

		mockGetSensitiveFunc := newMockGetSensitiveConfigFunc(t)
		mockLocalConfigFunc := newMockGetLocalConfigFunc(t)

		provider := NewClassicProvider()
		provider.getDogus = func() ([]string, error) {
			return mockGetAllFunc.Execute()
		}

		provider.getNormalConfig = func(dogu string, ignoreKeys []string) ([]core.KeyValue, error) {
			return mockNormalConfigFunc.Execute(dogu, ignoreKeys)
		}

		provider.getSensitiveConfig = func(dogu string, ignoreKeys []string) ([]core.KeyValue, error) {
			return mockGetSensitiveFunc.Execute(dogu, ignoreKeys)
		}

		provider.getLocalConfig = func(dogu string, ignoreKeys []string) ([]core.KeyValue, error) {
			return mockLocalConfigFunc.Execute(dogu, ignoreKeys)
		}

		r, err := provider.getDoguConfigs(nil)
		require.Error(t, err)
		assert.Len(t, r, 0)
		assert.Contains(t, err.Error(), "failed to get normal config of dogu d1")
		assert.Contains(t, err.Error(), "testerror")
	})

	t.Run("fail on get all dogus", func(t *testing.T) {
		mockGetAllFunc := newMockGetAllDogusFunc(t)
		mockGetAllFunc.EXPECT().Execute().Return([]string{"d1", "d2"}, fmt.Errorf("testerror"))

		mockNormalConfigFunc := newMockGetNormalConfigFunc(t)

		mockGetSensitiveFunc := newMockGetSensitiveConfigFunc(t)
		mockLocalConfigFunc := newMockGetLocalConfigFunc(t)

		provider := NewClassicProvider()
		provider.getDogus = func() ([]string, error) {
			return mockGetAllFunc.Execute()
		}

		provider.getNormalConfig = func(dogu string, ignoreKeys []string) ([]core.KeyValue, error) {
			return mockNormalConfigFunc.Execute(dogu, ignoreKeys)
		}

		provider.getSensitiveConfig = func(dogu string, ignoreKeys []string) ([]core.KeyValue, error) {
			return mockGetSensitiveFunc.Execute(dogu, ignoreKeys)
		}

		provider.getLocalConfig = func(dogu string, ignoreKeys []string) ([]core.KeyValue, error) {
			return mockLocalConfigFunc.Execute(dogu, ignoreKeys)
		}

		r, err := provider.getDoguConfigs(nil)
		require.Error(t, err)
		assert.Len(t, r, 0)
		assert.Contains(t, err.Error(), "failed to get all dogus")
		assert.Contains(t, err.Error(), "testerror")
	})
}
