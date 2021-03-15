package stats

import (
	statsapi "k8s-lx1036/k8s/kubelet/pkg/apis/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/kubelet/util"
)

// SummaryProvider provides summaries of the stats from Kubelet.
type SummaryProvider interface {
	// Get provides a new Summary with the stats from Kubelet,
	// and will update some stats if updateStats is true
	Get(updateStats bool) (*statsapi.Summary, error)
	// GetCPUAndMemoryStats provides a new Summary with the CPU and memory stats from Kubelet,
	GetCPUAndMemoryStats() (*statsapi.Summary, error)
}

// NewSummaryProvider returns a SummaryProvider using the stats provided by the
// specified statsProvider.
func NewSummaryProvider(statsProvider Provider) SummaryProvider {
	kubeletCreationTime := metav1.Now()
	bootTime, err := util.GetBootTime()
	if err != nil {
		// bootTime will be zero if we encounter an error getting the boot time.
		klog.Warningf("Error getting system boot time.  Node metrics will have an incorrect start time: %v", err)
	}

	return &summaryProviderImpl{
		kubeletCreationTime: kubeletCreationTime,
		systemBootTime:      metav1.NewTime(bootTime),
		provider:            statsProvider,
	}
}
