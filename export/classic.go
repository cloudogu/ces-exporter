package export

import (
	"context"
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/cloudogu/ces-exporter/etcd"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"log/slog"
	"path"
	"strings"
)

type execClient interface {
	GetAllDogus() ([]string, error)
	ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error)
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)
}

type ExecClient struct{}

type ClassicExportModeProvider struct {
	currentExportDogu string
	execClient        execClient
	config            core.Configuration
}

func NewClassicExportModeProvider(conf core.Configuration) *ClassicExportModeProvider {
	return &ClassicExportModeProvider{config: conf, execClient: &ExecClient{}}
}

// GetExportDogu gets the dogu.name currently set in the ces-exporter-dogu-exporter service
func (c *ClassicExportModeProvider) GetExportDogu(_ context.Context) (*doguExport, error) {
	dogu, err := c.checkDogu(c.currentExportDogu)
	if err != nil {
		return nil, err
	}

	dE := &doguExport{
		Dogu:         dogu,
		VolumePath:   path.Join(dataVolumePath, dogu),
		ExporterPort: c.config.ClassicExportPort,
	}
	return dE, nil
}

// SetExportDogu sets the given dogu as dogu.name in the ces-exporter-dogu-exporter service
func (c *ClassicExportModeProvider) SetExportDogu(_ context.Context, doguName string) (*doguExport, error) {
	dogu, err := c.checkDogu(doguName)
	if err != nil {
		return nil, err
	}
	c.currentExportDogu = dogu

	dE := &doguExport{
		Dogu:         dogu,
		VolumePath:   path.Join(dataVolumePath, dogu),
		ExporterPort: c.config.ClassicExportPort,
	}
	return dE, nil
}

func (c *ClassicExportModeProvider) GetExportMode(ctx context.Context) (*exportModeStatus, error) {
	// Get all dogus
	dogus, err := c.execClient.GetAllDogus()
	if err != nil {
		return nil, err
	}

	// Filter dogus that are installed but not started - nether healthy nor unhealthy
	var dockercontainers = make(map[string]bool)
	containers, _ := c.execClient.ContainerList(ctx, container.ListOptions{})
	for _, con := range containers {
		containername := strings.Split(strings.Join(con.Names, ","), "/")[1]
		dockercontainers[containername] = true
	}

	healthy := true
	for _, dogu := range dogus {
		_, ok := dockercontainers[dogu]
		if !ok {
			slog.Info(fmt.Sprintf("skipping %s because it is not running", dogu))
			continue
		}

		cj, err := c.execClient.ContainerInspect(ctx, dogu)
		if err != nil {
			slog.Error("error getting container json", "err", err)
			healthy = false
			break
		}

		slog.Info(fmt.Sprintf("%s: %s", dogu, cj.State.Health.Status))
		if "healthy" != cj.State.Health.Status {
			healthy = false
			// break on first unhealthy dogu
			slog.Warn(fmt.Sprintf("dogu %s is %s", dogu, cj.State.Health.Status))
			break
		}
	}
	// since we did not step out until now - the global export-mode-status is true
	return &exportModeStatus{IsActive: healthy}, nil

}

func (c *ClassicExportModeProvider) checkDogu(dogu string) (string, error) {
	dogus, err := c.execClient.GetAllDogus()
	if err != nil {
		return "", fmt.Errorf("failed to get dogu list: %w", err)
	}
	found := false
	for _, d := range dogus {
		if d == dogu {
			found = true
			break
		}
	}
	if !found {
		return "", fmt.Errorf("can not get dogu %s", dogu)
	}
	return dogu, nil
}

func (e *ExecClient) GetAllDogus() ([]string, error) {
	return etcd.GetAllDogus()
}

func (e *ExecClient) ContainerList(ctx context.Context, options container.ListOptions) ([]types.Container, error) {
	// Get Docker client
	docker, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		slog.Error("error getting docker client", "err", err)
	}
	return docker.ContainerList(ctx, container.ListOptions{})
}

func (e *ExecClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	// Get Docker client
	docker, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		slog.Error("error getting docker client", "err", err)
	}
	return docker.ContainerInspect(ctx, containerID)
}
