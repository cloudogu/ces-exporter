package configuration

import (
	"errors"
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/cloudogu/ces-exporter/etcd"
	"log/slog"
	"slices"
	"strings"
)

const (
	scheduledBackupName = "scheduled-backup"
)

type backupScheduleProvider struct {
	getConfig getConfigFunc
}

func newBackupScheduleProvider(getConfig getConfigFunc) backupScheduleProvider {
	return backupScheduleProvider{
		getConfig: getConfig,
	}
}

// convertTimeToCron converts a time in format HH:MM to a cron pattern
func convertTimeToCron(time string) string {
	parts := strings.Split(time, ":")
	if len(parts) != 2 {
		return ""
	}
	hour := parts[0]
	minute := parts[1]
	return fmt.Sprintf("%s %s * * *", minute, hour)
}

func (e *backupScheduleProvider) getBackupSchedules() ([]core.BackupSchedule, error) {
	keys, err := e.getConfig("backup", []string{})
	if err != nil {
		if errors.Is(err, etcd.ErrDoguNotFound) {
			return []core.BackupSchedule{}, nil
		}

		return nil, fmt.Errorf("failed to get backup config: %w", err)
	}

	activeBackup := slices.ContainsFunc(keys.NormalConfig, func(kv core.KeyValue) bool {
		return kv.Key == "/active" && kv.Value == "true"
	})

	if !activeBackup {
		slog.Info("backup is not active, skipping backup schedule.")
		return []core.BackupSchedule{}, nil
	}

	for _, kv := range keys.NormalConfig {
		if kv.Key == "/time" {
			cron := convertTimeToCron(kv.Value)
			if cron != "" {
				return []core.BackupSchedule{
					{
						Name:     scheduledBackupName,
						Schedule: cron,
					},
				}, nil
			}
		}
	}

	return make([]core.BackupSchedule, 0), nil
}
