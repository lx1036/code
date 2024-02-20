//go:build linux

package main

import (
    "github.com/cilium/ebpf"
    "github.com/cilium/ebpf/link"
    "github.com/cilium/ebpf/rlimit"
    "github.com/sirupsen/logrus"
    "golang.org/x/sys/unix"
    "log"
    "net"
    "os"
    "os/signal"
    "path/filepath"
    "syscall"
    "time"
)

// /root/linux-5.10.142/tools/testing/selftests/bpf/prog_tests/tcp_rtt.c

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -tags "linux" bpf tcprtt_sockops.c -- -I.

// go generate .
func main() {
    stopCh := make(chan os.Signal, 1)
    signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

    // Allow the current process to lock memory for eBPF resources.
    if err := rlimit.RemoveMemlock(); err != nil {
        log.Fatal(err)
    }

    serverFd := startServer()
    defer unix.Close(serverFd)

    // Find the path to a cgroup enabled to version 2
    cgroupPath, err := findCgroupPath()
    if err != nil {
        log.Fatal(err)
    }

    // Load pre-compiled programs and maps into the kernel.
    objs := bpfObjects{}
    if err := loadBpfObjects(&objs, nil); err != nil {
        log.Fatalf("loading objects: %v", err)
    }
    defer objs.Close()

    // Attach ebpf program to a cgroupv2
    l, err := link.AttachCgroup(link.CgroupOptions{
        Path:    cgroupPath,
        Program: objs.bpfPrograms.BpfSockopsCb,
        Attach:  ebpf.AttachCGroupSockOps,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer l.Close()

    clientFd := connectToFd(serverFd)
    defer unix.Close(clientFd)

    verifySk(objs, clientFd, "syn-ack", bpfTcpRttStorage{
        Invoked:         1,
        DsackDups:       0,
        Delivered:       1,
        DeliveredCe:     0,
        IcskRetransmits: 0,
    })

    sendBytes(clientFd)
    waitForAck(clientFd, 100)

    verifySk(objs, clientFd, "first payload byte", bpfTcpRttStorage{
        Invoked:         2,
        DsackDups:       0,
        Delivered:       2,
        DeliveredCe:     0,
        IcskRetransmits: 0,
    })
}

func findCgroupPath() (string, error) {
    cgroupPath := "/sys/fs/cgroup"

    var st syscall.Statfs_t
    err := syscall.Statfs(cgroupPath, &st)
    if err != nil {
        return "", err
    }
    isCgroupV2Enabled := st.Type == unix.CGROUP2_SUPER_MAGIC
    if !isCgroupV2Enabled {
        cgroupPath = filepath.Join(cgroupPath, "unified")
    }
    return cgroupPath, nil
}

func startServer() int {
    fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
    if err != nil {
        log.Fatal(err)
    }

    setSocketTimeout(fd, 0)

    ip := net.ParseIP("127.0.0.1")
    sa := &unix.SockaddrInet4{
        Port: 8000,
        Addr: [4]byte{},
    }
    copy(sa.Addr[:], ip)
    err = unix.Bind(fd, sa)
    if err != nil {
        log.Fatal(err)
    }

    err = unix.Listen(fd, 1)
    if err != nil {
        log.Fatal(err)
    }

    return fd
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
        log.Fatal(err)
    }

    err = unix.SetsockoptTimeval(fd, unix.SOL_SOCKET, unix.SO_SNDTIMEO, timeVal)
    if err != nil {
        log.Fatal(err)
    }
}

func connectToFd(serverFd int) int {
    serverSockAddr, err := unix.Getsockname(serverFd)
    if err != nil {
        log.Fatal(err)
    }

    clientFd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
    if err != nil {
        log.Fatal(err)
    }
    setSocketTimeout(clientFd, 0)
    // 非阻塞的
    err = unix.Connect(clientFd, serverSockAddr)
    if err != nil {
        log.Fatal(err)
    }

    return clientFd
}

func sendBytes(clientFd int) {
    data := "hello"
    _, err := unix.Write(clientFd, []byte(data))
    if err != nil {
        log.Fatal(err)
    }
}

func waitForAck(clientFd, retries int) {
    for i := 0; i < retries; i++ {
        tcpInfo, err := unix.GetsockoptTCPInfo(clientFd, unix.SOL_TCP, unix.TCP_INFO)
        if err != nil {
            logrus.Errorf("GetsockoptTCPInfo: %v", err)
            continue
        }

        if tcpInfo.Unacked == 0 {
            return
        }

        time.Sleep(time.Millisecond * 10)
    }

    return
}

func verifySk(objs bpfObjects, clientFd int, msg string, expected bpfTcpRttStorage) {
    var storage bpfTcpRttStorage
    err := objs.bpfMaps.SocketStorageMap.Lookup(clientFd, &storage)
    if err != nil {
        log.Fatal(err)
    }

    if storage.Invoked != expected.Invoked {
        logrus.Errorf("%s: tcp rtt Invoked not expected: %d actual: %d", msg, expected.Invoked, storage.Invoked)
    }
    if storage.DsackDups != expected.DsackDups {
        logrus.Errorf("%s: tcp rtt DsackDups not expected: %d actual: %d", msg, expected.DsackDups, storage.DsackDups)
    }
    if storage.Delivered != expected.Delivered {
        logrus.Errorf("%s: tcp rtt Delivered not expected: %d actual: %d", msg, expected.Delivered, storage.Delivered)
    }
    if storage.DeliveredCe != expected.DeliveredCe {
        logrus.Errorf("%s: tcp rtt DeliveredCe not expected: %d actual: %d", msg, expected.DeliveredCe, storage.DeliveredCe)
    }
    if storage.IcskRetransmits != expected.IcskRetransmits {
        logrus.Errorf("%s: tcp rtt IcskRetransmits not expected: %d actual: %d", msg, expected.IcskRetransmits, storage.IcskRetransmits)
    }
}
