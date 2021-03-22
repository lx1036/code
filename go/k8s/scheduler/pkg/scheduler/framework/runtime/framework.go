package runtime

import (
	framework "k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/v1alpha1"

	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/events"
)

type frameworkOptions struct {
	clientSet            clientset.Interface
	eventRecorder        events.EventRecorder
	informerFactory      informers.SharedInformerFactory
	snapshotSharedLister framework.SharedLister
	metricsRecorder      *metricsRecorder
	profileName          string
	podNominator         framework.PodNominator
	extenders            []framework.Extender
	runAllFilters        bool
}

// Option for the frameworkImpl.
type Option func(*frameworkOptions)

// WithPodNominator sets podNominator for the scheduling frameworkImpl.
func WithPodNominator(nominator framework.PodNominator) Option {
	return func(o *frameworkOptions) {
		o.podNominator = nominator
	}
}
