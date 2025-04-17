package configuration

import (
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	"github.com/cloudogu/ces-exporter/etcd"
	"strings"
)

const (
	scheduledBackupName = "Scheduled Backup"
)

// convertTimeToCron converts a time in format HH:MM to a cron pattern
func convertTimeToCron(time string) string {
	parts := strings.Split(time, ":")
	if len(parts) != 2 {
		return "" // or handle error
	}
	hour := parts[0]
	minute := parts[1]
	return fmt.Sprintf("%s %s * * *", minute, hour)
}

func getBackupSchedules() ([]core.BackupSchedule, error) {
	keys, err := etcd.GetConfig("backup", []string{})
	if err != nil {
		return nil, fmt.Errorf("failed to get backup config: %w", err)
	}

	for _, kv := range keys.NormalConfig {
		if kv.Key == "/time" {
			return []core.BackupSchedule{{
				Name:     scheduledBackupName,
				Schedule: convertTimeToCron(kv.Value),
			}}, nil
		}
	}

	return make([]core.BackupSchedule, 0), nil
}
