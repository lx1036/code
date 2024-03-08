package socket_bypass_tcpip

import (
    "fmt"
    "github.com/cilium/ebpf"
    "github.com/cilium/ebpf/link"
    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/suite"
    "golang.org/x/sys/unix"
    "net"
    "os"
    "os/signal"
    "syscall"
    "testing"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go bpf test_socket_bypass_tcpip.c -- -I.

// go generate .

const (
    CgroupPath = "/sys/fs/cgroup/socket_service"

    PinPath1 = "/sys/fs/bpf/socket_bypass"

    /* External (address, port) pairs the client sends packets to. */
    EXT_IP4  = "127.0.0.1"
    EXT_PORT = 7007
)

/**
没有验证成功!!!
*/

func init() {
    logrus.SetReportCaller(true)
}

type SocketBypassTCPIPSuite struct {
    suite.Suite

    objs       *bpfObjects
    cgroupLink link.Link
    skMsgLink  *ProgAttachSkMsg
}

func TestSocketBypassTCPIPSuite(t *testing.T) {
    suite.Run(t, new(SocketBypassTCPIPSuite))
}

func (s *SocketBypassTCPIPSuite) SetupSuite() {
    cgroupPath := joinCgroup("socket_bypass")
    //defer cleanupCgroup()
    objs := bpfObjects{}
    opts := &ebpf.CollectionOptions{
        Programs: ebpf.ProgramOptions{
            LogLevel: ebpf.LogLevelInstruction,
            LogSize:  64 * 1024 * 1024, // 64M
        },
        Maps: ebpf.MapOptions{
            PinPath: PinPath1, // pin 下 map
        },
    }
    err := loadBpfObjects(&objs, opts)
    if err != nil {
        logrus.Errorf("loadBpfObjects err: %v", err)
        return
    }
    s.objs = &objs

    l1, err := link.AttachCgroup(link.CgroupOptions{
        Path:    cgroupPath,
        Program: objs.bpfPrograms.BpfSockopsV4,
        Attach:  ebpf.AttachCGroupSockOps,
    })
    if err != nil {
        logrus.Errorf("AttachCgroup err: %v", err)
        return
    }
    s.cgroupLink = l1

    // `bpftool prog attach id progID msg_verdict id mapID` attach a sockops map
    // https://github.com/cilium/cilium/blob/1c466d26ff0edfb5021d024f755d4d00bc744792/pkg/sockops/sockops.go#L47-L60
    l2, err := AttachSkMsg(objs.bpfPrograms.BpfTcpipBypass, objs.bpfMaps.SockOpsMap)
    if err != nil {
        logrus.Errorf("AttachSkMsg err: %v", err)
        return
    }
    s.skMsgLink = l2
}

func (s *SocketBypassTCPIPSuite) TearDownSuite() {
    // debug
    stopCh := make(chan os.Signal, 1)
    signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)
    <-stopCh

    if s.objs != nil {
        s.objs.Close()
    }
    if s.cgroupLink != nil {
        s.cgroupLink.Close()
    }
    if s.skMsgLink != nil {
        s.skMsgLink.Close()
    }
}

// CGO_ENABLED=0 go test -v -testify.m ^TestSocketBypass$ .
// `netstat -tulpn`
func (s *SocketBypassTCPIPSuite) TestSocketBypass() {
    // only TCP, listen at 127.0.0.1:7007
    serverFd, err := makeServer(unix.SOCK_STREAM, nil, EXT_IP4, EXT_PORT)
    if err != nil {
        return
    }
    defer unix.Close(serverFd)

    clientFd := makeClient(unix.SOCK_STREAM, EXT_IP4, EXT_PORT)
    defer unix.Close(clientFd)

    tcpEcho(clientFd, serverFd, "testing")
}

