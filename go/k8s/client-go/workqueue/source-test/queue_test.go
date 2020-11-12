package source_test

import (
	"sync"
	"time"
	"testing"
	
	"k8s.io/client-go/util/workqueue"
)

func TestBasic(t *testing.T) {
	// If something is seriously wrong this test will never complete.
	q := workqueue.New()
	
	// Start producers
	const producers = 5
	producerWG := sync.WaitGroup{}
	producerWG.Add(producers)
	for i := 0; i < producers; i++ {
		go func(i int) {
			defer producerWG.Done()
			for j := 0; j < 2; j++ {
				q.Add(i)
				time.Sleep(time.Millisecond)
			}
		}(i)
	}
	
	// Start consumers
	const consumers = 3
	consumerWG := sync.WaitGroup{}
	consumerWG.Add(consumers)
	for i := 0; i < consumers; i++ {
		go func(i int) {
			defer consumerWG.Done()
			for {
				item, quit := q.Get()
				if item == "added after shutdown!" {
					t.Errorf("Got an item added after shutdown.")
				}
				if quit {
					return
				}
				//t.Logf("Worker %v: begin processing %v", i, item)
				time.Sleep(3 * time.Millisecond)
				t.Logf("Worker %v: done processing %v", i, item)
				q.Done(item)
			}
		}(i)
	}
	
	producerWG.Wait()
	q.ShutDown()
	q.Add("added after shutdown!")
	consumerWG.Wait()
}
