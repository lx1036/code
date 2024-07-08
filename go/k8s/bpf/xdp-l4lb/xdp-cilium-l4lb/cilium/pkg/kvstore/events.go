package kvstore

import "sync"

// Watcher represents a KVstore watcher
type Watcher struct {
	// Events is the channel to which change notifications will be sent to
	Events EventChan `json:"-"`

	Name      string `json:"name"`
	Prefix    string `json:"prefix"`
	stopWatch stopChan

	// stopOnce guarantees that Stop() is only called once
	stopOnce sync.Once

	// stopWait is the wait group to wait for watchers to exit gracefully
	stopWait sync.WaitGroup
}
