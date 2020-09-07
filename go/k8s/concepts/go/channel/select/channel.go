package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	//demo1()
	demo2()
}

func demo1() {
	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-time.Tick(time.Second * 2):
		fmt.Println("hello")
	case <-stop:
		fmt.Println("world")
	}

	<-stop
}

func demo2() {
	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	opened := false
	for {
		select {
		case <-time.Tick(time.Second * 2):
			fmt.Println("hello")
		case <-stop:
			fmt.Println("world")
			close(stop)
			opened = true
		}

		if opened {
			break
		}
	}
}
