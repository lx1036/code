package main

import "fmt"

func main()  {
    defer fmt.Println("hello")
    defer fmt.Println("world")

    defer func() {
        fmt.Println("this is a test")
    }()

    fmt.Println("!!!")

    func() {
        for i := 0; i < 3; i++ {
            defer fmt.Println("a:", i)
        }
    }()

    func() {
        for i := 0; i < 3;i++  {
            defer func() {
                fmt.Println("b:", i)
            }()
        }
    }()
}
