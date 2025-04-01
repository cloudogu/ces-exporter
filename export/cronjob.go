package export

import (
	"context"
	"fmt"
	"github.com/adhocore/gronx"
	"github.com/adhocore/gronx/pkg/tasker"
	"log/slog"
)

const ExportCronJobEnv = "EXPORT_CRON"

type CronJob struct {
	expr string
}

func NewCronJob(expr string) *CronJob {
	return &CronJob{expr: expr}
}

func (cj *CronJob) IsValid() bool {
	return gronx.IsValid(cj.expr)
}

func (cj *CronJob) Run() error {
	if !cj.IsValid() {
		return fmt.Errorf("Configured exporter cron expression %s is invalid", cj.expr)
	}

	taskr := tasker.New(tasker.Option{
		Verbose: true,
	})

	taskr.Task(cj.expr, func(ctx context.Context) (int, error) {
		slog.Info("hello i run")
		return 0, nil
	})

	taskr.Run()

	if !taskr.Running() {
		return fmt.Errorf("Failed to start ExporterCronJob")
	}

	return nil
}
