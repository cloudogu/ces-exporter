package export

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDoguClient(t *testing.T) {
	t.Run("create a new doguclient", func(t *testing.T) {
		doguclient := NewDoguClient("ecosystem", nil)
		require.Equal(t, "ecosystem", doguclient.namespace)
	})
}

func TestServiceClient(t *testing.T) {
	t.Run("create a new serviceclient", func(t *testing.T) {
		doguclient := NewServiceClient("ecosystem", nil)
		require.Equal(t, "ecosystem", doguclient.namespace)
	})
}
