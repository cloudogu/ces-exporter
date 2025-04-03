package export

import (
	"bytes"
	"fmt"
	v2 "github.com/cloudogu/k8s-dogu-operator/v3/api/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestInvalidCron(t *testing.T) {
	t.Run("cron expression is invalid", func(t *testing.T) {
		expr := "invalid"
		cronjob := NewCronJob(expr, nil, "ecosystem")

		err := cronjob.Run()
		require.Contains(t, err.Error(), "configured exporter cron expression 'invalid' is invalid")
	})
}

func TestCallCronJob(t *testing.T) {
	t.Run("call cronjob without dogus", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, nil))
		slog.SetDefault(logger)
		expr := "* * * * *"
		doguClient := NewMockDoguClientInterface(t)
		doguClient.EXPECT().List(mock.Anything).Return(&v2.DoguList{
			Items: []v2.Dogu{},
		}, nil)
		cronjob := NewCronJob(expr, doguClient, "ecosystem")

		cronjob.callCronJob()
		require.Contains(t, buf.String(), "start export mode cronjob due to timetable")
	})
	t.Run("call cronjob with dogus in export mode", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, nil))
		slog.SetDefault(logger)
		expr := "* * * * *"
		doguClient := NewMockDoguClientInterface(t)
		doguClient.EXPECT().List(mock.Anything).Return(&v2.DoguList{
			Items: []v2.Dogu{
				{
					Spec: v2.DoguSpec{
						Name:       "test_A",
						ExportMode: true,
					},
				},
				{
					Spec: v2.DoguSpec{
						Name:       "test_B",
						ExportMode: true,
					},
				},
			},
		}, nil)
		cronjob := NewCronJob(expr, doguClient, "ecosystem")

		_, _ = cronjob.callCronJob()

		require.Contains(t, buf.String(), "start export mode cronjob due to timetable")
		require.NotContains(t, buf.String(), "Activate export mode for dogu")
	})
	t.Run("call cronjob with dogus not in export mode", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, nil))
		slog.SetDefault(logger)
		expr := "* * * * *"
		doguClient := NewMockDoguClientInterface(t)
		doguClient.EXPECT().List(mock.Anything).Return(&v2.DoguList{
			Items: []v2.Dogu{
				{
					Spec: v2.DoguSpec{
						Name:       "test_A",
						ExportMode: false,
					},
				},
				{
					Spec: v2.DoguSpec{
						Name:       "test_B",
						ExportMode: true,
					},
				},
			},
		}, nil)
		// no error
		doguClient.EXPECT().Update(mock.Anything, mock.Anything).Return(nil, nil)

		cronjob := NewCronJob(expr, doguClient, "ecosystem")

		_, _ = cronjob.callCronJob()
		require.Contains(t, buf.String(), "start export mode cronjob due to timetable")
		require.Contains(t, buf.String(), "Activate export mode for dogu 'test_A'")
		require.NotContains(t, buf.String(), "Activate export mode for dogu 'test_B'")
	})
	t.Run("call cronjob with error on updating dogu", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, nil))
		slog.SetDefault(logger)
		expr := "* * * * *"
		doguClient := NewMockDoguClientInterface(t)
		doguClient.EXPECT().List(mock.Anything).Return(&v2.DoguList{
			Items: []v2.Dogu{
				{
					Spec: v2.DoguSpec{
						Name:       "test_A",
						ExportMode: false,
					},
				},
				{
					Spec: v2.DoguSpec{
						Name:       "test_B",
						ExportMode: true,
					},
				},
			},
		}, nil)
		// no error
		doguClient.EXPECT().Update(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("testerror"))

		cronjob := NewCronJob(expr, doguClient, "ecosystem")

		_, _ = cronjob.callCronJob()
		require.Contains(t, buf.String(), "start export mode cronjob due to timetable")
		require.Contains(t, buf.String(), "Activate export mode for dogu 'test_A'")
		require.NotContains(t, buf.String(), "Activate export mode for dogu 'test_B'")
		require.Contains(t, buf.String(), "Error while activating export mode\" err=testerror")
	})
}

func TestRunCronJob(t *testing.T) {
	t.Run("start tasker within one minute", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, nil))
		slog.SetDefault(logger)
		expr := "* * * * *"
		doguClient := NewMockDoguClientInterface(t)
		doguClient.EXPECT().List(mock.Anything).Return(&v2.DoguList{
			Items: []v2.Dogu{
				{
					Spec: v2.DoguSpec{
						Name:       "test_A",
						ExportMode: false,
					},
				},
				{
					Spec: v2.DoguSpec{
						Name:       "test_B",
						ExportMode: true,
					},
				},
			},
		}, nil)

		doguClient.EXPECT().Update(mock.Anything, mock.Anything).Return(nil, nil)

		cronjob := NewCronJob(expr, doguClient, "ecosystem")

		_ = os.Setenv(CronJobVerboseEnv, "true")

		go cronjob.Run()

		time.Sleep(65 * time.Second) // slightly more than 1 minute

		require.Contains(t, buf.String(), "Activate export mode for dogu 'test_A'")
		cronjob.Stop()
	})

	t.Run("start tasker and fail on first execute", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, nil))
		slog.SetDefault(logger)
		expr := "* * * * *"
		doguClient := NewMockDoguClientInterface(t)
		doguClient.EXPECT().List(mock.Anything).Return(nil, fmt.Errorf("testerror"))

		cronjob := NewCronJob(expr, doguClient, "ecosystem")

		_ = os.Setenv(CronJobVerboseEnv, "true")

		go func() {
			_ = cronjob.Run()
		}()

		time.Sleep(65 * time.Second) // slightly more than 1 minute

		require.NotContains(t, buf.String(), "Activate export mode for dogu 'test_A'")
		require.Contains(t, buf.String(), "Error while getting dogu list")
		cronjob.Stop()
	})
}
