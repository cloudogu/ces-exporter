package export

import (
	"context"
	"fmt"
	v2 "github.com/cloudogu/k8s-dogu-operator/v3/api/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"testing"
	"time"
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
	testCtx := context.Background()
	t.Run("should set export dogu", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)
		doguClient := newMockDoguClient(t)
		serviceClient := newMockServiceClient(t)
		sshBannerReader := newMockSshBannerReader(t)
		provider := &MultinodeExportModeProvider{
			namespace:       "ecosystem",
			configMaps:      configMaps,
			doguclient:      doguClient,
			serviceclient:   serviceClient,
			sshBannerReader: sshBannerReader,
		}

		sshBannerReader.EXPECT().readSSHBanner("fqdn:8080").Return("SSH server test_A\n", nil)

		configMaps.EXPECT().List(testCtx, mock.Anything).Return(&corev1.ConfigMapList{
			Items: []corev1.ConfigMap{
				{Data: map[string]string{
					"config.yaml": "fqdn: fqdn",
				}},
			},
		}, nil)

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

		dogu, err := provider.SetExportDogu(testCtx, "test_A")

		require.NoError(t, err)
		require.Equal(t, 8080, dogu.ExporterPort)
		require.Equal(t, "test_A", dogu.Dogu)
		require.Equal(t, "/data", dogu.VolumePath)
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
	t.Run("should fail to set export dogu on error getting fqdn", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)
		doguClient := newMockDoguClient(t)
		serviceClient := newMockServiceClient(t)
		sshBannerReader := newMockSshBannerReader(t)
		provider := &MultinodeExportModeProvider{
			namespace:       "ecosystem",
			configMaps:      configMaps,
			doguclient:      doguClient,
			serviceclient:   serviceClient,
			sshBannerReader: sshBannerReader,
		}

		configMaps.EXPECT().List(testCtx, mock.Anything).Return(nil, assert.AnError)

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

		_, err := provider.SetExportDogu(testCtx, "test_A")

		require.Error(t, err)
		require.ErrorContains(t, err, "failed to fqdn while setting export-mode:")
	})
	t.Run("should fail set export dogu for error waiting for sidecar", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)
		doguClient := newMockDoguClient(t)
		serviceClient := newMockServiceClient(t)
		sshBannerReader := newMockSshBannerReader(t)
		provider := &MultinodeExportModeProvider{
			namespace:       "ecosystem",
			configMaps:      configMaps,
			doguclient:      doguClient,
			serviceclient:   serviceClient,
			sshBannerReader: sshBannerReader,
		}

		configMaps.EXPECT().List(testCtx, mock.Anything).Return(&corev1.ConfigMapList{
			Items: []corev1.ConfigMap{
				{Data: map[string]string{
					"config.yaml": "fqdn: fqdn",
				}},
			},
		}, nil)

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

		originalMaxTries := getMaxTriesWaitForDoguSidecar
		getMaxTriesWaitForDoguSidecar = func() int {
			return 0
		}
		defer func() {
			getMaxTriesWaitForDoguSidecar = originalMaxTries
		}()

		_, err := provider.SetExportDogu(testCtx, "test_A")

		require.Error(t, err)
		require.ErrorContains(t, err, "failed to wait for endpoints to update: maxTries [0] reached while waiting for exporter-sidecar for dogu \"test_A\"")
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
					Status: v2.DoguStatus{
						ExportMode: true,
						Health:     v2.AvailableHealthStatus,
					},
				},
				{
					Spec: v2.DoguSpec{
						Name:       "test_B",
						ExportMode: true,
					},
					Status: v2.DoguStatus{
						ExportMode: true,
						Health:     v2.AvailableHealthStatus,
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
	t.Run("should return get export mode as false when exportMode-status is false", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)

		doguClient := newMockDoguClient(t)
		doguClient.EXPECT().List(mock.Anything, mock.Anything).Return(&v2.DoguList{
			Items: []v2.Dogu{
				{
					Spec: v2.DoguSpec{
						Name:       "test_A",
						ExportMode: true,
					},
					Status: v2.DoguStatus{
						ExportMode: true,
						Health:     v2.AvailableHealthStatus,
					},
				},
				{
					Spec: v2.DoguSpec{
						Name:       "test_B",
						ExportMode: true,
					},
					Status: v2.DoguStatus{
						ExportMode: false,
						Health:     v2.AvailableHealthStatus,
					},
				},
			},
		}, nil)

		serviceClient := newMockServiceClient(t)

		provider := NewMultinodeExportModeProvider("ecosystem", configMaps, doguClient, serviceClient)

		mode, _ := provider.GetExportMode(context.Background())

		require.False(t, mode.IsActive)
	})

	t.Run("should return get export mode as false when dogu is unhealthy", func(t *testing.T) {
		configMaps := newMockConfigMaps(t)

		doguClient := newMockDoguClient(t)
		doguClient.EXPECT().List(mock.Anything, mock.Anything).Return(&v2.DoguList{
			Items: []v2.Dogu{
				{
					Spec: v2.DoguSpec{
						Name:       "test_A",
						ExportMode: true,
					},
					Status: v2.DoguStatus{
						ExportMode: true,
						Health:     v2.AvailableHealthStatus,
					},
				},
				{
					Spec: v2.DoguSpec{
						Name:       "test_B",
						ExportMode: true,
					},
					Status: v2.DoguStatus{
						ExportMode: true,
						Health:     v2.UnavailableHealthStatus,
					},
				},
			},
		}, nil)

		serviceClient := newMockServiceClient(t)

		provider := NewMultinodeExportModeProvider("ecosystem", configMaps, doguClient, serviceClient)

		mode, _ := provider.GetExportMode(context.Background())

		require.False(t, mode.IsActive)
	})
}

