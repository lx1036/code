package prometheus

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

const (
	Module = "self"
)

type Options struct {
	AppName          string
	Idc              string
	WatchPath        map[string]struct{}
	HistogramBuckets []float64
}

type LatencyRecord struct {
	Time   float64
	Api    string
	Module string
	Method string
	Code   int
}

type QpsRecord struct {
	Times  float64
	Api    string
	Module string
	Method string
	Code   int
}

type Prometheus struct {
	AppName   string
	Idc       string
	WatchPath map[string]struct{}
	Counter   *prometheus.CounterVec
	Histogram *prometheus.HistogramVec
}

var Metrics *Prometheus

func GetWrapper() *Prometheus {
	return Metrics
}

func Init(options Options) {
	if strings.TrimSpace(options.AppName) == "" || strings.TrimSpace(options.Idc) == "" || len(options.HistogramBuckets) == 0 {
		panic(options.AppName + " or " + options.Idc + " or HistogramBuckets is empty.")
	}

	Metrics = &Prometheus{
		AppName:   options.AppName,
		Idc:       options.Idc,
		WatchPath: options.WatchPath,
		Counter: prometheus.NewCounterVec( // QPS and failure ratio
			prometheus.CounterOpts{
				Name: "module_responses",
				Help: "used to calculate qps and failure ratio",
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
	//prometheus2.MustRegister(Metrics.Histogram)
}

// QPS
func (metrics *Prometheus) QpsCounterLog(record QpsRecord) {
	if strings.TrimSpace(record.Module) == "" {
		record.Module = Module
	}

	// "app", "module", "api", "method", "code", "idc"
	metrics.Counter.WithLabelValues(
		metrics.AppName,
		record.Module,
		record.Api,
		record.Method,
		strconv.Itoa(record.Code),
		metrics.Idc,
	).Add(record.Times)
}

// P95/P99
func (metrics *Prometheus) LatencyLog(record LatencyRecord) {
	if strings.TrimSpace(record.Module) == "" {
		record.Module = Module
	}

	// "app", "module", "api", "method", "idc"
	metrics.Histogram.WithLabelValues(
		metrics.AppName,
		record.Module,
		record.Api,
		record.Method,
		metrics.Idc,
	).Observe(record.Time)
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
		times := time.Now()
		context.Next()
		latency := float64(time.Since(times).Milliseconds())
		log.Info(latency)
	}
}

func MiddlewarePrometheusAccessLogger() gin.HandlerFunc {
	return func(context *gin.Context) {
		start := time.Now()
		context.Next()

		go func() {
			if GetWrapper() != nil {
				if _, ok := Metrics.WatchPath[context.Request.URL.Path]; ok {
					// QPS
					Metrics.QpsCounterLog(QpsRecord{
						Times:  1,
						Api:    context.Request.URL.Path,
						Module: Module,
						Method: context.Request.Method,
						Code:   context.Writer.Status(),
					})

					// P95/P99
					latency := float64(time.Since(start).Milliseconds())
					Metrics.LatencyLog(LatencyRecord{
						Time:   latency,
						Api:    context.Request.URL.Path,
						Module: Module,
						Method: context.Request.Method,
						Code:   context.Writer.Status(),
					})

					log.WithFields(log.Fields{
						"method":  context.Request.Method,
						"path":    context.Request.URL.Path,
						"status":  context.Writer.Status(),
						"latency": latency,
					}).Info("[app level]prometheus access logger")
				}
			} else {
				log.Warn("need to init prometheus!!!")
			}
		}()

		if _, ok := Metrics.WatchPath[context.Request.URL.Path]; ok {
			latency := float64(time.Since(start).Milliseconds())
			/*Metrics.LatencyLog(LatencyRecord{
				Time:   latency,
				Api:    context.Request.URL.Path,
				Method: context.Request.Method,
				Code:   context.Writer.Status(),
			})*/

			Metrics.QpsCounterLog(QpsRecord{
				Api:    context.Request.URL.Path,
				Method: context.Request.Method,
				Code:   context.Writer.Status(),
			})

			fmt.Println(context.Request.URL.Path, context.Request.Method, context.Writer.Status(), latency)
		}
	}
}
