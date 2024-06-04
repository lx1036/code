package agent

import (
    "context"

    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/api/v1/models"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/lock"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/monitor/agent/consumer"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/monitor/agent/listener"

    "github.com/cilium/ebpf"
    "github.com/cilium/ebpf/perf"
)

// Agent structure for centralizing the responsibilities of the main events
// reader.
// There is some racey-ness around perfReaderCancel since it replaces on every
// perf reader start. In the event that a MonitorListener from a previous
// generation calls its cleanup after the start of the new perf reader, we
// might call the new, and incorrect, cancel function. We guard for this by
// checking the number of listeners during the cleanup call. The perf reader
// must have at least one MonitorListener (since it started) so no cancel is called.
// If it doesn't, the cancel is the correct behavior (the older generation
// cancel must have been called for us to get this far anyway).
type Agent struct {
    lock.Mutex
    models.MonitorStatus

    ctx              context.Context
    perfReaderCancel context.CancelFunc

    // listeners are external cilium monitor clients which receive raw
    // gob-encoded payloads
    listeners map[listener.MonitorListener]struct{}
    // consumers are internal clients which receive decoded messages
    consumers map[consumer.MonitorConsumer]struct{}

    events        *ebpf.Map
    monitorEvents *perf.Reader
}

// NewAgent starts a new monitor agent instance which distributes monitor events to registered listeners.
// Once the datapath is set up, AttachToEventsMap needs to be called to receive events from the perf ring buffer.
// Otherwise, only user space events received via SendEvent are distributed registered listeners.
// Internally, the agent spawns a singleton goroutine reading events from
// the BPF perf ring buffer and provides an interface to pass in non-BPF events.
// The instance can be stopped by cancelling ctx, which will stop the perf reader
// go routine and close all registered listeners.
// Note that the perf buffer reader is started only when listeners are
// connected.
func NewAgent(ctx context.Context) *Agent {
    return &Agent{
        ctx:              ctx,
        listeners:        make(map[listener.MonitorListener]struct{}),
        consumers:        make(map[consumer.MonitorConsumer]struct{}),
        perfReaderCancel: func() {}, // no-op to avoid doing null checks everywhere
    }
}
