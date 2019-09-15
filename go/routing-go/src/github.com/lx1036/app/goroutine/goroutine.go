package main

import (
    "log"
    "math/rand"
    "runtime"
    "sync"
    "time"
)

var wg sync.WaitGroup

func main()  {
    log.Println(runtime.NumCPU())
    //os.Exit(1)
    rand.Seed(time.Now().UnixNano())
    log.SetFlags(0)
    wg.Add(2)
    go SayHello("hello", 3)
    go SayHello("world", 3)
    wg.Wait()
    time.Sleep(time.Second * 3)
}

func SayHello(word string, times int)  {
    for i:=0; i < times; i++ {
        log.Println(word)
        time.Sleep(time.Second * time.Duration(rand.Intn(5)))
    }

    wg.Wait()
    //wg.Done()
}
