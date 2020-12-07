package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

// select 语句的语法：
// 1. 每个 case 都必须是一个通信
// 2. 所有 channel 表达式都会被求值, 所有被发送的表达式都会被求值
// 3. 如果任意某个通信可以进行，它就执行，其他被忽略; 如果有多个 case 都可以运行，Select 会随机公平地选出一个执行。其他不会执行
// 4. 如果有 default 子句，则执行该语句; 如果没有 default 子句，select 将阻塞，直到某个通信可以运行；Go 不会重新对 channel 或值进行求值
func TestSelect(test *testing.T) {
	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	
	select {
	case <-time.Tick(time.Second * 2):
		fmt.Println("hello")
	case <-stop:
		fmt.Println("world")
	}
	
	//<-stop
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
