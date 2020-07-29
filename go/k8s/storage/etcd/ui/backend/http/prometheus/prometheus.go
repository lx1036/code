package prometheus

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	Module = "self"
)

type Opts struct {
	AppName         string
	Idc             string
	WatchPath       map[string]struct{}
	HistogramBucket []float64
}

var Wrapper *prom

func Init(opts Opts) {
	if strings.TrimSpace(opts.AppName) == "" {
		panic("Prometheus Opts.AppName Can't Be Empty")
	}

	if strings.TrimSpace(opts.Idc) == "" {
		panic("Prometheus Opts.Idc Can't Be Empty")
	}

	if len(opts.HistogramBucket) == 0 {
		panic("Prometheus Opts.HistogramBucket Can't Be Empty")
	}

	p := &prom{
		Appname:   opts.AppName,
		Idc:       opts.Idc,
		WatchPath: opts.WatchPath,
		counter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "module_responses",
				Help: "used to calculate qps, failure ratio",
			},
			[]string{"app", "module", "api", "method", "code", "idc"},
		),
		histogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "response_duration_milliseconds",
				Help:    "HTTP latency distributions",
				Buckets: opts.HistogramBucket,
			},
			[]string{"app", "module", "api", "method", "idc"},
		),
	}

	prometheus.MustRegister(p.counter)
	prometheus.MustRegister(p.histogram)

	Wrapper = p
}

func MetricsServerStart(path string, port int) {
	// prometheus metrics path
	go func() {
		http.Handle(path, promhttp.Handler())
		http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
		fmt.Printf("Prometheus start with path '/metrics' and port on %d\n", port)
	}()
}

func GetWrapper() *prom {
	return Wrapper
}

type prom struct {
	Appname   string
	Idc       string
	WatchPath map[string]struct{}
	counter   *prometheus.CounterVec
	histogram *prometheus.HistogramVec
}

type QPSRecord struct {
	Times  float64
	Api    string
	Module string
	Method string
	Code   int
}

func (p *prom) QpsCountLog(r QPSRecord) (ret bool, err error) {
	if strings.TrimSpace(r.Api) == "" {
		return ret, errors.New("QPSRecord.Api Can't Be Empty")
	}

	if r.Times <= 0 {
		r.Times = 1
	}

	if strings.TrimSpace(r.Module) == "" {
		r.Module = "self"
	}

	if strings.TrimSpace(r.Method) == "" {
		r.Method = "GET"
	}

	if r.Code == 0 {
		r.Code = 200
	}

	p.counter.WithLabelValues(p.Appname, r.Module, r.Api, r.Method, strconv.Itoa(r.Code), p.Idc).Add(r.Times)

	return true, nil
}

type LatencyRecord struct {
	Time   float64
	Api    string
	Module string
	Method string
}

func (p *prom) LatencyLog(r LatencyRecord) (ret bool, err error) {
	if r.Time <= 0 {
		return ret, errors.New("LatencyRecord.Time Must Greater Than 0")
	}

	if strings.TrimSpace(r.Module) == "" {
		r.Module = "self"
	}

	if strings.TrimSpace(r.Method) == "" {
		r.Method = "GET"
	}

	p.histogram.WithLabelValues(p.Appname, r.Module, r.Api, r.Method, p.Idc).Observe(r.Time)

	return true, nil
}
