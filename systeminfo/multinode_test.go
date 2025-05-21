package systeminfo

import (
	"context"
	"fmt"
	v1 "github.com/cloudogu/k8s-component-operator/pkg/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestMultinodeSystemInfoProvider(t *testing.T) {
	t.Run("isMultinode() always returns true", func(t *testing.T) {
		provider := NewMultinodeSystemInfoProvider(nil, nil, "", nil)
		assert.True(t, provider.isMultinode())
	})

	t.Run("getComponents()", func(t *testing.T) {
		t.Run("can find components", func(t *testing.T) {
			lister := newMockComponentLister(t)
			lister.EXPECT().List(mock.Anything, mock.Anything).Return(&v1.ComponentList{
				Items: []v1.Component{
					{
						Spec: v1.ComponentSpec{Name: "c1", Version: "v1"},
					},
					{
						Spec: v1.ComponentSpec{Name: "c2", Version: "v2"},
					},
				},
			}, nil)
			provider := NewMultinodeSystemInfoProvider(nil, nil, "", lister)

			expectedComponents := []component{
				{Name: "c1", Version: "v1"},
				{Name: "c2", Version: "v2"},
			}
			components, err := provider.getComponents(context.Background())
			assert.NoError(t, err)

			assert.Equal(t, expectedComponents, components)
		})

		t.Run("fail on component list", func(t *testing.T) {
			lister := newMockComponentLister(t)
			lister.EXPECT().List(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("testerror"))
			provider := NewMultinodeSystemInfoProvider(nil, nil, "", lister)

			components, err := provider.getComponents(context.Background())
			assert.Error(t, err)
			assert.Nil(t, components)
		})
	})

	t.Run("getDogus()", func(t *testing.T) {
		t.Run("can find dogus", func(t *testing.T) {
			cm := newMockConfigMaps(t)
			cm.EXPECT().List(mock.Anything, mock.Anything).Return(&corev1.ConfigMapList{
				Items: []corev1.ConfigMap{
					{ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"dogu.name": "mydogu",
						},
					}, Immutable: nil, Data: map[string]string{
						"current": "1.0.0-1",
					}, BinaryData: nil},
				},
			}, nil)
			cm.EXPECT().Get(mock.Anything, "dogu-spec-mydogu", mock.Anything).Return(&corev1.ConfigMap{
				Data: map[string]string{
					"1.0.0-1": "{\n  \"Name\": \"official/mydogu\",\n  \"Version\": \"1.0.0-1\"\n}",
				},
			}, nil)
			pvc := newMockPvcClient(t)
			pvc.EXPECT().Get(mock.Anything, "mydogu", mock.Anything).Return(&corev1.PersistentVolumeClaim{
				Status: corev1.PersistentVolumeClaimStatus{
					Capacity: map[corev1.ResourceName]resource.Quantity{
						"storage": resource.MustParse("1Gi"),
					},
				},
			}, nil)
			provider := NewMultinodeSystemInfoProvider(cm, pvc, "namespace", nil)
			expectedDogus := []dogu{
				{Name: "official/mydogu", Version: "1.0.0-1", Volume: volume{SizeInBytes: 1073741824}},
			}
			dogus, err := provider.getDogus(context.Background())
			assert.NoError(t, err)

			assert.Equal(t, expectedDogus, dogus)
		})

		t.Run("fail on dogu list", func(t *testing.T) {
			cm := newMockConfigMaps(t)
			cm.EXPECT().List(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("testerror"))
			provider := NewMultinodeSystemInfoProvider(cm, nil, "", nil)

			components, err := provider.getDogus(context.Background())
			assert.Error(t, err)
			assert.Nil(t, components)
		})

		t.Run("fail on get dogu-descriptor", func(t *testing.T) {
			cm := newMockConfigMaps(t)
			cm.EXPECT().List(mock.Anything, mock.Anything).Return(&corev1.ConfigMapList{
				Items: []corev1.ConfigMap{
					{ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"dogu.name": "mydogu",
						},
					}, Immutable: nil, Data: map[string]string{
						"current": "1.0.0-1",
					}, BinaryData: nil},
				},
			}, nil)
			cm.EXPECT().Get(mock.Anything, "dogu-spec-mydogu", mock.Anything).Return(nil, assert.AnError)

			provider := NewMultinodeSystemInfoProvider(cm, nil, "namespace", nil)

			_, err := provider.getDogus(context.Background())
			require.Error(t, err)

			assert.ErrorContains(t, err, "failed to get dogu descriptor for dogu mydogu: failed to get dogu descriptor config map for dogu \"mydogu\": assert.AnError")
		})

		t.Run("fail on get volume size - no error => volume size zero", func(t *testing.T) {
			cm := newMockConfigMaps(t)
			cm.EXPECT().List(mock.Anything, mock.Anything).Return(&corev1.ConfigMapList{
				Items: []corev1.ConfigMap{
					{ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"dogu.name": "mydogu",
						},
					}, Immutable: nil, Data: map[string]string{
						"current": "1.0.0-1",
					}, BinaryData: nil},
				},
			}, nil)
			cm.EXPECT().Get(mock.Anything, "dogu-spec-mydogu", mock.Anything).Return(&corev1.ConfigMap{
				Data: map[string]string{
					"1.0.0-1": "{\n  \"Name\": \"official/mydogu\",\n  \"Version\": \"1.0.0-1\"\n}",
				},
			}, nil)
			pvc := newMockPvcClient(t)
			pvc.EXPECT().Get(mock.Anything, "mydogu", mock.Anything).Return(nil, fmt.Errorf(""))
			provider := NewMultinodeSystemInfoProvider(cm, pvc, "", nil)

			dogus, err := provider.getDogus(context.Background())
			assert.NoError(t, err)
			expectedDogus := []dogu{
				{Name: "official/mydogu", Version: "1.0.0-1", Volume: volume{SizeInBytes: 0}},
			}

			assert.Equal(t, expectedDogus, dogus)
		})

	})

	t.Run("getFqdn()", func(t *testing.T) {
		t.Run("can get fqdn from configmap", func(t *testing.T) {
			cm := newMockConfigMaps(t)
			cm.EXPECT().List(mock.Anything, mock.Anything).Return(&corev1.ConfigMapList{
				Items: []corev1.ConfigMap{
					{Data: map[string]string{
						"config.yaml": "fqdn: fqdn",
					}},
				},
			}, nil)
			provider := NewMultinodeSystemInfoProvider(cm, nil, "", nil)
			fqdn, err := provider.getFqdn(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, "fqdn", fqdn)
		})

		t.Run("fail on get global config repo", func(t *testing.T) {
			cm := newMockConfigMaps(t)
			cm.EXPECT().List(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("testerror"))
			provider := NewMultinodeSystemInfoProvider(cm, nil, "", nil)
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
			provider := NewMultinodeSystemInfoProvider(cm, nil, "", nil)
			_, err := provider.getFqdn(context.Background())
			assert.Error(t, err)
			assert.Equal(
				t,
				"critical error: no fqdn is configured in registry",
				err.Error(),
			)
		})
	})
}
