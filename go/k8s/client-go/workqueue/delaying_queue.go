package workqueue

import "time"

type DelayingQueue interface {
	Queue
	
	AddAfter(item interface{}, duration time.Duration)
}
