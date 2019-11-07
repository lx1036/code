package concurrency

// https://mp.weixin.qq.com/s/ZJkkllp-jj7PK2-qfHKwhg

import (
	"fmt"
	"testing"
)

func printHello(channel chan int) {
	fmt.Println("hello goroutine")
	<-channel
}

func printNums(channel chan int) {
	for i := 0; i < 10; i++ {
		channel <- i
	}

	close(channel)
}

func printNums2(channel chan int) {
	fmt.Println(<-channel)
}

func TestChannel(test *testing.T) {
	channel := make(chan int)
	channel2 := make(chan int)
	fmt.Printf("channel type is %T value is %v \n", channel, channel) // channel type is chan int value is 0xc0000682a0

	//channel2 <- 2 // block

	go printNums(channel2)

	/*for  {
	   value, ok := <-channel2
	   if ok == false {
	       fmt.Println(value, ok)
	       break
	   }
	   fmt.Println(value)
	}*/

	for value := range channel2 {
		fmt.Println(value)
	}

	//channel <- data
	go printHello(channel)
	//time.Sleep(time.Second)
	channel <- 1 // main goroutine block, scheduler run hello goroutine
	fmt.Println("main goroutine")

	channel3 := make(chan int, 3)
	go printNums2(channel3)
	channel3 <- 1
	channel3 <- 2
	channel3 <- 3
	channel3 <- 4
}
