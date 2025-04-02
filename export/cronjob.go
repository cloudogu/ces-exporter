package export

import (
	"context"
	"fmt"
	"github.com/adhocore/gronx"
	"github.com/adhocore/gronx/pkg/tasker"
	"log/slog"
)

const ExportCronJobEnv = "EXPORT_CRON"

type CronJobFunction func() (int, error)

type CronJob struct {
	expr string
}

func NewCronJob(expr string) *CronJob {
	return &CronJob{expr: expr}
}

func (cj *CronJob) IsValid() bool {
	return gronx.IsValid(cj.expr)
}

func (cj *CronJob) Run(call CronJobFunction) error {
	if !cj.IsValid() {
		return fmt.Errorf("Configured exporter cron expression %s is invalid", cj.expr)
	}

	taskr := tasker.New(tasker.Option{
		Verbose: false,
	})

	taskr.Task(cj.expr, func(ctx context.Context) (int, error) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("Panic at the disco: ", "err", err)
			}
		}()
		return call()
	})

	taskr.Run()

	return nil
}
