package schedulerinterface

import v1 "k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/apis/sparkoperator.k9s.io/v1"

type BatchScheduler interface {
	Name() string

	ShouldSchedule(app *v1.SparkApplication) bool
	DoBatchSchedulingOnSubmission(app *v1.SparkApplication) error
	CleanupOnCompletion(app *v1.SparkApplication) error
}
