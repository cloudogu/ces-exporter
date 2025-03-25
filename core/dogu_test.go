package core

import (
	"context"
	"github.com/cloudogu/ces-commons-lib/dogu"
	"github.com/cloudogu/cesapp-lib/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestGetInstalledDogus(t *testing.T) {
	t.Run("will query installed dogus", func(t *testing.T) {
		cm := newMockConfigmaps(t)
		cm.EXPECT().List(mock.Anything, mock.Anything).Return(&corev1.ConfigMapList{
			Items: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"dogu.name": "d1",
						},
					},
					Data: map[string]string{
						"current": "1.0.0-1",
					},
				},
			},
		}, nil)
		dogus, err := GetInstalledDogus(context.TODO(), cm)
		require.NoError(t, err)
		v, err := core.ParseVersion("1.0.0-1")
		require.Nil(t, err)
		assert.Equal(t, []dogu.SimpleNameVersion{
			{
				Name:    "d1",
				Version: v,
			},
		}, dogus)
	})
}
