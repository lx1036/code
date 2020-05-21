package sync

import (
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// Implements SynchronizerManager interface.
type synchronizerManager struct {
	client kubernetes.Interface
}

// Secret implements synchronizer manager. See SynchronizerManager interface for more information.
func (self *synchronizerManager) Secret(namespace, name string) syncApi.Synchronizer {
	return &secretSynchronizer{
		namespace:      namespace,
		name:           name,
		client:         self.client,
		actionHandlers: make(map[watch.EventType][]syncApi.ActionHandlerFunction),
	}
}

// NewSynchronizerManager creates new instance of SynchronizerManager.
func NewSynchronizerManager(client kubernetes.Interface) syncApi.SynchronizerManager {
	return &synchronizerManager{client: client}
}
