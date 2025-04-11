package configuration

import (
	"context"
	"github.com/cloudogu/ces-commons-lib/dogu"
	bup "github.com/cloudogu/k8s-backup-operator/pkg/api/v1"
	"github.com/cloudogu/k8s-registry-lib/config"
)

type keyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type globalConfig []keyValue

type doguConfig struct {
	Name            string     `json:"name"`
	NormalConfig    []keyValue `json:"normal"`
	LocalConfig     []keyValue `json:"local"`
	SensitiveConfig []keyValue `json:"sensitive"`
}

type backupSchedule struct {
	Name     string `json:"name"`
	Schedule string `json:"schedule"`
}

type configuration struct {
	GlobalConfig    globalConfig     `json:"global"`
	DoguConfigs     []doguConfig     `json:"dogus"`
	BackupSchedules []backupSchedule `json:"backupSchedules"`
}

type doguVersionRegistry interface {
	GetCurrentOfAll(ctx context.Context) ([]dogu.SimpleNameVersion, error)
}

type backupScheduleRuntimeClient interface {
	ListBackupSchedules(ctx context.Context) (*bup.BackupScheduleList, error)
}

type doguConfigRepository interface {
	Get(ctx context.Context, name dogu.SimpleName) (config.DoguConfig, error)
}

type globalConfigRepository interface {
	Get(ctx context.Context) (config.GlobalConfig, error)
}
