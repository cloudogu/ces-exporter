package export

import (
	"context"
	"fmt"
	"testing"

	"github.com/cloudogu/ces-exporter/core"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCGetExportDogu(t *testing.T) {
	t.Run("should return get export dogu", func(t *testing.T) {
		config := core.Configuration{ClassicExportPort: 7022}
		execClient := newMockExecClient(t)

		execClient.EXPECT().GetAllDogus().Return([]string{"test_A"}, nil)

		provider := ClassicExportModeProvider{config: config, execClient: execClient}

		provider.currentExportDogu = "test_A"

		dogu, _ := provider.GetExportDogu(context.Background())

		require.Equal(t, 7022, dogu.ExporterPort)
		require.Equal(t, "test_A", dogu.Dogu)
		require.Equal(t, "/data/test_A/volumes", dogu.VolumePath)
	})
}

func TestCSetExportDogu(t *testing.T) {
	t.Run("should set export dogu", func(t *testing.T) {
		config := core.Configuration{ClassicExportPort: 7022}
		execClient := newMockExecClient(t)

		execClient.EXPECT().GetAllDogus().Return([]string{"test_A"}, nil)
		provider := ClassicExportModeProvider{config: config, execClient: execClient}

		dogu, _ := provider.SetExportDogu(context.Background(), "test_A")

		require.Equal(t, 7022, dogu.ExporterPort)
		require.Equal(t, "test_A", dogu.Dogu)
		require.Equal(t, "/data/test_A/volumes", dogu.VolumePath)
	})
	t.Run("fail on export non existent dogu", func(t *testing.T) {
		config := core.Configuration{ClassicExportPort: 7022}
		execClient := newMockExecClient(t)

		execClient.EXPECT().GetAllDogus().Return([]string{"test_A"}, nil)
		provider := ClassicExportModeProvider{config: config, execClient: execClient}

		_, err := provider.SetExportDogu(context.Background(), "test_B")

		require.Contains(t, err.Error(), "can not get dogu test_B")
	})
}

func TestCGetExportMode(t *testing.T) {
	t.Run("should return get export mode as false", func(t *testing.T) {
		config := core.Configuration{ClassicExportPort: 7022}
		execClient := newMockExecClient(t)

		execClient.EXPECT().GetAllDogus().Return([]string{"test_A"}, nil)
		execClient.EXPECT().ContainerList(mock.Anything, mock.Anything).Return(client.ContainerListResult{Items: []container.Summary{
			{Names: []string{"/test_A"}},
		}}, nil)
		var result = client.ContainerInspectResult{
			Container: container.InspectResponse{
				State: &container.State{
					Health: &container.Health{
						Status: "unknown",
					},
				},
			},
		}
		execClient.EXPECT().ContainerInspect(mock.Anything, "test_A").Return(result, nil)

		provider := ClassicExportModeProvider{config: config, execClient: execClient}

		mode, _ := provider.GetExportMode(context.Background())

		require.Equal(t, false, mode.IsActive)
	})
	t.Run("should return get export mode as true", func(t *testing.T) {
		config := core.Configuration{ClassicExportPort: 7022}
		execClient := newMockExecClient(t)

		execClient.EXPECT().GetAllDogus().Return([]string{"test_A", "test_B"}, nil)
		execClient.EXPECT().ContainerList(mock.Anything, mock.Anything).Return(client.ContainerListResult{Items: []container.Summary{
			{Names: []string{"/test_A"}},
		}}, nil)
		var result = client.ContainerInspectResult{
			Container: container.InspectResponse{
				State: &container.State{
					Health: &container.Health{
						Status: "healthy",
					},
				},
			},
		}
		execClient.EXPECT().ContainerInspect(mock.Anything, "test_A").Return(result, nil)

		provider := ClassicExportModeProvider{config: config, execClient: execClient}

		mode, _ := provider.GetExportMode(context.Background())

		require.Equal(t, true, mode.IsActive)
	})
	t.Run("should return error on export mode", func(t *testing.T) {
		config := core.Configuration{ClassicExportPort: 7022}
		execClient := newMockExecClient(t)

		execClient.EXPECT().GetAllDogus().Return([]string{"test_A"}, fmt.Errorf("testerror"))

		provider := ClassicExportModeProvider{config: config, execClient: execClient}

		_, err := provider.GetExportMode(context.Background())

		require.Contains(t, "error getting dogu list: testerror", err.Error())
	})
	t.Run("should return error on export mode in docker inspect", func(t *testing.T) {
		config := core.Configuration{ClassicExportPort: 7022}
		execClient := newMockExecClient(t)

		execClient.EXPECT().GetAllDogus().Return([]string{"test_A", "test_B"}, nil)
		execClient.EXPECT().ContainerList(mock.Anything, mock.Anything).Return(client.ContainerListResult{Items: []container.Summary{
			{Names: []string{"/test_A"}},
		}}, nil)

		execClient.EXPECT().ContainerInspect(mock.Anything, "test_A").Return(client.ContainerInspectResult{}, fmt.Errorf("testerror"))

		provider := ClassicExportModeProvider{config: config, execClient: execClient}

		mode, _ := provider.GetExportMode(context.Background())

		require.Equal(t, false, mode.IsActive)
	})

}