func TestMultinodeExportModeProvider_getFqdn(t *testing.T) {
	t.Run("can get fqdn from configmap", func(t *testing.T) {
		cm := newMockConfigMaps(t)
		cm.EXPECT().List(mock.Anything, mock.Anything).Return(&corev1.ConfigMapList{
			Items: []corev1.ConfigMap{
				{Data: map[string]string{
					"config.yaml": "fqdn: fqdn",
				}},
			},
		}, nil)
		provider := &MultinodeExportModeProvider{
			configMaps: cm,
		}

		fqdn, err := provider.getFqdn(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, "fqdn", fqdn)
	})

	t.Run("fail on get global config repo", func(t *testing.T) {
		cm := newMockConfigMaps(t)
		cm.EXPECT().List(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("testerror"))
		provider := &MultinodeExportModeProvider{
			configMaps: cm,
		}

		_, err := provider.getFqdn(context.Background())
		assert.Error(t, err)
		assert.Equal(
			t,
			"failed to get global config: could not get global config: unable to get data 'global-config' from cluster: unable to list config-map from cluster: testerror",
			err.Error(),
		)
	})

	t.Run("fail on get global config key for fqdn", func(t *testing.T) {
		cm := newMockConfigMaps(t)
		cm.EXPECT().List(mock.Anything, mock.Anything).Return(&corev1.ConfigMapList{
			Items: []corev1.ConfigMap{
				{Data: map[string]string{
					"config.yaml": "{}",
				}},
			},
		}, nil)
		provider := &MultinodeExportModeProvider{
			configMaps: cm,
		}

		_, err := provider.getFqdn(context.Background())
		assert.Error(t, err)
		assert.Equal(
			t,
			"critical error: no fqdn is configured in registry",
			err.Error(),
		)
	})
}

func Test_waitForDoguSidecar(t *testing.T) {
	t.Run("should wait for sidecar", func(t *testing.T) {
		address := "test:7022"
		mockBannerReader := newMockSshBannerReader(t)
		mockBannerReader.EXPECT().readSSHBanner(address).RunAndReturn(func(addr string) (string, error) {
			return "SSH server ldap\n", nil
		})

		m := &MultinodeExportModeProvider{
			sshBannerReader: mockBannerReader,
		}

		err := m.waitForDoguSidecar(address, "ldap", 10, 10*time.Millisecond)
		require.NoError(t, err)
	})

	t.Run("should continue waiting for sidecar on error", func(t *testing.T) {
		address := "test:7022"
		mockBannerReader := newMockSshBannerReader(t)
		i := 0
		mockBannerReader.EXPECT().readSSHBanner(address).RunAndReturn(func(addr string) (string, error) {
			i++
			assert.Less(t, i, 4)

			if i == 1 {
				return "", assert.AnError
			}
			if i == 2 {
				return "SSH server otherDogu\n", nil
			}

			return "SSH server ldap\n", nil
		})

		m := &MultinodeExportModeProvider{
			sshBannerReader: mockBannerReader,
		}

		err := m.waitForDoguSidecar(address, "ldap", 3, 10*time.Millisecond)
		require.NoError(t, err)
	})

	t.Run("should wait for max retries", func(t *testing.T) {
		address := "test:7022"
		mockBannerReader := newMockSshBannerReader(t)
		mockBannerReader.EXPECT().readSSHBanner(address).RunAndReturn(func(addr string) (string, error) {
			return "SSH server otherDogu\n", nil
		})

		m := &MultinodeExportModeProvider{
			sshBannerReader: mockBannerReader,
		}

		err := m.waitForDoguSidecar(address, "ldap", 3, 10*time.Millisecond)
		require.Error(t, err)
		assert.ErrorContains(t, err, "maxTries [3] reached while waiting for exporter-sidecar for dogu \"ldap\"")
	})
}
