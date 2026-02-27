package maintenance

import (
	"context"
	"testing"

	"github.com/cloudogu/k8s-registry-lib/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testCtx = context.Background()

func TestMNGetMaintenanceMode(t *testing.T) {
	t.Run("should return maintenance mode is inactive", func(t *testing.T) {
		adapter := newMockMaintenanceAdapter(t)
		adapter.EXPECT().GetStatus(testCtx).Return(repository.MaintenanceModeDescription{}, false, nil)
		provider := NewMultinodeProvider(adapter)

		status, err := provider.GetMaintenanceMode(testCtx)
		require.NoError(t, err)
		assert.False(t, status.IsActive)
	})

	t.Run("should return maintenance mode is active", func(t *testing.T) {
		adapter := newMockMaintenanceAdapter(t)
		adapter.EXPECT().GetStatus(testCtx).Return(repository.MaintenanceModeDescription{}, true, nil)
		provider := NewMultinodeProvider(adapter)

		status, err := provider.GetMaintenanceMode(testCtx)
		require.NoError(t, err)
		assert.True(t, status.IsActive)
	})

	t.Run("should fail on get global config", func(t *testing.T) {
		adapter := newMockMaintenanceAdapter(t)
		adapter.EXPECT().GetStatus(testCtx).Return(repository.MaintenanceModeDescription{}, false, assert.AnError)
		provider := NewMultinodeProvider(adapter)

		_, err := provider.GetMaintenanceMode(testCtx)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "failed to get maintenance status")
	})
}

func TestMNDeactivateMaintenanceMode(t *testing.T) {
	t.Run("should return maintenance mode is inactive", func(t *testing.T) {
		adapter := newMockMaintenanceAdapter(t)
		adapter.EXPECT().Deactivate(testCtx, true).Return(nil)
		provider := NewMultinodeProvider(adapter)
		mmReq := maintenanceModeRequest{
			Activate: false,
			Message: Message{
				Title: "test",
				Text:  "testmessage",
			},
		}

		status, err := provider.SetMaintenanceMode(mmReq, testCtx)
		require.NoError(t, err)
		require.False(t, status.IsActive)
	})

	t.Run("should return error", func(t *testing.T) {
		adapter := newMockMaintenanceAdapter(t)
		adapter.EXPECT().Deactivate(testCtx, true).Return(assert.AnError)
		provider := NewMultinodeProvider(adapter)
		mmReq := maintenanceModeRequest{
			Activate: false,
			Message: Message{
				Title: "test",
				Text:  "testmessage",
			},
		}

		_, err := provider.SetMaintenanceMode(mmReq, testCtx)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "deactivate maintenance mode")
	})
}

func TestMNActivateMaintenanceMode(t *testing.T) {
	t.Run("should return maintenance mode is active", func(t *testing.T) {
		adapter := newMockMaintenanceAdapter(t)
		adapter.EXPECT().Activate(testCtx, repository.MaintenanceModeDescription{
			Title: "test",
			Text:  "testmessage",
		}, true).Return(nil)
		provider := NewMultinodeProvider(adapter)
		req := maintenanceModeRequest{
			Activate: true,
			Message: Message{
				Title: "test",
				Text:  "testmessage",
			},
		}
		status, err := provider.SetMaintenanceMode(req, testCtx)
		require.NoError(t, err)
		assert.True(t, status.IsActive)
	})

	t.Run("should return error", func(t *testing.T) {
		adapter := newMockMaintenanceAdapter(t)
		adapter.EXPECT().Activate(testCtx, repository.MaintenanceModeDescription{
			Title: "test",
			Text:  "testmessage",
		}, true).Return(assert.AnError)
		provider := NewMultinodeProvider(adapter)
		req := maintenanceModeRequest{
			Activate: true,
			Message: Message{
				Title: "test",
				Text:  "testmessage",
			},
		}
		_, err := provider.SetMaintenanceMode(req, testCtx)
		assert.ErrorIs(t, err, assert.AnError)
		assert.ErrorContains(t, err, "activate maintenance mode")
	})
}
