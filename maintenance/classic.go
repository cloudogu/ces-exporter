package maintenance

import (
	"context"
	"fmt"
	client2 "go.etcd.io/etcd/client/v2"
	"log/slog"
	"os/exec"
	"strings"
)

const (
	maintenanceModeEtcdKey = "/config/_global/maintenance"
)

type etcdConfigRepo interface {
	Get(string) (string, error)
	Update(key string, value string) (string, error)
	Delete(string) error
}

type ClassicMaintenanceModeProvider struct {
	etcdConfigRepo etcdConfigRepo
}

type EtcdConfigRepo struct {
	etcdConfigRepo
	Exec func(name string, arg ...string) *exec.Cmd
}

func NewClassicProvider(repo etcdConfigRepo) *ClassicMaintenanceModeProvider {
	return &ClassicMaintenanceModeProvider{repo}
}

// SetMaintenanceMode activates or deactivates the maintenance mode by adding/removing the key maintenance to the global-config
func (m ClassicMaintenanceModeProvider) SetMaintenanceMode(mReq maintenanceModeRequest, ctx context.Context) (*MaintenanceModeStatus, error) {
	status := &MaintenanceModeStatus{
		IsActive: mReq.Activate,
	}

	if status.IsActive {
		reqJson, err := BuildMaintenanceJSON(mReq)
		if err != nil {
			return nil, err
		}
		out, err := m.etcdConfigRepo.Update(maintenanceModeEtcdKey, reqJson.String())
		if err != nil {
			return nil, fmt.Errorf("could not set maintenance mode: %w", err)
		}

		slog.Info(fmt.Sprintf("Maintenance-Mode activated: %s", out))
	} else {
		err := m.etcdConfigRepo.Delete(maintenanceModeEtcdKey)
		if err != nil {
			return nil, fmt.Errorf("failed to remove maintenance-etcd key: %w", err)
		}
		slog.Info("Maintenance-Mode deactivated")
	}
	return status, nil
}

// GetMaintenanceMode Get maintenance mode by checking if key maintenance exists in etcd
func (m ClassicMaintenanceModeProvider) GetMaintenanceMode(ctx context.Context) (*MaintenanceModeStatus, error) {

	status := &MaintenanceModeStatus{
		IsActive: false,
	}

	out, err := m.etcdConfigRepo.Get(maintenanceModeEtcdKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get etcd value: %w", err)
	}

	output := strings.TrimSpace(out)
	if output != "" {
		status.IsActive = true
	}

	return status, nil
}

func (e *EtcdConfigRepo) Get(key string) (string, error) {
	client, err := GetEtcdClient()
	if err != nil {
		return "", err
	}
	resp, err := client.Get(context.Background(), key, &client2.GetOptions{})
	if err != nil {
		return "", err
	}
	return resp.Node.Value, nil
}

func (e *EtcdConfigRepo) Update(key string, value string) (string, error) {
	client, err := GetEtcdClient()
	if err != nil {
		return "", err
	}
	resp, err := client.Set(context.Background(), key, value, &client2.SetOptions{})
	if err != nil {
		return "", err
	}
	return resp.Node.Value, nil
}

func (e *EtcdConfigRepo) Delete(key string) error {
	client, err := GetEtcdClient()
	if err != nil {
		return err
	}
	_, err = client.Delete(context.Background(), key, &client2.DeleteOptions{})
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	return nil
}
