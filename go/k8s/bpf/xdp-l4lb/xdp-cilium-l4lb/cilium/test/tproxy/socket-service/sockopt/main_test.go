package main

import (
    "fmt"
    "github.com/cilium/ebpf"
    "github.com/cilium/ebpf/link"
    "github.com/sirupsen/logrus"
    "golang.org/x/sys/unix"
    "net"
    "os"
    "strconv"
    "testing"
)

const (
    CgroupPath = "/sys/fs/cgroup/sockopt"

    PinPath1 = "/sys/fs/bpf/socket_service/sockopt"
    PinPath2 = "/sys/fs/bpf/socket_service/sock_inherit"

    SOL_CUSTOM      = 0xdeadbeef
    CUSTOM_INHERIT1 = 0
    CUSTOM_INHERIT2 = 1
    CUSTOM_LISTENER = 2

    INADDR_LOOPBACK = "127.0.0.1"
)

func TestSockOptInherit(test *testing.T) {
    cgroupPath := joinCgroup("sockopt_inherit")
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
    err := loadBpfObjects(&objs, opts)
    if err != nil {
        logrus.Errorf("loadBpfObjects err: %v", err)
        return
    }
    defer objs.Close()

    //2.Attach programs to cgroup
    l1, err := link.AttachCgroup(link.CgroupOptions{
        Path:    cgroupPath,
        Program: objs.bpfPrograms.Setsockopt2,
        Attach:  ebpf.AttachCGroupSetsockopt,
    })
    if err != nil {
        logrus.Fatal(err)
    }
    defer l1.Close()

    l2, err := link.AttachCgroup(link.CgroupOptions{
        Path:    cgroupPath,
        Program: objs.bpfPrograms.Getsockopt2,
        Attach:  ebpf.AttachCGroupGetsockopt,
    })
    if err != nil {
        logrus.Fatal(err)
    }
    defer l2.Close()

    serverFd := makeServer()
    defer unix.Close(serverFd)
    err = unix.Listen(serverFd, 1)
    verifySockOpt(serverFd, CUSTOM_INHERIT1, 0x01, "listen")
    verifySockOpt(serverFd, CUSTOM_INHERIT2, 0x01, "listen")
    verifySockOpt(serverFd, CUSTOM_LISTENER, 0x01, "listen")

    go func(serverFd int) {
        clientServerFd, _, err := unix.Accept(serverFd)
        if err != nil {
            logrus.Errorf("Accept err: %v", err)
            return
        }
        defer unix.Close(clientServerFd)
        verifySockOpt(clientServerFd, CUSTOM_INHERIT1, 0x01, "accept")
        verifySockOpt(clientServerFd, CUSTOM_INHERIT2, 0x01, "accept")
        verifySockOpt(clientServerFd, CUSTOM_LISTENER, 0x01, "accept")
    }(serverFd)

    clientFd := connectToFd(serverFd)
    defer unix.Close(clientFd)
    verifySockOpt(clientFd, CUSTOM_INHERIT1, 0, "connect")
    verifySockOpt(clientFd, CUSTOM_INHERIT2, 0, "connect")
    verifySockOpt(clientFd, CUSTOM_LISTENER, 0, "connect")
}

