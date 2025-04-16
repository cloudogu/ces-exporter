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
		etcdAccess := newMockExecWrapper(t)

		etcdAccess.EXPECT().GetAllDogus().Return([]string{"test_A"}, nil)

		provider := NewClassicExportModeProvider(config, etcdAccess)

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
		etcdAccess := newMockExecWrapper(t)

		etcdAccess.EXPECT().GetAllDogus().Return([]string{"test_A"}, nil)
		provider := NewClassicExportModeProvider(config, etcdAccess)

		dogu, _ := provider.SetExportDogu("test_A", context.Background())

		require.Equal(t, 7022, dogu.ExporterPort)
		require.Equal(t, "test_A", dogu.Dogu)
		require.Equal(t, "/data", dogu.VolumePath)
	})
	t.Run("fail on export non existent dogu", func(t *testing.T) {
		config := core.Configuration{ClassicExportPort: 7022}
		etcdAccess := newMockExecWrapper(t)

		etcdAccess.EXPECT().GetAllDogus().Return([]string{"test_A"}, nil)
		provider := NewClassicExportModeProvider(config, etcdAccess)

		_, err := provider.SetExportDogu("test_B", context.Background())

		require.Contains(t, err.Error(), "can not get dogu test_B")
	})
}

func TestCGetExportMode(t *testing.T) {
	t.Run("should return get export mode as false", func(t *testing.T) {
		config := core.Configuration{ClassicExportPort: 7022}
		remoteAccess := newMockExecWrapper(t)

		remoteAccess.EXPECT().GetAllDogus().Return([]string{"test_A"}, nil)
		remoteAccess.EXPECT().ContainerList(mock.Anything, mock.Anything).Return([]types.Container{types.Container{
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
		remoteAccess.EXPECT().ContainerInspect(mock.Anything, "test_A").Return(state, nil)

		provider := NewClassicExportModeProvider(config, remoteAccess)

		mode, _ := provider.GetExportMode(context.Background())

		require.Equal(t, false, mode.IsActive)
	})
	t.Run("should return get export mode as true", func(t *testing.T) {
		config := core.Configuration{ClassicExportPort: 7022}
		remoteAccess := newMockExecWrapper(t)

		remoteAccess.EXPECT().GetAllDogus().Return([]string{"test_A", "test_B"}, nil)
		remoteAccess.EXPECT().ContainerList(mock.Anything, mock.Anything).Return([]types.Container{types.Container{
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
		remoteAccess.EXPECT().ContainerInspect(mock.Anything, "test_A").Return(state, nil)

		provider := NewClassicExportModeProvider(config, remoteAccess)

		mode, _ := provider.GetExportMode(context.Background())

		require.Equal(t, true, mode.IsActive)
	})
	t.Run("should return error on export mode", func(t *testing.T) {
		config := core.Configuration{ClassicExportPort: 7022}
		remoteAccess := newMockExecWrapper(t)

		remoteAccess.EXPECT().GetAllDogus().Return([]string{"test_A"}, fmt.Errorf("testerror"))

		provider := NewClassicExportModeProvider(config, remoteAccess)

		_, err := provider.GetExportMode(context.Background())

		require.Contains(t, "error getting dogu list: testerror", err.Error())
	})
	t.Run("should return error on export mode in docker inspect", func(t *testing.T) {
		config := core.Configuration{ClassicExportPort: 7022}
		remoteAccess := newMockExecWrapper(t)

		remoteAccess.EXPECT().GetAllDogus().Return([]string{"test_A", "test_B"}, nil)
		remoteAccess.EXPECT().ContainerList(mock.Anything, mock.Anything).Return([]types.Container{types.Container{
			Names: []string{"/test_A"},
		}}, nil)

		remoteAccess.EXPECT().ContainerInspect(mock.Anything, "test_A").Return(types.ContainerJSON{}, fmt.Errorf("testerror"))

		provider := NewClassicExportModeProvider(config, remoteAccess)

		mode, _ := provider.GetExportMode(context.Background())

		require.Equal(t, false, mode.IsActive)
	})

}
