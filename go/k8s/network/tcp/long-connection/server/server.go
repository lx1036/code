package main

import "net"

// https://victoriest.gitbooks.io/golang-tcp-server/content/chapter2.html
func main() {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:2020")
	listener, _ := net.ListenTCP("tcp", tcpAddr)
	defer listener.Close()
	
	for  {
		connection, err := listener.AcceptTCP()
		if err != nil {
			continue
		}
		
		go func() {
			defer connection.Close()
			remoteAddr := connection.RemoteAddr().String()
			
			
		}()
	}
}