func verifySockOpt(serverFd, optName, expected int, msg string) {
    value, err := unix.GetsockoptInt(serverFd, SOL_CUSTOM, optName)
    if err != nil {
        logrus.Fatal(err)
    }

    if value != expected {
        logrus.Errorf("%s: unexpected getsockopt value %d != %d", msg, value, expected)
    }
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

func TestSockOpt(test *testing.T) {
    cgroupPath := joinCgroup("sockopt_sk")
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

    //2.Attach programs to cgroup
    l1, err := link.AttachCgroup(link.CgroupOptions{
        Path:    cgroupPath,
        Program: objs.bpfPrograms.Setsockopt1,
        Attach:  ebpf.AttachCGroupSetsockopt,
    })
    if err != nil {
        logrus.Fatal(err)
    }
    defer l1.Close()

    l2, err := link.AttachCgroup(link.CgroupOptions{
        Path:    cgroupPath,
        Program: objs.bpfPrograms.Getsockopt1,
        Attach:  ebpf.AttachCGroupGetsockopt,
    })
    if err != nil {
        logrus.Fatal(err)
    }
    defer l2.Close()

    GetSetSockOpt()

}

func GetSetSockOpt() {

    serverFd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)

    /* IP_TOS - BPF bypass */
    err = unix.SetsockoptInt(serverFd, unix.SOL_IP, unix.IP_TOS, 0x08)
    if err != nil {

    }
    value, err := unix.GetsockoptInt(serverFd, unix.SOL_IP, unix.IP_TOS)
    if err != nil {

    }
    if value != 0x08 {
        logrus.Errorf("Unexpected getsockopt(IP_TOS) optval 0x%x != 0x08", value)
    }

    /* IP_TTL - EPERM */
    err = unix.SetsockoptInt(serverFd, unix.SOL_IP, unix.IP_TTL, 0x01)
    if err == nil {
        logrus.Errorf("Unexpected success from setsockopt(IP_TTL)")
    } else {
        logrus.Infof("unix.IP_TTL expected err:%v", err)
    }

    /* SOL_CUSTOM - handled by BPF */
    err = unix.SetsockoptInt(serverFd, SOL_CUSTOM, 0, 0x01)
    if err != nil {

    }
    value2, err := unix.GetsockoptInt(serverFd, SOL_CUSTOM, 0)
    if value2 != 0x01 {

    }

    /* IP_FREEBIND - BPF can't access optval past PAGE_SIZE */
    pagesize := unix.Getpagesize()
    bigBuf := make([]byte, pagesize*2)
    err = unix.SetsockoptString(serverFd, unix.SOL_IP, unix.IP_FREEBIND, string(bigBuf))

    value3, err := unix.GetsockoptString(serverFd, unix.SOL_IP, unix.IP_FREEBIND)
    if value3 != strconv.Itoa(0x55) { // "85"

    }

    /* SO_SNDBUF is overwritten */
    err = unix.SetsockoptInt(serverFd, unix.SOL_SOCKET, unix.SO_SNDBUF, 0x01010101)
    if err != nil {
    }
    value4, err := unix.GetsockoptInt(serverFd, unix.SOL_SOCKET, unix.SO_SNDBUF)
    if value4 != 0x55AA*2 {

    }

    /* TCP_CONGESTION can extend the string */
    err = unix.SetsockoptString(serverFd, unix.IPPROTO_TCP, unix.TCP_CONGESTION, "nv")
    if err != nil {
    }
    value5, err := unix.GetsockoptString(serverFd, unix.IPPROTO_TCP, unix.TCP_CONGESTION)
    if value5 != "cubic" {
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
    if err := os.WriteFile(file, []byte(fmt.Sprintf("%d\n", pid)), 0644); err != nil {
        logrus.Fatalf("os.WriteFile err: %v", err)
    }

    return cgroupPath
}

func cleanupCgroup() {
    os.RemoveAll(CgroupPath)
}

func makeServer() int {
    serverFd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
    if err != nil {
        logrus.Fatal(err)
    }

    err = unix.SetsockoptInt(serverFd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
    if err != nil {
        logrus.Fatalf("unix.SO_REUSEADDR error: %v", err)
    }

    // 赋值 SOL_CUSTOM
    for i := CUSTOM_INHERIT1; i <= CUSTOM_LISTENER; i++ {
        buf := 0x01
        err = unix.SetsockoptInt(serverFd, SOL_CUSTOM, i, buf)
        if err != nil {
            logrus.Fatal(err)
        }
    }

    ip := net.ParseIP(INADDR_LOOPBACK)
    sa := &unix.SockaddrInet4{
        //Port: 5432,
        Addr: [4]byte{},
    }
    copy(sa.Addr[:], ip)
    err = unix.Bind(serverFd, sa)
    if err != nil {
        logrus.Fatal(err)
    }

    return serverFd
}