// 127.0.0.1:5432 > 127.0.0.1:7007
func makeClient(socketType int, ip string, port int) int {
    var err error
    var sockfd int
    defer func() {
        if err != nil && sockfd > 0 {
            unix.Close(sockfd)
        }
    }()

    sockfd, err = unix.Socket(unix.AF_INET, socketType, 0)
    if err != nil {
        logrus.Fatal(err)
    }
    setSocketTimeout(sockfd, 1000)

    err = unix.SetsockoptInt(sockfd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
    if err != nil {
        logrus.Errorf("unix.SO_REUSEADDR error: %v", err)
        return 0
    }

    // client bind ip:port
    // bind ip 单元测试时会报错 "address already in use"
    ip1 := net.ParseIP(EXT_IP4)
    sa1 := &unix.SockaddrInet4{
        Port: 5432, // client 源端口
        Addr: [4]byte{},
    }
    copy(sa1.Addr[:], ip1)
    err = unix.Bind(sockfd, sa1)
    if err != nil {
        logrus.Fatal(err)
    }

    ipAddr := net.ParseIP(ip)
    sa := &unix.SockaddrInet4{
        Port: port,
        Addr: [4]byte{},
    }
    copy(sa.Addr[:], ipAddr)
    // 非阻塞的, tcp/udp 都 connect()
    err = unix.Connect(sockfd, sa)
    if err != nil {
        logrus.Errorf("unix.Connect error: %v", err)
        return 0
    }

    return sockfd
}

func makeServer(socketType int, reuseportProg *ebpf.Program, ip string, port int) (int, error) {
    var err error
    var sockfd int

    sockfd, err = unix.Socket(unix.AF_INET, socketType, 0)
    if err != nil {
        logrus.Errorf("unix.Socket error: %v", err)
        return 0, err
    }
    setSocketTimeout(sockfd, 1000)
    // INFO: 注意 udp IP_RECVORIGDSTADDR 和 tcp SO_REUSEADDR
    if socketType == unix.SOCK_DGRAM {
        err = unix.SetsockoptInt(sockfd, unix.SOL_IP, unix.IP_RECVORIGDSTADDR, 1)
        if err != nil {
            logrus.Errorf("unix.IP_RECVORIGDSTADDR error: %v", err)
            return 0, err
        }
    }

    // ignores TIME-WAIT state using SO_REUSEADDR option
    // https://serverfault.com/questions/329845/how-to-forcibly-close-a-socket-in-time-wait
    err = unix.SetsockoptInt(sockfd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
    if err != nil {
        logrus.Errorf("unix.SO_REUSEADDR error: %v", err)
        return 0, err
    }

    if reuseportProg != nil {
        err = unix.SetsockoptInt(sockfd, unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
        if err != nil {
            logrus.Errorf("unix.SO_REUSEPORT error: %v", err)
            return 0, err
        }
    }

    // bind 127.0.0.1:7007
    ipAddr := net.ParseIP(ip)
    sa := &unix.SockaddrInet4{
        Port: port,
        Addr: [4]byte{},
    }
    copy(sa.Addr[:], ipAddr)
    err = unix.Bind(sockfd, sa)
    if err != nil {
        logrus.Errorf("unix.Bind error: %v", err)
        return 0, err
    }
    if socketType == unix.SOCK_STREAM {
        err = unix.Listen(sockfd, 1)
        if err != nil {
            logrus.Errorf("unix.Listen error: %v", err)
            return 0, err
        }
    }

    // INFO: 注意这里 2 个 socket_fd 都挂载了 reuseport_ebpf
    if reuseportProg != nil {
        err = unix.SetsockoptInt(sockfd, unix.SOL_SOCKET, unix.SO_ATTACH_REUSEPORT_EBPF, reuseportProg.FD())
        if err != nil {
            logrus.Errorf("unix.SO_ATTACH_REUSEPORT_EBPF error: %v", err)
            return 0, err
        }
    }

    return sockfd, nil
}

func setSocketTimeout(fd, timeoutMs int) {
    var timeVal *unix.Timeval
    if timeoutMs > 0 {
        timeVal = &unix.Timeval{
            Sec:  int64(timeoutMs / 1000),
            Usec: int64(timeoutMs % 1000 * 1000),
        }
    } else {
        timeVal = &unix.Timeval{
            Sec: 3,
        }
    }

    err := unix.SetsockoptTimeval(fd, unix.SOL_SOCKET, unix.SO_RCVTIMEO, timeVal)
    if err != nil {
        logrus.Fatal(err)
    }

    err = unix.SetsockoptTimeval(fd, unix.SOL_SOCKET, unix.SO_SNDTIMEO, timeVal)
    if err != nil {
        logrus.Fatal(err)
    }
}

func tcpEcho(clientFd, serverFd int, echoData string) {
    // 这里不使用 Sendto(fd int, p []byte, flags int, to Sockaddr) 原因是已经调用 Connect(server)
    err := unix.Send(clientFd, []byte(echoData), 0)
    if err != nil {
        logrus.Errorf("unix.Send err: %v", err)
        return
    }

    clientServerFd, clientSockAddr, err := unix.Accept(serverFd)
    if err != nil {
        logrus.Errorf("unix.Accept err: %v", err)
        return
    }
    defer unix.Close(clientServerFd)
    // 保持源端口，无需改造 server IP_TRANSPARENT(https://powerdns.org/tproxydoc/tproxy.md.html) 就可以获得源端口
    clientPort := clientSockAddr.(*unix.SockaddrInet4).Port
    logrus.Infof("client port %d", clientPort) // 5432
    cbuf := make([]byte, 1024)
    n, _, _, _, err := unix.Recvmsg(clientServerFd, cbuf, nil, 0)
    if err != nil {
        logrus.Errorf("unix.Recvmsg clientServerFd err: %v", err)
        return
    }
    cbuf = cbuf[:n]
    logrus.Infof("server recvmsg from client: %s", string(cbuf))

    err = unix.Send(clientServerFd, cbuf, 0) // server > client
    if err != nil {
        logrus.Errorf("unix.Send err: %v", err)
        return
    }
    cbuf2 := make([]byte, 1024)
    n2, _, _, _, err := unix.Recvmsg(clientFd, cbuf2, nil, 0)
    if err != nil {
        logrus.Errorf("unix.Recvmsg clientFd err: %v", err)
        return
    }
    cbuf2 = cbuf2[:n2]
    logrus.Infof("client recvmsg from server: %s", string(cbuf2))

    if string(cbuf2) == echoData {
        logrus.Infof("tcp echo successfully")
    } else {
        logrus.Errorf("fail to tcp echo")
    }
}

type ProgAttachSkMsg struct {
    mapId      ebpf.MapID
    program    *ebpf.Program
    attachType ebpf.AttachType
}

func (skMsg *ProgAttachSkMsg) Close() error {
    err := link.RawDetachProgram(link.RawDetachProgramOptions{
        Target:  int(skMsg.mapId),
        Program: skMsg.program,
        Attach:  skMsg.attachType,
    })
    if err != nil {
        return fmt.Errorf("close cgroup: %s", err)
    }
    return nil
}

func AttachSkMsg(prog *ebpf.Program, bpfMap *ebpf.Map) (*ProgAttachSkMsg, error) {
    if t := prog.Type(); t != ebpf.SkMsg {
        return nil, fmt.Errorf("invalid program type %s, expected SkMsg", t)
    }

    info, err := bpfMap.Info()
    if err != nil {
        return nil, err
    }
    mapId, ok := info.ID()
    if !ok {
        return nil, fmt.Errorf("invalid map id: %d", mapId)
    }

    err = link.RawAttachProgram(link.RawAttachProgramOptions{
        Target:  int(mapId),
        Program: prog,
        Attach:  ebpf.AttachSkMsgVerdict,
        Flags:   0,
    })

    skMsg := &ProgAttachSkMsg{
        mapId:      mapId,
        program:    prog,
        attachType: ebpf.AttachSkMsgVerdict,
    }

    return skMsg, nil
}

func joinCgroup(path string) string {
    if len(path) == 0 {
        logrus.Fatalf("path is empty")
    }

    cgroupPath := fmt.Sprintf("%s/%s", CgroupPath, path)
    if err := os.MkdirAll(cgroupPath, 0777); err != nil {
        logrus.Fatalf("os.Mkdir err: %v", err)
    }
    pid := os.Getpid()
    file := fmt.Sprintf("%s/cgroup.procs", cgroupPath)
    if err := os.WriteFile(file, []byte(fmt.Sprintf("%d\n", pid)), 0644); err != nil {
        logrus.Fatalf("os.WriteFile err: %v", err)
    }

    return cgroupPath
}

func cleanupCgroup() {
    os.RemoveAll(CgroupPath)
}
