package pluginwatcher

import (
	"path/filepath"
	"testing"

	"github.com/fsnotify/fsnotify"

	"k8s.io/klog/v2"
)

// INFO: https://github.com/fsnotify/fsnotify , 支持 macos

func TestFsNotify(test *testing.T) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		klog.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				klog.Infof("event: %v", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					klog.Infof("modified file: %s", event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				klog.Infof("error: %v", err)
			}
		}
	}()

	watchPath, err := filepath.Abs("./tmp")
	if err != nil {
		klog.Fatal(err)
	}
	err = watcher.Add(watchPath)
	if err != nil {
		klog.Fatal(err)
	}

	<-done
}
