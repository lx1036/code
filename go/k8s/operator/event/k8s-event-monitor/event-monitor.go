package main

import (
	"flag"
	"fmt"
	"k8s-lx1036/k8s/operator/event/k8s-event-monitor/common/flags"
	"k8s-lx1036/k8s/operator/event/k8s-event-monitor/sources"
	"k8s.io/klog"
	"net"
	"net/http"
	"os"
	"strconv"
)

var (
	argSources flags.Uris
	// alertmanager includes inhibition rules, notification routing and notification receivers
	argReceivers   string
	argHealthzIP   = flag.String("healthz-ip", "0.0.0.0", "ip eventer health check service uses")
	argHealthzPort = flag.Uint("healthz-port", 8084, "port eventer health check listens on")

	debug bool
)

func init() {
	// --sources="kubernetes:http://<k8s-api-server-address>:<port>?inClusterConfig=false"
	// --sources="k8s:http://localhost:8080/abc?key1=value1" --sources="k9s:http://localhost:8090/abc?key1=value1"
	flag.Var(&argSources, "sources", "source(s) to read events from")
	flag.StringVar(&argReceivers, "receivers", "", "external notification receivers that receive events")

	flag.BoolVar(&debug, "debug", false, "debug application")
}

// go run ./event-monitor.go --sources="kubernetes:https://192.168.64.32:8443?inClusterConfig=false&insecure=true"
func main() {
	flag.Parse()

	srcs, err := sources.NewSourceFactory().Build(argSources)
	if err != nil {
		klog.Errorf("Failed to create source, because of %s", err.Error())
	}

	source := srcs[0]
	events := source.GetEvents()

	fmt.Println(events.Timestamp, len(events.Events))

	for _, event := range events.Events {
		fmt.Println(fmt.Sprintf("message: %s, reason: %s, type: %s", event.Message, event.Reason, event.Type))
	}

	select {}
	if debug {
		fmt.Println("exit.")
		os.Exit(0)
	}

	receiver := receivers.NewReceiverFactory().BuildAll(argReceivers)
	receiverManager := receivers.NewReceiverManager(receiver)

	mgr := manager.NewManager(source, receiverManager)
	mgr.Start()

	go startHTTPServer()
}

func startHTTPServer() {
	http.ListenAndServe(net.JoinHostPort(*argHealthzIP, strconv.Itoa(int(*argHealthzPort))), nil)
}
