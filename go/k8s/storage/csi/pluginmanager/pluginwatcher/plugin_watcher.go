package pluginwatcher

import (
	"fmt"
	"os"
	"strings"
	"time"

	"k8s-lx1036/k8s/storage/csi/pluginmanager/cache"

	"github.com/fsnotify/fsnotify"

	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/kubelet/util"
	"k8s.io/kubernetes/pkg/util/filesystem"
)

// Watcher is the plugin watcher
type Watcher struct {
	path                string
	fs                  filesystem.Filesystem
	fsWatcher           *fsnotify.Watcher
	stopped             chan struct{}
	desiredStateOfWorld cache.DesiredStateOfWorld
}

// NewWatcher provides a new watcher for socket registration
func NewWatcher(sockDir string, desiredStateOfWorld cache.DesiredStateOfWorld) *Watcher {
	return &Watcher{
		path:                sockDir,
		fs:                  &filesystem.DefaultFs{},
		desiredStateOfWorld: desiredStateOfWorld,
	}
}

func (w *Watcher) init() error {
	klog.Infof("Ensuring Plugin directory at %s ", w.path)

	if err := w.fs.MkdirAll(w.path, 0755); err != nil {
		return fmt.Errorf("error (re-)creating root %s: %v", w.path, err)
	}

	return nil
}

// Walks through the plugin directory discover any existing plugin sockets.
// Ignore all errors except root dir not being walkable
func (w *Watcher) traversePluginDir(dir string) error {
	return w.fs.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if path == dir {
				return fmt.Errorf("error accessing path: %s error: %v", path, err)
			}

			klog.Errorf("error accessing path: %s error: %v", path, err)
			return nil
		}

		switch mode := info.Mode(); {
		case mode.IsDir():
			if err := w.fsWatcher.Add(path); err != nil {
				return fmt.Errorf("failed to watch %s, err: %v", path, err)
			}

			klog.Infof("filesystem watcher path %s", path)
		case mode&os.ModeSocket != 0:
			event := fsnotify.Event{
				Name: path,
				Op:   fsnotify.Create,
			}
			//TODO: Handle errors by taking corrective measures
			if err := w.handleCreateEvent(event); err != nil {
				klog.Errorf("error %v when handling create event: %s", err, event)
			}
		default:
			klog.Infof("Ignoring file %s with mode %v", path, mode)
		}

		return nil
	})
}

// Handle filesystem notify event.
// Files names:
// - MUST NOT start with a '.'
func (w *Watcher) handleCreateEvent(event fsnotify.Event) error {
	klog.Infof("Handling create event: %v", event)

	fileInfo, err := os.Stat(event.Name)
	if err != nil {
		return fmt.Errorf("stat file %s failed: %v", event.Name, err)
	}

	if strings.HasPrefix(fileInfo.Name(), ".") {
		klog.Infof("Ignoring file (starts with '.'): %s", fileInfo.Name())
		return nil
	}

	if !fileInfo.IsDir() {
		isSocket, err := util.IsUnixDomainSocket(util.NormalizePath(event.Name))
		if err != nil {
			return fmt.Errorf("failed to determine if file: %s is a unix domain socket: %v", event.Name, err)
		}
		if !isSocket {
			klog.Infof("Ignoring non socket file %s", fileInfo.Name())
			return nil
		}

		return w.handlePluginRegistration(event.Name)
	}

	return w.traversePluginDir(event.Name)
}

func (w *Watcher) handlePluginRegistration(socketPath string) error {
	klog.Infof("Adding socket path or updating timestamp %s to desired state cache", socketPath)
	err := w.desiredStateOfWorld.AddOrUpdatePlugin(socketPath)
	if err != nil {
		return fmt.Errorf("error adding socket path %s or updating timestamp to desired state cache: %v", socketPath, err)
	}
	return nil
}

func (w *Watcher) handleDeleteEvent(event fsnotify.Event) {
	klog.Infof("Handling delete event: %v", event)

	socketPath := event.Name
	klog.Infof("Removing socket path %s from desired state cache", socketPath)
	w.desiredStateOfWorld.RemovePlugin(socketPath)
}

// @see pkg/util/filesystem/watcher.go

// Start watches for the creation and deletion of plugin sockets at the path
func (w *Watcher) Start(stopCh <-chan struct{}) error {
	klog.Infof("Plugin Watcher Start at %s", w.path)

	w.stopped = make(chan struct{})

	// Creating the directory to be watched if it doesn't exist yet,
	// and walks through the directory to discover the existing plugins.
	if err := w.init(); err != nil {
		return err
	}

	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to start plugin fsWatcher, err: %v", err)
	}
	w.fsWatcher = fsWatcher

	// Traverse plugin dir and add filesystem watchers before starting the plugin processing goroutine.
	if err := w.traversePluginDir(w.path); err != nil {
		klog.Errorf("failed to traverse plugin socket path %q, err: %v", w.path, err)
	}

	go func(fsWatcher *fsnotify.Watcher) {
		defer close(w.stopped)
		for {
			select {
			case event := <-fsWatcher.Events:
				//TODO: Handle errors by taking corrective measures
				if event.Op&fsnotify.Create == fsnotify.Create {
					err := w.handleCreateEvent(event)
					if err != nil {
						klog.Errorf("error %v when handling create event: %s", err, event)
					}
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					w.handleDeleteEvent(event)
				}
				continue
			case err := <-fsWatcher.Errors:
				if err != nil {
					klog.Errorf("fsWatcher received error: %v", err)
				}
				continue
			case <-stopCh:
				// In case of plugin watcher being stopped by plugin manager, stop
				// probing the creation/deletion of plugin sockets.
				// Also give all pending go routines a chance to complete
				select {
				case <-w.stopped:
				case <-time.After(11 * time.Second):
					klog.Errorf("timeout on stopping watcher")
				}
				w.fsWatcher.Close()
				return
			}
		}
	}(fsWatcher)

	return nil
}
