package endpoint

import (
    "context"
    "runtime"
    "sync"
    "time"
    "unsafe"

    "github.com/sirupsen/logrus"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/controller"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/datapath/link"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/defaults"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/endpoint/regeneration"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/eventqueue"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/identity"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/identity/cache"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/labels"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging/logfields"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/mac"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/metrics"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/option"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/policy"
)

var (
    EndpointMutableOptionLibrary = option.GetEndpointMutableOptionLibrary()
)

type policyRepoGetter interface {
    GetPolicyRepository() *policy.Repository
}

type namedPortsGetter interface {
    GetNamedPorts() (npm policy.NamedPortMultiMap)
}

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
    // nodeMAC is the MAC of the node (agent). The MAC is different for every endpoint.
    nodeMAC mac.MAC

    // DatapathConfiguration is the endpoint's datapath configuration as
    // passed in via the plugin that created the endpoint, e.g. the CNI
    // plugin which performed the plumbing will enable certain datapath
    // features according to the mode selected.
    //DatapathConfiguration models.EndpointDatapathConfiguration

    desiredPolicy  *policy.EndpointPolicy
    realizedPolicy *policy.EndpointPolicy

    // policyGetter can get the policy.Repository object.
    policyGetter policyRepoGetter

    // namedPortsGetter can get the ipcache.IPCache object.
    namedPortsGetter namedPortsGetter

    proxy EndpointProxy

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

    regenFailedChan chan struct{}

    allocator cache.IdentityAllocator

    // logLimiter rate limits potentially repeating warning logs
    logLimiter logging.Limiter

    noTrackPort uint16

    // createdAt stores the time the endpoint was created. This value is
    // recalculated on endpoint restore.
    createdAt time.Time
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
        e.getLogger().WithFields(logrus.Fields{
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

// SetDefaultOpts initializes the endpoint Options and configures the specified
// options.
func (e *Endpoint) SetDefaultOpts(opts *option.IntOptions) {
    if e.Options == nil {
        e.Options = option.NewIntOptions(&EndpointMutableOptionLibrary)
    }
    if e.Options.Library == nil {
        e.Options.Library = &EndpointMutableOptionLibrary
    }
    if e.Options.Opts == nil {
        e.Options.Opts = option.OptionMap{}
    }

    if opts != nil {
        epOptLib := option.GetEndpointMutableOptionLibrary()
        for k := range epOptLib {
            e.Options.SetValidated(k, opts.GetValue(k))
        }
    }
    if option.Config.Debug {
        e.Options.SetValidated(option.DebugPolicy, option.OptionEnabled)
    }

    e.UpdateLogger(nil)
}

func createEndpoint(owner regeneration.Owner, policyGetter policyRepoGetter, namedPortsGetter namedPortsGetter,
    proxy EndpointProxy, allocator cache.IdentityAllocator, ID uint16, ifName string) *Endpoint {
    ep := &Endpoint{
        owner:            owner,
        policyGetter:     policyGetter,
        namedPortsGetter: namedPortsGetter,
        ID:               ID,
        createdAt:        time.Now(),
        proxy:            proxy,
        ifName:           ifName,
        OpLabels:         labels.NewOpLabels(),
        //DNSRules:         nil,
        //DNSHistory:       fqdn.NewDNSCacheWithLimit(option.Config.ToFQDNsMinTTL, option.Config.ToFQDNsMaxIPsPerHost),
        //DNSZombies:       fqdn.NewDNSZombieMappings(option.Config.ToFQDNsMaxDeferredConnectionDeletes, option.Config.ToFQDNsMaxIPsPerHost),
        state:           "",
        status:          NewEndpointStatus(),
        hasBPFProgram:   make(chan struct{}, 0),
        desiredPolicy:   policy.NewEndpointPolicy(policyGetter.GetPolicyRepository()),
        controllers:     controller.NewManager(),
        regenFailedChan: make(chan struct{}, 1),
        allocator:       allocator,
        logLimiter:      logging.NewLimiter(10*time.Second, 3), // 1 log / 10 secs, burst of 3
        noTrackPort:     0,
    }

    //ep.initDNSHistoryTrigger()

    ctx, cancel := context.WithCancel(context.Background())
    ep.aliveCancel = cancel
    ep.aliveCtx = ctx

    ep.realizedPolicy = ep.desiredPolicy

    ep.SetDefaultOpts(option.Config.Opts)

    return ep
}

// CreateHostEndpoint creates the endpoint corresponding to the host.
// bpf_host.c
func CreateHostEndpoint(owner regeneration.Owner, policyGetter policyRepoGetter, namedPortsGetter namedPortsGetter, proxy EndpointProxy, allocator cache.IdentityAllocator) (*Endpoint, error) {
    // cilium_host device
    hostDeviceMac, err := link.GetHardwareAddr(defaults.HostDevice)
    if err != nil {
        return nil, err
    }

    ep := createEndpoint(owner, policyGetter, namedPortsGetter, proxy, allocator, 0, defaults.HostDevice)
    ep.isHost = true
    ep.mac = hostDeviceMac
    ep.nodeMAC = hostDeviceMac
    //ep.DatapathConfiguration = NewDatapathConfiguration()

    ep.setState(StateWaitingForIdentity, "Endpoint creation")

    return ep, nil
}
