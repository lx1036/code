package workqueue

import (
	"sync"
	"k8s.io/apimachinery/pkg/util/clock"
	"time"
)

type queueMetrics interface {
	add(item job)
	get(item job)
	done(item job)
	updateUnfinishedWork()
}

// GaugeMetric represents a single numerical value that can arbitrarily go up
// and down.
type GaugeMetric interface {
	Inc()
	Dec()
}
// CounterMetric represents a single numerical value that only ever
// goes up.
type CounterMetric interface {
	Inc()
}
// HistogramMetric counts individual observations.
type HistogramMetric interface {
	Observe(float64)
}
// SettableGaugeMetric represents a single numerical value that can arbitrarily go up
// and down. (Separate from GaugeMetric to preserve backwards compatibility.)
type SettableGaugeMetric interface {
	Set(float64)
}
// MetricsProvider generates various metrics used by the queue.
type MetricsProvider interface {
	NewDepthMetric(name string) GaugeMetric
	NewAddsMetric(name string) CounterMetric
	NewLatencyMetric(name string) HistogramMetric
	NewWorkDurationMetric(name string) HistogramMetric
	NewUnfinishedWorkSecondsMetric(name string) SettableGaugeMetric
	NewLongestRunningProcessorSecondsMetric(name string) SettableGaugeMetric
	NewRetriesMetric(name string) CounterMetric
}

type noopMetric struct{}
func (noopMetric) Inc()            {}
func (noopMetric) Dec()            {}
func (noopMetric) Set(float64)     {}
func (noopMetric) Observe(float64) {}
type noopMetricsProvider struct{}
func (_ noopMetricsProvider) NewDepthMetric(name string) GaugeMetric {
	return noopMetric{}
}

func (_ noopMetricsProvider) NewAddsMetric(name string) CounterMetric {
	return noopMetric{}
}

func (_ noopMetricsProvider) NewLatencyMetric(name string) HistogramMetric {
	return noopMetric{}
}

func (_ noopMetricsProvider) NewWorkDurationMetric(name string) HistogramMetric {
	return noopMetric{}
}

func (_ noopMetricsProvider) NewUnfinishedWorkSecondsMetric(name string) SettableGaugeMetric {
	return noopMetric{}
}

func (_ noopMetricsProvider) NewLongestRunningProcessorSecondsMetric(name string) SettableGaugeMetric {
	return noopMetric{}
}

func (_ noopMetricsProvider) NewRetriesMetric(name string) CounterMetric {
	return noopMetric{}
}
type noMetrics struct{}

func (noMetrics) add(item job)            {}
func (noMetrics) get(item job)            {}
func (noMetrics) done(item job)           {}
func (noMetrics) updateUnfinishedWork() {}
// defaultQueueMetrics expects the caller to lock before setting any metrics.
type defaultQueueMetrics struct {
	clock clock.Clock
	
	// current depth of a workqueue
	depth GaugeMetric
	// total number of adds handled by a workqueue
	adds CounterMetric
	// how long an item stays in a workqueue
	latency HistogramMetric
	// how long processing an item from a workqueue takes
	workDuration         HistogramMetric
	addTimes             map[job]time.Time
	processingStartTimes map[job]time.Time
	
	// how long have current threads been working?
	unfinishedWorkSeconds   SettableGaugeMetric
	longestRunningProcessor SettableGaugeMetric
}
func (m *defaultQueueMetrics) add(item job) {
	if m == nil {
		return
	}
	
	m.adds.Inc()
	m.depth.Inc()
	if _, exists := m.addTimes[item]; !exists {
		m.addTimes[item] = m.clock.Now()
	}
}
func (m *defaultQueueMetrics) get(item job) {
	if m == nil {
		return
	}
	
	m.depth.Dec()
	m.processingStartTimes[item] = m.clock.Now()
	if startTime, exists := m.addTimes[item]; exists {
		m.latency.Observe(m.sinceInSeconds(startTime))
		delete(m.addTimes, item)
	}
}
func (m *defaultQueueMetrics) done(item job) {
	if m == nil {
		return
	}
	
	if startTime, exists := m.processingStartTimes[item]; exists {
		m.workDuration.Observe(m.sinceInSeconds(startTime))
		delete(m.processingStartTimes, item)
	}
}
func (m *defaultQueueMetrics) updateUnfinishedWork() {
	// Note that a summary metric would be better for this, but prometheus
	// doesn't seem to have non-hacky ways to reset the summary metrics.
	var total float64
	var oldest float64
	for _, t := range m.processingStartTimes {
		age := m.sinceInSeconds(t)
		total += age
		if age > oldest {
			oldest = age
		}
	}
	m.unfinishedWorkSeconds.Set(total)
	m.longestRunningProcessor.Set(oldest)
}
// Gets the time since the specified start in seconds.
func (m *defaultQueueMetrics) sinceInSeconds(start time.Time) float64 {
	return m.clock.Since(start).Seconds()
}
var globalMetricsFactory = queueMetricsFactory{
	metricsProvider: noopMetricsProvider{},
}
type queueMetricsFactory struct {
	metricsProvider MetricsProvider
	
	onlyOnce sync.Once
}
func (f *queueMetricsFactory) newQueueMetrics(name string, clock clock.Clock) queueMetrics {
	mp := f.metricsProvider
	if len(name) == 0 || mp == (noopMetricsProvider{}) {
		return noMetrics{}
	}
	return &defaultQueueMetrics{
		clock:                   clock,
		depth:                   mp.NewDepthMetric(name),
		adds:                    mp.NewAddsMetric(name),
		latency:                 mp.NewLatencyMetric(name),
		workDuration:            mp.NewWorkDurationMetric(name),
		unfinishedWorkSeconds:   mp.NewUnfinishedWorkSecondsMetric(name),
		longestRunningProcessor: mp.NewLongestRunningProcessorSecondsMetric(name),
		addTimes:                map[job]time.Time{},
		processingStartTimes:    map[job]time.Time{},
	}
}
