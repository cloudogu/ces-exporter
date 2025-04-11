package export

import (
	"context"
	"fmt"
	"github.com/adhocore/gronx"
	"github.com/adhocore/gronx/pkg/tasker"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log/slog"
)

type CronJobFunction func() (int, error)

type CronJob struct {
	namespace   string
	doguClient  doguClient
	expr        string
	taskr       *tasker.Tasker
	verboseCron bool
}

func NewCronJob(expr string, ecosystemClient doguClient, namespace string, verboseCron bool) *CronJob {
	return &CronJob{
		namespace:   namespace,
		doguClient:  ecosystemClient,
		expr:        expr,
		verboseCron: verboseCron,
	}
}

func (cj *CronJob) Run() error {
	if !gronx.IsValid(cj.expr) {
		return fmt.Errorf("configured exporter cron expression '%s' is invalid", cj.expr)
	}
	cj.taskr = tasker.New(tasker.Option{
		Verbose: cj.verboseCron,
	})

	cj.taskr.Task(cj.expr, func(ctx context.Context) (int, error) {
		return cj.callCronJob(ctx)
	})

	cj.taskr.Run()

	return nil
}

func (cj *CronJob) Stop() {
	if cj.taskr != nil && cj.taskr.Running() {
		cj.taskr.Stop()
	}
}

/* this handles the actual exporter */
func (cj *CronJob) callCronJob(ctx context.Context) (int, error) {
	slog.Info("start export mode cronjob due to timetable ")
	dogus, err := cj.doguClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		slog.Error("Error while getting dogu list", "err", err)
		return 0, nil
	}
	for _, d := range dogus.Items {
		if !d.Spec.ExportMode {
			slog.Info(fmt.Sprintf("Activate export mode for dogu '%s'", d.Spec.Name))
			d.Spec.ExportMode = true
			_, err := cj.doguClient.Update(ctx, &d, metav1.UpdateOptions{})
			if err != nil {
				slog.Error("Error while activating export mode", "err", err)
			}
		}
	}

	// the cron job do not fail. All errors will be logged
	slog.Info("export mode cron job finished")
	return 0, nil
}
