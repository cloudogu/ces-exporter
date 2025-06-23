package systeminfo

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/cloudogu/ces-exporter/etcd"
	cesLibCore "github.com/cloudogu/cesapp-lib/core"
	"math"
	"os/exec"
	"strings"
)

var (
	errFqdnGlobalConfigKeyNotFound = errors.New("global config key for fqdn not found")
)

var _ SystemInfoProvider = (*SingleNodeSystemInfoProvider)(nil)

type getGlobalConfig func(ignoreKeys []string) (core.GlobalConfig, error)

// doguGetter defines methods to retrieve installed Dogus and their specifications.
type doguGetter interface {
	GetAllDogus() ([]string, error)
	GetDoguSpec(dogu string) (cesLibCore.Dogu, error)
}

// singleNodeDoguGetter implements doguGetter by using the etcd backend.
type singleNodeDoguGetter struct{}

func (s singleNodeDoguGetter) GetAllDogus() ([]string, error) {
	return etcd.GetAllDogus()
}

func (s singleNodeDoguGetter) GetDoguSpec(dogu string) (cesLibCore.Dogu, error) {
	return etcd.GetDoguSpec(dogu)
}

type volumeSizeGetter interface {
	GetVolumeSizeInBytes(dogu string) (int64, error)
}

type commandExecutor interface {
	Output() ([]byte, error)
}

type singleNodeVolumeSizeGetter struct {
	volumeBasePath string
	execCommand    func(command string, args ...string) commandExecutor
}

func newSingleNodeVolumeSizeGetter(volumeBasePath string) *singleNodeVolumeSizeGetter {
	return &singleNodeVolumeSizeGetter{
		volumeBasePath: volumeBasePath,
		execCommand: func(command string, args ...string) commandExecutor {
			return exec.Command(command, args...)
		},
	}
}

// GetVolumeSizeInBytes returns the total referenced size of a Dogu volume using `btrfs filesystem du --raw` command.
func (s singleNodeVolumeSizeGetter) GetVolumeSizeInBytes(doguName string) (int64, error) {
	path := fmt.Sprintf("%s/%s", s.volumeBasePath, doguName)
	cmd := s.execCommand("btrfs", "filesystem", "du", "-s", "--raw", path)

	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to run btrfs filesystem du: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("unexpected output for btrfs command (not enough lines): %s", string(out))
	}

	// Skip header, parse second line
	fields := strings.Fields(lines[1])
	if len(fields) < 1 {
		return 0, fmt.Errorf("unexpected data format: %v", fields)
	}

	// read the column for total size
	var sizeInBytes int64
	_, err = fmt.Sscanf(fields[0], "%d", &sizeInBytes)
	if err != nil {
		return 0, fmt.Errorf("failed to parse size in bytes: %w", err)
	}

	return sizeInBytes, nil
}

// SingleNodeSystemInfoProvider provides system information for single-node CES installations.
type SingleNodeSystemInfoProvider struct {
	getGlobalConfig
	doguGetter
	volumeSizeGetter
	volumeIncreaseFactor float32
}

// NewSingleNodeSystemInfoProvider creates a new SingleNodeSystemInfoProvider using etcd and Btrfs.
func NewSingleNodeSystemInfoProvider(volumeBasePath string, volumeIncreaseFactor float32) *SingleNodeSystemInfoProvider {
	return &SingleNodeSystemInfoProvider{
		getGlobalConfig:      etcd.GetGlobalConfig,
		doguGetter:           &singleNodeDoguGetter{},
		volumeSizeGetter:     newSingleNodeVolumeSizeGetter(volumeBasePath),
		volumeIncreaseFactor: volumeIncreaseFactor,
	}
}

func (sp SingleNodeSystemInfoProvider) getComponents(_ context.Context) ([]component, error) {
	return []component{}, nil
}

func (sp SingleNodeSystemInfoProvider) isMultinode() bool {
	return false
}

func (sp SingleNodeSystemInfoProvider) getFqdn(_ context.Context) (string, error) {
	globalCfg, err := sp.getGlobalConfig(nil)
	if err != nil {
		return "", fmt.Errorf("error getting global config: %w", err)
	}

	for _, kvPair := range globalCfg {
		if kvPair.Key == globalConfigKeyFqdn {
			return kvPair.Value, nil
		}
	}

	return "", errFqdnGlobalConfigKeyNotFound
}

func (sp SingleNodeSystemInfoProvider) getDogus(_ context.Context) ([]dogu, error) {
	doguList, err := sp.GetAllDogus()
	if err != nil {
		return nil, fmt.Errorf("error getting currently installed dogus: %w", err)
	}

	doguSystemInfoList := make([]dogu, 0, len(doguList))

	for _, d := range doguList {
		doguSpec, lErr := sp.GetDoguSpec(d)
		if lErr != nil {
			return nil, fmt.Errorf("error getting doguSpec for dogu %s: %w", d, lErr)
		}

		volumeSize, lErr := sp.GetVolumeSizeInBytes(d)
		if lErr != nil {
			return nil, fmt.Errorf("error getting volume size for dogu %s: %w", d, lErr)
		}

		targetVolumeSize := calculateTargetVolumeSize(sp.volumeIncreaseFactor, volumeSize)

		doguSystemInfoList = append(doguSystemInfoList, dogu{
			Name:    doguSpec.Name,
			Version: doguSpec.Version,
			Volume:  volume{SizeInBytes: targetVolumeSize},
		})
	}

	return doguSystemInfoList, nil
}

// calculateTargetVolumeSize increases the actual volume size by the given factor,
// and rounds the result up to the next full GiB.
func calculateTargetVolumeSize(increaseFactor float32, actualSize int64) int64 {
	const GiB = 1024 * 1024 * 1024

	estimated := float64(actualSize) * float64(increaseFactor)

	if estimated == 0 {
		return GiB
	}

	// Round up to the next GiB
	rounded := math.Ceil(estimated/float64(GiB)) * float64(GiB)

	return int64(rounded)
}
