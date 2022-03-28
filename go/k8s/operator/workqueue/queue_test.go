package workqueue

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// 测试多个producers和多个consumers
func TestMultipleProducersAndMultipleConsumers(test *testing.T) {
	queue := New()
	producerWg := sync.WaitGroup{}
	producerNumber := 50
	producerWg.Add(producerNumber)
	for i := 0; i < producerNumber; i++ {
		go func() {
			defer producerWg.Done()
			for j := 0; j < 10; j++ {
				queue.Add(j)
				time.Sleep(time.Millisecond * 2)
			}
		}()
	}

	consumerWg := sync.WaitGroup{}
	consumerNumber := 10
	consumerWg.Add(consumerNumber)
	for i := 0; i < consumerNumber; i++ {
		go func() {
			defer consumerWg.Done()
			for {
				item, quit := queue.Get()
				if item == "added after shutdown" {
					test.Errorf("get an item after shutdown")
				}
				if quit {
					return
				}

				fmt.Println(item)
				time.Sleep(time.Millisecond * 4)

				queue.Done(item)
			}
		}()
	}

	producerWg.Wait()
	queue.ShutDown()
	queue.Add("added after shutdown")
	consumerWg.Wait()
}
