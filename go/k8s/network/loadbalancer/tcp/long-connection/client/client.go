package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	stopChan := make(chan bool)
	tcpAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:2020")
	connection, _ := net.DialTCP("tcp", nil, tcpAddr)
	defer connection.Close()

	go func() {
		reader := bufio.NewReader(connection)
		for {
			message, err := reader.ReadString('\n')

			fmt.Println(message)

			if err != nil {
				stopChan <- true
				return
			}

			//time.Sleep(time.Second * 5)
			//_, _ = connection.Write([]byte(message))
		}
	}()

	//<- stopChan

	var msg string
	for {
		select {
		case <-stopChan:
			return
		default:
			fmt.Scanln(&msg)
			if msg == "quit" {
				return
			}
			_, _ = connection.Write([]byte(msg + "\n"))
		}
	}

	/*for  {
		fmt.Scanln(&msg)
		if msg == "quit" {
			break
		}
		_ , _ = connection.Write([]byte(msg + "\n"))
	}

	<- stopChan*/
}
