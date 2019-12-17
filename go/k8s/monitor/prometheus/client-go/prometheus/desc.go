package prometheus

import "k8s-lx1036/app/k8s/prometheus/client-go/prometheus/metrics"

type Desc struct {
	// fqName has been built from Namespace, Subsystem, and Name.
	fqName string
	// help provides some helpful information about this metric.
	help string
	// constLabelPairs contains precalculated DTO label pairs based on
	// the constant labels.
	constLabelPairs []*metrics.LabelPair
	// VariableLabels contains names of labels for which the metric
	// maintains variable values.
	variableLabels []string
	// id is a hash of the values of the ConstLabels and fqName. This
	// must be unique among all registered descriptors and can therefore be
	// used as an identifier of the descriptor.
	id uint64
	// dimHash is a hash of the label names (preset and variable) and the
	// Help string. Each Desc with the same fqName must have the same
	// dimHash.
	dimHash uint64
	// err is an error that occurred during construction. It is reported on
	// registration time.
	err error
}

func NewDesc(fqName, help string, variableLabels []string, constLabels Labels) *Desc {
	desc := &Desc{
		fqName:         fqName,
		help:           help,
		variableLabels: variableLabels,
	}

	return desc
}
