package main

import (
    "fmt"
    "github.com/cilium/ebpf"
    "github.com/cilium/ebpf/link"
    "github.com/containernetworking/plugins/pkg/ns"
    "github.com/sirupsen/logrus"
    "golang.org/x/sys/unix"
    "net"
    "path/filepath"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go bpf test_sk_lookup.c -- -I.

const (
    SOMAXCONN = 4096

    /*MAX_SERVERS enum {
        SERVER_A = 0,
        SERVER_B,
    };*/
    MAX_SERVERS = 2

    /* External (address, port) pairs the client sends packets to. */
    EXT_IP4  = "127.0.0.1"
    EXT_PORT = 7007

    /* Internal (address, port) pairs the server listens/receives at. */
    INT_IP4  = "127.0.0.2"
    INT_PORT = 8008
)

const (
    ServerA = iota
    ServerB
)

const (
    Prog1 = iota
    Prog2
)

// go generate .
// CGO_ENABLED=0 go run .
func main() {
    logrus.SetReportCaller(true)

    netnsPath := "/proc/self/ns/net"
    bpfFsPath := "/sys/fs/bpf"
    netns, pinPath, err := openNetNS(netnsPath, bpfFsPath)
    if err != nil {
        logrus.Errorf("openNetNS err: %v", err)
        return
    }
    defer netns.Close()

    // 1.Load pre-compiled programs and maps into the kernel.
    objs := bpfObjects{}
    opts := &ebpf.CollectionOptions{
        Programs: ebpf.ProgramOptions{
            LogLevel: ebpf.LogLevelInstruction,
            LogSize:  64 * 1024 * 1024, // 64M
        },
        Maps: ebpf.MapOptions{
            PinPath: pinPath, // pin 下 map
        },
    }
    err = loadBpfObjects(&objs, opts)
    if err != nil {
        logrus.Errorf("loadBpfObjects err: %v", err)
        return
    }
    defer objs.Close()

    // 2.attach sk_lookup/redir_port into netns
    l, err := link.AttachNetNs(int(netns.Fd()), objs.bpfPrograms.RedirPort)
    if err != nil {
        logrus.Errorf("attach program to netns %s: %s", netns.Path(), err)
        return
    }
    defer l.Close()

    // 3.update map
    serverFd, err := makeServer(unix.SOCK_DGRAM, nil, EXT_IP4, INT_PORT)
    if err != nil {
        return
    }
    //serverFds := makeServer(unix.SOCK_STREAM, nil)
    defer unix.Close(serverFd)
    key := uint32(0)
    value := uint64(serverFd)
    err = objs.bpfMaps.RedirMap.Put(key, value)

    //reuseportHasConns := false
    //if reuseportHasConns {
    //
    //}

    // 4.echo server test
    clientFd := makeClient(unix.SOCK_DGRAM, EXT_IP4, EXT_PORT)
    //clientFd := makeClient(unix.SOCK_STREAM)
    defer unix.Close(clientFd)
    //tcpEcho(clientFd, serverFd)
    udpEcho(clientFd, serverFd, getEchoData(0))
}

func udpEcho(clientFd, serverFd int, echoData string) {
    err := unix.Send(clientFd, []byte(echoData), 0)
    if err != nil {
        logrus.Errorf("unix.Send err: %v", err)
        return
    }
    cbuf := make([]byte, 1024)
    n, _, _, from, err := unix.Recvmsg(serverFd, cbuf, nil, 0)
    if err != nil {
        logrus.Errorf("unix.Recvmsg err: %v", err)
        return
    }
    cbuf = cbuf[:n]
    logrus.Infof("server unix.Recvmsg from client: %s", string(cbuf))

    clientPort := from.(*unix.SockaddrInet4).Port
    logrus.Infof("client port %d", clientPort) // 58442

    // INFO: 由于 serverFd 是 redirect 之后的 socket_fd，所以不能直接 unix.Sendto(serverFd, cbuf, 0, from)，会 block
    //  只能新建一个 socket_fd, bind 到 original dest address，然后 unix.Sendmsg(clientServerFd, cbuf, nil, from, 0)

    // INFO: 类似于 TCP Accept() 作用
    serverSocketAddr, err := unix.Getpeername(clientFd)
    if err != nil {
        logrus.Errorf("unix.Getpeername err: %v", err)
        return
    }
    clientServerFd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
    if err != nil {
        logrus.Errorf("unix.Socket: %v", err)
        return
    }
    defer unix.Close(clientServerFd)
    err = unix.SetsockoptInt(clientServerFd, unix.SOL_IP, unix.IP_RECVORIGDSTADDR, 1)
    if err != nil {
        logrus.Errorf("unix.IP_RECVORIGDSTADDR error: %v", err)
        return
    }
    // udp 必须加上 unix.SO_REUSEADDR, 否则 socket bind 也会报错 "already in use"
    err = unix.SetsockoptInt(clientServerFd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
    if err != nil {
        logrus.Errorf("unix.IP_RECVORIGDSTADDR error: %v", err)
        return
    }
    err = unix.Bind(clientServerFd, serverSocketAddr) // INFO: 必须是 original dest server，否则 block
    if err != nil {
        logrus.Errorf("unix.Bind err: %v", err)
        return
    }

    // 这里不使用 Sendto(fd int, p []byte, flags int, to Sockaddr) 原因是还没有 Connect(server)
    err = unix.Sendmsg(clientServerFd, cbuf, nil, from, 0)
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
    logrus.Infof("client unix.Recvmsg from server: %s", string(cbuf2))

    if string(cbuf2) == echoData {
        logrus.Infof("udp echo successfully")
    } else {
        logrus.Errorf("fail to tcp echo")
    }
}

// INFO: 目前验证，TCP 保持目的源端口，无需改造 server tproxy+IP_TRANSPARENT(https://powerdns.org/tproxydoc/tproxy.md.html) 就可以获得源端口
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

func getEchoData(server int) string {
    switch server {
    case ServerA:
        return "ServerA"
    case ServerB:
        return "ServerB"
    default:
        return ""
    }
}

// 127.0.0.1:5432 connect 127.0.0.1:7007
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

    setSocketTimeout(sockfd, 1000)
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

// listen at 127.0.0.1:8008
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

    // bind 127.0.0.1:8008
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
        err = unix.Listen(sockfd, SOMAXCONN)
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

func socketCookie(fd int) int {
    cookie, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_COOKIE)
    if err != nil {
        logrus.Errorf("unix.SO_COOKIE err: %v", err)
        return 0
    }

    return cookie
}

func openNetNS(nsPath, bpfFsPath string) (ns.NetNS, string, error) {
    var fs unix.Statfs_t
    err := unix.Statfs(bpfFsPath, &fs)
    if err != nil || fs.Type != unix.BPF_FS_MAGIC {
        return nil, "", fmt.Errorf("invalid BPF filesystem path: %s", bpfFsPath)
    }

    netNs, err := ns.GetNS(nsPath)
    if err != nil {
        return nil, "", err
    }

    var stat unix.Stat_t
    if err := unix.Fstat(int(netNs.Fd()), &stat); err != nil {
        return nil, "", fmt.Errorf("stat netns: %s", err)
    }

    dir := fmt.Sprintf("%d_dispatcher", stat.Ino)
    return netNs, filepath.Join(bpfFsPath, dir), nil
}
