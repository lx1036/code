// go run client.go
package main

import (
	"fmt"
	"log"
	"net"
	"sync"
)

const (
	addr      = "localhost:6868"
	connCount = 100
)

func connect(data string) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Fatalf("ResolveTCPAddr failed; %v", err)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatalf("Dial failed; %v", err)
	}

	_, err = conn.Write([]byte(data))
	if err != nil {
		log.Fatalf("Write to server failed; %v", err)
	}
	log.Printf("Done sending %s", data)
}

func main() {
	var wg sync.WaitGroup
	for i := 0; i < connCount; i++ {
		wg.Add(1)
		data := fmt.Sprintf("data from %d", i)
		go func() {
			defer wg.Done()
			connect(data)
		}()
	}
	wg.Wait()
}
