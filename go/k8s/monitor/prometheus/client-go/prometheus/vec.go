package prometheus

import "sync"

type metricVec struct {
	*metricMap

	curry []curriedLabelValue

	// hashAdd and hashAddByte can be replaced for testing collision handling.
	hashAdd     func(h uint64, s string) uint64
	hashAddByte func(h uint64, b byte) uint64
}

func newMetricVec(desc *Desc, newMetric func(lvs ...string) Metric) *metricVec {
	return &metricVec{
		metricMap: &metricMap{
			desc:      desc,
			newMetric: newMetric,
		},
	}
}

type metricMap struct {
	mtx       sync.RWMutex // Protects metrics.
	metrics   map[uint64][]metricWithLabelValues
	desc      *Desc
	newMetric func(labelValues ...string) Metric
}

func (metric *metricMap) Describe(ch chan<- *Desc) {

}

func (metric *metricMap) Collect(ch chan<- Metric) {

}

type curriedLabelValue struct {
	index int
	value string
}
type metricWithLabelValues struct {
	values []string
	metric Metric
}
