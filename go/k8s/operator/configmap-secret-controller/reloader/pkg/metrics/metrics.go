package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
)

type Collectors struct {
	Counter *prometheus.CounterVec
}

func NewCollectors() Collectors {
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "reloader",
			Name:      "reload_executed_total",
			Help:      "Counter of reloads executed by Reloader.",
		},
		[]string{"success"},
	)

	//set 0 as default value
	counter.With(prometheus.Labels{"success": "true"}).Add(0)
	counter.With(prometheus.Labels{"success": "false"}).Add(0)

	return Collectors{
		Counter: counter,
	}
}

func SetupPrometheusEndpoint() Collectors {
	collectors := NewCollectors()
	prometheus.MustRegister(collectors.Counter)

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		logrus.Fatal(http.ListenAndServe(":9090", nil))
	}()

	return collectors
}
