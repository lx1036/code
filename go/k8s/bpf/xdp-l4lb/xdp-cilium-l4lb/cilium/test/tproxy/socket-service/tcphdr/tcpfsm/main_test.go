package main

// /root/linux-5.10.142/tools/testing/selftests/bpf/test_tcpbpf_user.c

import (
    "fmt"
    "github.com/cilium/ebpf"
    "github.com/cilium/ebpf/link"
    "github.com/sirupsen/logrus"
    "github.com/stretchr/testify/assert"
    "golang.org/x/sys/unix"
    "net"
    "os"
    "os/signal"
    "syscall"
    "testing"
    "time"
    "unsafe"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go bpf test_tcp_fsm.c -- -I.

// go generate .

/**
验证通过!!!
*/

const (
    CgroupPath = "/sys/fs/cgroup"

    PinPath1 = "/sys/fs/bpf/socket_service/tcp_fsm"

    InaddrLoopback = "127.0.0.1"
    SeverPort      = 8080

    /* 3 comes from one listening socket + both ends of the connection */
    ExpectedCloseEvents = 3
)

// /root/linux-5.10.142/tools/include/uapi/linux/bpf.h
const (
    BPF_SOCK_OPS_VOID = iota
    BPF_SOCK_OPS_TIMEOUT_INIT
    BPF_SOCK_OPS_RWND_INIT
    BPF_SOCK_OPS_TCP_CONNECT_CB
    BPF_SOCK_OPS_ACTIVE_ESTABLISHED_CB
    BPF_SOCK_OPS_PASSIVE_ESTABLISHED_CB
    BPF_SOCK_OPS_NEEDS_ECN
    BPF_SOCK_OPS_BASE_RTT
    BPF_SOCK_OPS_RTO_CB
    BPF_SOCK_OPS_RETRANS_CB
    BPF_SOCK_OPS_STATE_CB
    BPF_SOCK_OPS_TCP_LISTEN_CB
    BPF_SOCK_OPS_RTT_CB
    BPF_SOCK_OPS_PARSE_HDR_OPT_CB
    BPF_SOCK_OPS_HDR_OPT_LEN_CB
    BPF_SOCK_OPS_WRITE_HDR_OPT_CB
)

// CGO_ENABLED=0 go test -v -count=1 -run ^TestTcpFsm$ .
// tcpdump -i lo -nneevv port 8080 -w tcp_fsm.pcap
func TestTcpFsm(test *testing.T) {
    logrus.SetReportCaller(true)

    stopCh := make(chan os.Signal, 1)
    signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

    cgroupPath := joinCgroup("tcp_fsm2")
    //defer cleanupCgroup()

    //1.Load pre-compiled programs and maps into the kernel.
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

    l1, err := link.AttachCgroup(link.CgroupOptions{
        Path:    cgroupPath,
        Program: objs.bpfPrograms.BpfTcpFsm,
        Attach:  ebpf.AttachCGroupSockOps,
    })
    if err != nil {
        logrus.Fatal(err)
    }
    defer l1.Close()

    serverFd := makeServer()
    //defer unix.Close(serverFd)
    clientFd := connectToFd(serverFd)
    //defer unix.Close(clientFd)
    // client<=>server
    echoData := []byte(makeBytes())
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
    //logrus.Infof("server recvmsg from client: %s", string(cbuf))

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
    //logrus.Infof("client recvmsg from server: %s", string(cbuf2))

    // close socket
    unix.Close(clientServerFd)
    unix.Close(clientFd)
    unix.Close(serverFd)

    time.Sleep(time.Second * 3)
    var result bpfTcpbpfGlobals
    for i := 0; i < 10; i++ {
        err = objs.bpfMaps.GlobalMap.Lookup(uint32(0), &result)
        if err != nil {
            logrus.Fatal(err)
        }
        if result.NumCloseEvents != ExpectedCloseEvents {
            logrus.Errorf("Unexpected number of close events (%d), retrying!", result.NumCloseEvents)
            time.Sleep(time.Millisecond * 100)
            continue
        }

        break
    }
    expectedEvents := (1 << BPF_SOCK_OPS_TIMEOUT_INIT) |
        (1 << BPF_SOCK_OPS_RWND_INIT) |
        (1 << BPF_SOCK_OPS_TCP_CONNECT_CB) |
        (1 << BPF_SOCK_OPS_ACTIVE_ESTABLISHED_CB) |
        (1 << BPF_SOCK_OPS_PASSIVE_ESTABLISHED_CB) |
        (1 << BPF_SOCK_OPS_NEEDS_ECN) |
        (1 << BPF_SOCK_OPS_STATE_CB) |
        (1 << BPF_SOCK_OPS_TCP_LISTEN_CB)
    expected2 := bpfTcpbpfGlobals{
        EventMap:       uint32(expectedEvents), // 0x0c7e
        TotalRetrans:   0,
        DataSegsIn:     1,
        DataSegsOut:    1,
        BadCbTestRv:    0x80,
        GoodCbTestRv:   0,
        BytesReceived:  1001,
        BytesAcked:     1002,
        NumListen:      1,
        NumCloseEvents: ExpectedCloseEvents,
    }
    assert.Equal(test, expected2, result)

    /* check setsockopt for SAVE_SYN */
    var val int
    // val 是 int, 必须 unsafe.Pointer(), 否则报错: "can't marshal key: binary.Write: invalid type int"
    err = objs.bpfMaps.SockoptResults.Lookup(uint32(0), unsafe.Pointer(&val))
    if err != nil {
        logrus.Fatal(err)
    }
    expected := 0
    if val != expected {
        logrus.Errorf("expected %d, actual %d", expected, val)
    } else {
        logrus.Infof("expected value %d success", expected)
    }
    /* check getsockopt for SAVED_SYN */
    err = objs.bpfMaps.SockoptResults.Lookup(uint32(1), unsafe.Pointer(&val))
    if err != nil {
        logrus.Fatal(err)
    }
    expected = 1
    if val != expected {
        logrus.Errorf("expected %d, actual %d", expected, val)
    } else {
        logrus.Infof("expected value %d success", expected)
    }

    // Wait
    <-stopCh
}

func makeBytes() string {
    bytes := ""
    for i := 0; i < 1000; i++ {
        bytes += "+"
    }
    return bytes
}

// tcp listen 127.0.0.1:8080
// /root/linux-5.10.142/tools/testing/selftests/bpf/tcp_server.py
func makeServer() int {
    serverFd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
    if err != nil {
        logrus.Fatal(err)
    }
    setSocketTimeout(serverFd, 5000)

    err = unix.SetsockoptInt(serverFd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
    if err != nil {
        logrus.Fatalf("unix.SO_REUSEADDR error: %v", err)
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

// 把当前进程 pid 写到新建的 connect_force_port cgroup
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
    if err := os.WriteFile(file, []byte(fmt.Sprintf("%d\n", pid)), os.ModeAppend); err != nil {
        logrus.Fatalf("os.WriteFile err: %v", err)
    }

    return cgroupPath
}

func cleanupCgroup() {
    os.RemoveAll(CgroupPath)
}
