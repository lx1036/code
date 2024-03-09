package main

import (
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/sirupsen/logrus"
	"net"
)

// $BPF_CLANG and $BPF_CFLAGS are set by the Makefile.
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc "$CLANG" -strip "$STRIP" -makebase "$MAKEDIR" xdp ./bpf/xdp.c -- -mcpu=v2 -nostdinc -Wall -Werror -Wno-compare-distinct-pointer-types -I./bpf/include

type lbArguments struct {
	DstMac [6]uint8
	_      [2]byte // 必须 20 bytes???
	Daddr  uint32
	Saddr  uint32
	Vip    uint32
}

// go run . --dst-mac="ee:ff:ff:ff:ff:ff" --endpoint="172.16.0.154" --node-ip="172.16.0.153" --vip="192.168.10.1"
func main() {
	var (
		dstMac   string
		endpoint string
		iface    string
		nodeIP   string
		vip      string
	)

	flag.StringVar(&dstMac, "dst-mac", "", "the mac address of the next hop")
	flag.StringVar(&endpoint, "endpoint", "", "the ip of the endpoint")
	flag.StringVar(&nodeIP, "node-ip", "", "the ip of the lb")
	flag.StringVar(&iface, "iface", "eth0", "the interface to attach this program to")
	flag.StringVar(&vip, "vip", "", "the virtual ip")
	flag.Parse()

	ifaceObj, err := net.InterfaceByName(iface)
	if err != nil {
		logrus.Fatalf("lookup network iface %q: %s", iface, err)
	}
	logrus.Infof("interface index: %d", ifaceObj.Index)

	objs := xdpObjects{}
	if err = loadXdpObjects(&objs, &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: "/sys/fs/bpf",
		},
	}); err != nil {
		logrus.Fatalf("fail to load xdp objs into kernel err: %v", err)
	}
	defer objs.Close()

	l, err := link.AttachXDP(link.XDPOptions{
		Program:   objs.xdpPrograms.BpfXdpEntry, // 挂载 xdp_drop_func xdp
		Interface: ifaceObj.Index,
		Flags:     link.XDPGenericMode, // ecs eth0 貌似不支持 XDPDriverMode
	})
	if err != nil {
		logrus.Fatalf("could not attach XDP program: %s", err)
	}
	defer l.Close()

	// update bpf map
	mac, err := macToBytes(dstMac)
	if err != nil {
		logrus.Fatal(err)
	}
	var macArray [6]uint8
	copy(macArray[:], mac)
	intDest, err := IPv4ToInt(net.ParseIP(endpoint))
	if err != nil {
		logrus.Fatal(err)
	}
	intSrc, err := IPv4ToInt(net.ParseIP(nodeIP))
	if err != nil {
		logrus.Fatal(err)
	}
	intVip, err := IPv4ToInt(net.ParseIP(vip))
	if err != nil {
		logrus.Fatal(err)
	}
	args := lbArguments{
		Daddr:  intDest,
		Saddr:  intSrc,
		DstMac: macArray,
		Vip:    intVip,
	}
	err = objs.XdpParamsArray.Put(uint32(0), args)
	if err != nil {
		logrus.Fatalf("update map err: %v", err)
	}

	select {}
}

var ErrNotIPv4Address = errors.New("not an IPv4 addres")

func IPv4ToInt(ipaddr net.IP) (uint32, error) {
	if ipaddr.To4() == nil {
		return 0, ErrNotIPv4Address
	}
	return binary.BigEndian.Uint32(ipaddr.To4()), nil
}

func macToBytes(mac string) ([]uint8, error) {
	//mac = strings.Replace(mac, ":", "", -1)
	// Parse the MAC address string into a hardware address
	macAddr, err := net.ParseMAC(mac)
	if err != nil {
		return nil, fmt.Errorf("error parsing MAC address: %w", err)
	}

	// Convert the hardware address to an array of bytes
	macBytes := macAddr[:]
	return macBytes, nil
}
