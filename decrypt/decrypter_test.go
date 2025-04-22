package decrypt

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetPrivateKeyPath(t *testing.T) {
	t.Run("will format correctly", func(t *testing.T) {
		assert.Equal(t, "/data/d1/volumes/_private/private.pem", getPrivateKeyPath("d1"))
	})
}
