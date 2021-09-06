package main

import (
	"bufio"
	"fmt"
	"net"
)

// 聊天功能
// https://victoriest.gitbooks.io/golang-tcp-server/content/chapter2.html
func main() {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:2020")
	listener, _ := net.ListenTCP("tcp", tcpAddr)
	defer listener.Close()

	var ConnectionMap = map[string]*net.TCPConn{}
	stopC := make(chan struct{})
	for {
		connection, err := listener.AcceptTCP()
		if err != nil {
			continue
		}
		
		select {
		case <-stopC:
			return
		default:
		}

		ConnectionMap[connection.RemoteAddr().String()] = connection

		go func(connection *net.TCPConn) {
			defer connection.Close()
			//remoteAddr := connection.RemoteAddr().String()
			reader := bufio.NewReader(connection)
			for {
				message, err := reader.ReadString('\n')
				if err != nil {
					return
				}

				fmt.Println(message)

				for _, connection := range ConnectionMap {
					_, _ = connection.Write([]byte(message))
				}
			}
		}(connection)
	}
}
