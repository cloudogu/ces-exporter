package export

import (
	"context"
	"fmt"
	v2 "github.com/cloudogu/k8s-dogu-operator/v3/api/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		require.Equal(t, "/data", dogu.VolumePath)
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

		serviceClient.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

		doguClient.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(&v2.Dogu{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test_A",
			},
		}, nil)

		dogu, _ := provider.SetExportDogu("test_A", context.Background())

		require.Equal(t, 8080, dogu.ExporterPort)
		require.Equal(t, "test_A", dogu.Dogu)
		require.Equal(t, "test_A-data", dogu.VolumePath)
	})
	t.Run("should return get error on getting dogu for name", func(t *testing.T) {
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

		doguClient.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("testerror"))

		_, err := provider.SetExportDogu("test_A", context.Background())

		require.Contains(t, "could not get dogu resource for current export dogu: testerror", err.Error())
	})
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

		serviceClient.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("testerror"))

		doguClient.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(&v2.Dogu{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test_A",
			},
		}, nil)

		_, err := provider.SetExportDogu("test_A", context.Background())

		require.Contains(t, "could not get dogu resource for current export dogu: testerror", err.Error())
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
