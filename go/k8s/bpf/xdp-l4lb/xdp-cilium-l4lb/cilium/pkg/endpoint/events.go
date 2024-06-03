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

// EndpointRegenerationResult contains the results of an endpoint regeneration.
type EndpointRegenerationResult struct {
    err error
}

// Handle handles the regeneration event for the endpoint.
func (ev *EndpointRegenerationEvent) Handle(res chan interface{}) {
    e := ev.ep
    regenContext := ev.regenContext

    err := e.rlockAlive()
    if err != nil {
        e.logDisconnectedMutexAction(err, "before regeneration")
        res <- &EndpointRegenerationResult{
            err: err,
        }

        return
    }
    e.runlock()

    // We should only queue the request after we use all the endpoint's
    // lock/unlock. Otherwise this can get a deadlock if the endpoint is
    // being deleted at the same time. More info PR-1777.
    doneFunc, err := e.owner.QueueEndpointBuild(regenContext.parentContext, uint64(e.ID))
    if err != nil {
        e.getLogger().WithError(err).Warning("unable to queue endpoint build")
    } else if doneFunc != nil {
        e.getLogger().Debug("Dequeued endpoint from build queue")

        regenContext.DoneFunc = doneFunc

        err = ev.ep.regenerate(ev.regenContext)

        doneFunc()
        e.notifyEndpointRegeneration(err)
    } else {
        // If another build has been queued for the endpoint, that means that
        // that build will be able to take care of all of the work needed to
        // regenerate the endpoint at this current point in time; queueing
        // another build is a waste of resources.
        e.getLogger().Debug("build not queued for endpoint because another build has already been queued")
    }

    res <- &EndpointRegenerationResult{
        err: err,
    }
    return
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
