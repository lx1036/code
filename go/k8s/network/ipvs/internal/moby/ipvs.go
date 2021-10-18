//go:build linux
// +build linux

package moby

import (
	"fmt"
	"github.com/vishvananda/netlink/nl"
	"github.com/vishvananda/netns"
	"golang.org/x/sys/unix"
	"net"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	netlinkRecvSocketsTimeout = 3 * time.Second
	netlinkSendSocketTimeout  = 30 * time.Second
)

// define a ipvs service
// e.g. &{117.50.107.43 6 80 0 rr 2 0 4294967295 2  {0 0 0 0 0 0 0 0 0 0}}
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
	Stats         SvcStats
}

// define real server
// e.g. :
// &{192.168.1.37 80 1 0 2 0 0 0 0 {0 0 0 0 0 0 0 0 0 0}}
// &{192.168.1.142 80 1 0 2 0 0 0 0 {0 0 0 0 0 0 0 0 0 0}}
type Destination struct {
	Address             net.IP
	Port                uint16
	Weight              int
	ConnectionFlags     uint32
	AddressFamily       uint16
	UpperThreshold      uint32
	LowerThreshold      uint32
	ActiveConnections   int
	InactiveConnections int
	Stats               DstStats
}
type DstStats SvcStats

// SvcStats defines an IPVS service statistics
type SvcStats struct {
	Connections uint32
	PacketsIn   uint32
	PacketsOut  uint32
	BytesIn     uint64
	BytesOut    uint64
	CPS         uint32
	BPSOut      uint32
	PPSIn       uint32
	PPSOut      uint32
	BPSIn       uint32
}

// a specific namespace handler
type Runner struct {
	seq  uint32
	sock *nl.NetlinkSocket
}

// new a ipvs handle
func New(path string) (*Runner, error) {
	Setup()

	//simple way to handle network namespaces
	ns := netns.None()
	if path != "" {
		var err error
		ns, err = netns.GetFromPath(path)
		if err != nil {
			return nil, err
		}
	}
	defer ns.Close()

	sock, err := nl.GetNetlinkSocketAt(ns, netns.None(), unix.NETLINK_GENERIC)
	if err != nil {
		return nil, err
	}
	// Add operation timeout to avoid deadlocks
	tv := unix.NsecToTimeval(netlinkSendSocketTimeout.Nanoseconds())
	if err := sock.SetSendTimeout(&tv); err != nil {
		return nil, err
	}
	tv = unix.NsecToTimeval(netlinkRecvSocketsTimeout.Nanoseconds())
	if err := sock.SetReceiveTimeout(&tv); err != nil {
		return nil, err
	}

	return &Runner{sock: sock}, nil
}

func (handle *Runner) GetServices() ([]*Service, error) {
	return handle.GetServicesCmd(nil)
}

func (handle *Runner) GetDestinations(service *Service) ([]*Destination, error) {
	return handle.GetDestinationCmd(service, nil)
}

func (handle *Runner) GetServicesCmd(service *Service) ([]*Service, error) {
	var services []*Service
	messages, err := handle.doCmdWithResponse(service, nil, ipvsCmdGetService)
	if err != nil {
		return nil, err
	}

	for _, message := range messages {
		service, err := handle.parseService(message)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	return services, nil
}

func (handle *Runner) GetDestinationCmd(service *Service, destination *Destination) ([]*Destination, error) {
	var destionations []*Destination
	messages, err := handle.doCmdWithResponse(service, destination, ipvsCmdGetDestination)
	if err != nil {
		return nil, err
	}

	for _, message := range messages {
		destionation, err := handle.parseDestination(message)
		if err != nil {
			return nil, err
		}
		destionations = append(destionations, destionation)
	}

	return destionations, nil
}

func (handle *Runner) parseService(message []byte) (*Service, error) {

	//Remove General header for this message and parse the NetLink message
	hdr := deserializeGenlMsg(message)
	NetLinkAttrs, err := nl.ParseRouteAttr(message[hdr.Len():])
	if err != nil {
		return nil, err
	}
	if len(NetLinkAttrs) == 0 {
		return nil, fmt.Errorf("error no valid netlink message found while parsing service record")
	}

	//Now Parse and get IPVS related attributes messages packed in this message.
	ipvsAttrs, err := nl.ParseRouteAttr(NetLinkAttrs[0].Value)
	if err != nil {
		return nil, err
	}

}

func (handle *Runner) parseDestination(message []byte) (*Destination, error) {

}

func (handle *Runner) doCmdWithResponse(service *Service, destination *Destination, cmd uint8) ([][]byte, error) {
	request := nl.NewNetlinkRequest(ipvsFamily, syscall.NLM_F_ACK)
	request.AddData(&genlMsgHdr{cmd: cmd, version: 1})
	request.Seq = atomic.AddUint32(&handle.seq, 1)

	if service == nil {
		request.Flags |= syscall.NLM_F_DUMP                    //Flag to dump all messages
		request.AddData(nl.NewRtAttr(ipvsCmdAttrService, nil)) //Add a dummy attribute
	} else {
		request.AddData(fillService(service))
	}

	if destination == nil {
		if cmd == ipvsCmdGetDestination {
			request.Flags |= syscall.NLM_F_DUMP
		}

	} else {
		request.AddData(fillDestination(destination))
	}

	response, err := execute(handle.sock, request, 0)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func newGenlRequest(familyID int, cmd uint8) *nl.NetlinkRequest {
	request := nl.NewNetlinkRequest(familyID, syscall.NLM_F_ACK)
	request.AddData(&genlMsgHdr{cmd: cmd, version: 1})
	return request
}

func newIPVSRequest(cmd uint8) *nl.NetlinkRequest {
	return newGenlRequest(ipvsFamily, cmd)
}
