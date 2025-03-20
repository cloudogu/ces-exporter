package systeminfo

import (
	"context"
	"fmt"
	v1 "github.com/cloudogu/k8s-component-operator/pkg/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestMultinodeSystemInfoProvider(t *testing.T) {
	t.Run("isMultinode() always returns true", func(t *testing.T) {
		provider := NewMultinodeSystemInfoProvider(nil, "", nil)
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
			provider := NewMultinodeSystemInfoProvider(nil, "", lister)

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
			provider := NewMultinodeSystemInfoProvider(nil, "", lister)

			components, err := provider.getComponents(context.Background())
			assert.Error(t, err)
			assert.Nil(t, components)
		})
	})
}
