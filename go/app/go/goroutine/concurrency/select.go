// https://studygolang.gitbook.io/learn-go-with-tests/go-ji-chu/select

package concurrency

import (
    "fmt"
    "net/http"
    "time"
)

// 进程同步
// select 可帮助你同时在多个 channel 上等待
func Racer(a string, b string, timeout time.Duration) (winner string, error error) {
    /*startA := time.Now()
    http.Get(a)
    aDuration := time.Since(startA)

    startB := time.Now()
    http.Get(b)
    bDuration := time.Since(startB)*/

    /*aDuration := measureResponseTime(a)
    bDuration := measureResponseTime(b)
    if aDuration < bDuration {
        return a
    }
    return b*/

    select {
    case <-ping(a):
        return a, nil
    case <-ping(b):
        return b, nil
    case <-time.After(timeout * time.Second):
        return "", fmt.Errorf("timed out waiting for %s and %s", a, b)
    }
}
func ping(url string) chan bool {
    ch := make(chan bool)
    go func() {
        http.Get(url)
        ch <- true
    }()
    return ch
}

func measureResponseTime(url string) time.Duration {
    start := time.Now()
    http.Get(url)
    return time.Since(start)
}
