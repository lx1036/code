package agent

import (
    "context"
    "errors"
    "github.com/sirupsen/logrus"
    "golang.org/x/sys/unix"
    "os"
    "time"

    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/api/v1/models"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/bpf"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/lock"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging/logfields"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/monitor/agent/consumer"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/monitor/agent/listener"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/monitor/payload"

    "github.com/cilium/ebpf"
    "github.com/cilium/ebpf/perf"
)

/*
struct {
	__uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
	__uint(key_size, sizeof(__u32));
	__uint(value_size, sizeof(__u32));
	__uint(pinning, LIBBPF_PIN_BY_NAME);
	__uint(max_entries, __NR_CPUS__);
} EVENTS_MAP __section_maps_btf;
*/

const eventsMapName = "cilium_events"

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

    // 数据源，读取自 bpf_perf_event_output() 的数据
    events        *ebpf.Map
    monitorEvents *perf.Reader
}

func (a *Agent) Context() context.Context {
    return a.ctx
}

// AttachToEventsMap opens the events perf ring buffer and makes it ready for
// consumption, such that any subscribed consumers may receive events
// from it. This function is to be called once the events map has been set up.
func (a *Agent) AttachToEventsMap(nPages int) error {
    a.Lock()
    defer a.Unlock()

    if a.events != nil {
        return errors.New("events map already attached")
    }

    // assert that we can actually connect the monitor
    path := bpf.MapPath(eventsMapName)
    eventsMap, err := ebpf.LoadPinnedMap(path, nil)
    if err != nil {
        return err
    }

    a.events = eventsMap
    a.MonitorStatus = models.MonitorStatus{
        Cpus:     int64(eventsMap.MaxEntries()),
        Npages:   int64(nPages), // 64
        Pagesize: int64(os.Getpagesize()),
    }

    // start the perf reader if we already have subscribers
    if a.hasSubscribersLocked() {
        a.startPerfReaderLocked()
    }

    return nil
}

// hasSubscribersLocked returns true if there are listeners subscribed to the
// agent right now.
func (a *Agent) hasSubscribersLocked() bool {
    return len(a.listeners)+len(a.consumers) != 0
}

// startPerfReaderLocked starts the perf reader. This should only be
// called if there are no other readers already running.
// The goroutine is spawned with a context derived from m.Context() and the
// cancelFunc is assigned to perfReaderCancel. Note that cancelling m.Context()
// (e.g. on program shutdown) will also cancel the derived context.
// Note: it is critical to hold the lock for this operation.
func (a *Agent) startPerfReaderLocked() {
    if a.events == nil {
        return // not attached to events map yet
    }

    a.perfReaderCancel() // don't leak any old readers, just in case.
    perfEventReaderCtx, cancel := context.WithCancel(a.ctx)
    a.perfReaderCancel = cancel
    go a.handleEvents(perfEventReaderCtx)
}

// handleEvents reads events from the perf buffer and processes them. It
// will exit when stopCtx is done. Note, however, that it will block in the
// Poll call but assumes enough events are generated that these blocks are
// short.
func (a *Agent) handleEvents(stopCtx context.Context) {
    scopedLog := log.WithField(logfields.StartTime, time.Now())
    scopedLog.Info("Beginning to read perf buffer")
    defer scopedLog.Info("Stopped reading perf buffer")

    bufferSize := int(a.Pagesize * a.Npages) // pagesize * 64
    monitorEvents, err := perf.NewReader(a.events, bufferSize)
    if err != nil {
        scopedLog.WithError(err).Fatal("Cannot initialise BPF perf ring buffer sockets")
    }
    defer func() {
        monitorEvents.Close()
        a.Lock()
        a.monitorEvents = nil
        a.Unlock()
    }()

    a.Lock() // lock???
    a.monitorEvents = monitorEvents
    a.Unlock()

    // block for poll read bpf events
    for !isCtxDone(stopCtx) {
        record, err := monitorEvents.Read()
        switch {
        case isCtxDone(stopCtx):
            return
        case err != nil:
            if perf.IsUnknownEvent(err) {
                a.Lock()
                a.MonitorStatus.Unknown++
                a.Unlock()
            } else {
                scopedLog.WithError(err).Warn("Error received while reading from perf buffer")
                if errors.Is(err, unix.EBADFD) {
                    return
                }
            }
            continue
        }

        a.processPerfRecord(scopedLog, record)
    }
}

