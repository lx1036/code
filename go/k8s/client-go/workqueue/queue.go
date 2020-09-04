package workqueue

import "sync"

type Interface interface {
	Add(item interface{})
	Len() int
	Get() (item interface{}, shutdown bool)
	Done(item interface{})
	ShutDown()
	ShuttingDown() bool
}

type Queue struct {
	jobs []job
	
	// 标记该队列是否正处于关闭状态
	shuttingDown bool
	
	cond *sync.Cond
	
	// 需要被处理的job
	dirty set
	// 正在被处理的job
	processing set
	
	metrics queueMetrics
}

type job interface {}
type empty struct {}
type set map[job]empty

func (s set) has(item job) bool {
	_, exists := s[item]
	return exists
}
func (s set) insert(item job)  {
	s[item] = empty{}
}
func (s set) delete(item job)  {
	delete(s, item)
}

func (queue *Queue) Add(item interface {})  {
	queue.cond.L.Lock()
	defer queue.cond.L.Unlock()
	
	if queue.shuttingDown {
		return
	}
	
	if queue.dirty.has(item) {
		return
	}
	
	queue.metrics.add(item)
	
	queue.dirty.insert(item)
	if queue.processing.has(item) {
		return
	}
	
	queue.jobs = append(queue.jobs, item)
	queue.cond.Signal()
}

func (queue *Queue) Len() int {
	queue.cond.L.Lock()
	defer queue.cond.L.Unlock()
	
	return len(queue.jobs)
}

func (queue *Queue) Get() (item interface{}, shutdown bool)  {
	queue.cond.L.Lock()
	defer queue.cond.L.Unlock()
	
	for len(queue.jobs) == 0 && !queue.shuttingDown {
		queue.cond.Wait()
	}
	
	if len(queue.jobs) == 0 {
		return nil, true
	}
	
	item, queue.jobs = queue.jobs[0], queue.jobs[1:]
	
	queue.metrics.get(item)
	
	queue.processing.insert(item)
	queue.dirty.delete(item)
	
	return item, false
}

func (queue *Queue) Done(item interface{})  {
	queue.cond.L.Lock()
	defer queue.cond.L.Unlock()
	
	queue.metrics.done(item)
	
	queue.processing.delete(item)
	if queue.dirty.has(item) {
		queue.jobs = append(queue.jobs, item)
		queue.cond.Signal()
	}
}

func (queue *Queue) ShutDown()  {
	queue.cond.L.Lock()
	defer queue.cond.L.Unlock()
	
	queue.shuttingDown = true
	queue.cond.Broadcast()
}

func (queue *Queue) ShuttingDown() bool {
	queue.cond.L.Lock()
	defer queue.cond.L.Unlock()
	
	return queue.shuttingDown
}


func New() *Queue {
	queue := &Queue{
		jobs:         nil,
		shuttingDown: false,
		cond:         nil,
		dirty:        nil,
		processing:   nil,
		metrics:      nil,
	}
	return queue
}

