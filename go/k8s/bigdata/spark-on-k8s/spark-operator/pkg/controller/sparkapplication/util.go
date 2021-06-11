package sparkapplication

import (
	"encoding/json"
	"fmt"

	v1 "k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/apis/sparkoperator.k9s.io/v1"
	"k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/config"
)

func getDriverPodName(app *v1.SparkApplication) string {
	name := app.Spec.Driver.PodName
	if name != nil && len(*name) > 0 {
		return *name
	}

	sparkConf := app.Spec.SparkConf
	if sparkConf[config.SparkDriverPodNameKey] != "" {
		return sparkConf[config.SparkDriverPodNameKey]
	}

	return fmt.Sprintf("%s-driver", app.Name)
}

// INFO: 这个函数可以复用
func printStatus(status *v1.SparkApplicationStatus) (string, error) {
	marshalled, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return "", err
	}
	return string(marshalled), nil
}
