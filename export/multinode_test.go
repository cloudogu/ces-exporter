package export

import (
	"context"
	"fmt"
	v2 "github.com/cloudogu/k8s-dogu-operator/v3/api/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestMNGetExportDogu(t *testing.T) {
	t.Run("should return get export dogu", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)
		doguClient := newMockDoguClient(t)
		serviceClient := newMockServiceClient(t)
		provider := NewMultinodeExportModeProvider("ecosystem", configMaps, doguClient, serviceClient)

		serviceClient.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(&corev1.Service{
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{{
					Port: 8080,
				}},
				Selector: map[string]string{"dogu.name": "test_A"},
			},
		}, nil)

		dogu, _ := provider.GetExportDogu(context.Background())

		require.Equal(t, 8080, dogu.ExporterPort)
		require.Equal(t, "test_A", dogu.Dogu)
		require.Equal(t, "/data/test_A", dogu.VolumePath)
	})
}
func TestMNSetExportDogu(t *testing.T) {
	t.Run("should set export dogu", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)
		doguClient := newMockDoguClient(t)
		serviceClient := newMockServiceClient(t)
		provider := NewMultinodeExportModeProvider("ecosystem", configMaps, doguClient, serviceClient)

		serviceClient.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(&corev1.Service{
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{{
					Port: 8080,
				}},
				Selector: map[string]string{"dogu.name": "test_A"},
			},
		}, nil)
		doguClient.EXPECT().Get(mock.Anything, "test_A", mock.Anything).Return(nil, nil)

		serviceClient.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

		dogu, _ := provider.SetExportDogu(context.Background(), "test_A")

		require.Equal(t, 8080, dogu.ExporterPort)
		require.Equal(t, "test_A", dogu.Dogu)
		require.Equal(t, "/data/test_A", dogu.VolumePath)
	})
	t.Run("fail on error with service update", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)
		doguClient := newMockDoguClient(t)
		serviceClient := newMockServiceClient(t)
		provider := NewMultinodeExportModeProvider("ecosystem", configMaps, doguClient, serviceClient)

		serviceClient.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(&corev1.Service{
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{{
					Port: 8080,
				}},
				Selector: map[string]string{"dogu.name": "test_A"},
			},
		}, nil)
		doguClient.EXPECT().Get(mock.Anything, "test_A", mock.Anything).Return(nil, nil)

		serviceClient.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("testerror"))

		_, err := provider.SetExportDogu(context.Background(), "test_A")

		require.Contains(t, err.Error(), "failed to update exporter service: testerror")
	})
	t.Run("fail on export non existent dogu", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)
		doguClient := newMockDoguClient(t)
		serviceClient := newMockServiceClient(t)
		provider := NewMultinodeExportModeProvider("ecosystem", configMaps, doguClient, serviceClient)

		serviceClient.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(&corev1.Service{
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{{
					Port: 8080,
				}},
				Selector: map[string]string{"dogu.name": "test_A"},
			},
		}, nil)
		doguClient.EXPECT().Get(mock.Anything, "test_A", mock.Anything).Return(nil, fmt.Errorf("testerror"))

		_, err := provider.SetExportDogu(context.Background(), "test_A")

		require.Contains(t, err.Error(), "could not get dogu resource for current export dogu: testerror")
	})
}

func TestMNGetExportMode(t *testing.T) {
	t.Run("should return get export mode as false", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)

		doguClient := newMockDoguClient(t)
		doguClient.EXPECT().List(mock.Anything, mock.Anything).Return(&v2.DoguList{
			Items: []v2.Dogu{
				{
					Spec: v2.DoguSpec{
						Name:       "test_A",
						ExportMode: false,
					},
				},
				{
					Spec: v2.DoguSpec{
						Name:       "test_B",
						ExportMode: true,
					},
				},
			},
		}, nil)

		serviceClient := newMockServiceClient(t)
		provider := NewMultinodeExportModeProvider("ecosystem", configMaps, doguClient, serviceClient)

		mode, _ := provider.GetExportMode(context.Background())

		require.Equal(t, false, mode.IsActive)
	})
	t.Run("should return get export mode as true", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)

		doguClient := newMockDoguClient(t)
		doguClient.EXPECT().List(mock.Anything, mock.Anything).Return(&v2.DoguList{
			Items: []v2.Dogu{
				{
					Spec: v2.DoguSpec{
						Name:       "test_A",
						ExportMode: true,
					},
				},
				{
					Spec: v2.DoguSpec{
						Name:       "test_B",
						ExportMode: true,
					},
				},
			},
		}, nil)

		serviceClient := newMockServiceClient(t)

		provider := NewMultinodeExportModeProvider("ecosystem", configMaps, doguClient, serviceClient)

		mode, _ := provider.GetExportMode(context.Background())

		require.Equal(t, true, mode.IsActive)
	})
	t.Run("should return error on export mode", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)

		doguClient := newMockDoguClient(t)
		doguClient.EXPECT().List(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("testerror"))

		serviceClient := newMockServiceClient(t)
		provider := NewMultinodeExportModeProvider("ecosystem", configMaps, doguClient, serviceClient)

		_, err := provider.GetExportMode(context.Background())

		require.Contains(t, "error getting dogu list: testerror", err.Error())
	})
}
