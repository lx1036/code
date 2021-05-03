package watcher

// SubcontainerEventType indicates an addition or deletion event.
type ContainerEventType int

const (
	ContainerAdd ContainerEventType = iota
	ContainerDelete
)

type ContainerWatchSource int

const (
	Raw ContainerWatchSource = iota
)

// ContainerEvent represents a
type ContainerEvent struct {
	// The type of event that occurred.
	EventType ContainerEventType

	// The full container name of the container where the event occurred.
	Name string

	// The watcher that detected this change event
	WatchSource ContainerWatchSource
}

type ContainerWatcher interface {
	// Registers a channel to listen for events affecting subcontainers (recursively).
	Start(events chan ContainerEvent) error

	// Stops watching for subcontainer changes.
	Stop() error
}
