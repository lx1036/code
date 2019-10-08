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

