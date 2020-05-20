package main

import (
	"flag"
	"fmt"
	eventloop "k8s-lx1036/k8s/network/event-loop-network/event-loop"
	"log"
	"strings"
)

// telnet localhost 5000
func main() {
	var port int
	var loops int
	var udp bool
	var trace bool
	var reuseport bool
	var stdlib bool
	var host string

	flag.StringVar(&host, "host", "localhost", "server host")
	flag.IntVar(&port, "port", 5000, "server port")
	flag.BoolVar(&udp, "udp", false, "listen on udp")
	flag.BoolVar(&reuseport, "reuseport", false, "reuseport (SO_REUSEPORT)")
	flag.BoolVar(&trace, "trace", false, "print packets to console")
	flag.IntVar(&loops, "loops", 0, "num loops")
	flag.BoolVar(&stdlib, "stdlib", false, "use stdlib")
	flag.Parse()

	var events eventloop.Events
	events.NumLoops = loops
	events.Serving = func(server eventloop.Server) (action eventloop.Action) {
		log.Printf("echo server started on port %d (loops: %d)", port, server.NumLoops)
		if reuseport {
			log.Println("reuseport")
		}
		if stdlib {
			log.Println("stdlib")
		}
		return
	}
	events.Data = func(c eventloop.Conn, in []byte) (out []byte, action eventloop.Action) {
		if trace {
			log.Printf("%s", strings.TrimSpace(string(in)))
		}
		out = in
		return
	}
	scheme := "tcp"
	if udp {
		scheme = "udp"
	}
	if stdlib {
		scheme += "-net"
	}

	log.Fatal(eventloop.Serve(events, fmt.Sprintf("%s://%s:%d?reuseport=%t", scheme, host, port, reuseport)))
}
