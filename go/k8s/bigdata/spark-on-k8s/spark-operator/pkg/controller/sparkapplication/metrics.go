package sparkapplication

import v1 "k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/apis/sparkoperator.k9s.io/v1"

type sparkAppMetrics struct {
}

func (metrics *sparkAppMetrics) exportMetrics(oldApp, newApp *v1.SparkApplication) {

}
