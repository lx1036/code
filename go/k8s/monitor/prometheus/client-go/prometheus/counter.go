package prometheus

type Counter interface {
	Metric
	Collector
	Inc()
	Add(float64)
}

type CounterVec struct {
	*metricVec
}

/*func (c *CounterVec) Describe(chan<- *Desc) {
	panic("implement me")
}

func (c *CounterVec) Collect(chan<- Metric) {
	panic("implement me")
}*/

func (c *CounterVec) WithLabelValues(labelValues ...string) Counter {

}

type CounterOpts Options

func NewCounterVec(opts CounterOpts, labelNames []string) *CounterVec {
	desc := NewDesc(
		BuildFQName(opts.Namespace, opts.Subsystem, opts.Name),
		opts.Help,
		labelNames,
		opts.ConstLabels,
	)
	return &CounterVec{
		metricVec: newMetricVec(desc, func(lvs ...string) Metric {
			if len(lvs) != len(desc.variableLabels) {
				panic(makeInconsistentCardinalityError(desc.fqName, desc.variableLabels, lvs))
			}
			result := &counter{desc: desc, labelPairs: makeLabelPairs(desc, lvs)}
			result.init(result) // Init self-collection.
			return result
		}),
	}
}
