package netlink

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink/nl"
	"github.com/vishvananda/netns"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

var (
	native     = nl.NativeEndian()
	ipvsOnce   sync.Once
	ipvsFamily int
)

type NetLinkMessageHeader struct {
	cmd      uint8
	version  uint8
	reserved uint16
}

func (hdr *NetLinkMessageHeader) Serialize() []byte {
	return (*(*[unsafe.Sizeof(*hdr)]byte)(unsafe.Pointer(hdr)))[:]
}
func (hdr *NetLinkMessageHeader) Len() int {
	return int(unsafe.Sizeof(*hdr))
}

// Netlink套接字是用以实现用户进程与内核进程通信的一种特殊的进程间通信(IPC) ,也是网络应用程序与内核通信的最常用的接口。
func getIpvsFamily() (int, error) {
	// 1. 构建 netlink socket 对象，打开 netlink 套接字
	// @see github.com/vishvananda/netlink/nl::getNetlinkSocket()
	// fd, err := unix.Socket(unix.AF_NETLINK, unix.SOCK_RAW|unix.SOCK_CLOEXEC, protocol) 系统调用创建 netlink socket 套接字
	sock, err := nl.GetNetlinkSocketAt(netns.None(), netns.None(), syscall.NETLINK_GENERIC) // 通用 netlink
	if err != nil {
		return 0, err
	}
	defer sock.Close()

	// 2. 构建 request 对象
	request := nl.NewNetlinkRequest(genlCtrlID, syscall.NLM_F_ACK)
	request.AddData(&NetLinkMessageHeader{cmd: genlCtrlCmdGetFamily, version: 1})
	request.AddData(nl.NewRtAttr(genlCtrlAttrFamilyName, nl.ZeroTerminated("IPVS")))

	// 3. 发送 request 对象
	messages, err := execute(sock, request, 0)
	if err != nil {
		return 0, err
	}
	for _, message := range messages {
		hdr := deserializeGenlMsg(message)
		attrs, err := nl.ParseRouteAttr(message[hdr.Len():])
		if err != nil {
			return 0, err
		}

		for _, attr := range attrs {
			fmt.Println("attr.Value: " + string(attr.Value))

			switch int(attr.Attr.Type) {
			case genlCtrlAttrFamilyID:
				return int(native.Uint16(attr.Value[0:2])), nil
			}
		}
	}

	return 0, fmt.Errorf("no family id in the netlink response")
}

func execute(socket *nl.NetlinkSocket, request *nl.NetlinkRequest, resType uint16) ([][]byte, error) {
	if err := socket.Send(request); err != nil {
		return nil, err
	}
	pid, err := socket.GetPid()
	if err != nil {
		return nil, err
	}

	var res [][]byte

done:
	for {
		messages, _, err := socket.Receive()
		if err != nil {
			if socket.GetFd() == -1 {
				return nil, fmt.Errorf("socket got closed on receive")
			}
			if err == syscall.EAGAIN {
				// timeout fired
				continue
			}
			return nil, err
		}

		for _, message := range messages {
			if message.Header.Seq != request.Seq {
				continue
			}
			if message.Header.Pid != pid {
				return nil, fmt.Errorf("wrong pid %d, expected %d", message.Header.Pid, pid)
			}
			if message.Header.Type == syscall.NLMSG_DONE {
				break done
			}
			if message.Header.Type == syscall.NLMSG_ERROR {
				err := int32(native.Uint32(message.Data[0:4]))
				if err == 0 {
					break done
				}
				return nil, syscall.Errno(-err)
			}
			if resType != 0 && message.Header.Type != resType {
				continue
			}
			res = append(res, message.Data)
			if message.Header.Flags&syscall.NLM_F_MULTI == 0 {
				break done
			}
		}
	}

	return res, nil
}

func deserializeGenlMsg(b []byte) (hdr *NetLinkMessageHeader) {
	return (*NetLinkMessageHeader)(unsafe.Pointer(&b[0:unsafe.Sizeof(*hdr)][0]))
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

func Setup() {
	// check if ip_vs is loaded
	ipvsOnce.Do(initDependencies)
}
