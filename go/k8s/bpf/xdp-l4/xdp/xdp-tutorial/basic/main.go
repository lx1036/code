package main

import (
	"flag"
	"k8s-lx1036/k8s/bpf/xdp-l4/xdp/xdp-tutorial/basic/bpf"
	"log"
	"net"
	"time"
)

// https://github.com/cilium/ebpf/blob/main/examples/xdp/main.go

var (
	ifaceName = flag.String("iface", "eth0", "specify a network interface")
)

/*
使用 veth-pair 来测试 xdp program
*/

// ./setup-env.sh setup --legacy-ip --name veth-basic02
// go run . --iface=veth-basic02

// ip netns exec veth-basic02 bash
// ping 10.11.1.1
func main() {
	flag.Parse()

	iface, err := net.InterfaceByName(*ifaceName)
	if err != nil {
		log.Fatalf("lookup network iface %q: %s", *ifaceName, err)
	}

	xdpObjects, err := bpf.LoadAndAttachXdp(iface.Index)
	if err != nil {
		log.Fatalf("LoadAndAttachXdp err: %v", err)
	}
	defer xdpObjects.Close()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		log.Printf("ok\n")
	}
}
