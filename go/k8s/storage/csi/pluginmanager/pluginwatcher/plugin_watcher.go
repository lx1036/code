package pluginwatcher

import (
	"k8s-lx1036/k8s/storage/csi/pluginmanager/cache"

	"github.com/fsnotify/fsnotify"

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
