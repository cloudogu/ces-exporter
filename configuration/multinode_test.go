package configuration

import (
	"context"
	"fmt"
	"github.com/cloudogu/ces-commons-lib/dogu"
	"github.com/cloudogu/ces-exporter/core"
	v1 "github.com/cloudogu/k8s-backup-operator/pkg/api/v1"
	"github.com/cloudogu/k8s-registry-lib/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestGetGlobalConfigs(t *testing.T) {
	t.Run("get global configs successfully", func(t *testing.T) {
		globalConfigRepo := newMockGlobalConfigRepository(t)
		globalConfigRepo.EXPECT().Get(mock.Anything).Return(config.CreateGlobalConfig(map[config.Key]config.Value{
			"key": "value",
		}), nil)
		provider := NewMultinodeConfigurationProvider("", nil, nil, globalConfigRepo, nil, nil)
		configs, err := provider.getGlobalConfigs(context.TODO())
		assert.NoError(t, err)
		assert.Equal(t, []core.KeyValue{
			{
				Key:   "key",
				Value: "value",
			},
		}, configs)
	})
	t.Run("fail on get global configs", func(t *testing.T) {
		globalConfigRepo := newMockGlobalConfigRepository(t)
		globalConfigRepo.EXPECT().Get(mock.Anything).Return(config.GlobalConfig{}, fmt.Errorf("testerror"))
		provider := NewMultinodeConfigurationProvider("", nil, nil, globalConfigRepo, nil, nil)
		configs, err := provider.getGlobalConfigs(context.TODO())
		assert.Errorf(t, err, "asd")
		assert.Equal(t, 0, len(configs))
	})
}

func TestGetDoguConfigs(t *testing.T) {
	t.Run("get dogu configs successfully", func(t *testing.T) {
		dc := newMockDoguConfigRepository(t)
		dc.EXPECT().Get(mock.Anything, mock.Anything).Return(config.CreateDoguConfig("d1", map[config.Key]config.Value{
			"a": "b",
		}), nil)
		sc := newMockDoguConfigRepository(t)
		sc.EXPECT().Get(mock.Anything, mock.Anything).Return(config.CreateDoguConfig("d1", map[config.Key]config.Value{
			"e": "f",
		}), nil)
		dvc := newMockDoguVersionRegistry(t)
		dvc.EXPECT().GetCurrentOfAll(mock.Anything).Return([]dogu.SimpleNameVersion{{
			Name: "d1",
		}}, nil)
		provider := NewMultinodeConfigurationProvider("", sc, dc, nil, dvc, nil)
		configs, err := provider.getDoguConfigs(context.TODO())
		assert.NoError(t, err)
		assert.Equal(t, []core.DoguConfig{
			{
				Name: "d1",
				NormalConfig: []core.KeyValue{
					{
						Key:   "a",
						Value: "b",
					},
				},
				LocalConfig: []core.KeyValue{},
				SensitiveConfig: []core.KeyValue{
					{
						Key:   "e",
						Value: "f",
					},
				},
			},
		}, configs)
	})

	t.Run("fail on get sensitive dogu config", func(t *testing.T) {
		dc := newMockDoguConfigRepository(t)
		dc.EXPECT().Get(mock.Anything, mock.Anything).Return(config.CreateDoguConfig("d1", map[config.Key]config.Value{
			"a": "b",
		}), nil)
		sc := newMockDoguConfigRepository(t)
		sc.EXPECT().Get(mock.Anything, mock.Anything).Return(config.DoguConfig{}, fmt.Errorf("testerror"))
		dvc := newMockDoguVersionRegistry(t)
		dvc.EXPECT().GetCurrentOfAll(mock.Anything).Return([]dogu.SimpleNameVersion{{
			Name: "d1",
		}}, nil)
		provider := NewMultinodeConfigurationProvider("", sc, dc, nil, dvc, nil)
		configs, err := provider.getDoguConfigs(context.TODO())
		assert.Errorf(t, err, "failed to get sensitive dogu config for dogu d1: could not get config for "+
			"dogu d1: unable to get data 'd1-config' from cluster: unable to list config-map from cluster: testerror")
		assert.Equal(t, 0, len(configs))
	})

	t.Run("fail on get dogu config", func(t *testing.T) {
		dc := newMockDoguConfigRepository(t)
		dc.EXPECT().Get(mock.Anything, mock.Anything).Return(config.DoguConfig{}, fmt.Errorf("testerror"))
		dvc := newMockDoguVersionRegistry(t)
		dvc.EXPECT().GetCurrentOfAll(mock.Anything).Return([]dogu.SimpleNameVersion{{
			Name: "d1",
		}}, nil)
		provider := NewMultinodeConfigurationProvider("", nil, dc, nil, dvc, nil)
		configs, err := provider.getDoguConfigs(context.TODO())
		assert.Error(t, err)
		assert.Equal(t, "failed to get dogu config for dogu d1: testerror", err.Error())
		assert.Equal(t, 0, len(configs))
	})

	t.Run("fail on get installed dogus", func(t *testing.T) {
		dvc := newMockDoguVersionRegistry(t)
		dvc.EXPECT().GetCurrentOfAll(mock.Anything).Return([]dogu.SimpleNameVersion{}, fmt.Errorf("testerror"))
		provider := NewMultinodeConfigurationProvider("", nil, nil, nil, dvc, nil)
		configs, err := provider.getDoguConfigs(context.TODO())
		assert.Error(t, err)
		assert.Equal(t, "failed to get installed dogus: testerror", err.Error())
		assert.Equal(t, 0, len(configs))
	})
}

func TestGetBackupSchedules(t *testing.T) {
	t.Run("get backup schedules successfully", func(t *testing.T) {
		client := newMockBackupScheduleRuntimeClient(t)
		client.EXPECT().ListBackupSchedules(mock.Anything).Return(&v1.BackupScheduleList{
			Items: []v1.BackupSchedule{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "schedule",
					},
					Spec: v1.BackupScheduleSpec{
						Schedule: "12345",
					},
				},
			},
		}, nil)
		provider := NewMultinodeConfigurationProvider("", nil, nil, nil, nil, client)
		configs, err := provider.getBackupSchedules(context.TODO())
		assert.NoError(t, err)
		require.Equal(t, 1, len(configs))
		assert.Equal(t, core.BackupSchedule{
			Name:     "schedule",
			Schedule: "12345",
		}, configs[0])
	})

	t.Run("fail on list schedules", func(t *testing.T) {
		client := newMockBackupScheduleRuntimeClient(t)
		client.EXPECT().ListBackupSchedules(mock.Anything).Return(nil, fmt.Errorf("testerror"))
		provider := NewMultinodeConfigurationProvider("", nil, nil, nil, nil, client)
		configs, err := provider.getBackupSchedules(context.TODO())
		assert.Error(t, err)
		assert.Equal(t, "failed to list backup schedules: testerror", err.Error())
		require.Equal(t, 0, len(configs))
	})
}
