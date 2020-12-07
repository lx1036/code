package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
)

type Collectors struct {
	ConfigMapCounter  *prometheus.CounterVec
	SecretCounter     *prometheus.CounterVec
	DeploymentCounter *prometheus.CounterVec
}

func NewCollectors() Collectors {
	configMapCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "trigger",
			Name:      "configMap",
			Help:      "Counter of configMap",
		},
		[]string{"configMap"},
	)
	//set 0 as default value
	configMapCounter.With(prometheus.Labels{"configMap": "success"}).Add(0)
	configMapCounter.With(prometheus.Labels{"configMap": "fail"}).Add(0)

	secretCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "trigger",
			Name:      "secret",
			Help:      "Counter of secret",
		},
		[]string{"secret"},
	)
	secretCounter.With(prometheus.Labels{"secret": "success"}).Add(0)
	secretCounter.With(prometheus.Labels{"secret": "fail"}).Add(0)

	deploymentCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "trigger",
			Name:      "deployment",
			Help:      "Counter of deployment",
		},
		[]string{"deployment"},
	)
	deploymentCounter.With(prometheus.Labels{"deployment": "success"}).Add(0)
	deploymentCounter.With(prometheus.Labels{"deployment": "fail"}).Add(0)

	return Collectors{
		ConfigMapCounter:  configMapCounter,
		SecretCounter:     secretCounter,
		DeploymentCounter: deploymentCounter,
	}
}

func SetupPrometheusEndpoint() Collectors {
	collectors := NewCollectors()
	prometheus.MustRegister(collectors.ConfigMapCounter)
	prometheus.MustRegister(collectors.SecretCounter)
	prometheus.MustRegister(collectors.DeploymentCounter)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		logrus.Fatal(http.ListenAndServe(":8001", nil))
	}()

	return collectors
}
