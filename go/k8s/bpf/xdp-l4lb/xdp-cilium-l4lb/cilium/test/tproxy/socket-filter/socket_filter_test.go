package socket_filter

import (
    "github.com/cilium/ebpf"
    "github.com/sirupsen/logrus"
    "golang.org/x/sys/unix"
    "net"
    "os"
    "os/signal"
    "syscall"
    "testing"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go bpf test_socket_filter.c -- -I.

// go generate .

/**
没有验证成功，一旦 attack socket_filter bpf 程序，server 没法 ack client 的数据包
`tcpdump -i lo -nneevv port 8080`
*/

const (
    PinPath1 = "/sys/fs/bpf/socket_filter"

    InaddrLoopback = "127.0.0.1"
    SeverPort      = 8080
)

// CGO_ENABLED=0 go test -v -count=1 -run ^TestSocketFilter$ .
// tcpdump -i lo -nneevv port 8080 -w tcp_fsm.pcap
func TestSocketFilter(test *testing.T) {
    logrus.SetReportCaller(true)

    stopCh := make(chan os.Signal, 1)
    signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

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
    defer objs.Close()

    key := uint32(0)
    //value := make([]uint64, runtime.NumCPU())
    //for i := 0; i < runtime.NumCPU(); i++ {
    //    value[i] = 0
    //}
    value := uint64(0)
    err = objs.bpfMaps.CounterMap.Put(key, value)
    if err != nil {
        logrus.Errorf("CounterMap.Put err: %v", err)
        return
    }

    serverFd := makeServer(objs.bpfPrograms.PacketCounter)
    //defer unix.Close(serverFd)
    clientFd := connectToFd(serverFd)
    //defer unix.Close(clientFd)

    echoData := []byte("testing")
    err = unix.Send(clientFd, echoData, 0)
    if err != nil {
        logrus.Errorf("unix.Send err: %v", err)
        return
    }
    clientServerFd, clientSockAddr, err := unix.Accept(serverFd)
    if err != nil {
        logrus.Errorf("unix.Accept err: %v", err)
        return
    }
    //defer unix.Close(clientServerFd)
    clientPort := clientSockAddr.(*unix.SockaddrInet4).Port
    logrus.Infof("client port %d", clientPort)
    cbuf := make([]byte, 1024)
    n, _, _, _, err := unix.Recvmsg(clientServerFd, cbuf, nil, 0)
    if err != nil {
        logrus.Errorf("unix.Recvmsg clientServerFd err: %v", err)
        return
    }
    cbuf = cbuf[:n]
    logrus.Infof("server recvmsg from client: %s", string(cbuf))

    err = unix.Send(clientServerFd, cbuf, 0)
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

    err = DetachSocketFilter(serverFd)
    if err != nil {
        logrus.Errorf("DetachSocketFilter err: %v", err)
    }

    // close socket
    unix.Close(clientServerFd)
    unix.Close(clientFd)
    unix.Close(serverFd)

    // Wait
    <-stopCh
}

// @see github.com/cilium/ebpf/link::AttachSocketFilter()
func AttachSocketFilter(sockFd int, program *ebpf.Program) error {
    return unix.SetsockoptInt(sockFd, unix.SOL_SOCKET, unix.SO_ATTACH_BPF, program.FD())
}

func DetachSocketFilter(sockFd int) error {
    return unix.SetsockoptInt(sockFd, unix.SOL_SOCKET, unix.SO_DETACH_BPF, 0)
}

func makeServer(prog *ebpf.Program) int {
    serverFd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
    if err != nil {
        logrus.Fatal(err)
    }
    setSocketTimeout(serverFd, 5000)

    err = unix.SetsockoptInt(serverFd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
    if err != nil {
        logrus.Fatalf("unix.SO_REUSEADDR error: %v", err)
    }

    err = AttachSocketFilter(serverFd, prog)
    if err != nil {
        logrus.Errorf("AttachSocketFilter err: %v", err)
    }

    ip := net.ParseIP(InaddrLoopback)
    sa := &unix.SockaddrInet4{
        Port: SeverPort,
        Addr: [4]byte{},
    }
    copy(sa.Addr[:], ip)
    err = unix.Bind(serverFd, sa)
    if err != nil {
        logrus.Fatal(err)
    }

    err = unix.Listen(serverFd, 128)
    if err != nil {
        logrus.Fatal(err)
    }

    return serverFd
}

// /root/linux-5.10.142/tools/testing/selftests/bpf/tcp_client.py
func connectToFd(serverFd int) int {
    clientFd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0) // unix.SOCK_STREAM
    if err != nil {
        logrus.Fatal(err)
    }
    setSocketTimeout(clientFd, 5000)

    ip1 := net.ParseIP(InaddrLoopback)
    sa1 := &unix.SockaddrInet4{
        Port: 5432, // client 源端口
        Addr: [4]byte{},
    }
    copy(sa1.Addr[:], ip1)
    err = unix.Bind(clientFd, sa1)
    if err != nil {
        logrus.Fatal(err)
    }

    serverSockAddr, err := unix.Getsockname(serverFd)
    if err != nil {
        logrus.Fatal(err)
    }

    // 非阻塞的
    err = unix.Connect(clientFd, serverSockAddr)
    if err != nil {
        logrus.Fatal(err)
    }

    return clientFd
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
