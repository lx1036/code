package main

import (
    "fmt"
    "github.com/cilium/ebpf"
    "github.com/cilium/ebpf/link"
    "github.com/sirupsen/logrus"
    "golang.org/x/sys/unix"
    "net"
    "os"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go bpf connect_force_port4.c -- -I.

const (
    CgroupPath = "/sys/fs/cgroup/connect_force_port"

    Loopback = "127.0.0.1"

    PinPath = "/sys/fs/bpf/socket_service"
)

// go generate .
// CGO_ENABLED=0 go run .
func main() {
    logrus.SetReportCaller(true)

    joinCgroup()
    //defer cleanupCgroup()

    // tcp listen 127.0.0.1:60123, 作为 backend
    serverFd := makeServer(unix.SOCK_STREAM, 60123)
    defer unix.Close(serverFd)

    //1.Load pre-compiled programs and maps into the kernel.
    objs := bpfObjects{}
    opts := &ebpf.CollectionOptions{
        Programs: ebpf.ProgramOptions{
            LogLevel: ebpf.LogLevelInstruction,
            LogSize:  64 * 1024 * 1024, // 64M
        },
        Maps: ebpf.MapOptions{
            PinPath: PinPath, // pin 下 map
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
        Path:    CgroupPath,
        Program: objs.bpfPrograms.Connect4,
        Attach:  ebpf.AttachCGroupInet4Connect,
    })
    if err != nil {
        logrus.Fatal(err)
    }
    defer l1.Close()

    l2, err := link.AttachCgroup(link.CgroupOptions{
        Path:    CgroupPath,
        Program: objs.bpfPrograms.Getpeername4,
        Attach:  ebpf.AttachCgroupInet4GetPeername,
    })
    if err != nil {
        logrus.Fatal(err)
    }
    defer l2.Close()

    l3, err := link.AttachCgroup(link.CgroupOptions{
        Path:    CgroupPath,
        Program: objs.bpfPrograms.Getsockname4,
        Attach:  ebpf.AttachCgroupInet4GetSockname,
    })
    if err != nil {
        logrus.Fatal(err)
    }
    defer l3.Close()

    clientFd := connectToFd(serverFd)
    defer unix.Close(clientFd)

    verifyPorts(clientFd, 22222, 60000)
}

func verifyPorts(clientFd, expectedLocalPort, expectedPeerPort int) {
    logrus.Infof("check local/peer port")

    serverSockAddr, err := unix.Getsockname(clientFd)
    if err != nil {
        logrus.Fatal(err)
    }
    localPort := serverSockAddr.(*unix.SockaddrInet4).Port
    if localPort != expectedLocalPort {
        logrus.Errorf("Unexpected local port %d, expected %d", localPort, expectedLocalPort)
    }

    peerSockAddr, err := unix.Getpeername(clientFd)
    if err != nil {
        logrus.Fatal(err)
    }
    peerPort := peerSockAddr.(*unix.SockaddrInet4).Port
    if peerPort != expectedPeerPort {
        logrus.Errorf("Unexpected peer port %d, expected %d", peerPort, expectedPeerPort)
    }
}

func makeServer(socketType, port int) int {
    serverFd, err := unix.Socket(unix.AF_INET, socketType, 0)
    if err != nil {
        logrus.Errorf("unix.Socket error: %v", err)
    }
    setSocketTimeout(serverFd, 5000)

    err = unix.SetsockoptInt(serverFd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
    if err != nil {
        logrus.Fatalf("unix.SO_REUSEADDR error: %v", err)
    }

    ip := net.ParseIP(Loopback)
    sa := &unix.SockaddrInet4{
        Port: port,
        Addr: [4]byte{},
    }
    copy(sa.Addr[:], ip)
    err = unix.Bind(serverFd, sa)

    if socketType == unix.SOCK_STREAM {
        err = unix.Listen(serverFd, 1)
        if err != nil {
            logrus.Errorf("unix.Listen error: %v", err)
        }
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

func connectToFd(serverFd int) int {
    socketType, err := unix.GetsockoptInt(serverFd, unix.SOL_SOCKET, unix.SO_TYPE)
    if err != nil {
        logrus.Fatal(err)
    }

    // INFO: 注意这里 bpf hook 返回的是 :60123 -> 1.2.3.4:60000
    serverSockAddr, err := unix.Getsockname(serverFd)
    if err != nil {
        logrus.Fatal(err)
    }

    clientFd, err := unix.Socket(unix.AF_INET, socketType, 0) // unix.SOCK_STREAM
    if err != nil {
        logrus.Fatal(err)
    }
    setSocketTimeout(clientFd, 5000)

    // 非阻塞的
    err = unix.Connect(clientFd, serverSockAddr) // client_fd -> 1.2.3.4:60000
    if err != nil {
        logrus.Fatal(err)
    }

    return clientFd
}

// 把当前进程 pid 写到新建的 connect_force_port cgroup
func joinCgroup() {
    if err := os.Mkdir(CgroupPath, 0777); err != nil {
        logrus.Fatalf("os.Mkdir err: %v", err)
        return
    }
    pid := os.Getpid()
    file := fmt.Sprintf("%s/cgroup.procs", CgroupPath)
    if err := os.WriteFile(file, []byte(fmt.Sprintf("%d\n", pid)), 0644); err != nil {
        logrus.Fatalf("os.WriteFile err: %v", err)
        return
    }
}

func cleanupCgroup() {
    os.RemoveAll(CgroupPath)
}
