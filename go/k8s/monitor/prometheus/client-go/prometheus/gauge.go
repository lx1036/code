package prometheus

import "k8s-lx1036/app/prometheus/client-go/prometheus/metrics"

type Gauge interface {
	Metric
	Collector
	Set(float64)
	Inc()
	Dec()
	Add(float64)
	Sub(float64)
	SetToCurrentTime()
}

type GaugeOptions Options

type gauge struct {
	valBits unit64
	selfCollector
	desc       *Desc
	labelPairs []*metrics.LabelPair
}

func NewGauge(options GaugeOptions) Gauge {
	desc := NewDesc(BuildFQName(opts.Namespace, opts.Subsystem, opts.Name), opts.Help, nil, opts.ConstLabels)
	result := &gauge{
		valBits:       nil,
		selfCollector: nil,
		desc:          desc,
		labelPairs:    desc.constLabelPairs,
	}

	result.init(result)

	return result
}
