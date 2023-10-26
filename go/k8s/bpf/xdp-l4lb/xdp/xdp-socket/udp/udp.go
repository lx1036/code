package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"math"
	"net"
	"time"

	xdp_socket "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp/xdp-socket"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/vishvananda/netlink"
)

var (
	NIC         string
	QueueID     int
	SrcMAC      string
	DstMAC      string
	SrcIP       string
	DstIP       string
	SrcPort     uint
	DstPort     uint
	PayloadSize uint
)

/*
sendudp pre-generates a frame with a UDP packet with a payload of the given
size and starts sending it in and endless loop to given destination as fast as
possible.
*/

// go run . --interface=eth0 --srcmac=00163e3a0d9c --srcip=172.16.0.154 --dstmac=00163e37ad84 --dstip=172.16.0.153
/*
sending UDP packets from 172.16.0.154 (00:16:3e:3a:0d:9c) to 172.16.0.153 (00:16:3e:37:ad:84)...
304130 packets/s (3508 Mb/s)
299995 packets/s (3460 Mb/s)
299998 packets/s (3460 Mb/s)
300068 packets/s (3461 Mb/s)
300051 packets/s (3461 Mb/s)
300251 packets/s (3463 Mb/s)
300068 packets/s (3461 Mb/s)
300262 packets/s (3463 Mb/s)
300015 packets/s (3460 Mb/s)
*/

// 在 172.16.0.153 上抓包:
// tcpdump -i eth0 -nneevv -A udp and port 1234

func main() {
	flag.StringVar(&NIC, "interface", "eth0", "Network interface to attach to.")
	flag.IntVar(&QueueID, "queue", 0, "The queue on the network interface to attach to.")
	flag.StringVar(&SrcMAC, "srcmac", "b2968175b211", "Source MAC address to use in sent frames.")
	flag.StringVar(&DstMAC, "dstmac", "ffffffffffff", "Destination MAC address to use in sent frames.")
	flag.StringVar(&SrcIP, "srcip", "192.168.111.10", "Source IP address to use in sent frames.")
	flag.StringVar(&DstIP, "dstip", "192.168.111.1", "Destination IP address to use in sent frames.")
	flag.UintVar(&SrcPort, "srcport", 1234, "Source UDP port.")
	flag.UintVar(&DstPort, "dstport", 1234, "Destination UDP port.")
	flag.UintVar(&PayloadSize, "payloadsize", 1400, "Size of the UDP payload.")
	flag.Parse()

	// Initialize the XDP socket.
	link, err := netlink.LinkByName(NIC)
	if err != nil {
		panic(err)
	}

	xsk, err := xdp_socket.NewSocket(link.Attrs().Index, QueueID, nil)
	if err != nil {
		panic(err)
	}

	// Pre-generate a frame containing a DNS query.
	srcMAC, _ := hex.DecodeString(SrcMAC)
	dstMAC, _ := hex.DecodeString(DstMAC)
	eth := &layers.Ethernet{
		SrcMAC:       net.HardwareAddr(srcMAC),
		DstMAC:       net.HardwareAddr(dstMAC),
		EthernetType: layers.EthernetTypeIPv4,
	}
	ip := &layers.IPv4{
		Version:  4,
		IHL:      5,
		TTL:      64,
		Id:       0,
		Protocol: layers.IPProtocolUDP,
		SrcIP:    net.ParseIP(SrcIP).To4(),
		DstIP:    net.ParseIP(DstIP).To4(),
	}
	udp := &layers.UDP{
		SrcPort: layers.UDPPort(SrcPort),
		DstPort: layers.UDPPort(DstPort),
	}
	_ = udp.SetNetworkLayerForChecksum(ip)
	payload := make([]byte, PayloadSize)
	for i := 0; i < len(payload); i++ {
		payload[i] = byte(i) // [0 1 2 3 4 5 6 7 8 9 ...]
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}
	err = gopacket.SerializeLayers(buf, opts, eth, ip, udp, gopacket.Payload(payload))
	if err != nil {
		panic(err)
	}
	frameLen := len(buf.Bytes())

	// Fill all the frames in UMEM with the pre-generated UDP packet.
	descs := xsk.GetDescs(math.MaxInt32, false)
	for i := range descs {
		frameLen = copy(xsk.GetFrame(descs[i]), buf.Bytes())
	}

	// Start sending the pre-generated frame as quickly as possible in an
	// endless loop printing statistics of the number of sent frames and
	// the number of sent bytes every second.
	fmt.Printf("sending UDP packets from %v (%v) to %v (%v)...\n", ip.SrcIP, eth.SrcMAC, ip.DstIP, eth.DstMAC)

	go func() {
		var err error
		var prev xdp_socket.Stats
		var cur xdp_socket.Stats
		var numPkts uint64
		for i := uint64(0); ; i++ {
			time.Sleep(time.Duration(1) * time.Second)
			cur, err = xsk.Stats()
			if err != nil {
				panic(err)
			}
			numPkts = cur.Completed - prev.Completed
			fmt.Printf("%d packets/s (%d Mb/s)\n", numPkts, (numPkts*uint64(frameLen)*8)/(1000*1000))
			prev = cur
		}
	}()

	for {
		descs2 := xsk.GetDescs(xsk.NumFreeTxSlots(), false)
		for i := range descs2 {
			descs2[i].Len = uint32(frameLen)
		}
		xsk.Transmit(descs2)

		_, _, err = xsk.Poll(-1)
		if err != nil {
			panic(err)
		}
	}
}
