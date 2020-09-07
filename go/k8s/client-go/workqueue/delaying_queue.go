package workqueue

import "time"

type DelayingQueue interface {
	Interface
	
	AddAfter(item interface{}, duration time.Duration)
}
