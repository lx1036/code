package demo_node_status

import (
	"errors"
	"sync"
	"time"
)

var (
	// ErrNetworkUnknown indicates the network state is unknown
	ErrNetworkUnknown = errors.New("network state unknown")
)

type runtimeState struct {
	sync.RWMutex
	lastBaseRuntimeSync      time.Time
	baseRuntimeSyncThreshold time.Duration
	networkError             error
	runtimeError             error
	storageError             error
	cidr                     string
	healthChecks             []*healthCheck
}

// A health check function should be efficient and not rely on external
// components (e.g., container runtime).
type healthCheckFnType func() (bool, error)

type healthCheck struct {
	name string
	fn   healthCheckFnType
}

func (s *runtimeState) setNetworkState(err error) {
	s.Lock()
	defer s.Unlock()
	s.networkError = err
}

func newRuntimeState(runtimeSyncThreshold time.Duration) *runtimeState {
	return &runtimeState{
		lastBaseRuntimeSync:      time.Time{},
		baseRuntimeSyncThreshold: runtimeSyncThreshold,
		networkError:             ErrNetworkUnknown,
	}
}
