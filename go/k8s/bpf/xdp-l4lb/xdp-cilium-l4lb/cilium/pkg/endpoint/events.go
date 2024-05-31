package endpoint

import (
    "fmt"

    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/eventqueue"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging/logfields"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/option"
)

// EndpointRegenerationEvent contains all fields necessary to regenerate an endpoint.
type EndpointRegenerationEvent struct {
    regenContext *regenerationContext
    ep           *Endpoint
}

// Handle handles the regeneration event for the endpoint.
func (ev *EndpointRegenerationEvent) Handle(res chan interface{}) {

}

// InitEventQueue initializes the endpoint's event queue. Note that this
// function does not begin processing events off the queue, as that's left up
// to the caller to call Expose in order to allow other subsystems to access
// the endpoint. This function assumes that the endpoint ID has already been
// allocated!
//
// Having this be a separate function allows us to prepare
// the event queue while the endpoint is being validated (during restoration)
// so that when its metadata is resolved, events can be enqueued (such as
// visibility policy and bandwidth policy).
func (e *Endpoint) InitEventQueue() {
    e.eventQueue = eventqueue.NewEventQueueBuffered(fmt.Sprintf("endpoint-%d", e.ID), option.Config.EndpointQueueSize)
}

// Start assigns a Cilium Endpoint ID to the endpoint and prepares it to
// receive events from other subsystems.
//
// The endpoint must not already be exposed via the endpointmanager prior to
// calling Start(), as it assumes unconditional access over the Endpoint
// object.
func (e *Endpoint) Start(id uint16) {
    // No need to check liveness as an endpoint can only be deleted via the
    // API after it has been inserted into the manager.
    // 'e.ID' written below, read lock is not enough.
    e.unconditionalLock()
    defer e.unlock()

    e.ID = id
    e.UpdateLogger(map[string]interface{}{
        logfields.EndpointID: e.ID,
    })

    // Start goroutines that are responsible for handling events.
    e.startRegenerationFailureHandler()
    if e.eventQueue == nil {
        e.InitEventQueue()
    }
    e.eventQueue.Run()
    e.getLogger().Info("New endpoint")
}
