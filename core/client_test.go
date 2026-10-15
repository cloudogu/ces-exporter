package core

import (
	"context"
	"fmt"
	"testing"

	bup "github.com/cloudogu/k8s-backup-lib/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	rclient "sigs.k8s.io/controller-runtime/pkg/client"
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

func TestListBackupSchedulesUnavailableAPI(t *testing.T) {
	resource := schema.GroupResource{Group: "k8s.cloudogu.com", Resource: "backupschedules"}
	kind := &meta.NoKindMatchError{GroupKind: schema.GroupKind{Group: resource.Group, Kind: "BackupSchedule"}}
	for _, tc := range []struct {
		name string
		err  error
	}{
		{name: "no matching kind found", err: kind},
		{name: "no matching resource found", err: &meta.NoResourceMatchError{PartialResource: resource.WithVersion("v1")}},
		{name: "resource not found", err: apierrors.NewNotFound(resource, "")},
		{name: "wrapped no match error returned", err: fmt.Errorf("discovery failed: %w", kind)},
		{name: "wrapped not found error returned", err: fmt.Errorf("request failed: %w", apierrors.NewNotFound(resource, ""))},
		{name: "installed API, but no schedules"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rtc := NewMockRuntimeClient(t)
			rtc.EXPECT().List(mock.Anything, mock.Anything, mock.Anything).Return(tc.err).Once()
			scheduleClient := NewBackupScheduleRuntimeClient(rtc, "namespace")

			schedules, err := scheduleClient.ListBackupSchedules(context.Background())

			require.NoError(t, err)
			require.NotNil(t, schedules)
			assert.Empty(t, schedules.Items)
		})
	}
}

func TestListBackupSchedulesPreservesErrors(t *testing.T) {
	resource := schema.GroupResource{Group: "k8s.cloudogu.com", Resource: "backupschedules"}
	for _, tc := range []struct {
		name string
		err  error
	}{
		{name: "API error: forbidden", err: apierrors.NewForbidden(resource, "", assert.AnError)},
		{name: "API error: unauthorized", err: apierrors.NewUnauthorized("authentication failed")},
		{name: "API error: server error", err: apierrors.NewInternalError(assert.AnError)},
		{name: "API error: service unavailable", err: apierrors.NewServiceUnavailable("server unavailable")},
		{name: "API error: timeout", err: apierrors.NewTimeoutError("request timed out", 1)},
		{name: "connection error", err: assert.AnError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rtc := NewMockRuntimeClient(t)
			rtc.EXPECT().List(mock.Anything, mock.Anything, mock.Anything).Return(tc.err).Once()
			scheduleClient := NewBackupScheduleRuntimeClient(rtc, "namespace")

			schedules, err := scheduleClient.ListBackupSchedules(context.Background())

			require.Error(t, err)
			assert.ErrorIs(t, err, tc.err)
			assert.ErrorContains(t, err, "failed to list backup schedules")
			assert.Nil(t, schedules)
		})
	}
}
