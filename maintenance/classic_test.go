package maintenance

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.etcd.io/etcd/client/v2"
	"testing"
)

func TestCGetMaintenanceMode(t *testing.T) {
	t.Run("should return maintenance mode is inactive", func(t *testing.T) {
		etcdRepoMock := newMockEtcdConfigRepo(t)

		etcdRepoMock.EXPECT().Get(maintenanceModeEtcdKey).Return("", nil)

		provider := NewClassicProvider(etcdRepoMock)

		status, err := provider.GetMaintenanceMode(context.TODO())
		require.NoError(t, err)
		require.Equal(t, status.IsActive, false)
	})

	t.Run("should return maintenance mode is active", func(t *testing.T) {
		etcdRepoMock := newMockEtcdConfigRepo(t)

		etcdRepoMock.EXPECT().Get(maintenanceModeEtcdKey).Return("some value", nil)

		provider := NewClassicProvider(etcdRepoMock)

		status, err := provider.GetMaintenanceMode(context.TODO())
		require.NoError(t, err)
		require.Equal(t, status.IsActive, true)
	})

	t.Run("should fail on get global config", func(t *testing.T) {
		etcdRepoMock := newMockEtcdConfigRepo(t)

		etcdRepoMock.EXPECT().Get(maintenanceModeEtcdKey).Return("some value", fmt.Errorf("testerror"))

		provider := NewClassicProvider(etcdRepoMock)

		_, err := provider.GetMaintenanceMode(context.TODO())
		require.Contains(t, err.Error(), "failed to get etcd value:")
	})
}

func TestCDeactivateMaintenanceMode(t *testing.T) {
	t.Run("should return maintenance mode is inactive", func(t *testing.T) {
		etcdRepoMock := newMockEtcdConfigRepo(t)
		mmReq := maintenanceModeRequest{
			Activate: false,
			Message: Message{
				Title: "test",
				Text:  "testmessage",
			},
		}

		etcdRepoMock.EXPECT().Delete(maintenanceModeEtcdKey).Return(nil)
		provider := NewClassicProvider(etcdRepoMock)

		status, err := provider.SetMaintenanceMode(mmReq, context.TODO())
		require.NoError(t, err)
		require.Equal(t, status.IsActive, false)
	})

	t.Run("should return error", func(t *testing.T) {
		etcdRepoMock := newMockEtcdConfigRepo(t)
		mmReq := maintenanceModeRequest{
			Activate: false,
			Message: Message{
				Title: "test",
				Text:  "testmessage",
			},
		}
		etcdRepoMock.EXPECT().Delete(maintenanceModeEtcdKey).Return(fmt.Errorf("testerror"))

		provider := NewClassicProvider(etcdRepoMock)

		_, err := provider.SetMaintenanceMode(mmReq, context.TODO())
		require.Contains(t, err.Error(), "failed to remove maintenance-etcd key: testerror")
	})
}

func TestCActivateMaintenanceMode(t *testing.T) {
	t.Run("should return maintenance mode is active", func(t *testing.T) {
		etcdRepoMock := newMockEtcdConfigRepo(t)

		etcdRepoMock.EXPECT().Update(maintenanceModeEtcdKey, mock.Anything).Return("some value", nil)

		provider := NewClassicProvider(etcdRepoMock)
		req := maintenanceModeRequest{
			Activate: true,
		}
		status, err := provider.SetMaintenanceMode(req, context.TODO())
		require.NoError(t, err)
		require.Equal(t, status.IsActive, true)
	})

	t.Run("should return error", func(t *testing.T) {
		etcdRepoMock := newMockEtcdConfigRepo(t)

		etcdRepoMock.EXPECT().Update(maintenanceModeEtcdKey, mock.Anything).Return("some value", fmt.Errorf("testerror"))

		provider := NewClassicProvider(etcdRepoMock)

		req := maintenanceModeRequest{
			Activate: true,
		}
		_, err := provider.SetMaintenanceMode(req, context.TODO())
		require.Contains(t, err.Error(), "could not set maintenance mode: testerror")
	})

}

func TestEtcdCommandWrapper(t *testing.T) {
	t.Run("test getter ok", func(t *testing.T) {
		mockApi := NewMockKeysApi(t)
		mockApi.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(&client.Response{
			Node: &client.Node{Value: "this is my value"},
		}, nil)
		repo := &EtcdConfigRepo{Etcdclient: mockApi}
		get, _ := repo.Get(maintenanceModeEtcdKey)
		require.Contains(t, get, "this is my value")
	})
	t.Run("test getter key not found", func(t *testing.T) {
		mockApi := NewMockKeysApi(t)
		mockApi.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("Key not found"))
		repo := &EtcdConfigRepo{Etcdclient: mockApi}
		get, _ := repo.Get(maintenanceModeEtcdKey)
		require.Equal(t, get, "")
	})
	t.Run("test getter error", func(t *testing.T) {
		mockApi := NewMockKeysApi(t)
		mockApi.EXPECT().Get(mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("testerror"))
		repo := &EtcdConfigRepo{Etcdclient: mockApi}
		get, _ := repo.Get(maintenanceModeEtcdKey)
		require.Equal(t, get, "")
	})
	t.Run("test setter ok", func(t *testing.T) {
		mockApi := NewMockKeysApi(t)
		mockApi.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&client.Response{
			Node: &client.Node{Value: "{\"title\": \"newTitle\"}"},
		}, nil)
		repo := &EtcdConfigRepo{Etcdclient: mockApi}
		update, _ := repo.Update(maintenanceModeEtcdKey, "{\"title\": \"newTitle\"}")
		require.Contains(t, update, "{\"title\": \"newTitle\"}")
	})
	t.Run("test setter error", func(t *testing.T) {
		mockApi := NewMockKeysApi(t)
		mockApi.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("testerror"))
		repo := &EtcdConfigRepo{Etcdclient: mockApi}
		update, _ := repo.Update(maintenanceModeEtcdKey, "{\"title\": \"newTitle\"}")
		require.Equal(t, update, "")
	})
	t.Run("test delete ok", func(t *testing.T) {
		mockApi := NewMockKeysApi(t)
		mockApi.EXPECT().Delete(mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
		repo := &EtcdConfigRepo{Etcdclient: mockApi}
		err := repo.Delete(maintenanceModeEtcdKey)
		require.NoError(t, err)
	})
	t.Run("test delete error", func(t *testing.T) {
		mockApi := NewMockKeysApi(t)
		mockApi.EXPECT().Delete(mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("testerror"))
		repo := &EtcdConfigRepo{Etcdclient: mockApi}
		err := repo.Delete(maintenanceModeEtcdKey)
		require.Error(t, err)
	})
}
