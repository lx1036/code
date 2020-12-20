package promhttp

import (
	"net/http"
)

func InstrumentHandlerCounter(counter *prometheus.CounterVec, next http.Handler) http.HandlerFunc {
	code, method := checkLabels(counter)

	if code {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			d := newDelegator(w, nil)
			next.ServeHTTP(d, r)
			counter.With(labels(code, method, r.Method, d.Status())).Inc()
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		counter.With(labels(code, method, r.Method, 0)).Inc()
	})
}

func InstrumentHandlerInFlight(g prometheus.Gauge, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		g.Inc()
		defer g.Dec()
		next.ServeHTTP(w, r)
	})
}

func checkLabels(c prometheus.Collector) (code bool, method bool) {
	// TODO: Remove this hacky way to check for instance labels
	// once Descriptors can have their dimensionality queried.
	var (
		desc *prometheus.Desc
		m    prometheus.Metric
		pm   metrics.Metric
		lvs  []string
	)

	// Get the Desc from the Collector.
	descc := make(chan *prometheus.Desc, 1)
	c.Describe(descc)

	select {
	case desc = <-descc:
	default:
		panic("no description provided by collector")
	}

	select {
	case <-descc:
		panic("more than one description provided by collector")
	default:
	}

	close(descc)
}
