package prometheus

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Options struct {
	AppName          string
	Idc              string
	WatchPath        map[string]struct{}
	HistogramBuckets []float64
}

type Prometheus struct {
	AppName   string
	Idc       string
	WatchPath map[string]struct{}
	Counter   *prometheus.CounterVec
	Histogram *prometheus.HistogramVec
}

var Metrics *Prometheus

func Init(options Options) {
	if strings.TrimSpace(options.AppName) == "" || strings.TrimSpace(options.Idc) == "" || len(options.HistogramBuckets) == 0 {
		panic(options.AppName + " or " + options.Idc + " or HistogramBuckets is empty.")
	}

	Metrics = &Prometheus{
		AppName:   options.AppName,
		Idc:       options.Idc,
		WatchPath: options.WatchPath,
		Counter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "module_responses",
				Help: "calculate qps",
			},
			[]string{"app", "module", "api", "method", "code", "idc"},
		),
		Histogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace:   "",
				Subsystem:   "",
				Name:        "response_duration_milliseconds",
				Help:        "HTTP latency distributions",
				ConstLabels: nil,
				//Buckets:     options.HistogramBuckets,
			},
			[]string{"app", "module", "api", "method", "idc"},
		),
	}

	prometheus.MustRegister(Metrics.Counter)
	prometheus.MustRegister(Metrics.Histogram)
}

type LatencyRecord struct {
	Time   float64
	Api    string
	Module string
	Method string
	Code   int
}

type QpsRecord struct {
	Api    string
	Module string
	Method string
	Code   int
}

func (metrics *Prometheus) LatencyLog(record LatencyRecord) {
	if strings.TrimSpace(record.Module) == "" {
		record.Module = "self"
	}

	metrics.Histogram.WithLabelValues(
		metrics.AppName,
		record.Module,
		record.Api,
		record.Method,
		metrics.Idc,
	).Observe(record.Time)
}

func (metrics *Prometheus) QpsCounterLog(record QpsRecord) {
	if strings.TrimSpace(record.Module) == "" {
		record.Module = "self"
	}

	metrics.Counter.WithLabelValues(
		metrics.AppName,
		record.Module,
		record.Api,
		record.Method,
		strconv.Itoa(record.Code),
		metrics.Idc,
	).Inc()
}

func MetricsServerStart(path string, port int) {
	go func() {
		http.Handle(path, promhttp.Handler())
		fmt.Println(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
	}()
}

func MiddlewareTest() gin.HandlerFunc {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.JSONFormatter{})

	return func(context *gin.Context) {
		log.Info("afsadfadsfafsadfadsf")
		times := time.Now()
		context.Next()
		latency := float64(time.Since(times).Milliseconds())
		log.Info(latency)
	}
}

func MiddlewarePrometheusAccessLogger() gin.HandlerFunc {
	return func(context *gin.Context) {
		times := time.Now()

		context.Next()

		if _, ok := Metrics.WatchPath[context.Request.URL.Path]; ok {
			latency := float64(time.Since(times).Milliseconds())
			Metrics.LatencyLog(LatencyRecord{
				Time:   latency,
				Api:    context.Request.URL.Path,
				Method: context.Request.Method,
				Code:   context.Writer.Status(),
			})

			Metrics.QpsCounterLog(QpsRecord{
				Api:    context.Request.URL.Path,
				Method: context.Request.Method,
				Code:   context.Writer.Status(),
			})

			fmt.Println(context.Request.URL.Path, context.Request.Method, context.Writer.Status(), latency)
		}
	}
}
