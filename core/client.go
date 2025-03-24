package core

import (
	"context"
	"fmt"
	bup "github.com/cloudogu/k8s-backup-operator/pkg/api/v1"
	rclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type RuntimeClient interface {
	rclient.Client
}

func NewBackupScheduleRuntimeClient(rclient RuntimeClient, namespace string) *BackupScheduleRuntimeClient {
	return &BackupScheduleRuntimeClient{
		rclient,
		namespace,
	}
}

type BackupScheduleRuntimeClient struct {
	rclient   RuntimeClient
	namespace string
}

func (b *BackupScheduleRuntimeClient) ListBackupSchedules() (*bup.BackupScheduleList, error) {
	var backupSchedules bup.BackupScheduleList

	err := b.rclient.List(context.TODO(), &backupSchedules, rclient.InNamespace(b.namespace))
	if err != nil {
		return nil, fmt.Errorf("failed to list backup schedules: %w", err)
	}

	return &backupSchedules, nil
}
