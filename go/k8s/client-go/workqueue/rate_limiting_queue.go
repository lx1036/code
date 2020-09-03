package workqueue

type RateLimitingQueue interface {
	DelayingQueue
	
	//
	AddRateLimited(item interface{})
}

