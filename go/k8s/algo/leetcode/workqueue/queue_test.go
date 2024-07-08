package workqueue

import (
	"k8s.io/client-go/util/workqueue"
	"sync"
	"testing"
	"time"
)

func TestBasic(test *testing.T) {
	/**
	  多个 producer 向 workqueue 里添加 task, 多个 consumer 消费 task
	*/

	queue := workqueue.New()

	// Start producers
	const producers = 50
	producerWG := &sync.WaitGroup{}
	producerWG.Add(producers)
	for i := 0; i < producers; i++ {
		go func(i int) {
			defer producerWG.Done()
			for j := 0; j < 50; j++ {
				queue.Add(i) // add task 50 times
				time.Sleep(time.Millisecond)
			}
		}(i)
	}

	// Start consumers
	const consumers = 5
	consumerWG := sync.WaitGroup{}
	consumerWG.Add(consumers)
	for i := 0; i < consumers; i++ {
		go func(i int) {
			defer consumerWG.Done()
			for {
				item, quit := queue.Get()

				if quit {
					return
				}

				test.Logf("Worker %v: begin processing %v", i, item)
				time.Sleep(3 * time.Millisecond) // do work
				test.Logf("Worker %v: done processing %v", i, item)
				queue.Done(item)
			}
		}(i)
	}

	producerWG.Wait()
	consumerWG.Wait()
	if queue.Len() != 0 {
		test.Errorf("Expected the queue to be empty, had: %v items", queue.Len())
	}
}
