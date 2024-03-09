package main

import (
	"flag"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"net"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -tags "linux" -type iptnl_info -type vip bpf xdp.c -- -I.

// go generate .
func main() {
	iface := flag.String("iface", "eth0", "interface name")
	flag.Parse()

	// Allow the current process to lock memory for eBPF resources.
	if err := rlimit.RemoveMemlock(); err != nil {
		logrus.Fatal(err)
	}

	device, err := net.InterfaceByName(*iface)
	if err != nil {
		logrus.Fatal(err)
	}

	// Load pre-compiled programs and maps into the kernel.
	objs := bpfObjects{}
	if err := loadBpfObjects(&objs, nil); err != nil {
		logrus.Fatalf("loading objects: %v", err)
	}
	defer objs.Close()

	l, err := link.AttachXDP(link.XDPOptions{
		Program:   objs.XdpTxIptunnel,
		Interface: device.Index,
		Flags:     link.XDPDriverMode,
	})
	if err != nil {
		logrus.Fatal(err)
	}
	defer l.Close()

	vip := bpfVip{
		Daddr: struct {
			V6 [4]uint32
		}{},
		Dport:    0,
		Family:   unix.AF_INET,
		Protocol: 0,
	}
	objs.bpfMaps.Vip2tnl.Put()

}
