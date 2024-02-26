package main

import (
    "fmt"
    "github.com/cilium/ebpf"
    "github.com/cilium/ebpf/link"
    "github.com/sirupsen/logrus"
    "golang.org/x/sys/unix"
    "net"
    "os"
    "os/signal"
    "syscall"
    "testing"
    "unsafe"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go bpf socket_lookup_cgroup.c -- -I.

// go generate .

const (
    CgroupPath = "/sys/fs/cgroup/socket-lookup-cgroup"

    PinPath2 = "/sys/fs/bpf/socket-lookup-cgroup/ingress-lookup"

    InaddrLoopback = "127.0.0.1"

    PARENT_CGROUP = "test-bpf-sock-fields"
    CHILD_CGROUP  = "test-bpf-sock-fields/child"

    DATA = "Hello BPF!"
)

// CGO_ENABLED=0 go test -v -run ^TestIngressLookup$ .
func TestIngressLookup(test *testing.T) {
    logrus.SetReportCaller(true)

    stopCh := make(chan os.Signal, 1)
    signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

    /** Create a socket before joining testing cgroup so that its cgroup id
     * differs from that of testing cgroup. Moving selftests process to
     * testing cgroup won't change cgroup id of an already created socket.
     */
    outClientFd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
    if err != nil {
        logrus.Errorf("unix.Socket err: %v", err)
        return
    }

    // `cat /sys/fs/cgroup/sockopt/sockopt_inherit/cgroup.procs`
    cgroupPath := joinCgroup("ingress-lookup")
    //defer cleanupCgroup()

    //1.Load pre-compiled programs and maps into the kernel.
    objs := bpfObjects{}
    opts := &ebpf.CollectionOptions{
        Programs: ebpf.ProgramOptions{
            LogLevel: ebpf.LogLevelInstruction,
            LogSize:  64 * 1024 * 1024, // 64M
        },
        Maps: ebpf.MapOptions{
            PinPath: PinPath2, // pin 下 map
        },
    }
    err = loadBpfObjects(&objs, opts)
    if err != nil {
        logrus.Errorf("loadBpfObjects err: %v", err)
        return
    }
    defer objs.Close()

    //2.Attach programs to cgroup
    l1, err := link.AttachCgroup(link.CgroupOptions{
        Path:    cgroupPath,
        Program: objs.bpfPrograms.IngressLookup,
        Attach:  ebpf.AttachCGroupInetIngress,
    })
    if err != nil {
        logrus.Fatal(err)
    }
    defer l1.Close()

    serverFd := makeServer()
    //defer unix.Close(serverFd)

    /* Client outside of test cgroup should fail to connect by timeout. */
    connectFdToFd(outClientFd, serverFd)

    /* Client inside test cgroup should connect just fine. */
    clientFd := connectToFd(serverFd)
    //defer unix.Close(clientFd)

    // connect 已经建立 tcp connection，从全队列(accept)里取出一个 connection socket
    clientServerFd, _, err := unix.Accept(serverFd)
    if err != nil {
        logrus.Fatalf("Accept err: %v", err)
    }

    unix.Close(clientServerFd)
    unix.Close(clientFd)
    unix.Close(serverFd)

    <-stopCh
}

// CGO_ENABLED=0 go test -v -run ^TestLoadBytesRelative$ .
func TestLoadBytesRelative(test *testing.T) {
    logrus.SetReportCaller(true)

    stopCh := make(chan os.Signal, 1)
    signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

    cgroupPath := joinCgroup("load_bytes_relative")
    //defer cleanupCgroup()

    serverFd := makeServer()
    //defer unix.Close(serverFd)

    objs := bpfObjects{}
    opts := &ebpf.CollectionOptions{
        Programs: ebpf.ProgramOptions{
            LogLevel: ebpf.LogLevelInstruction,
            LogSize:  64 * 1024 * 1024, // 64M
        },
        Maps: ebpf.MapOptions{
            PinPath: PinPath2, // pin 下 map
        },
    }
    err := loadBpfObjects(&objs, opts)
    if err != nil {
        logrus.Errorf("loadBpfObjects err: %v", err)
        return
    }
    defer objs.Close()

    //2.Attach programs to cgroup
    l1, err := link.AttachCgroup(link.CgroupOptions{
        Path:    cgroupPath,
        Program: objs.bpfPrograms.LoadBytesRelative,
        Attach:  ebpf.AttachCGroupInetEgress,
    })
    if err != nil {
        logrus.Fatal(err)
    }
    defer l1.Close()

    clientFd := connectToFd(serverFd)
    defer unix.Close(clientFd)

    var val uint32
    err = objs.bpfMaps.TestResult.Lookup(uint32(0), &val)
    if err != nil {
        logrus.Fatal(err)
    }
    if val != uint32(1) {
        logrus.Errorf("expected 1, actual %d", val)
    } else {
        logrus.Infof("expected 1")
    }

    <-stopCh
}

// CGO_ENABLED=0 go test -v -run ^TestSockFields$ .
func TestSockFields(test *testing.T) {

    parentCgroupPath := joinCgroup(PARENT_CGROUP)
    childCgroupPath := joinCgroup(CHILD_CGROUP)
    //defer cleanupCgroup()

    objs := bpfObjects{}
    opts := &ebpf.CollectionOptions{
        Programs: ebpf.ProgramOptions{
            LogLevel: ebpf.LogLevelInstruction,
            LogSize:  64 * 1024 * 1024, // 64M
        },
        Maps: ebpf.MapOptions{
            PinPath: PinPath2, // pin 下 map
        },
    }
    err := loadBpfObjects(&objs, opts)
    if err != nil {
        logrus.Errorf("loadBpfObjects err: %v", err)
        return
    }
    defer objs.Close()

    //2.Attach programs to cgroup
    l1, err := link.AttachCgroup(link.CgroupOptions{
        Path:    childCgroupPath,
        Program: objs.bpfPrograms.EgressReadSockFields,
        Attach:  ebpf.AttachCGroupInetEgress,
    })
    if err != nil {
        logrus.Fatal(err)
    }
    defer l1.Close()

    l2, err := link.AttachCgroup(link.CgroupOptions{
        Path:    childCgroupPath,
        Program: objs.bpfPrograms.IngressReadSockFields,
        Attach:  ebpf.AttachCGroupInetIngress,
    })
    if err != nil {
        logrus.Fatal(err)
    }
    defer l2.Close()

    l3, err := link.AttachCgroup(link.CgroupOptions{
        Path:    childCgroupPath,
        Program: objs.bpfPrograms.ReadSkDstPort,
        Attach:  ebpf.AttachCGroupInetIngress,
    })
    if err != nil {
        logrus.Fatal(err)
    }
    defer l3.Close()

    serverFd := makeServer()
    //defer unix.Close(serverFd)
    serverSockAddr, err := unix.Getsockname(serverFd)
    if err != nil {
        logrus.Fatal(err)
    }

    // TODO: rewrite consts

    clientFd := connectToFd(serverFd)
    clientSockAddr, err := unix.Getsockname(clientFd)
    if err != nil {
        logrus.Fatal(err)
    }

    // connect 已经建立 tcp connection，从全队列(accept)里取出一个 connection socket
    clientServerFd, _, err := unix.Accept(serverFd)
    if err != nil {
        logrus.Fatalf("Accept err: %v", err)
    }
    val := bpfBpfSpinlockCnt{
        Cnt: 0xeB9F,
    }
    err = objs.bpfMaps.SkPktOutCnt.Put(clientServerFd, val)
    if err != nil {
        logrus.Fatal(err)
    }
    err = objs.bpfMaps.SkPktOutCnt10.Put(clientServerFd, val)
    if err != nil {
        logrus.Fatal(err)
    }

    for i := 0; i < 2; i++ {
        /* Send some data from accept_fd to cli_fd server->client
         * MSG_EOR to stop kernel from coalescing two pkts.
         */
        err = unix.Send(clientServerFd, []byte(DATA), unix.MSG_EOR)
        cbuf := make([]byte, len(DATA))
        n, _, _, _, err := unix.Recvmsg(clientFd, cbuf, nil, 0)
        if err != nil {
            logrus.Errorf("unix.Recvmsg clientServerFd err: %v", err)
            continue
        }
        cbuf = cbuf[:n]
        logrus.Infof("client recvmsg from server: %s", string(cbuf))
    }

    unix.Shutdown(clientFd, unix.SHUT_WR) // clientFd
    cbuf := make([]byte, 1)
    n, _, _, _, err := unix.Recvmsg(clientServerFd, cbuf, nil, 0)
    if err != nil {
        logrus.Fatal(err)
    }
    cbuf = cbuf[:n]
    logrus.Infof("server recvmsg from client for Fin: %s", string(cbuf))
    unix.Shutdown(clientServerFd, unix.SHUT_WR)
    n, _, _, _, err = unix.Recvmsg(clientFd, cbuf, nil, 0)
    if err != nil {
        logrus.Fatal(err)
    }
    cbuf = cbuf[:n]
    logrus.Infof("client recvmsg from server for Fin: %s", string(cbuf))

    val1, val2 := bpfBpfSpinlockCnt{}, bpfBpfSpinlockCnt{}
    val1.Cnt = 0
    val2.Cnt = 0
    err = objs.bpfMaps.SkPktOutCnt.Lookup(unsafe.Pointer(&clientServerFd), &val1)
    if err != nil {
        err = objs.bpfMaps.SkPktOutCnt10.Lookup(unsafe.Pointer(&clientServerFd), &val2)
        if err != nil {
            logrus.Fatal(err)
        }
    }

    /* The bpf prog only counts for fullsock and
     * passive connection did not become fullsock until 3WHS
     * had been finished, so the bpf prog only counted two data
     * packet out.
     */
    if val1.Cnt < 0xeB9F+2 || val2.Cnt < 0xeB9F+20 {
        logrus.Errorf("bpf_map_lookup_elem(sk_pkt_out_cnt, &accept_fd), pkt_out_cnt:%d pkt_out_cnt10:%d", val1.Cnt, val2.Cnt)
    }

    val1.Cnt = 0
    val2.Cnt = 0
    err = objs.bpfMaps.SkPktOutCnt.Lookup(unsafe.Pointer(&clientFd), &val1)
    if err != nil {
        logrus.Fatal(err)
    }
    /* Active connection is fullsock from the beginning.
     * 1 SYN and 1 ACK during 3WHS
     * 2 Acks on data packet.
     *
     * The bpf_prog initialized it to 0xeB9F.
     */
    if val1.Cnt < 0xeB9F+4 || val2.Cnt < 0xeB9F+40 {
        logrus.Errorf("bpf_map_lookup_elem(sk_pkt_out_cnt, &accept_fd), pkt_out_cnt:%d pkt_out_cnt10:%d", val1.Cnt, val2.Cnt)
    }

    checkResult()
}

func checkResult() {

}

func connectFdToFd(outClientFd, serverFd int) {
    setSocketTimeout(outClientFd, 5000)

    serverSockAddr, err := unix.Getsockname(serverFd)
    if err != nil {
        logrus.Fatal(err)
    }

    // 非阻塞的
    err = unix.Connect(outClientFd, serverSockAddr)
    if err != nil {
        logrus.Errorf("%+v", err)
    }

    logrus.Infof("clientFd outside of test cgroup should fail to connect by timeout")
}

func connectToFd(serverFd int) int {
    clientFd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0) // unix.SOCK_STREAM
    if err != nil {
        logrus.Fatal(err)
    }
    setSocketTimeout(clientFd, 5000)

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

// tcp listen 127.0.0.1:60123
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
        Port: 60123,
        Addr: [4]byte{},
    }
    copy(sa.Addr[:], ip)
    err = unix.Bind(serverFd, sa)
    if err != nil {
        logrus.Fatal(err)
    }

    err = unix.Listen(serverFd, 1)
    if err != nil {
        logrus.Fatal(err)
    }

    return serverFd
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
