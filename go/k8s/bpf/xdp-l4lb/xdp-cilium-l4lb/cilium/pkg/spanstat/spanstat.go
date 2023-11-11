package spanstat

import (
	"github.com/cilium/cilium/pkg/lock"
	"github.com/cilium/cilium/pkg/safetime"
	"time"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging/logfields"
)

var (
	subSystem = "spanstat"
	log       = logging.DefaultLogger.WithField(logfields.LogSubsys, subSystem)
)

// SpanStat measures the total duration of all time spent in between Start()
// and Stop() calls.
type SpanStat struct {
	mutex           lock.RWMutex
	spanStart       time.Time
	successDuration time.Duration
	failureDuration time.Duration
}

func Start() *SpanStat {
	s := &SpanStat{}
	return s.Start()
}

// Start starts a new span
func (s *SpanStat) Start() *SpanStat {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.spanStart = time.Now()
	return s
}

func (s *SpanStat) End(success bool) *SpanStat {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.end(success)
}

func (s *SpanStat) end(success bool) *SpanStat {
	if !s.spanStart.IsZero() {
		d, _ := safetime.TimeSinceSafe(s.spanStart, log)
		if success {
			s.successDuration += d
		} else {
			s.failureDuration += d
		}
	}
	s.spanStart = time.Time{}
	return s
}
