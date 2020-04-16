package main

import (
	"flag"
	"k8s-lx1036/k8s-ui/backend/kubernetes/plugins/event/k8s-event-monitor/receivers"
	"k8s-lx1036/k8s-ui/backend/kubernetes/plugins/event/k8s-event-monitor/sources"
	"net"
	"net/http"
	"strconv"
)

var (
	argSources string
	// alertmanager includes inhibition rules, notification routing and notification receivers
	argReceivers   string
	argHealthzIP   = flag.String("healthz-ip", "0.0.0.0", "ip eventer health check service uses")
	argHealthzPort = flag.Uint("healthz-port", 8084, "port eventer health check listens on")
)

func init() {
	flag.StringVar(&argSources, "sources", "", "source(s) to read events from")
	flag.StringVar(&argReceivers, "receivers", "", "external notification receivers that receive events")
}

func main() {
	flag.Parse()

	source := sources.NewSourceFactory().BuildAll(argSources)
	receiver := receivers.NewReceiverFactory().BuildAll(argReceivers)
	receiverManager := receivers.NewReceiverManager(receiver)

	mgr := NewManager(source, receiverManager)
	mgr.Start()

	go startHTTPServer()
}

func startHTTPServer() {
	http.ListenAndServe(net.JoinHostPort(*argHealthzIP, strconv.Itoa(int(*argHealthzPort))), nil)
}
