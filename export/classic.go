package export

import (
	"context"
	"fmt"
	"github.com/cloudogu/ces-exporter/configuration"
	dtypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"log/slog"
	"strings"
)

const ()

type ClassicExportModeProvider struct {
	currentExportDogu string
}

func NewClassicExportModeProvider() *ClassicExportModeProvider {
	return &ClassicExportModeProvider{}
}

// GetExportDogu gets the dogu.name currently set in the ces-exporter-dogu-exporter service
func (c *ClassicExportModeProvider) GetExportDogu(ctx context.Context) (*doguExport, error) {
	dogu, err := checkDogu(c.currentExportDogu)
	if err != nil {
		return nil, err
	}

	doguExport := doguExport{
		Dogu:         dogu,
		VolumePath:   dataVolumePath,
		ExporterPort: 7000,
	}
	return &doguExport, nil
}

// SetExportDogu sets the given dogu as dogu.name in the ces-exporter-dogu-exporter service
func (c *ClassicExportModeProvider) SetExportDogu(doguName string, ctx context.Context) (*doguExport, error) {
	dogu, err := checkDogu(doguName)
	if err != nil {
		return nil, err
	}
	c.currentExportDogu = dogu

	doguExport := doguExport{
		Dogu:         dogu,
		VolumePath:   dataVolumePath,
		ExporterPort: 7000,
	}
	return &doguExport, nil
}

func (c *ClassicExportModeProvider) GetExportMode(ctx context.Context) (*exportModeStatus, error) {
	// Get all dogus
	dogus, err := configuration.GetAllDogus()
	if err != nil {
		return nil, err
	}

	// Get Docker client
	docker, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		slog.Error("error getting docker client", "err", err)
	}

	// Filter dogus that are installed but not started - nether healthy nor unhealthy
	var dockercontainers = make(map[string]bool)
	containers, _ := docker.ContainerList(ctx, dtypes.ListOptions{})
	for _, container := range containers {
		containername := strings.Split(strings.Join(container.Names, ","), "/")[1]
		dockercontainers[containername] = true
	}

	healthy := true
	for _, dogu := range dogus {
		_, ok := dockercontainers[dogu]
		if !ok {
			slog.Info(fmt.Sprintf("skipping %s because it is not running", dogu))
			continue
		}

		cj, err := docker.ContainerInspect(ctx, dogu)
		if err != nil {
			slog.Error("error getting container json", "err", err)
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

func checkDogu(dogu string) (string, error) {
	dogus, err := configuration.GetAllDogus()
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
