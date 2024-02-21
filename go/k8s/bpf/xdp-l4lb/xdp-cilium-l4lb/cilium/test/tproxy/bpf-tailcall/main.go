package main

import (
    "github.com/cilium/ebpf"
    "github.com/cilium/ebpf/rlimit"
    "github.com/sirupsen/logrus"
    "golang.org/x/sys/unix"
    tc_program_attach "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/test/tc-program-attach"
    "net"
    "os"
    "os/signal"
    "syscall"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go bpf tailcall1.c -- -I.

const (
    PinPath = "/sys/fs/bpf/tail_call"

    BindPort = 1234
)

func init() {
    // Allow the current process to lock memory for eBPF resources.
    if err := rlimit.RemoveMemlock(); err != nil {
        logrus.Fatal(err)
    }
}

// go generate .
// CGO_ENABLED=0 go run .
func main() {
    logrus.SetReportCaller(true)

    stopCh := make(chan os.Signal, 1)
    signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

    serverFd := startServer()
    //defer unix.Close(serverFd)

    // 1.Load pre-compiled programs and maps into the kernel.
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

    // 2.put maps
    //for i := 0; i < int(objs.bpfMaps.JmpTable.MaxEntries()); i++ {}
    err = objs.bpfMaps.JmpTable.Put(uint32(0), uint32(objs.bpfPrograms.BpfFunc0.FD()))
    if err != nil {
        logrus.Errorf("JmpTable Put err: %v", err)
    }
    err = objs.bpfMaps.JmpTable.Put(uint32(1), uint32(objs.bpfPrograms.BpfFunc1.FD()))
    if err != nil {
        logrus.Errorf("JmpTable Put err: %v", err)
    }
    err = objs.bpfMaps.JmpTable.Put(uint32(2), uint32(objs.bpfPrograms.BpfFunc2.FD()))
    if err != nil {
        logrus.Errorf("JmpTable Put err: %v", err)
    }

    // 3.attach tc ingress to lo
    info, err := objs.bpfPrograms.Entry.Info()
    if err != nil {
        logrus.Errorf("program Entry Info err: %v", err)
        return
    }
    attachParams := &tc_program_attach.TcAttachParams{
        Interface:      "lo",
        ProgramName:    info.Name, // entry
        ProgramFd:      objs.bpfPrograms.Entry.FD(),
        Direction:      tc_program_attach.TcDirectionIngress,
        DirectAction:   true,
        ClobberIngress: true,
    }
    program := tc_program_attach.NewTcSchedClsProgram()
    err = program.Attach(attachParams)
    defer program.Detach()

    // 4.check

    //`tc filter show dev lo ingress`
    //`tc qdisc del dev lo clsact`
    //filter protocol all pref 1 bpf chain 0
    //filter protocol all pref 1 bpf chain 0 handle 0x1600 entry direct-action not_in_hw id 263 tag 30954aa0a540a213 jited

    clientFd := connectToFd()
    //defer unix.Close(clientFd)

    clientServerFd, _, err := unix.Accept(serverFd)
    if err != nil {
        logrus.Fatalf("Accept err: %v", err)
    }
    defer unix.Close(clientServerFd)
    if _, err = unix.Write(clientFd, []byte("testing")); err != nil {
        logrus.Fatalf("unix.Write err: %v", err)
    }
    cbuf := make([]byte, unix.CmsgSpace(4))
    n, _, _, _, err := unix.Recvmsg(clientServerFd, cbuf, nil, 0)
    if err != nil {
        logrus.Fatalf("unix.Recvmsg err: %v", err)
    }
    cbuf = cbuf[:n]
    logrus.Infof("%s", string(cbuf)) // testing

    <-stopCh

    // client 先发 fin 包给 server 来关闭，否则 defer，server 会先发 fin 包给 client
    unix.Close(clientFd)
    unix.Close(serverFd)
}

func startServer() int {
    var err error
    var fd int
    defer func() {
        if err != nil && fd > 0 {
            unix.Close(fd)
        }
    }()

    fd, err = unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
    if err != nil {
        logrus.Fatal(err)
    }

    setSocketTimeout(fd, 5000)

    // INFO: 解决报错 "address already in use", 因为 client/server tcp connection 没有正常关闭，会导致 server tcp 状态机进入 TIME_WAIT
    //  状态，需要等待 2 * Maximum Segment Lifetime=4min 时间后才能释放，测试也不是4min，大概几十秒
    //  `netstat | grep 1234`
    //  tcp        0      0 localhost:postgresql    localhost:1234          TIME_WAIT
    // ignores TIME-WAIT state using SO_REUSEADDR option
    // https://serverfault.com/questions/329845/how-to-forcibly-close-a-socket-in-time-wait
    err = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
    if err != nil {
        logrus.Fatalf("unix.SO_REUSEADDR error: %v", err)
    }

    ip := net.ParseIP("127.0.0.1")
    sa := &unix.SockaddrInet4{
        Port: BindPort,
        Addr: [4]byte{},
    }
    copy(sa.Addr[:], ip)
    err = unix.Bind(fd, sa)
    if err != nil {
        logrus.Fatal(err)
    }

    err = unix.Listen(fd, 1)
    if err != nil {
        logrus.Fatal(err)
    }

    return fd
}

func connectToFd() int {
    var err error
    var clientFd int
    defer func() {
        if err != nil && clientFd > 0 {
            unix.Close(clientFd)
        }
    }()

    clientFd, err = unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
    if err != nil {
        logrus.Fatal(err)
    }
    setSocketTimeout(clientFd, 5000)

    // INFO: 这里客户端指定 ip:port
    ip1 := net.ParseIP("127.0.0.1")
    sa1 := &unix.SockaddrInet4{
        Port: 5432,
        Addr: [4]byte{},
    }
    copy(sa1.Addr[:], ip1)
    err = unix.Bind(clientFd, sa1)
    if err != nil {
        logrus.Fatal(err)
    }

    ip := net.ParseIP("127.0.0.1")
    sa := &unix.SockaddrInet4{
        Port: BindPort,
        Addr: [4]byte{},
    }
    copy(sa.Addr[:], ip)
    // 非阻塞的, TCP 建联
    err = unix.Connect(clientFd, sa)
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
