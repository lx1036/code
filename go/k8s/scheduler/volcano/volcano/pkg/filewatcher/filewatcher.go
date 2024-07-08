package filewatcher

import "github.com/fsnotify/fsnotify"

// FileWatcher is an interface watching the underlying OS file path.
type FileWatcher interface {
	Events() chan fsnotify.Event
	Errors() chan error
	Close()
}

type fileWatcher struct {
	watcher *fsnotify.Watcher
}

func (w *fileWatcher) Events() chan fsnotify.Event {
	if w == nil || w.watcher == nil {
		return nil
	}

	return w.watcher.Events
}

func (w *fileWatcher) Errors() chan error {
	if w == nil || w.watcher == nil {
		return nil
	}

	return w.watcher.Errors
}

func (w *fileWatcher) Close() {
	if w == nil || w.watcher == nil {
		return
	}

	w.watcher.Close()
}

func NewFileWatcher(path string) (FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	err = watcher.Add(path)
	if err != nil {
		return nil, err
	}

	return &fileWatcher{
		watcher: watcher,
	}, nil
}
