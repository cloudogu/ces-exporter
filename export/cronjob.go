package export

import (
	"context"
	"fmt"
	"github.com/adhocore/gronx"
	"github.com/adhocore/gronx/pkg/tasker"
	"log/slog"
	"os"
)

const CronJobEnv = "EXPORT_CRON"
const CronJobVerboseEnv = "EXPORT_CRON_VERBOSE"

type CronJobFunction func() (int, error)

type CronJob struct {
	namespace  string
	doguClient DoguClientInterface
	expr       string
}

func NewCronJob(expr string, doguClient DoguClientInterface, namespace string) *CronJob {
	return &CronJob{
		namespace:  namespace,
		doguClient: doguClient,
		expr:       expr,
	}
}

func (cj *CronJob) Run() error {
	if !gronx.IsValid(cj.expr) {
		return fmt.Errorf("configured exporter cron expression '%s' is invalid", cj.expr)
	}
	verbose := os.Getenv(CronJobVerboseEnv) == "true"
	taskr := tasker.New(tasker.Option{
		Verbose: verbose,
	})

	taskr.Task(cj.expr, func(ctx context.Context) (int, error) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("Error while setting export mode: ", "err", err)
			}
		}()
		return cj.callCronJob()
	})

	taskr.Run()

	return nil
}

/* this handles the actual exporter */
func (cj *CronJob) callCronJob() (int, error) {
	slog.Info("start export mode cronjob due to timetable ")
	ctx := context.Background()
	dogus, _ := cj.doguClient.List(ctx)
	for _, d := range dogus.Items {
		if !d.Spec.ExportMode {
			slog.Info(fmt.Sprintf("Activate export mode for dogu '%s'", d.Name))
			d.Spec.ExportMode = true
			_, err := cj.doguClient.Update(ctx, &d)
			if err != nil {
				slog.Error("Error while activating export mode", "err", err)
			}
		}
	}

	// the cron job do not fail. All errors will be logged
	return 0, nil
}
