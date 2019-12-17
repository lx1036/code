package promhttp

import (
	"k8s-lx1036/app/prometheus/client-go/prometheus"
	"net/http"
	"time"
)

func Handler() http.Handler {
	return InstrumentMetricHandler(prometheus.DefaultRegisterer, HandlerFor(prometheus.DefaultGatherer, HandlerOpts{}))
}

func InstrumentMetricHandler(reg prometheus.Registerer, handler http.Handler) http.Handler {
	counterVec := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "promhttp_metric_handler_requests_total",
			Help: "Total number of scrapes by HTTP status code.",
		},
		[]string{"code"},
	)
	counterVec.WithLabelValues("200")
	counterVec.WithLabelValues("500")
	counterVec.WithLabelValues("503")
	if err := reg.Register(counterVec); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			counterVec = are.ExistingCollector.(*prometheus.CounterVec)
		} else {
			panic(err)
		}
	}

	gauge := prometheus.NewGauge(prometheus.GaugeOptions{
		Name: "promhttp_metric_handler_requests_in_flight",
		Help: "Current number of scrapes being served.",
	})
	if err := reg.Register(gauge); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			gauge = are.ExistingCollector.(prometheus.Gauge)
		} else {
			panic(err)
		}
	}

	return InstrumentHandlerCounter(counterVec, InstrumentHandlerInFlight(gauge, handler))
}

type Logger interface {
	Println(v ...interface{})
}

type HandlerErrorHandling int

type HandlerOpts struct {
	ErrorLog            Logger
	ErrorHandling       HandlerErrorHandling
	DisableCompression  bool
	MaxRequestsInFlight int
	Timeout             time.Duration
}

func HandlerFor(reg prometheus.Gatherer, opts HandlerOpts) http.Handler {

}
