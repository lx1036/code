package main

import (
	"encoding/binary"
	"github.com/cilium/ebpf"
	"github.com/sirupsen/logrus"
	"net"
)

const (
	MapPinFile = "/sys/fs/bpf/tc/globals/svc_map"
)

type Backends struct {
	Be1        uint32
	Be2        uint32
	Targetport uint16
}

// CGO_ENABLED="0" go run .
// bpftool map dump pinned /sys/fs/bpf/tc/globals/svc_map -j | jq
func main() {
	logrus.SetReportCaller(true)

	fileName := MapPinFile
	m, err := ebpf.LoadPinnedMap(fileName, nil)
	if err != nil {
		logrus.Fatalf("LoadPinnedMap err: %v", err)
	}
	defer m.Close()

	key1 := uint16(31000)
	ip1 := "10.240.1.2"
	ip2 := "10.240.1.3"
	value1 := Backends{
		Be1:        binary.BigEndian.Uint32(net.ParseIP(ip1).To4()),
		Be2:        binary.BigEndian.Uint32(net.ParseIP(ip2).To4()),
		Targetport: uint16(80),
	}

	err = m.Put(key1, value1)
	if err != nil {
		logrus.Fatalf("Put err: %v", err)
	}
}
