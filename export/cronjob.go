package export

import (
	"context"
	"fmt"
	"github.com/adhocore/gronx"
	"github.com/adhocore/gronx/pkg/tasker"
	ecoSystemV2 "github.com/cloudogu/k8s-dogu-operator/v3/api/ecoSystem"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log/slog"
	"os"
)

const CronJobEnv = "EXPORT_CRON"
const CronJobVerboseEnv = "EXPORT_CRON_VERBOSE"

type CronJobFunction func() (int, error)

type CronJob struct {
	namespace  string
	doguClient ecoSystemV2.EcoSystemV2Interface
	expr       string
}

func NewCronJob(expr string, doguClient ecoSystemV2.EcoSystemV2Interface, namespace string) *CronJob {
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
				slog.Error("Panic at the disco: ", "err", err)
			}
		}()
		return cj.callCronJob()
	})

	taskr.Run()

	return nil
}

// this handles the actual exporter
func (cj *CronJob) callCronJob() (int, error) {
	slog.Info("start export mode cronjob due to timetable ")
	dogus, _ := cj.doguClient.Dogus(cj.namespace).List(context.Background(), metav1.ListOptions{})
	for _, d := range dogus.Items {
		if !d.Spec.ExportMode {
			slog.Info(fmt.Sprintf("Activate export mode for dogu '%s'", d.Name))
			d.Spec.ExportMode = true
			_, err := cj.doguClient.Dogus(cj.namespace).Update(context.Background(), &d, metav1.UpdateOptions{})
			if err != nil {
				slog.Error("Error while activating export mode", err)
			}
		}
	}

	// the cron job do not fail. All errors will be logged
	return 0, nil
}
