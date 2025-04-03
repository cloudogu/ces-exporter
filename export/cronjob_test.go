package export

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"log/slog"
	"testing"
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
	t.Run("call cronjob", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, nil))
		slog.SetDefault(logger)

		slog.Info("hallo")
		fmt.Println(buf.String())
		expr := "* * * * *"
		doguClient := NewMockDoguClientInterface(t)
		doguClient.EXPECT().List(mock.Anything).Return(nil, nil)
		cronjob := NewCronJob(expr, doguClient, "ecosystem")

		cronjob.callCronJob()

	})
}
