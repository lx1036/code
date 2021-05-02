package events

import (
	"sort"
	"strings"
	"sync"
	"time"

	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/utils"
)

// Request holds a set of parameters by which Event objects may be screened.
// The caller may want events that occurred within a specific timeframe
// or of a certain type, which may be specified in the *Request object
// they pass to an EventManager function
type Request struct {
	// events falling before StartTime do not satisfy the request. StartTime
	// must be left blank in calls to WatchEvents
	StartTime time.Time
	// events falling after EndTime do not satisfy the request. EndTime
	// must be left blank in calls to WatchEvents
	EndTime time.Time
	// EventType is a map that specifies the type(s) of events wanted
	EventType map[v1.EventType]bool
	// allows the caller to put a limit on how many
	// events to receive. If there are more events than MaxEventsReturned
	// then the most chronologically recent events in the time period
	// specified are returned. Must be >= 1
	MaxEventsReturned int
	// the absolute container name for which the event occurred
	ContainerName string
	// if IncludeSubcontainers is false, only events occurring in the specific
	// container, and not the subcontainers, will be returned
	IncludeSubcontainers bool
}

// returns a pointer to an initialized Request object
func NewRequest() *Request {
	return &Request{
		EventType:            map[v1.EventType]bool{},
		IncludeSubcontainers: false,
		MaxEventsReturned:    10,
	}
}

type EventChannel struct {
	// Watch ID. Can be used by the caller to request cancellation of watch events.
	watchID int
	// Channel on which the caller can receive watch events.
	channel chan *v1.Event
}

// initialized by a call to WatchEvents(), a watch struct will then be added
// to the events slice of *watch objects. When AddEvent() finds an event that
// satisfies the request parameter of a watch object in events.watchers,
// it will send that event out over the watch object's channel. The caller that
// called WatchEvents will receive the event over the channel provided to
// WatchEvents
type watch struct {
	// request parameters passed in by the caller of WatchEvents()
	request *Request
	// a channel used to send event back to the caller.
	eventChannel *EventChannel
}

// EventManager is implemented by Events. It provides two ways to monitor
// events and one way to add events
type EventManager interface {
	// WatchEvents() allows a caller to register for receiving events based on the specified request.
	// On successful registration, an EventChannel object is returned.
	WatchEvents(request *Request) (*EventChannel, error)
	// GetEvents() returns all detected events based on the filters specified in request.
	GetEvents(request *Request) ([]*v1.Event, error)
	// AddEvent allows the caller to add an event to an EventManager
	// object
	AddEvent(event *v1.Event) error
	// Cancels a previously requested watch event.
	StopWatch(watchID int)
}

// Policy specifying how many events to store.
// MaxAge is the max duration for which to keep events.
// MaxNumEvents is the max number of events to keep (-1 for no limit).
type StoragePolicy struct {
	// Defaults limites, used if a per-event limit is not set.
	DefaultMaxAge       time.Duration
	DefaultMaxNumEvents int

	// Per-event type limits.
	PerTypeMaxAge       map[v1.EventType]time.Duration
	PerTypeMaxNumEvents map[v1.EventType]int
}

func DefaultStoragePolicy() StoragePolicy {
	return StoragePolicy{
		DefaultMaxAge:       24 * time.Hour,
		DefaultMaxNumEvents: 100000,
		PerTypeMaxAge:       make(map[v1.EventType]time.Duration),
		PerTypeMaxNumEvents: make(map[v1.EventType]int),
	}
}

// events provides an implementation for the EventManager interface.
type events struct {
	// eventStore holds the events by event type.
	eventStore map[v1.EventType]*utils.TimedStore
	// map of registered watchers keyed by watch id.
	watchers map[int]*watch
	// lock guarding the eventStore.
	eventsLock sync.RWMutex
	// lock guarding watchers.
	watcherLock sync.RWMutex
	// last allocated watch id.
	lastID int
	// Event storage policy.
	storagePolicy StoragePolicy
}

func (e *events) WatchEvents(request *Request) (*EventChannel, error) {
	panic("implement me")
}

// determines if an event occurs within the time set in the request object and is the right type
func checkIfEventSatisfiesRequest(request *Request, event *v1.Event) bool {
	startTime := request.StartTime
	endTime := request.EndTime
	eventTime := event.Timestamp
	if !startTime.IsZero() {
		if startTime.After(eventTime) {
			return false
		}
	}
	if !endTime.IsZero() {
		if endTime.Before(eventTime) {
			return false
		}
	}
	if !request.EventType[event.EventType] {
		return false
	}
	if request.ContainerName != "" {
		return isSubcontainer(request, event)
	}
	return true
}

// If the request wants all subcontainers, this returns if the request's
// container path is a prefix of the event container path.  Otherwise,
// it checks that the container paths of the event and request are
// equivalent
func isSubcontainer(request *Request, event *v1.Event) bool {
	if request.IncludeSubcontainers {
		return request.ContainerName == "/" || strings.HasPrefix(event.ContainerName+"/", request.ContainerName+"/")
	}
	return event.ContainerName == request.ContainerName
}

type byTimestamp []*v1.Event

// functions necessary to implement the sort interface on the Events struct
func (e byTimestamp) Len() int {
	return len(e)
}

func (e byTimestamp) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e byTimestamp) Less(i, j int) bool {
	return e[i].Timestamp.Before(e[j].Timestamp)
}

// sorts and returns up to the last MaxEventsReturned chronological elements
func getMaxEventsReturned(request *Request, eSlice []*v1.Event) []*v1.Event {
	sort.Sort(byTimestamp(eSlice))
	n := request.MaxEventsReturned
	if n >= len(eSlice) || n <= 0 {
		return eSlice
	}
	return eSlice[len(eSlice)-n:]
}

// method of Events object that screens Event objects found in the eventStore
// attribute and if they fit the parameters passed by the Request object,
// adds it to a slice of *Event objects that is returned. If both MaxEventsReturned
// and StartTime/EndTime are specified in the request object, then only
// up to the most recent MaxEventsReturned events in that time range are returned.
func (e *events) GetEvents(request *Request) ([]*v1.Event, error) {
	returnEventList := []*v1.Event{}
	e.eventsLock.RLock()
	defer e.eventsLock.RUnlock()
	for eventType, fetch := range request.EventType {
		if !fetch {
			continue
		}
		timedStore, ok := e.eventStore[eventType]
		if !ok {
			continue
		}

		res := timedStore.InTimeRange(request.StartTime, request.EndTime, request.MaxEventsReturned)
		for _, in := range res {
			e := in.(*v1.Event)
			if checkIfEventSatisfiesRequest(request, e) {
				returnEventList = append(returnEventList, e)
			}
		}
	}
	returnEventList = getMaxEventsReturned(request, returnEventList)

	return returnEventList, nil
}

func (e *events) AddEvent(event *v1.Event) error {
	panic("implement me")
}

func (e *events) StopWatch(watchID int) {
	panic("implement me")
}

// returns a pointer to an initialized Events object.
func NewEventManager(storagePolicy StoragePolicy) EventManager {
	return &events{
		eventStore:    make(map[v1.EventType]*utils.TimedStore),
		watchers:      make(map[int]*watch),
		storagePolicy: storagePolicy,
	}
}