// processPerfRecord processes a record from the datapath and sends it to any
// registered subscribers
func (a *Agent) processPerfRecord(scopedLog *logrus.Entry, record perf.Record) {
    a.Lock()
    defer a.Unlock()

    if record.LostSamples > 0 {
        a.MonitorStatus.Lost += int64(record.LostSamples)
        a.notifyPerfEventLostLocked(record.LostSamples, record.CPU)
        a.sendToListenersLocked(&payload.Payload{
            CPU:  record.CPU,
            Lost: record.LostSamples,
            Type: payload.RecordLost,
        })
    } else {
        a.notifyPerfEventLocked(record.RawSample, record.CPU)
        a.sendToListenersLocked(&payload.Payload{
            CPU:  record.CPU,
            Data: record.RawSample,
            Type: payload.EventSample,
        })
    }
}

// notifyEventToConsumersLocked notifies all consumers about lost events.
// The caller must hold the monitor lock.
func (a *Agent) notifyPerfEventLostLocked(numLostEvents uint64, cpu int) {
    for mc := range a.consumers {
        mc.NotifyPerfEventLost(numLostEvents, cpu)
    }
}

// notifyPerfEventLocked notifies all consumers about a perf event.
// The caller must hold the monitor lock.
func (a *Agent) notifyPerfEventLocked(data []byte, cpu int) {
    for mc := range a.consumers {
        mc.NotifyPerfEvent(data, cpu)
    }
}

// sendToListenersLocked enqueues the payload to all listeners while holding the monitor lock.
func (a *Agent) sendToListenersLocked(pl *payload.Payload) {
    for ml := range a.listeners {
        ml.Enqueue(pl)
    }
}

// RegisterNewListener adds the new MonitorListener to the global list.
// It also spawns a singleton goroutine to read and distribute the events.
func (a *Agent) RegisterNewListener(newListener listener.MonitorListener) {
    if a == nil {
        return
    }

    a.Lock()
    defer a.Unlock()

    if isCtxDone(a.ctx) {
        log.Debug("RegisterNewListener called on stopped monitor")
        newListener.Close()
        return
    }

    // If this is the first listener, start the perf reader. ???
    if !a.hasSubscribersLocked() {
        a.startPerfReaderLocked()
    }

    version := newListener.Version()
    switch version {
    case listener.Version1_2:
        a.listeners[newListener] = struct{}{}

    default:
        newListener.Close()
        log.WithField("version", version).Error("Closing listener from unsupported monitor client version")
    }

    log.WithFields(logrus.Fields{
        "count.listener": len(a.listeners),
        "version":        version,
    }).Debug("New listener connected")
}

// RemoveListener deletes the MonitorListener from the list, closes its queue,
// and stops perfReader if this is the last subscriber
func (a *Agent) RemoveListener(ml listener.MonitorListener) {
    if a == nil {
        return
    }

    a.Lock()
    defer a.Unlock()

    // Remove the listener and close it.
    delete(a.listeners, ml)
    log.WithFields(logrus.Fields{
        "count.listener": len(a.listeners),
        "version":        ml.Version(),
    }).Debug("Removed listener")
    ml.Close()

    // If this was the final listener, shutdown the perf reader and unmap our
    // ring buffer readers. This tells the kernel to not emit this data.
    // Note: it is critical to hold the lock and check the number of listeners.
    // This guards against an older generation listener calling the
    // current generation perfReaderCancel
    if !a.hasSubscribersLocked() {
        a.perfReaderCancel()
    }
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

// isCtxDone is a utility function that returns true when the context's Done()
// channel is closed. It is intended to simplify goroutines that need to check
// this multiple times in their loop.
func isCtxDone(ctx context.Context) bool {
    select {
    case <-ctx.Done():
        return true
    default:
        return false
    }
}
