/*
dumpframes demostrates how to receive frames from a network link using
github.com/asavie/xdp package, it sets up an XDP socket attached to a
particular network link and dumps all frames it receives to standard output.
*/
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/vishvananda/netlink"
	xdp_socket "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp/xdp-socket"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp/xdp-socket/dumpframes/bpf"
	"log"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// 该程序作用：在网卡 eth0 queue=0 注册个 xdp 程序，然后把 tcp/udp 包 redirect 到 xdp_socket 用户态程序上

// TCP: go run . --ipproto=6 --linkname=eth1 然后 ip addr 查看 eth1 网卡注册了 xdp generic 程序
// UDP: go run . --ipproto=17
// 测试没法在 eth0 attach ip-proto xdp 程序，很奇怪
// 但是新建一个 eth1 可以，应该是 eth0 网卡配置有问题
// ip link add eth1 type dummy
// ip link set eth3 up
// ip addr add 100.200.0.126 dev eth1
func main() {
	var linkName string
	var queueID int
	var protocol int64

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	flag.StringVar(&linkName, "linkname", "eth0", "The network link on which rebroadcast should run on.")
	flag.IntVar(&queueID, "queueid", 0, "The ID of the Rx queue to which to attach to on the network link.")
	flag.Int64Var(&protocol, "ipproto", 0, "If greater than 0 and less than or equal to 255, limit xdp bpf_redirect_map to packets with the specified IP protocol number.")
	flag.Parse()

	link, err := netlink.LinkByName(linkName)
	if err != nil {
		fmt.Printf("error: failed to fetch network link: %v\n", err)
		return
	}
	Ifindex := link.Attrs().Index

	var program *xdp_socket.Program

	// Create a new XDP eBPF program and attach it to our chosen network link.
	if protocol == 0 {
		program, err = xdp_socket.NewProgram(queueID + 1)
	} else {
		program, err = bpf.NewIPProtoProgram(uint32(protocol), nil)
	}
	if err != nil {
		fmt.Printf("error: failed to create xdp program: %v\n", err)
		return
	}
	defer program.Close()
	if err := program.Attach(Ifindex); err != nil {
		fmt.Printf("error: failed to attach xdp program to interface: %v\n", err)
		return
	}
	defer program.Detach(Ifindex)

	// Create and initialize an XDP socket attached to our chosen network link
	xsk, err := xdp_socket.NewSocket(Ifindex, queueID, nil)
	if err != nil {
		fmt.Printf("error: failed to create an XDP socket: %v\n", err)
		return
	}

	// Register our XDP socket file descriptor with the eBPF program so it can be redirected packets
	if err := program.Register(queueID, xsk.FD()); err != nil {
		fmt.Printf("error: failed to register socket in BPF map: %v\n", err)
		return
	}
	defer program.Unregister(queueID)

	for {
		// If there are any free slots on the Fill queue...
		if n := xsk.NumFreeFillSlots(); n > 0 {
			// ...then fetch up to that number of not-in-use
			// descriptors and push them onto the Fill ring queue
			// for the kernel to fill them with the received
			// frames.
			xsk.Fill(xsk.GetDescs(n, true))
		}

		// Wait for receive - meaning the kernel has
		// produced one or more descriptors filled with a received
		// frame onto the Rx ring queue.
		log.Printf("waiting for frame(s) to be received...")
		numRx, _, err := xsk.Poll(-1)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return
		}

		if numRx > 0 {
			// Consume the descriptors filled with received frames
			// from the Rx ring queue.
			rxDescs := xsk.Receive(numRx)

			// Print the received frames and also modify them
			// in-place replacing the destination MAC address with
			// broadcast address.
			for i := 0; i < len(rxDescs); i++ {
				pktData := xsk.GetFrame(rxDescs[i])
				pkt := gopacket.NewPacket(pktData, layers.LayerTypeEthernet, gopacket.Default)
				log.Printf("received frame:\n%s%+v", hex.Dump(pktData[:]), pkt)
			}
		}
	}
}
