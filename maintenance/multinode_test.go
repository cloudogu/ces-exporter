package maintenance

import (
	"context"
	"fmt"
	"github.com/cloudogu/k8s-registry-lib/repository"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestMNGetMaintenanceMode(t *testing.T) {
	t.Run("should return maintenance mode is inactive", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)
		globalConfigRepo := repository.NewGlobalConfigRepository(configMaps)
		configMaps.EXPECT().List(mock.Anything, mock.Anything).Return(&corev1.ConfigMapList{
			Items: []corev1.ConfigMap{
				{Data: map[string]string{
					"config.yaml": "key: value",
				}},
			},
		}, nil)
		provider := NewMultinodeMaintenanceModeProvider(globalConfigRepo)

		status, err := provider.GetMaintenanceMode(context.TODO())
		require.NoError(t, err)
		require.Equal(t, status.IsActive, false)
	})

	t.Run("should return maintenance mode is active", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)
		globalConfigRepo := repository.NewGlobalConfigRepository(configMaps)
		configMaps.EXPECT().List(mock.Anything, mock.Anything).Return(&corev1.ConfigMapList{
			Items: []corev1.ConfigMap{
				{Data: map[string]string{
					"config.yaml": "maintenance: value",
				}},
			},
		}, nil)
		provider := NewMultinodeMaintenanceModeProvider(globalConfigRepo)

		status, err := provider.GetMaintenanceMode(context.TODO())
		require.NoError(t, err)
		require.Equal(t, status.IsActive, true)
	})

	t.Run("should fail on get global config", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)
		globalConfigRepo := repository.NewGlobalConfigRepository(configMaps)
		configMaps.EXPECT().List(mock.Anything, mock.Anything).Return(&corev1.ConfigMapList{
			Items: []corev1.ConfigMap{
				{Data: map[string]string{}},
			},
		}, nil)
		provider := NewMultinodeMaintenanceModeProvider(globalConfigRepo)

		_, err := provider.GetMaintenanceMode(context.TODO())
		require.Contains(t, err.Error(), "failed to get global config:")
	})
}

func TestMNBuildMaintenanceJSON(t *testing.T) {
	t.Run("should build maintenance mode json", func(t *testing.T) {
		mmReq := maintenanceModeRequest{
			Activate: false,
			Title:    "test",
			Text:     "testmessage",
		}
		status, _ := BuildMaintenanceJSON(mmReq)
		require.Equal(t, status.String(), "{\"activate\":false,\"title\":\"test\",\"text\":\"testmessage\"}")
	})

	t.Run("should build maintenance mode json with empty request", func(t *testing.T) {
		mmReq := maintenanceModeRequest{
			Activate: false,
			Title:    "",
			Text:     "",
		}
		status, _ := BuildMaintenanceJSON(mmReq)
		require.Equal(t, status.String(), "{\"activate\":false,\"title\":\"\",\"text\":\"\"}")
	})
}

