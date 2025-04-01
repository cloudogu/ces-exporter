package export

import (
	"github.com/adhocore/gronx"
)

type CronJob struct {
	expr string
}

func NewCronJob(expr string) *CronJob {
	return &CronJob{expr: expr}
}

func (cj *CronJob) IsValid() bool {
	return gronx.IsValid(cj.expr)
}
