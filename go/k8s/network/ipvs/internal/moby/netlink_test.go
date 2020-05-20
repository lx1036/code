// +build linux

package moby

import (
	"encoding/binary"
	"fmt"
	"github.com/vishvananda/netlink/nl"
	"github.com/vishvananda/netns"
	"golang.org/x/sys/unix"
	"gotest.tools/assert"
	"net"
	"syscall"
	"testing"
	"time"
)

const (
	netlinkRecvSocketsTimeout = 3 * time.Second
	netlinkSendSocketTimeout  = 30 * time.Second
)

func TestGetFamily(test *testing.T) {
	ipvsFamily, err := getIpvsFamily()
	assert.NilError(test, err)
	assert.Check(test, ipvsFamily != 0)
}

func TestGetService(test *testing.T) {
	Setup()

	// 1. 打开 netlink 套接字
	sock, err := nl.GetNetlinkSocketAt(netns.None(), netns.None(), syscall.NETLINK_GENERIC) // 通用 netlink
	if err != nil {
		panic(err)
	}
	defer sock.Close()
	// Add operation timeout to avoid deadlocks
	tv := unix.NsecToTimeval(netlinkSendSocketTimeout.Nanoseconds())
	if err := sock.SetSendTimeout(&tv); err != nil {
		panic(err)
	}
	tv = unix.NsecToTimeval(netlinkRecvSocketsTimeout.Nanoseconds())
	if err := sock.SetReceiveTimeout(&tv); err != nil {
		panic(err)
	}

	// 2. 构建 request 对象
	request := nl.NewNetlinkRequest(ipvsFamily, syscall.NLM_F_ACK)
	request.AddData(&NetLinkMessageHeader{cmd: ipvsCmdGetService, version: 1})
	request.Flags |= syscall.NLM_F_DUMP                    //Flag to dump all messages
	request.AddData(nl.NewRtAttr(ipvsCmdAttrService, nil)) //Add a dummy attribute
	//request.Seq =

	// 3. 发送 request 对象
	messages, err := execute(sock, request, 0)
	if err != nil {
		panic(err)
	}

	for _, message := range messages {
		hdr := deserializeGenlMsg(message)
		attrs, err := nl.ParseRouteAttr(message[hdr.Len():])
		if err != nil {
			panic(err)
		}
		if len(attrs) == 0 {
			continue
		}
		//Now Parse and get IPVS related attributes messages packed in this message.
		ipvsAttrs, err := nl.ParseRouteAttr(attrs[0].Value)
		if err != nil {
			panic(err)
		}

		// Service defines an IPVS service in its entirety.
		type Service struct {
			// Virtual service address.
			Address  net.IP
			Protocol uint16
			Port     uint16
			FWMark   uint32 // Firewall mark of the service.

			// Virtual service options.
			SchedName     string
			Flags         uint32
			Timeout       uint32
			Netmask       uint32
			AddressFamily uint16
			PEName        string
			//Stats         SvcStats
		}

		var s Service
		var addressBytes []byte
		for _, attr := range ipvsAttrs {
			attrType := int(attr.Attr.Type)
			switch attrType {
			case ipvsSvcAttrAddressFamily:
				s.AddressFamily = native.Uint16(attr.Value)
			case ipvsSvcAttrProtocol:
				s.Protocol = native.Uint16(attr.Value)
			case ipvsSvcAttrAddress:
				addressBytes = attr.Value
			case ipvsSvcAttrPort:
				s.Port = binary.BigEndian.Uint16(attr.Value)
			case ipvsSvcAttrFWMark:
				s.FWMark = native.Uint32(attr.Value)
			case ipvsSvcAttrSchedName:
				s.SchedName = nl.BytesToString(attr.Value)
			case ipvsSvcAttrFlags:
				s.Flags = native.Uint32(attr.Value)
			case ipvsSvcAttrTimeout:
				s.Timeout = native.Uint32(attr.Value)
			case ipvsSvcAttrNetmask:
				s.Netmask = native.Uint32(attr.Value)
				//case ipvsSvcAttrStats:
				//	stats, err := assembleStats(attr.Value)
				//	if err != nil {
				//		return nil, err
				//	}
				//	s.Stats = stats
			}
		}

		resIP := net.ParseIP("192.168.0.0")
		if addressBytes != nil {
			switch s.AddressFamily {
			case syscall.AF_INET:
				resIP = addressBytes[:4]
			case syscall.AF_INET6:
				resIP = addressBytes[:16]
			}
		}

		fmt.Println(resIP.String())
		fmt.Println(s)
	}
}
