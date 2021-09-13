package main

import (
	"k8s-lx1036/k8s/storage/log/filebeat/pkg/libbeat/publisher/pipeline"
	"time"
)

func main() {

	// make pipeline
	publisher := pipeline.NewPipeline()
	client := Connect()

	ticker := time.NewTicker(5 * time.Second)
	stopCh := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				client.Publish(Event{
					Timestamp: time.Now(),
					Fields: common.MapStr{
						"type":    "mock",
						"message": "Mockbeat is alive!",
					},
				})
			case <-stopCh:
				ticker.Stop()
			}
		}
	}()

	<-stopCh
}
