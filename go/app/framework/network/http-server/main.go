package main

import (
	"flag"
	"fmt"
	"k8s-lx1036/app/framework/network/net"
	"log"
	"os"
	"strings"
	"time"
)

var (
	res string
)

type httpServer struct {
	*net.EventServer
	port    int
	noparse bool
}

func (h httpServer) OnInitComplete(server net.Server) (action net.Action) {
	panic("implement me")
}

func (h httpServer) OnOpened(c net.Connection) (out []byte, action net.Action) {
	panic("implement me")
}

func (h httpServer) OnClosed(c net.Connection, err error) (action net.Action) {
	panic("implement me")
}

func (h httpServer) PreWrite() {
	panic("implement me")
}

func (h httpServer) React(c net.Connection) (out []byte, action net.Action) {
	panic("implement me")
}

func (h httpServer) Tick() (delay time.Duration, action net.Action) {
	panic("implement me")
}

func main()  {
	var port int
	var multicore bool
	var aaaa bool
	var noparse bool

	flag.IntVar(&port, "port", 8080, "server port")
	flag.BoolVar(&aaaa, "aaaa", false, "aaaaa....")
	flag.BoolVar(&noparse, "noparse", true, "do not parse requests")
	flag.BoolVar(&multicore, "multicore", true, "multicore")
	flag.Parse()

	if os.Getenv("NOPARSE") == "1" {
		noparse = true
	}

	if aaaa {
		res = strings.Repeat("a", 1024)
	} else {
		res = "Hello World!\r\n"
	}

	http := &httpServer{port: port, noparse: noparse}
	// We at least want the single http address.
	addr := fmt.Sprintf("tcp://:%d", port)
	// Start serving!
	log.Fatal(net.Serve(http, addr, net.WithMulticore(multicore)))
}
