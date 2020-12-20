package main

import (
	"context"
	"flag"
	"golang.org/x/sync/errgroup"
	"k8s-lx1036/k8s/network/ingress/ingress-controller/server"
	"k8s-lx1036/k8s/network/ingress/ingress-controller/watcher"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"os"
)

var (
	host          string
	port, tlsPort int
)

func main() {
	flag.StringVar(&host, "host", "0.0.0.0", "the host to bind")
	flag.IntVar(&port, "port", 80, "the insecure http port")
	flag.IntVar(&tlsPort, "tls-port", 443, "the secure https port")
	flag.Parse()

	log.SetOutput(os.Stdout)

	runtime.ErrorHandlers = []func(error){
		func(err error) {
			log.Println(err)
		},
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal("get k8s configuration failed")
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal("create k8s client failed")
	}

	srv := server.New(server.WithHost(host), server.WithPort(port), server.WithTLSPort(tlsPort))
	watch := watcher.New(client, func(payload *watcher.Payload) {
		srv.Update(payload)
	})

	var eg errgroup.Group
	eg.Go(func() error {
		return srv.Run(context.TODO())
	})
	eg.Go(func() error {
		return watch.Run(context.TODO())
	})
	if err := eg.Wait(); err != nil {
		log.Fatal(err)
	}
}
