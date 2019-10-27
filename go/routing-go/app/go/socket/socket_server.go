package main

// use Go to create TCP server easily

import (
    "bufio"
    "fmt"
    "log"
    "net"
    "os"
    "strings"
)

const (
    Address = "localhost:8081"
)

func server() {
    fmt.Println("Launching server")
    listener, _ := net.Listen("tcp", Address)

    for  {
        connection, _ := listener.Accept()
        message, _ := bufio.NewReader(connection).ReadString('\n')
        fmt.Printf("Received from client message: %s", message)
        _, _ = connection.Write([]byte(strings.ToUpper(message)))
    }
}

func client()  {
    fmt.Println("Launching client")
    connection, _ := net.Dial("tcp", Address)

    for {
        text, _ := bufio.NewReader(os.Stdin).ReadString('\n')
        fmt.Print(text)
        _, err := fmt.Fprintf(connection, text)
        if err != nil {
            log.Fatal(err)
        }
        message, _ := bufio.NewReader(connection).ReadString('\n')
        fmt.Printf("Received from server message: %s", message)
    }
}

func main()  {
    go server()
    go client()

    select {}
}
