package main

import (
	"encoding/binary"
	"github.com/cilium/ebpf"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/sirupsen/logrus"
	"net"
	"testing"
)

// CGO_ENABLED="0" go test -v -run ^TestNS$ .
func TestNS(test *testing.T) {
	ns1, err := ns.GetNS("/var/run/netns/pod1_ns")
	if err != nil {
		logrus.Fatalf("GetNS err: %v", err)
	}
	err = ns1.Do(func(netNS ns.NetNS) error {
		l, err := net.InterfaceByName("eth0")
		if err != nil {
			return err
		}
		logrus.Infof("IfIndex: %d, mac: %s", l.Index, l.HardwareAddr.String())
		return nil
	})
	if err != nil {
		logrus.Fatalf("GetNS Do err: %v", err)
	}
}

// CGO_ENABLED="0" go test -v -run ^TestMap$ .
func TestMap(test *testing.T) {
	m, err := ebpf.LoadPinnedMap(MapPinFile, nil)
	if err != nil {
		logrus.Fatalf("LoadPinnedMap err: %v", err)
	}
	defer m.Close()

	var key EndpointMapKey
	var value EndpointMapInfo
	mapIterator := m.Iterate()
	mapIterator.Next(&key, &value)
	logrus.Infof("key: %+v, value: %+v", key, value)
	mapIterator.Next(&key, &value)
	logrus.Infof("key: %+v, value: %+v", key, value)
}

// CGO_ENABLED="0" go test -v -run ^TestIP$ .
func TestIP(test *testing.T) {
	ipStr := "100.0.1.2"
	ipInt := binary.BigEndian.Uint32(net.ParseIP(ipStr).To4())
	logrus.Infof("ParseIP %+v", ipInt)

	ip := net.ParseIP(ipStr)
	ipUint32 := uint32(ip.To4()[0])<<24 | uint32(ip.To4()[1])<<16 | uint32(ip.To4()[2])<<8 | uint32(ip.To4()[3])
	logrus.Infof("ParseIP %+v", ipUint32)
}
