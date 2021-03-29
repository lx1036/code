package runtime

import (
	framework "k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/v1alpha1"
	"k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"

	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/events"
)

type frameworkOptions struct {
	clientSet            clientset.Interface
	eventRecorder        events.EventRecorder
	informerFactory      informers.SharedInformerFactory
	snapshotSharedLister v1alpha1.SharedLister
	metricsRecorder      *metricsRecorder
	profileName          string
	podNominator         v1alpha1.PodNominator
	extenders            []v1alpha1.Extender
	runAllFilters        bool
}

// Option for the frameworkImpl.
type Option func(*frameworkOptions)

// WithPodNominator sets podNominator for the scheduling frameworkImpl.
func WithPodNominator(nominator v1alpha1.PodNominator) Option {
	return func(o *frameworkOptions) {
		o.podNominator = nominator
	}
}

// frameworkImpl is the component responsible for initializing and running scheduler
// plugins.
type frameworkImpl struct {
	registry              Registry
	snapshotSharedLister  v1alpha1.SharedLister
	waitingPods           *waitingPodsMap
	pluginNameToWeightMap map[string]int
	queueSortPlugins      []v1alpha1.QueueSortPlugin
	preFilterPlugins      []v1alpha1.PreFilterPlugin
	filterPlugins         []v1alpha1.FilterPlugin
	postFilterPlugins     []v1alpha1.PostFilterPlugin
	preScorePlugins       []v1alpha1.PreScorePlugin
	scorePlugins          []v1alpha1.ScorePlugin
	reservePlugins        []v1alpha1.ReservePlugin
	preBindPlugins        []v1alpha1.PreBindPlugin
	bindPlugins           []v1alpha1.BindPlugin
	postBindPlugins       []v1alpha1.PostBindPlugin
	permitPlugins         []v1alpha1.PermitPlugin

	clientSet       clientset.Interface
	eventRecorder   events.EventRecorder
	informerFactory informers.SharedInformerFactory

	metricsRecorder *metricsRecorder
	profileName     string

	preemptHandle v1alpha1.PreemptHandle

	// Indicates that RunFilterPlugins should accumulate all failed statuses and not return
	// after the first failure.
	runAllFilters bool
}
