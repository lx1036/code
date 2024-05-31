package eventqueue

import (
    "github.com/sirupsen/logrus"
    "reflect"
    "sync"

    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/lock"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging/logfields"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/option"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/spanstat"
)

var (
    log = logging.DefaultLogger.WithField(logfields.LogSubsys, "eventqueue")
)

type eventStatistics struct {

    // waitEnqueue shows how long a given event was waiting on the queue before
    // it was actually processed.
    waitEnqueue spanstat.SpanStat

    // durationStat shows how long the actual processing of the event took. This
    // is the time for how long Handle() takes for the event.
    durationStat spanstat.SpanStat

    // waitConsumeOffQueue shows how long it took for the event to be consumed
    // plus the time it the event waited in the queue.
    waitConsumeOffQueue spanstat.SpanStat
}

type EventHandler interface {
    Handle(chan interface{})
}

// Event is an event that can be enqueued onto an EventQueue.
type Event struct {
    // Metadata is the information about the event which is sent
    // by its queuer. Metadata must implement the EventHandler interface in
    // order for the Event to be successfully processed by the EventQueue.
    Metadata EventHandler

    // eventResults is a channel on which the results of the event are sent.
    // It is populated by the EventQueue itself, not by the queuer. This channel
    // is closed if the event is cancelled.
    eventResults chan interface{}

    // cancelled signals that the given Event was not ran. This can happen
    // if the EventQueue processing this Event was closed before the Event was
    // Enqueued onto the Event queue, or if the Event was Enqueued onto an
    // EventQueue, and the EventQueue on which the Event was scheduled was
    // closed.
    cancelled chan struct{}

    // stats is a field which contains information about when this event is
    // enqueued, dequeued, etc.
    stats eventStatistics

    // enqueued is an atomic boolean that specifies whether this event has
    // been enqueued on an EventQueue.
    enqueued int32
}

func (ev *Event) printStats(q *EventQueue) {
    if option.Config.Debug {
        q.getLogger().WithFields(logrus.Fields{
            "eventType":                    reflect.TypeOf(ev.Metadata).String(),
            "eventHandlingDuration":        ev.stats.durationStat.Total(),
            "eventEnqueueWaitTime":         ev.stats.waitEnqueue.Total(),
            "eventConsumeOffQueueWaitTime": ev.stats.waitConsumeOffQueue.Total(),
        }).Debug("EventQueue event processing statistics")
    }
}

func NewEvent(meta EventHandler) *Event {
    return &Event{
        Metadata:     meta,
        eventResults: make(chan interface{}, 1),
        cancelled:    make(chan struct{}),
        stats:        eventStatistics{},
    }
}

// EventQueue is a structure which is utilized to handle Events in a first-in,
// first-out order. An EventQueue may be closed, in which case all events which
// are queued up, but have not been processed yet, will be cancelled (i.e., not
// ran). It is guaranteed that no events will be scheduled onto an EventQueue
// after it has been closed; if any event is attempted to be scheduled onto an
// EventQueue after it has been closed, it will be cancelled immediately. For
// any event to be processed by the EventQueue, it must implement the
// `EventHandler` interface. This allows for different types of events to be
// processed by anything which chooses to utilize an `EventQueue`.
type EventQueue struct {
    // events represents the queue of events. This should always be a buffered
    // channel.
    events chan *Event

    // close is closed once the EventQueue has been closed.
    close chan struct{}

    // drain is closed when the EventQueue is stopped. Any Event which is
    // Enqueued after this channel is closed will be cancelled / not processed
    // by the queue. If an Event has been Enqueued, but has not been processed
    // before this channel is closed, it will be cancelled and not processed
    // as well.
    drain chan struct{}

    // eventQueueOnce is used to ensure that the EventQueue business logic can
    // only be ran once.
    eventQueueOnce sync.Once

    // closeOnce is used to ensure that the EventQueue can only be closed once.
    closeOnce sync.Once

    // name is used to differentiate this EventQueue from other EventQueues that
    // are also running in logs
    name string

    eventsMu lock.RWMutex

    // eventsClosed is a channel that's closed when the event loop (Run())
    // terminates.
    eventsClosed chan struct{}
}

func (q *EventQueue) getLogger() *logrus.Entry {
    return log.WithFields(
        logrus.Fields{
            "name": q.name,
        })
}

// Run consumes events that have been queued for this EventQueue. It
// is presumed that the eventQueue is a buffered channel with a length of one
// (i.e., only one event can be processed at a time). All business logic for
// handling queued events is contained within this function. The events in the
// queue must implement the EventHandler interface. If the event queue is
// closed, then all events which were queued up, but not processed, are
// cancelled; any event which is currently being processed will not be
// cancelled.
func (q *EventQueue) Run() {
    if q.notSafeToAccess() {
        return
    }

    go q.run()
}

func (q *EventQueue) notSafeToAccess() bool {
    return q == nil || q.close == nil || q.drain == nil || q.events == nil
}

func (q *EventQueue) run() {
    q.eventQueueOnce.Do(func() {
        defer close(q.eventsClosed)
        for ev := range q.events {
            select {
            case <-q.drain:
                ev.stats.waitConsumeOffQueue.End(false)
                close(ev.cancelled)
                close(ev.eventResults)
                ev.printStats(q)
            default:
                ev.stats.waitConsumeOffQueue.End(true)
                ev.stats.durationStat.Start()
                ev.Metadata.Handle(ev.eventResults)
                // Always indicate success for now.
                ev.stats.durationStat.End(true)
                // Ensures that no more results can be sent as the event has
                // already been processed.
                ev.printStats(q)
                close(ev.eventResults)
            }
        }
    })
}

func (q *EventQueue) Stop() {
    if q.notSafeToAccess() {
        return
    }

    q.closeOnce.Do(func() {
        q.getLogger().Debug("stopping EventQueue")
        // Any event that is sent to the queue at this point will be cancelled
        // immediately in Enqueue().
        close(q.drain)

        // Signal that the queue has been drained.
        close(q.close)

        q.eventsMu.Lock()
        close(q.events)
        q.eventsMu.Unlock()
    })
}

// NewEventQueueBuffered returns an EventQueue with a capacity of,
// numBufferedEvents at a time, and all other needed fields initialized.
func NewEventQueueBuffered(name string, numBufferedEvents int) *EventQueue {
    log.WithFields(logrus.Fields{
        "name":              name,
        "numBufferedEvents": numBufferedEvents,
    }).Debug("creating new EventQueue")
    return &EventQueue{
        name: name,
        // Up to numBufferedEvents can be Enqueued until Enqueueing blocks.
        events:       make(chan *Event, numBufferedEvents),
        close:        make(chan struct{}),
        drain:        make(chan struct{}),
        eventsClosed: make(chan struct{}),
    }
}
