package core

import (
	"context"
	"fmt"
	bup "github.com/cloudogu/k8s-backup-lib/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rclient "sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func TestListBackupSchedules(t *testing.T) {
	t.Run("can list backup schedules successfully", func(t *testing.T) {
		rtc := NewMockRuntimeClient(t)
		rtc.EXPECT().List(mock.Anything, mock.Anything, mock.Anything).Run(func(ctx context.Context, list rclient.ObjectList, opts ...rclient.ListOption) {
			backupList, ok := list.(*bup.BackupScheduleList)
			if !ok {
				require.Fail(t, "invalid input param for backup list")
			}

			*backupList = bup.BackupScheduleList{
				Items: []bup.BackupSchedule{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "schedulename",
						},
						Spec: bup.BackupScheduleSpec{
							Schedule: "12345",
						},
					},
				},
			}
		}).Return(nil)
		scheduleClient := NewBackupScheduleRuntimeClient(rtc, "namespace")

		schedules, err := scheduleClient.ListBackupSchedules(context.TODO())
		assert.NoError(t, err)

		require.Len(t, schedules.Items, 1)
		assert.Equal(t, "schedulename", schedules.Items[0].Name)
		assert.Equal(t, "12345", schedules.Items[0].Spec.Schedule)
	})
	t.Run("fail on list backups", func(t *testing.T) {
		rtc := NewMockRuntimeClient(t)
		rtc.EXPECT().List(mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("testerror"))
		scheduleClient := NewBackupScheduleRuntimeClient(rtc, "namespace")

		schedules, err := scheduleClient.ListBackupSchedules(context.TODO())
		assert.Errorf(t, err, "")
		assert.Nil(t, schedules)
	})
}
