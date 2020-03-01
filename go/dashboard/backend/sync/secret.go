package sync

import (
	"fmt"
	syncApi "k8s-lx1036/dashboard/backend/sync/api"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"sync"
)

// Implements Synchronizer interface. See Synchronizer for more information.
type secretSynchronizer struct {
	namespace string
	name      string

	secret         *v1.Secret
	client         kubernetes.Interface
	actionHandlers map[watch.EventType][]syncApi.ActionHandlerFunction
	errChan        chan error
	poller         syncApi.Poller

	mux sync.Mutex
}

func (self *secretSynchronizer) Start() {
	panic("implement me")
}

func (self *secretSynchronizer) Error() chan error {
	panic("implement me")
}

func (self *secretSynchronizer) Create(runtime.Object) error {
	panic("implement me")
}

func (self *secretSynchronizer) Get() runtime.Object {
	panic("implement me")
}

func (self *secretSynchronizer) Update(runtime.Object) error {
	panic("implement me")
}

func (self *secretSynchronizer) Delete() error {
	panic("implement me")
}

func (self *secretSynchronizer) Refresh() {
	panic("implement me")
}

func (self *secretSynchronizer) RegisterActionHandler(syncApi.ActionHandlerFunction, ...watch.EventType) {
	panic("implement me")
}

func (self *secretSynchronizer) SetPoller(poller syncApi.Poller) {
	panic("implement me")
}

// Name implements Synchronizer interface. See Synchronizer for more information.
func (self *secretSynchronizer) Name() string {
	return fmt.Sprintf("%s-%s", self.name, self.namespace)
}
