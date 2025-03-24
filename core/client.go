package core

import (
	"context"
	"fmt"
	bup "github.com/cloudogu/k8s-backup-operator/pkg/api/v1"
	rclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// RuntimeClient wraps sigs.k8s.io/controller-runtime/pkg/client and is used to query custom resources
type RuntimeClient interface {
	rclient.Client
}

// NewBackupScheduleRuntimeClient creates a new NewBackupScheduleRuntimeClient instance
func NewBackupScheduleRuntimeClient(rclient RuntimeClient, namespace string) *BackupScheduleRuntimeClient {
	return &BackupScheduleRuntimeClient{
		rclient,
		namespace,
	}
}

// BackupScheduleRuntimeClient contains the RuntimeClient to query custom resources and is used to query the backup schedule custom resource
type BackupScheduleRuntimeClient struct {
	rclient   RuntimeClient
	namespace string
}

// ListBackupSchedules returns a list of all backup schedule custom resources
func (b *BackupScheduleRuntimeClient) ListBackupSchedules(ctx context.Context) (*bup.BackupScheduleList, error) {
	var backupSchedules bup.BackupScheduleList

	err := b.rclient.List(ctx, &backupSchedules, rclient.InNamespace(b.namespace))
	if err != nil {
		return nil, fmt.Errorf("failed to list backup schedules: %w", err)
	}

	return &backupSchedules, nil
}
