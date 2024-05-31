package endpoint

import (
    "context"
    "github.com/cilium/cilium/pkg/identity"
    "github.com/cilium/cilium/pkg/identity/cache"
    "github.com/cilium/cilium/pkg/logging"
    "github.com/cilium/cilium/pkg/metrics"
    "github.com/cilium/cilium/pkg/policy"
    "os"
    "runtime"
    "strconv"
    "strings"
    "sync"
    "time"
    "unsafe"

    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/controller"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/endpoint/regeneration"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/eventqueue"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/fqdn"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/labels"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging/logfields"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/mac"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/option"

    log "github.com/sirupsen/logrus"
)

// Endpoint represents a container or similar which can be individually
// addresses on L3 with its own IP addresses.
//
// The representation of the Endpoint which is serialized to disk for restore
// purposes is the serializableEndpoint type in this package.
type Endpoint struct {
    mutex sync.RWMutex

    owner regeneration.Owner

    // ID of the endpoint, unique in the scope of the node
    ID uint16
    // state is the state the endpoint is in. See setState()
    state string
    // status contains the last n state transitions this endpoint went through
    status *EndpointStatus

    // OpLabels is the endpoint's label configuration
    OpLabels labels.OpLabels

    // SecurityIdentity is the security identity of this endpoint. This is computed from
    // the endpoint's labels.
    SecurityIdentity *identity.Identity `json:"SecLabel"`

    hasBPFProgram chan struct{}

    // containerName is the name given to the endpoint by the container runtime
    containerName string

    // containerID is the container ID that docker has assigned to the endpoint
    containerID string

    // dockerNetworkID is the network ID of the libnetwork network if the
    // endpoint is a docker managed container which uses libnetwork
    dockerNetworkID string

    // dockerEndpointID is the Docker network endpoint ID if managed by
    // libnetwork
    dockerEndpointID string

    // ifName is the name of the host facing interface (veth pair) which
    // connects into the endpoint
    ifName string

    // ifIndex is the interface index of the host face interface (veth pair)
    ifIndex int

    // K8sPodName is the Kubernetes pod name of the endpoint
    K8sPodName string

    // K8sNamespace is the Kubernetes namespace of the endpoint
    K8sNamespace string

    // mac is the MAC address of the endpoint
    mac mac.MAC // Container MAC address.

    desiredPolicy  *policy.EndpointPolicy
    realizedPolicy *policy.EndpointPolicy

    // controllers is the list of async controllers syncing the endpoint to
    // other resources
    controllers *controller.Manager

    // Options determine the datapath configuration of the endpoint.
    Options *option.IntOptions

    aliveCtx    context.Context
    aliveCancel context.CancelFunc

    eventQueue *eventqueue.EventQueue

    // logger is a logrus object with fields set to report an endpoints information.
    // This must only be accessed with atomic.LoadPointer/StorePointer.
    // 'mutex' must be Lock()ed to synchronize stores. No lock needs to be held
    // when loading this pointer.
    logger unsafe.Pointer
    // policyLogger is a logrus object with fields set to report an endpoints information.
    // This must only be accessed with atomic LoadPointer/StorePointer.
    // 'mutex' must be Lock()ed to synchronize stores. No lock needs to be held
    // when loading this pointer.
    policyLogger unsafe.Pointer

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

// HasLabels returns whether endpoint e contains all labels l. Will return 'false'
// if any label in l is not in the endpoint's labels.
func (e *Endpoint) HasLabels(l labels.Labels) bool {
    e.unconditionalRLock()
    defer e.runlock()

    return e.hasLabelsRLocked(l)
}

// hasLabelsRLocked returns whether endpoint e contains all labels l. Will
// return 'false' if any label in l is not in the endpoint's labels.
// e.mutex must be RLock()ed.
func (e *Endpoint) hasLabelsRLocked(l labels.Labels) bool {
    allEpLabels := e.OpLabels.AllLabels()

    for _, v := range l {
        found := false
        for _, j := range allEpLabels {
            if j.Equals(&v) {
                found = true
                break
            }
        }
        if !found {
            return false
        }
    }

    return true
}

func createEndpoint(owner regeneration.Owner, policyGetter policyRepoGetter, namedPortsGetter namedPortsGetter, proxy EndpointProxy, allocator cache.IdentityAllocator, ID uint16, ifName string) *Endpoint {
    ep := &Endpoint{
        owner:            owner,
        policyGetter:     policyGetter,
        namedPortsGetter: namedPortsGetter,
        ID:               ID,
        createdAt:        time.Now(),
        proxy:            proxy,
        ifName:           ifName,
        OpLabels:         labels.NewOpLabels(),
        DNSRules:         nil,
        DNSHistory:       fqdn.NewDNSCacheWithLimit(option.Config.ToFQDNsMinTTL, option.Config.ToFQDNsMaxIPsPerHost),
        DNSZombies:       fqdn.NewDNSZombieMappings(option.Config.ToFQDNsMaxDeferredConnectionDeletes, option.Config.ToFQDNsMaxIPsPerHost),
        state:            "",
        status:           NewEndpointStatus(),
        hasBPFProgram:    make(chan struct{}, 0),
        desiredPolicy:    policy.NewEndpointPolicy(policyGetter.GetPolicyRepository()),
        controllers:      controller.NewManager(),
        regenFailedChan:  make(chan struct{}, 1),
        allocator:        allocator,
        logLimiter:       logging.NewLimiter(10*time.Second, 3), // 1 log / 10 secs, burst of 3
        noTrackPort:      0,
    }

    ep.initDNSHistoryTrigger()

    ctx, cancel := context.WithCancel(context.Background())
    ep.aliveCancel = cancel
    ep.aliveCtx = ctx

    ep.realizedPolicy = ep.desiredPolicy

    ep.SetDefaultOpts(option.Config.Opts)

    return ep
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
