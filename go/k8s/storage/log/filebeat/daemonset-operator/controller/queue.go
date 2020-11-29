package controller

import (
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
	"sync"
	"time"
)

type TaskQueue struct {
	// queue is the work queue the worker polls
	queue *workqueue.Type

	// syncTask is called for each item in the queue
	syncTask func([]interface{})

	// workerDone is closed when the worker exits
	workerDone chan struct{}

	lock sync.Mutex
}

func (t *TaskQueue) Enqueue(obj interface{}) {
	t.queue.Add(obj)
}

func (t *TaskQueue) Run(period time.Duration, stopCh <-chan struct{}) {
	wait.Until(t.worker, period, stopCh)
}

func (t *TaskQueue) worker() {
	for {
		var objs []interface{}
		for i := 0; i < t.queue.Len(); i++ {
			obj, quit := t.queue.Get()
			if quit {

			}

			objs = append(objs, obj)
		}

		t.syncTask(objs)

		for _, obj := range objs {
			t.queue.Done(obj)
		}

	}
}

// NewTaskQueue creates a new task queue with the given sync function.
// The sync function is called for every element inserted into the queue.
func NewTaskQueue(syncFn func([]interface{})) *TaskQueue {
	return &TaskQueue{
		queue:      workqueue.New(),
		syncTask:   syncFn,
		workerDone: make(chan struct{}),
	}
}
