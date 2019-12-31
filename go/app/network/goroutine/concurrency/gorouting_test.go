package concurrency

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestGroutine(test *testing.T) {
	runtime.GOMAXPROCS(1)
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		fmt.Println(1)
		fmt.Println(4)
		wg.Done()
	}()

	go func() {
		fmt.Println(2)
		time.Sleep(time.Second)
		fmt.Println(3)
		wg.Done()
	}()

	wg.Wait()
}

var balance int

func Deposit(amount int) {
	balance = balance + amount
}

func Balance() int {
	return balance
}
func TestBalance(test *testing.T) {
	go func() {
		Deposit(200)
		time.Sleep(time.Second)

		fmt.Printf("balance=%d \n", balance)
	}()

	go Deposit(100)

	var x []int
	go func() { x = make([]int, 10) }()
	go func() { x = make([]int, 1000000) }()

	fmt.Printf("%d cores \n", runtime.NumCPU())
	time.Sleep(time.Second * 2)
	x[999999] = 1
}
