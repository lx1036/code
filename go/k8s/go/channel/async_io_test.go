package channel

import (
	"log"
	"testing"
	"time"
)


func TestAsyncIO(test *testing.T) {
	type JsonResponse struct {
		code int
		data string
	}
	response := make(chan JsonResponse)

	go func() {
		for {
			select {
			case <-time.After(time.Second * 2):
				log.Printf("timeout")
				break
			case data := <-response:
				log.Println(data)
			}
		}
	}()

	go func() {
		message := JsonResponse{
			code: 0,
			data: "hello world",
		}

		response <- message
	}()

	select {}
}

