package bgp

import (
	"encoding/binary"
	"io"
	"net"
	"time"
)

func sendOpen(w io.Writer, asn uint32, routerID net.IP, holdTime time.Duration) error {
	if routerID.To4() == nil {
		panic("non-ipv4 address used as RouterID")
	}

	msg := struct {
		// Header
		Marker1, Marker2 uint64
		Len              uint16
		Type             uint8

		// OPEN
		Version  uint8
		ASN16    uint16
		HoldTime uint16
		RouterID [4]byte

		// Options (we only send one, capabilities)
		OptsLen uint8
		OptType uint8
		OptLen  uint8

		// Capabilities: multiprotocol extension for IPv4+IPv6
		// unicast, and 4-byte ASNs

		MP4Type uint8
		MP4Len  uint8
		AFI4    uint16
		SAFI4   uint16

		MP6Type uint8
		MP6Len  uint8
		AFI6    uint16
		SAFI6   uint16

		CapType uint8
		CapLen  uint8
		ASN32   uint32
	}{
		Marker1: 0xffffffffffffffff,
		Marker2: 0xffffffffffffffff,
		Len:     0, // Filled below
		Type:    1, // OPEN

		Version:  4,
		ASN16:    uint16(asn), // Possibly tweaked below
		HoldTime: uint16(holdTime.Seconds()),
		// RouterID filled below

		OptsLen: 20,
		OptType: 2, // Capabilities
		OptLen:  18,

		MP4Type: 1, // BGP Multi-protocol Extensions
		MP4Len:  4,
		AFI4:    1, // IPv4
		SAFI4:   1, // Unicast

		MP6Type: 1, // BGP Multi-protocol Extensions
		MP6Len:  4,
		AFI6:    2, // IPv6
		SAFI6:   1, // Unicast

		CapType: 65, // 4-byte ASN
		CapLen:  4,
		ASN32:   asn,
	}
	msg.Len = uint16(binary.Size(msg))
	if asn > 65535 {
		msg.ASN16 = 23456
	}
	copy(msg.RouterID[:], routerID.To4())

	return binary.Write(w, binary.BigEndian, msg)
}

func sendKeepalive(w io.Writer) error {
	msg := struct {
		Marker1, Marker2 uint64
		Len              uint16
		Type             uint8
	}{
		Marker1: 0xffffffffffffffff,
		Marker2: 0xffffffffffffffff,
		Len:     19,
		Type:    4,
	}
	return binary.Write(w, binary.BigEndian, msg)
}
