package workqueue

import (
	"k8s.io/utils/clock"
	"sync"
	"time"
)

const (
	defaultUnfinishedWorkUpdatePeriod = 500 * time.Millisecond
)

type Interface interface {
	Add(item interface{})
	Len() int
	Get() (item interface{}, shutdown bool)
	Done(item interface{})
	ShutDown()
	ShutDownWithDrain()
	ShuttingDown() bool
}

type t interface{}
type empty struct{}
type set map[t]empty

func (s set) has(item t) bool {
	_, exists := s[item]
	return exists
}

func (s set) insert(item t) {
	s[item] = empty{}
}

func (s set) delete(item t) {
	delete(s, item)
}

func (s set) len() int {
	return len(s)
}

type Type struct {
	cond *sync.Cond

	queue []t

	dirty set // 标记同一个 item 不被重复 add

	processing set // item 正在被消费

	shuttingDown bool
	drain        bool
}

func (q *Type) Add(item interface{}) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	if q.shuttingDown {
		return
	}

	if q.dirty.has(item) {
		return
	}

	//q.metrics.add(item)

	q.dirty.insert(item)
	if q.processing.has(item) {
		return
	}

	q.queue = append(q.queue, item)
	q.cond.Signal() // wakes one goroutine waiting on c, if there is any
}

// Get blocks until it can return an item to be processed.
func (q *Type) Get() (item interface{}, shutdown bool) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	for len(q.queue) == 0 && !q.shuttingDown {
		q.cond.Wait() // block
	}

	if len(q.queue) == 0 {
		// We must be shutting down.
		return nil, true
	}

	item = q.queue[0]
	// The underlying array still exists and reference this object, so the object will not be garbage collected.
	q.queue[0] = nil
	q.queue = q.queue[1:]

	//q.metrics.get(item)

	q.processing.insert(item)
	q.dirty.delete(item)

	return item, false
}

// Done marks item as done processing, and if it has been marked as dirty again
// while it was being processed, it will be re-added to the queue for
// re-processing.
func (q *Type) Done(item interface{}) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	//q.metrics.done(item)

	q.processing.delete(item)
	if q.dirty.has(item) {
		q.queue = append(q.queue, item)
		q.cond.Signal()
	} else if q.processing.len() == 0 {
		q.cond.Signal()
	}
}

func (q *Type) Len() int {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	return len(q.queue)
}

type QueueConfig struct {
	// Name for the queue. If unnamed, the metrics will not be registered.
	Name string

	// MetricsProvider optionally allows specifying a metrics provider to use for the queue
	// instead of the global provider.
	//MetricsProvider MetricsProvider

	// Clock ability to inject real or fake clock for testing purposes.
	Clock clock.WithTicker
}

func newQueueWithConfig(config QueueConfig, updatePeriod time.Duration) *Type {

	return newQueue(
		config.Clock,
		//metricsFactory.newQueueMetrics(config.Name, config.Clock),
		updatePeriod,
	)
}

func newQueue(c clock.WithTicker, updatePeriod time.Duration) *Type {
	t := &Type{
		//clock:                      c,
		dirty:      set{},
		processing: set{},
		cond:       sync.NewCond(&sync.Mutex{}),
		//metrics:                    metrics,
		//unfinishedWorkUpdatePeriod: updatePeriod,
	}

	return t
}

func New() *Type {
	return NewWithConfig(QueueConfig{
		Name: "",
	})
}

func NewWithConfig(config QueueConfig) *Type {
	return newQueueWithConfig(config, defaultUnfinishedWorkUpdatePeriod)
}
