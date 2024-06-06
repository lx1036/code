package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    clientmodel "github.com/prometheus/client_model/go"
)

type GaugeVec interface {
    WithLabelValues(lvls ...string) prometheus.Gauge
    prometheus.Collector
}

// GaugeVec

type gaugeVec struct {
    prometheus.Collector
}

func (gv *gaugeVec) WithLabelValues(lvls ...string) prometheus.Gauge {
    return NoOpGauge
}

// Gauge

type gauge struct {
    prometheus.Metric
    prometheus.Collector
}

func (g *gauge) Set(float64)       {}
func (g *gauge) Inc()              {}
func (g *gauge) Dec()              {}
func (g *gauge) Add(float64)       {}
func (g *gauge) Sub(float64)       {}
func (g *gauge) SetToCurrentTime() {}

type collector struct{}

func (c *collector) Describe(chan<- *prometheus.Desc) {}
func (c *collector) Collect(chan<- prometheus.Metric) {}

type metric struct{}

// Desc returns nil so do not register this metric into prometheus default register.
func (m *metric) Desc() *prometheus.Desc          { return nil }
func (m *metric) Write(*clientmodel.Metric) error { return nil }

var (
    NoOpGaugeVec GaugeVec         = &gaugeVec{NoOpCollector}
    NoOpGauge    prometheus.Gauge = &gauge{NoOpMetric, NoOpCollector}

    NoOpCollector prometheus.Collector = &collector{}
    NoOpMetric    prometheus.Metric    = &metric{}
)
