package export

import (
	"context"
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
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
		require.Equal(t, "/data", dogu.VolumePath)
	})
}
func TestCSetExportDogu(t *testing.T) {
	t.Run("should set export dogu", func(t *testing.T) {
		config := core.Configuration{ClassicExportPort: 7022}
		execClient := newMockExecClient(t)

		execClient.EXPECT().GetAllDogus().Return([]string{"test_A"}, nil)
		provider := ClassicExportModeProvider{config: config, execClient: execClient}

		dogu, _ := provider.SetExportDogu("test_A", context.Background())

		require.Equal(t, 7022, dogu.ExporterPort)
		require.Equal(t, "test_A", dogu.Dogu)
		require.Equal(t, "/data", dogu.VolumePath)
	})
	t.Run("fail on export non existent dogu", func(t *testing.T) {
		config := core.Configuration{ClassicExportPort: 7022}
		execClient := newMockExecClient(t)

		execClient.EXPECT().GetAllDogus().Return([]string{"test_A"}, nil)
		provider := ClassicExportModeProvider{config: config, execClient: execClient}

		_, err := provider.SetExportDogu("test_B", context.Background())

		require.Contains(t, err.Error(), "can not get dogu test_B")
	})
}

func TestCGetExportMode(t *testing.T) {
	t.Run("should return get export mode as false", func(t *testing.T) {
		config := core.Configuration{ClassicExportPort: 7022}
		execClient := newMockExecClient(t)

		execClient.EXPECT().GetAllDogus().Return([]string{"test_A"}, nil)
		execClient.EXPECT().ContainerList(mock.Anything, mock.Anything).Return([]types.Container{types.Container{
			Names: []string{"/test_A"},
		}}, nil)
		var state = types.ContainerJSON{
			ContainerJSONBase: &types.ContainerJSONBase{
				State: &types.ContainerState{
					Health: &types.Health{
						Status: "unknown",
					},
				},
			},
		}
		execClient.EXPECT().ContainerInspect(mock.Anything, "test_A").Return(state, nil)

		provider := ClassicExportModeProvider{config: config, execClient: execClient}

		mode, _ := provider.GetExportMode(context.Background())

		require.Equal(t, false, mode.IsActive)
	})
	t.Run("should return get export mode as true", func(t *testing.T) {
		config := core.Configuration{ClassicExportPort: 7022}
		execClient := newMockExecClient(t)

		execClient.EXPECT().GetAllDogus().Return([]string{"test_A", "test_B"}, nil)
		execClient.EXPECT().ContainerList(mock.Anything, mock.Anything).Return([]types.Container{types.Container{
			Names: []string{"/test_A"},
		}}, nil)
		var state = types.ContainerJSON{
			ContainerJSONBase: &types.ContainerJSONBase{
				State: &types.ContainerState{
					Health: &types.Health{
						Status: "healthy",
					},
				},
			},
		}
		execClient.EXPECT().ContainerInspect(mock.Anything, "test_A").Return(state, nil)

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
		execClient.EXPECT().ContainerList(mock.Anything, mock.Anything).Return([]types.Container{types.Container{
			Names: []string{"/test_A"},
		}}, nil)

		execClient.EXPECT().ContainerInspect(mock.Anything, "test_A").Return(types.ContainerJSON{}, fmt.Errorf("testerror"))

		provider := ClassicExportModeProvider{config: config, execClient: execClient}

		mode, _ := provider.GetExportMode(context.Background())

		require.Equal(t, false, mode.IsActive)
	})

}
