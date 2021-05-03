package common

import (
	"sync"

	"k8s.io/utils/inotify"
)

// Watcher for container-related inotify events in the cgroup hierarchy.
//
// Implementation is thread-safe.
type InotifyWatcher struct {
	// Underlying inotify watcher.
	watcher *inotify.Watcher

	// Map of containers being watched to cgroup paths watched for that container.
	containersWatched map[string]map[string]bool

	// Lock for all datastructure access.
	lock sync.Mutex
}

func NewInotifyWatcher() (*InotifyWatcher, error) {
	w, err := inotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &InotifyWatcher{
		watcher:           w,
		containersWatched: make(map[string]map[string]bool),
	}, nil
}
