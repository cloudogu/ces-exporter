package maintenance

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestMNGetMaintenanceMode(t *testing.T) {
	t.Run("should return maintenance mode is inactive", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)
		configMaps.EXPECT().List(mock.Anything, mock.Anything).Return(&corev1.ConfigMapList{
			Items: []corev1.ConfigMap{
				{Data: map[string]string{
					"config.yaml": "key: value",
				}},
			},
		}, nil)
		provider := NewMultinodeMaintenanceModeProvider(configMaps)

		status, err := provider.GetMaintenanceMode(context.TODO())
		require.NoError(t, err)
		require.Equal(t, status.IsActive, false)
	})

	t.Run("should return maintenance mode is active", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)
		configMaps.EXPECT().List(mock.Anything, mock.Anything).Return(&corev1.ConfigMapList{
			Items: []corev1.ConfigMap{
				{Data: map[string]string{
					"config.yaml": "maintenance: value",
				}},
			},
		}, nil)
		provider := NewMultinodeMaintenanceModeProvider(configMaps)

		status, err := provider.GetMaintenanceMode(context.TODO())
		require.NoError(t, err)
		require.Equal(t, status.IsActive, true)
	})

	t.Run("should fail on get global config", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)
		configMaps.EXPECT().List(mock.Anything, mock.Anything).Return(&corev1.ConfigMapList{
			Items: []corev1.ConfigMap{
				{Data: map[string]string{}},
			},
		}, nil)
		provider := NewMultinodeMaintenanceModeProvider(configMaps)

		_, err := provider.GetMaintenanceMode(context.TODO())
		require.Contains(t, err.Error(), "failed to get global config:")
	})
}

func TestMNBuildMaintenanceJSON(t *testing.T) {
	t.Run("should build maintenance mode json", func(t *testing.T) {
		mmReq := maintenanceModeRequest{
			Activate: false,
			Title:    "test",
			Message:  "testmessage",
		}
		status := BuildMaintenanceJSON(mmReq)
		require.Equal(t, status.String(), "{\"title\": \"test\", \"message\": \"testmessage\"}")
	})

	t.Run("should build maintenance mode json with empty request", func(t *testing.T) {
		mmReq := maintenanceModeRequest{
			Activate: false,
			Title:    "",
			Message:  "",
		}
		status := BuildMaintenanceJSON(mmReq)
		require.Equal(t, status.String(), "{\"title\": \"\", \"message\": \"\"}")
	})
}

func TestMNDeactivateMaintenanceMode(t *testing.T) {
	t.Run("should return maintenance mode is inactive", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)
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
		provider := NewMultinodeMaintenanceModeProvider(configMaps)

		status, err := provider.DeactivateMaintenanceMode(context.TODO())
		require.NoError(t, err)
		require.Equal(t, status.IsActive, false)
	})
	t.Run("should return error", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)
		configMaps.EXPECT().List(mock.Anything, mock.Anything).Return(&corev1.ConfigMapList{
			Items: []corev1.ConfigMap{
				{Data: map[string]string{
					"config.yaml": "maintenance: value",
				}},
			},
		}, nil)
		configMaps.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("testerror"))
		provider := NewMultinodeMaintenanceModeProvider(configMaps)

		_, err := provider.DeactivateMaintenanceMode(context.TODO())
		require.Contains(t, err.Error(), "failed to update global config:")
		require.Contains(t, err.Error(), "testerror")
	})
	t.Run("should fail on get global config", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)
		configMaps.EXPECT().List(mock.Anything, mock.Anything).Return(&corev1.ConfigMapList{
			Items: []corev1.ConfigMap{
				{Data: map[string]string{}},
			},
		}, nil)
		provider := NewMultinodeMaintenanceModeProvider(configMaps)

		_, err := provider.DeactivateMaintenanceMode(context.TODO())
		require.Contains(t, err.Error(), "failed to get global config:")
	})
}

func TestMNActivateMaintenanceMode(t *testing.T) {
	t.Run("should return maintenance mode is active", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)
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
		provider := NewMultinodeMaintenanceModeProvider(configMaps)
		req := maintenanceModeRequest{
			Activate: true,
		}
		status, err := provider.ActivateMaintenanceMode(req, context.TODO())
		require.NoError(t, err)
		require.Equal(t, status.IsActive, true)
	})

	t.Run("should return error", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)
		configMaps.EXPECT().List(mock.Anything, mock.Anything).Return(&corev1.ConfigMapList{
			Items: []corev1.ConfigMap{
				{Data: map[string]string{
					"config.yaml": "maintenance: value",
				}},
			},
		}, nil)
		configMaps.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("testerror"))
		provider := NewMultinodeMaintenanceModeProvider(configMaps)
		req := maintenanceModeRequest{
			Activate: true,
		}
		_, err := provider.ActivateMaintenanceMode(req, context.TODO())
		require.Contains(t, err.Error(), "failed to update global config:")
		require.Contains(t, err.Error(), "testerror")
	})
	t.Run("should fail on get global config", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)
		configMaps.EXPECT().List(mock.Anything, mock.Anything).Return(&corev1.ConfigMapList{
			Items: []corev1.ConfigMap{
				{Data: map[string]string{}},
			},
		}, nil)
		provider := NewMultinodeMaintenanceModeProvider(configMaps)
		req := maintenanceModeRequest{
			Activate: true,
		}
		_, err := provider.ActivateMaintenanceMode(req, context.TODO())
		require.Contains(t, err.Error(), "failed to get global config:")
	})
}
