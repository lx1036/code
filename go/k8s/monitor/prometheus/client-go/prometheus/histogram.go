package prometheus

type HistogramVec struct {
	*metricVec
}

type HistogramOpts struct {
	Namespace   string
	Subsystem   string
	Name        string
	Help        string
	ConstLabels Labels
	Buckets     []float64
}

/*func NewHistogramVec(options HistogramOpts, labelNames []string) *HistogramVec {
	desc := NewDesc(
		BuildFQName(options.Namespace, options.Subsystem, options.Name),
		options.Help,
		labelNames,
		options.ConstLabels,
	)
	return &HistogramVec{
		metricVec: newMetricVec(desc, func(lvs ...string) Metric {
			return newHistogram()
		})
	}
}*/
