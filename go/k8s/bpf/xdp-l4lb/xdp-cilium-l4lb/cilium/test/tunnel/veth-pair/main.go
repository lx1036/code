package main

import (
	"encoding/binary"
	"github.com/cilium/ebpf"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/sirupsen/logrus"
	"net"
)

const (
	MapPinFile = "/sys/fs/bpf/tc/globals/ding_lxc"
)

type EndpointMapKey struct {
	IP uint32
}

type EndpointMapInfo struct {
	IfIndex    uint32
	LxcIfIndex uint32 // 标记另一半的 ifindex
	// MAC        uint64
	// NodeMAC    uint64
	MAC     [8]byte
	NodeMAC [8]byte
}

// CGO_ENABLED="0" go run .
// bpftool map dump pinned /sys/fs/bpf/tc/globals/ding_lxc -j | jq
func main() {
	logrus.SetReportCaller(true)

	fileName := MapPinFile
	m, err := ebpf.LoadPinnedMap(fileName, nil)
	if err != nil {
		logrus.Fatalf("LoadPinnedMap err: %v", err)
	}
	defer m.Close()

	updateMap(m, "100.0.1.1", "/var/run/netns/pod1_ns", "pod1_veth")
	updateMap(m, "100.0.1.2", "/var/run/netns/pod2_ns", "pod2_veth")
}

func updateMap(m *ebpf.Map, ip, nsPath, ifName string) {
	epKey1 := EndpointMapKey{
		IP: binary.BigEndian.Uint32(net.ParseIP(ip).To4()),
	}
	l, err := net.InterfaceByName(ifName)
	if err != nil {
		logrus.Fatalf("InterfaceByName err: %v", err)
	}
	nsIndex, nsMac := getPodInterfaceIndexAndMac(nsPath)
	epValue1 := EndpointMapInfo{
		IfIndex:    uint32(nsIndex),
		LxcIfIndex: uint32(l.Index),
		MAC:        stuff8Byte(l.HardwareAddr),
		NodeMAC:    stuff8Byte(nsMac),
	}
	err = m.Put(epKey1, epValue1)
	if err != nil {
		logrus.Fatalf("Put err: %v", err)
	}
}

func getPodInterfaceIndexAndMac(nsPath string) (int, []byte) {
	ns1, err := ns.GetNS(nsPath) // "/var/run/netns/pod1_ns"
	if err != nil {
		logrus.Fatalf("GetNS err: %v", err)
	}
	defer ns1.Close()

	var index int
	var mac []byte
	err = ns1.Do(func(netNS ns.NetNS) error {
		l, err := net.InterfaceByName("eth0")
		if err != nil {
			return err
		}
		logrus.Infof("IfIndex: %d, mac: %s", l.Index, l.HardwareAddr.String())
		index = l.Index
		mac = l.HardwareAddr
		return nil
	})
	if err != nil {
		logrus.Fatalf("GetNS Do err: %v", err)
	}

	return index, mac
}

func stuff8Byte(b []byte) [8]byte {
	var res [8]byte
	if len(b) > 8 {
		b = b[0:9]
	}

	for index, _byte := range b {
		res[index] = _byte
	}
	return res
}
