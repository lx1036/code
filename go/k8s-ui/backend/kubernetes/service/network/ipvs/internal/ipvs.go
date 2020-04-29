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
	"syscall"
	"time"
	"unsafe"
)

const (
	netlinkRecvSocketsTimeout = 3 * time.Second
	netlinkSendSocketTimeout  = 30 * time.Second
)

var (
	ipvsOnce   sync.Once
	ipvsFamily int
)

type Service struct {
	Address net.IP
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

func (handle *Handle) GetServicesCmd(service *Service) ([]*Service, error) {
	var services []*Service
	messages, err := handle.doCmdwithResponse(service, nil, ipvsCmdGetService)
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

func (handle *Handle) parseService() (*Service, error) {

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
