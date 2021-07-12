package main

import (
	"net"
	"time"

	"k8s-lx1036/k8s/network/bgp/pkg/bgp"
)

func main() {
	stopCh := make(chan struct{})
	_, err := bgp.New("remote-ip:179", 65002,
		net.ParseIP("local-ip"), 65001,
		time.Second*10, "", "local-node-name")
	if err != nil {
		panic(err)
	}

	<-stopCh
}
