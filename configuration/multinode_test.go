package configuration

import (
	"context"
	"fmt"
	v1 "github.com/cloudogu/k8s-backup-operator/pkg/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestGetGlobalConfigs(t *testing.T) {
	t.Run("get global configs successfully", func(t *testing.T) {
		configMaps := newMockConfigMapsInterface(t)
		configMaps.EXPECT().List(mock.Anything, mock.Anything).Return(&corev1.ConfigMapList{
			Items: []corev1.ConfigMap{
				{Data: map[string]string{
					"config.yaml": "key: value",
				}},
			},
		}, nil)
		provider := NewMultinodeConfigurationProvider("", configMaps, nil, nil)
		configs, err := provider.getGlobalConfigs(context.TODO())
		assert.NoError(t, err)
		assert.Equal(t, []keyValue{
			{
				Key:   "key",
				Value: "value",
			},
		}, configs)
	})
	t.Run("fail on get global configs", func(t *testing.T) {
		configMaps := newMockConfigMapsInterface(t)
		configMaps.EXPECT().List(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("testerror"))
		provider := NewMultinodeConfigurationProvider("", configMaps, nil, nil)
		configs, err := provider.getGlobalConfigs(context.TODO())
		assert.Errorf(t, err, "asd")
		assert.Equal(t, 0, len(configs))
	})
}

func TestGetDoguConfigs(t *testing.T) {
	t.Run("get dogu configs successfully", func(t *testing.T) {
		configMaps := newMockConfigMapsInterface(t)
		configMaps.EXPECT().List(mock.Anything, mock.Anything).Return(&corev1.ConfigMapList{
			Items: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"dogu.name": "d1",
						},
					},
					Data: map[string]string{
						"current":     "1.0.0-1",
						"config.yaml": "a: b",
					},
				},
			},
		}, nil)
		configSecrets := newMockSecretsInterface(t)
		configSecrets.EXPECT().List(mock.Anything, mock.Anything).Return(&corev1.SecretList{
			Items: []corev1.Secret{
				{
					Data: map[string][]byte{
						"config.yaml": []byte("e: f"),
					},
				},
			},
		}, nil)
		provider := NewMultinodeConfigurationProvider("", configMaps, configSecrets, nil)
		configs, err := provider.getDoguConfigs(context.TODO())
		assert.NoError(t, err)
		assert.Equal(t, []doguConfig{
			{
				Name: "d1",
				NormalConfig: []keyValue{
					{
						Key:   "a",
						Value: "b",
					},
				},
				LocalConfig: []keyValue{},
				SensitiveConfig: []keyValue{
					{
						Key:   "e",
						Value: "f",
					},
				},
			},
		}, configs)
	})

	t.Run("fail on get sensitive dogu config", func(t *testing.T) {
		configMaps := newMockConfigMapsInterface(t)
		configMaps.EXPECT().List(mock.Anything, mock.Anything).Return(&corev1.ConfigMapList{
			Items: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"dogu.name": "d1",
						},
					},
					Data: map[string]string{
						"current":     "1.0.0-1",
						"config.yaml": "a: b",
					},
				},
			},
		}, nil)
		configSecrets := newMockSecretsInterface(t)
		configSecrets.EXPECT().List(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("testerror"))
		provider := NewMultinodeConfigurationProvider("", configMaps, configSecrets, nil)
		configs, err := provider.getDoguConfigs(context.TODO())
		assert.Errorf(t, err, "failed to get sensitive dogu config for dogu d1: could not get config for "+
			"dogu d1: unable to get data 'd1-config' from cluster: unable to list config-map from cluster: testerror")
		assert.Equal(t, 0, len(configs))
	})

	t.Run("fail on get dogu config", func(t *testing.T) {
		configMaps := newMockConfigMapsInterface(t)
		configMaps.EXPECT().List(mock.Anything, mock.Anything).Return(&corev1.ConfigMapList{
			Items: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"dogu.name": "d1",
						},
					},
					Data: map[string]string{
						"current":     "1.0.0-1",
						"config.yaml": "a: b",
					},
				},
			},
		}, nil).Once()
		configMaps.EXPECT().List(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("testerror"))
		configSecrets := newMockSecretsInterface(t)
		provider := NewMultinodeConfigurationProvider("", configMaps, configSecrets, nil)
		configs, err := provider.getDoguConfigs(context.TODO())
		assert.Error(t, err)
		assert.Equal(t, "failed to get dogu config for dogu d1: could not get config for dogu d1: unable to get data "+
			"'d1-config' from cluster: unable to list config-map from cluster: testerror", err.Error())
		assert.Equal(t, 0, len(configs))
	})

	t.Run("fail on get installed dogus", func(t *testing.T) {
		configMaps := newMockConfigMapsInterface(t)
		configMaps.EXPECT().List(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("testerror"))
		configSecrets := newMockSecretsInterface(t)
		provider := NewMultinodeConfigurationProvider("", configMaps, configSecrets, nil)
		configs, err := provider.getDoguConfigs(context.TODO())
		assert.Error(t, err)
		assert.Equal(t, "failed to get installed dogus: failed to get all cluster native"+
			" local dogu registries: testerror", err.Error())
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
		provider := NewMultinodeConfigurationProvider("", nil, nil, client)
		configs, err := provider.getBackupSchedules(context.TODO())
		assert.NoError(t, err)
		require.Equal(t, 1, len(configs))
		assert.Equal(t, backupSchedule{
			Name:     "schedule",
			Schedule: "12345",
		}, configs[0])
	})

	t.Run("fail on list schedules", func(t *testing.T) {
		client := newMockBackupScheduleRuntimeClient(t)
		client.EXPECT().ListBackupSchedules(mock.Anything).Return(nil, fmt.Errorf("testerror"))
		provider := NewMultinodeConfigurationProvider("", nil, nil, client)
		configs, err := provider.getBackupSchedules(context.TODO())
		assert.Error(t, err)
		assert.Equal(t, "failed to list backup schedules: testerror", err.Error())
		require.Equal(t, 0, len(configs))
	})
}
