package endpoint

import (
	"context"
	"github.com/cilium/cilium/pkg/controller"
	"github.com/cilium/cilium/pkg/identity"
	"github.com/cilium/cilium/pkg/logging/logfields"
	"github.com/cilium/cilium/pkg/metrics"
	"github.com/cilium/cilium/pkg/policy"
	log "github.com/sirupsen/logrus"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// Endpoint represents a container or similar which can be individually
// addresses on L3 with its own IP addresses.
//
// The representation of the Endpoint which is serialized to disk for restore
// purposes is the serializableEndpoint type in this package.
type Endpoint struct {
	mutex sync.RWMutex

	// ID of the endpoint, unique in the scope of the node
	ID uint16
	// state is the state the endpoint is in. See setState()
	state string
	// status contains the last n state transitions this endpoint went through
	status *EndpointStatus

	// SecurityIdentity is the security identity of this endpoint. This is computed from
	// the endpoint's labels.
	SecurityIdentity *identity.Identity `json:"SecLabel"`

	hasBPFProgram chan struct{}

	desiredPolicy  *policy.EndpointPolicy
	realizedPolicy *policy.EndpointPolicy

	// controllers is the list of async controllers syncing the endpoint to
	// other resources
	controllers *controller.Manager

	aliveCtx    context.Context
	aliveCancel context.CancelFunc

	isHost bool
}

// SetState modifies the endpoint's state. Returns true only if endpoints state
// was changed as requested
func (e *Endpoint) SetState(toState, reason string) bool {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	return e.setState(toState, reason)
}

func (e *Endpoint) setState(toState, reason string) bool {
	// Validate the state transition.
	fromState := e.state

	switch fromState { // From state
	case "": // Special case for capturing initial state transitions like
		// nil --> StateWaitingForIdentity, StateRestoring
		switch toState {
		case StateWaitingForIdentity, StateRestoring:
			goto OKState
		}
	case StateWaitingForIdentity:
		switch toState {
		case StateReady, StateDisconnecting, StateInvalid:
			goto OKState
		}
	case StateReady:
		switch toState {
		case StateWaitingForIdentity, StateDisconnecting, StateWaitingToRegenerate, StateRestoring:
			goto OKState
		}
	case StateDisconnecting:
		switch toState {
		case StateDisconnected:
			goto OKState
		}
	case StateDisconnected, StateInvalid:
		// No valid transitions, as disconnected and invalid are terminal
		// states for the endpoint.
	case StateWaitingToRegenerate:
		switch toState {
		// Note that transitions to StateWaitingToRegenerate are not allowed,
		// as callers of this function enqueue regenerations if 'true' is
		// returned. We don't want to return 'true' for the case of
		// transitioning to StateWaitingToRegenerate, as this means that a
		// regeneration is already queued up. Callers would then queue up
		// another unneeded regeneration, which is undesired.
		case StateWaitingForIdentity, StateDisconnecting, StateRestoring:
			goto OKState
		// Don't log this state transition being invalid below so that we don't
		// put warnings in the logs for a case which does not result in incorrect
		// behavior.
		case StateWaitingToRegenerate:
			return false
		}
	case StateRegenerating:
		switch toState {
		// Even while the endpoint is regenerating it is
		// possible that further changes require a new
		// build. In this case the endpoint is transitioned
		// from the regenerating state to
		// waiting-for-identity or waiting-to-regenerate state.
		case StateWaitingForIdentity, StateDisconnecting, StateWaitingToRegenerate, StateRestoring:
			goto OKState
		}
	case StateRestoring:
		switch toState {
		case StateDisconnecting, StateRestoring:
			goto OKState
		}
	}

	if toState != fromState {
		_, fileName, fileLine, _ := runtime.Caller(1)
		log.WithFields(log.Fields{
			logfields.EndpointState + ".from": fromState,
			logfields.EndpointState + ".to":   toState,
			"file":                            fileName,
			"line":                            fileLine,
		}).Info("Invalid state transition skipped")
	}
	return false

OKState:
	e.state = toState

	if fromState != "" {
		metrics.EndpointStateCount.WithLabelValues(fromState).Dec()
	}

	// Since StateDisconnected and StateInvalid are final states, after which
	// the endpoint is gone or doesn't exist, we should not increment metrics
	// for these states.
	if toState != "" && toState != StateDisconnected && toState != StateInvalid {
		metrics.EndpointStateCount.WithLabelValues(toState).Inc()
	}

	return true
}

func (e *Endpoint) IsHost() bool {
	return e.isHost
}

// FilterEPDir returns a list of directories' names that possible belong to an endpoint.
func FilterEPDir(dirFiles []os.FileInfo) []string {
	var eptsID []string
	for _, file := range dirFiles {
		if file.IsDir() {
			_, err := strconv.ParseUint(file.Name(), 10, 16)
			if err == nil || strings.HasSuffix(file.Name(), nextDirectorySuffix) || strings.HasSuffix(file.Name(), nextFailedDirectorySuffix) {
				eptsID = append(eptsID, file.Name())
			}
		}
	}

	return eptsID
}
