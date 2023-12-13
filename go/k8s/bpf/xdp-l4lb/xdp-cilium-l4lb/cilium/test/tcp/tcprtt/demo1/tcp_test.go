package main

import (
	"golang.org/x/sys/unix"
	"log"
	"net"
	"testing"
)

// go test -v -run ^TestTCP$ .
func TestTCP(test *testing.T) {
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
	if err != nil {
		log.Fatal(err)
	}

	setSocketTimeout(fd, 0)

	ip := net.ParseIP("127.0.0.1")
	sa := &unix.SockaddrInet4{
		Port: 8000,
		Addr: [4]byte{},
	}
	copy(sa.Addr[:], ip)
	err = unix.Bind(fd, sa)
	if err != nil {
		log.Fatal(err)
	}

	// 非阻塞的
	err = unix.Listen(fd, 1)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("success")
}
