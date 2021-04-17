package stats

import (
	"sync"
	"sync/atomic"
	"time"

	stats "k8s-lx1036/k8s/kubelet/pkg/apis/stats/v1alpha1"

	v1 "k8s.io/api/core/v1"
)

// volumeStatCalculator calculates volume metrics
// for a given pod periodically in the background and caches the result
type volumeStatCalculator struct {
	statsProvider Provider
	jitterPeriod  time.Duration
	pod           *v1.Pod
	stopChannel   chan struct{}
	startO        sync.Once
	stopO         sync.Once
	latest        atomic.Value
}

// PodVolumeStats encapsulates the VolumeStats for a pod.
// It consists of two lists, for local ephemeral volumes, and for persistent volumes respectively.
type PodVolumeStats struct {
	EphemeralVolumes  []stats.VolumeStats
	PersistentVolumes []stats.VolumeStats
}

// getLatest returns the most recent PodVolumeStats from the cache
func (s *volumeStatCalculator) GetLatest() (PodVolumeStats, bool) {
	result := s.latest.Load()
	if result == nil {
		return PodVolumeStats{}, false
	}
	return result.(PodVolumeStats), true
}