func TestMNDeactivateMaintenanceMode(t *testing.T) {
	t.Run("should return maintenance mode is inactive", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)
		globalConfigRepo := repository.NewGlobalConfigRepository(configMaps)
		mmReq := maintenanceModeRequest{
			Activate: false,
			Title:    "test",
			Text:     "testmessage",
		}
		configMaps.EXPECT().List(mock.Anything, mock.Anything).Return(&corev1.ConfigMapList{
			Items: []corev1.ConfigMap{
				{Data: map[string]string{
					"config.yaml": "maintenance: value",
				}},
			},
		}, nil)
		configMaps.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).Return(&corev1.ConfigMap{
			Data: map[string]string{
				"config.yaml": "key: value",
			},
		}, nil)
		provider := NewMultinodeMaintenanceModeProvider(globalConfigRepo)

		status, err := provider.SetMaintenanceMode(mmReq, context.TODO())
		require.NoError(t, err)
		require.Equal(t, status.IsActive, false)
	})
	t.Run("should return error", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)
		globalConfigRepo := repository.NewGlobalConfigRepository(configMaps)
		mmReq := maintenanceModeRequest{
			Activate: false,
			Title:    "test",
			Text:     "testmessage",
		}
		configMaps.EXPECT().List(mock.Anything, mock.Anything).Return(&corev1.ConfigMapList{
			Items: []corev1.ConfigMap{
				{Data: map[string]string{
					"config.yaml": "maintenance: value",
				}},
			},
		}, nil)
		configMaps.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("testerror"))
		provider := NewMultinodeMaintenanceModeProvider(globalConfigRepo)

		_, err := provider.SetMaintenanceMode(mmReq, context.TODO())
		require.Contains(t, err.Error(), "failed to update global config:")
		require.Contains(t, err.Error(), "testerror")
	})
	t.Run("should fail on get global config", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)
		globalConfigRepo := repository.NewGlobalConfigRepository(configMaps)
		mmReq := maintenanceModeRequest{
			Activate: false,
			Title:    "test",
			Text:     "testmessage",
		}
		configMaps.EXPECT().List(mock.Anything, mock.Anything).Return(&corev1.ConfigMapList{
			Items: []corev1.ConfigMap{
				{Data: map[string]string{}},
			},
		}, nil)
		provider := NewMultinodeMaintenanceModeProvider(globalConfigRepo)

		_, err := provider.SetMaintenanceMode(mmReq, context.TODO())
		require.Contains(t, err.Error(), "failed to get global config:")
	})
}

func TestMNActivateMaintenanceMode(t *testing.T) {
	t.Run("should return maintenance mode is active", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)
		globalConfigRepo := repository.NewGlobalConfigRepository(configMaps)
		configMaps.EXPECT().List(mock.Anything, mock.Anything).Return(&corev1.ConfigMapList{
			Items: []corev1.ConfigMap{
				{Data: map[string]string{
					"config.yaml": "maintenance: value",
				}},
			},
		}, nil)
		configMaps.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).Return(&corev1.ConfigMap{
			Data: map[string]string{
				"config.yaml": "maintenance: value",
			},
		}, nil)
		provider := NewMultinodeMaintenanceModeProvider(globalConfigRepo)
		req := maintenanceModeRequest{
			Activate: true,
		}
		status, err := provider.SetMaintenanceMode(req, context.TODO())
		require.NoError(t, err)
		require.Equal(t, status.IsActive, true)
	})

	t.Run("should return error", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)
		globalConfigRepo := repository.NewGlobalConfigRepository(configMaps)
		configMaps.EXPECT().List(mock.Anything, mock.Anything).Return(&corev1.ConfigMapList{
			Items: []corev1.ConfigMap{
				{Data: map[string]string{
					"config.yaml": "maintenance: value",
				}},
			},
		}, nil)
		configMaps.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("testerror"))
		provider := NewMultinodeMaintenanceModeProvider(globalConfigRepo)
		req := maintenanceModeRequest{
			Activate: true,
		}
		_, err := provider.SetMaintenanceMode(req, context.TODO())
		require.Contains(t, err.Error(), "failed to update global config:")
		require.Contains(t, err.Error(), "testerror")
	})
	t.Run("should fail on get global config", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)
		globalConfigRepo := repository.NewGlobalConfigRepository(configMaps)
		configMaps.EXPECT().List(mock.Anything, mock.Anything).Return(&corev1.ConfigMapList{
			Items: []corev1.ConfigMap{
				{Data: map[string]string{}},
			},
		}, nil)
		provider := NewMultinodeMaintenanceModeProvider(globalConfigRepo)
		req := maintenanceModeRequest{
			Activate: true,
		}
		_, err := provider.SetMaintenanceMode(req, context.TODO())
		require.Contains(t, err.Error(), "failed to get global config:")
	})
}
