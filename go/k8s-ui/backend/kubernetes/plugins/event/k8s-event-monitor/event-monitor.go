package main

import (
	"flag"
	"k8s-lx1036/k8s-ui/backend/kubernetes/plugins/event/k8s-event-monitor/common/flags"
	"k8s-lx1036/k8s-ui/backend/kubernetes/plugins/event/k8s-event-monitor/receivers"
	"k8s-lx1036/k8s-ui/backend/kubernetes/plugins/event/k8s-event-monitor/sources"
	"k8s.io/klog"
	"net"
	"net/http"
	"strconv"
)

var (
	argSources flags.Uris
	// alertmanager includes inhibition rules, notification routing and notification receivers
	argReceivers   string
	argHealthzIP   = flag.String("healthz-ip", "0.0.0.0", "ip eventer health check service uses")
	argHealthzPort = flag.Uint("healthz-port", 8084, "port eventer health check listens on")
)

func init() {
	// --sources="kubernetes:http://<k8s-api-server-address>:<port>?inClusterConfig=false"
	// --sources="k8s:http://localhost:8080/abc?key1=value1" --sources="k9s:http://localhost:8090/abc?key1=value1"
	flag.Var(&argSources, "sources", "source(s) to read events from")
	flag.StringVar(&argReceivers, "receivers", "", "external notification receivers that receive events")
}

func main() {
	flag.Parse()

	source, err := sources.NewSourceFactory().Build(argSources)
	if err != nil {
		klog.Errorf("Failed to create source, because of %s", err.Error())
	}





	receiver := receivers.NewReceiverFactory().BuildAll(argReceivers)
	receiverManager := receivers.NewReceiverManager(receiver)

	mgr := NewManager(source, receiverManager)
	mgr.Start()

	go startHTTPServer()
}

func startHTTPServer() {
	http.ListenAndServe(net.JoinHostPort(*argHealthzIP, strconv.Itoa(int(*argHealthzPort))), nil)
}
