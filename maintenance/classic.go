package maintenance

import (
	"context"
	"fmt"
	"github.com/cloudogu/ces-exporter/etcd"
	client2 "go.etcd.io/etcd/client/v2"
	"log/slog"
	"strings"
)

const (
	maintenanceModeEtcdKey = "/config/_global/maintenance"
)

type etcdGetter interface {
	Get(string, *client2.GetOptions) (string, error)
	Update(string, string, *client2.SetOptions) (string, error)
	Delete(string, *client2.DeleteOptions) error
}

type classicEtcdGetter struct{}

func (c classicEtcdGetter) Get(key string, options *client2.GetOptions) (string, error) {
	return etcd.Get(key, options)
}

func (c classicEtcdGetter) Update(key string, value string, options *client2.SetOptions) (string, error) {
	return etcd.Update(key, value, options)
}

func (c classicEtcdGetter) Delete(key string, options *client2.DeleteOptions) error {
	return etcd.Delete(key, options)
}

type ClassicMaintenanceModeProvider struct {
	etcdGetter
}

func NewClassicProvider() *ClassicMaintenanceModeProvider {
	return &ClassicMaintenanceModeProvider{etcdGetter: &classicEtcdGetter{}}
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
		out, err := m.etcdGetter.Update(maintenanceModeEtcdKey, reqJson.String(), &client2.SetOptions{})
		if err != nil {
			return nil, fmt.Errorf("could not set maintenance mode: %w", err)
		}

		slog.Info(fmt.Sprintf("Maintenance-Mode activated: %s", out))
	} else {
		err := m.etcdGetter.Delete(maintenanceModeEtcdKey, &client2.DeleteOptions{})
		// ignore key not found errors, as they just mean that the maintenance mode is already deactivated
		if err != nil && !client2.IsKeyNotFound(err) {
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

	out, err := m.etcdGetter.Get(maintenanceModeEtcdKey, &client2.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get etcd value: %w", err)
	}

	output := strings.TrimSpace(out)
	if output != "" {
		status.IsActive = true
	}

	return status, nil
}
