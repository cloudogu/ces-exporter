package maintenance

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCGetMaintenanceMode(t *testing.T) {
	t.Run("should return maintenance mode is inactive", func(t *testing.T) {
		etcdRepoMock := newMockEtcdGetter(t)

		etcdRepoMock.EXPECT().Get(maintenanceModeEtcdKey, mock.Anything).Return("", nil)

		provider := &ClassicMaintenanceModeProvider{
			etcdGetter: etcdRepoMock,
		}

		status, err := provider.GetMaintenanceMode(context.TODO())
		require.NoError(t, err)
		require.Equal(t, status.IsActive, false)
	})

	t.Run("should return maintenance mode is active", func(t *testing.T) {
		etcdRepoMock := newMockEtcdGetter(t)

		etcdRepoMock.EXPECT().Get(maintenanceModeEtcdKey, mock.Anything).Return("some value", nil)

		provider := &ClassicMaintenanceModeProvider{
			etcdGetter: etcdRepoMock,
		}

		status, err := provider.GetMaintenanceMode(context.TODO())
		require.NoError(t, err)
		require.Equal(t, status.IsActive, true)
	})

	t.Run("should fail on get global config", func(t *testing.T) {
		etcdRepoMock := newMockEtcdGetter(t)

		etcdRepoMock.EXPECT().Get(maintenanceModeEtcdKey, mock.Anything).Return("some value", fmt.Errorf("testerror"))

		provider := &ClassicMaintenanceModeProvider{
			etcdGetter: etcdRepoMock,
		}

		_, err := provider.GetMaintenanceMode(context.TODO())
		require.Contains(t, err.Error(), "failed to get etcd value:")
	})
}

func TestCDeactivateMaintenanceMode(t *testing.T) {
	t.Run("should return maintenance mode is inactive", func(t *testing.T) {
		etcdRepoMock := newMockEtcdGetter(t)
		mmReq := maintenanceModeRequest{
			Activate: false,
			Message: Message{
				Title: "test",
				Text:  "testmessage",
			},
		}

		etcdRepoMock.EXPECT().Delete(maintenanceModeEtcdKey, mock.Anything).Return(nil)
		provider := &ClassicMaintenanceModeProvider{
			etcdGetter: etcdRepoMock,
		}

		status, err := provider.SetMaintenanceMode(mmReq, context.TODO())
		require.NoError(t, err)
		require.Equal(t, status.IsActive, false)
	})

	t.Run("should return error", func(t *testing.T) {
		etcdRepoMock := newMockEtcdGetter(t)
		mmReq := maintenanceModeRequest{
			Activate: false,
			Message: Message{
				Title: "test",
				Text:  "testmessage",
			},
		}
		etcdRepoMock.EXPECT().Delete(maintenanceModeEtcdKey, mock.Anything).Return(fmt.Errorf("testerror"))

		provider := &ClassicMaintenanceModeProvider{
			etcdGetter: etcdRepoMock,
		}

		_, err := provider.SetMaintenanceMode(mmReq, context.TODO())
		require.Contains(t, err.Error(), "failed to remove maintenance-etcd key: testerror")
	})
}

func TestCActivateMaintenanceMode(t *testing.T) {
	t.Run("should return maintenance mode is active", func(t *testing.T) {
		etcdRepoMock := newMockEtcdGetter(t)

		etcdRepoMock.EXPECT().Update(maintenanceModeEtcdKey, mock.Anything, mock.Anything).Return("some value", nil)

		provider := &ClassicMaintenanceModeProvider{
			etcdGetter: etcdRepoMock,
		}
		req := maintenanceModeRequest{
			Activate: true,
		}
		status, err := provider.SetMaintenanceMode(req, context.TODO())
		require.NoError(t, err)
		require.Equal(t, status.IsActive, true)
	})

	t.Run("should return error", func(t *testing.T) {
		etcdRepoMock := newMockEtcdGetter(t)

		etcdRepoMock.EXPECT().Update(maintenanceModeEtcdKey, mock.Anything, mock.Anything).Return("some value", fmt.Errorf("testerror"))

		provider := &ClassicMaintenanceModeProvider{
			etcdGetter: etcdRepoMock,
		}

		req := maintenanceModeRequest{
			Activate: true,
		}
		_, err := provider.SetMaintenanceMode(req, context.TODO())
		require.Contains(t, err.Error(), "could not set maintenance mode: testerror")
	})

}

func TestCClassicConstructor(t *testing.T) {
	t.Run("create new instance", func(t *testing.T) {
		provider := NewClassicProvider()

		require.NotNil(t, provider.etcdGetter)
	})
}
