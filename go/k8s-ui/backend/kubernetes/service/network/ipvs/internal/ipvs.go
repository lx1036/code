// +build linux

package internal

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink/nl"
	"github.com/vishvananda/netns"
	"golang.org/x/sys/unix"
	"net"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

const (
	netlinkRecvSocketsTimeout = 3 * time.Second
	netlinkSendSocketTimeout  = 30 * time.Second
)

const (
	genlCtrlID = 0x10
)

// GENL control commands
const (
	genlCtrlCmdUnspec uint8 = iota
	genlCtrlCmdNewFamily
	genlCtrlCmdDelFamily
	genlCtrlCmdGetFamily
)

// GENL family attributes
const (
	genlCtrlAttrUnspec int = iota
	genlCtrlAttrFamilyID
	genlCtrlAttrFamilyName
)

// IPVS genl commands
const (
	ipvsCmdUnspec uint8 = iota
	ipvsCmdNewService
	ipvsCmdSetService
	ipvsCmdDelService
	ipvsCmdGetService
	ipvsCmdNewDestination
	ipvsCmdSetDestination
	ipvsCmdDelDestination
	ipvsCmdGetDestination
	ipvsCmdNewDaemon
	ipvsCmdDelDaemon
	ipvsCmdGetDaemon
	ipvsCmdSetConfig
	ipvsCmdGetConfig
	ipvsCmdSetInfo
	ipvsCmdGetInfo
	ipvsCmdZero
	ipvsCmdFlush
)

var (
	ipvsOnce   sync.Once
	ipvsFamily int
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
type Handle struct {
	seq  uint32
	sock *nl.NetlinkSocket
}

// new a ipvs handle
func New(path string) (*Handle, error) {
	setup()

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

	return &Handle{sock: sock}, nil
}

func (handle *Handle) GetServices() ([]*Service, error) {
	return handle.GetServicesCmd(nil)
}

func (handle *Handle) GetDestinations(service *Service) ([]*Destination, error) {
	return handle.GetDestinationCmd(service, nil)
}

func (handle *Handle) GetServicesCmd(service *Service) ([]*Service, error) {
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

func (handle *Handle) GetDestinationCmd(service *Service, destination *Destination) ([]*Destination, error) {
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

func (handle *Handle) parseService(message []byte) (*Service, error) {

	//Remove General header for this message and parse the NetLink message
	hdr := deserializeGenlMsg(msg)
	NetLinkAttrs, err := nl.ParseRouteAttr(msg[hdr.Len():])
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

func (handle *Handle) parseDestination(message []byte) (*Destination, error) {

}

func (handle *Handle) doCmdWithResponse(service *Service, destination *Destination, cmd uint8) ([][]byte, error) {
	request := newIPVSRequest(cmd)
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

func setup() {
	// check if ip_vs is loaded
	ipvsOnce.Do(initDependencies)
}

func initDependencies() {
	path, err := exec.LookPath("modprobe")
	if err != nil {
		log.Warnf("failed to find modprobe: %v", err)
		return
	}
	if out, err := exec.Command(path, "-va", "ip_vs").CombinedOutput(); err != nil {
		log.Warnf("failed to run [%s -va ip_vs] with messages: %s, error: %v", path, strings.TrimSpace(string(out)), err)
	}
	ipvsFamily, err = getIpvsFamily()
	if err != nil {
		log.Warnf("failed to get ipvs family information from the kernel: %v", err)
		return
	}
}

// Netlink套接字是用以实现用户进程与内核进程通信的一种特殊的进程间通信(IPC) ,也是网络应用程序与内核通信的最常用的接口。
func getIpvsFamily() (int, error) {
	sock, err := nl.GetNetlinkSocketAt(netns.None(), netns.None(), syscall.NETLINK_GENERIC) // 通用 netlink
	if err != nil {
		return 0, err
	}
	defer sock.Close()

	request := newGenlRequest(genlCtrlID, genlCtrlCmdGetFamily)
	request.AddData(nl.NewRtAttr(genlCtrlAttrFamilyName, nl.ZeroTerminated("IPVS")))
	messages, err := execute(sock, request, 0)
	if err != nil {
		return 0, err
	}
	for _, message := range messages {

	}
}

func execute(s *nl.NetlinkSocket, req *nl.NetlinkRequest, resType uint16) ([][]byte, error) {

}

func newGenlRequest(familyID int, cmd uint8) *nl.NetlinkRequest {
	request := nl.NewNetlinkRequest(familyID, syscall.NLM_F_ACK)
	request.AddData(&genlMsgHdr{cmd: cmd, version: 1})
	return request
}

func newIPVSRequest(cmd uint8) *nl.NetlinkRequest {
	return newGenlRequest(ipvsFamily, cmd)
}

type genlMsgHdr struct {
	cmd      uint8
	version  uint8
	reserved uint16
}

func (hdr *genlMsgHdr) Serialize() []byte {
	return (*(*[unsafe.Sizeof(*hdr)]byte)(unsafe.Pointer(hdr)))[:]
}
func (hdr *genlMsgHdr) Len() int {
	return int(unsafe.Sizeof(*hdr))
}
