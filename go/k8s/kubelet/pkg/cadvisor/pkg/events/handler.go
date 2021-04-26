package events

import (
	"time"

	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"
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

type EventChannel struct {
	// Watch ID. Can be used by the caller to request cancellation of watch events.
	watchID int
	// Channel on which the caller can receive watch events.
	channel chan *v1.Event
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
