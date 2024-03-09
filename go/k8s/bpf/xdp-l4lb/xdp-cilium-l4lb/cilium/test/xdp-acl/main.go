package main

import (
    "encoding/binary"
    "flag"
    "github.com/cilium/ebpf"
    "github.com/cilium/ebpf/btf"
    "github.com/cilium/ebpf/link"
    "github.com/sirupsen/logrus"
    "golang.org/x/sys/unix"
    "net"
    "os"
    "os/signal"
    "syscall"
    "unsafe"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go bpf acl.c -- -I.

const (
    // 经过测试: 本地运行换成 eth0-acl 网卡和 ip 不行，还是走的 127.0.0.1
    INADDR_TEST = "127.0.0.1"
    BindPort    = 9090

    DATA = "testing"
)

// go generate .
// CGO_ENABLED=0 go run . --action=1
// CGO_ENABLED=0 go run . --action=0
func main() {
    logrus.SetReportCaller(true)

    actionArg := flag.Int("action", 1, "for xdp")
    iface := flag.String("iface", "lo", "the interface to attach this program to")
    flag.Parse()

    stopCh := make(chan os.Signal, 1)
    signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

    serverFd := startServer()
    defer unix.Close(serverFd)
    clientFd := connectToFd(serverFd)
    defer unix.Close(clientFd)
    // connect 已经建立 tcp connection，从全队列(accept)里取出一个 connection socket
    clientServerFd, _, err := unix.Accept(serverFd)
    if err != nil {
        logrus.Fatalf("Accept err: %v", err)
    }
    defer unix.Close(clientServerFd)
    if _, err = unix.Write(clientFd, []byte(DATA)); err != nil {
        logrus.Fatalf("unix.Write err: %v", err)
    }
    cbuf := make([]byte, 1024)
    n, _, _, _, err := unix.Recvmsg(clientServerFd, cbuf, nil, 0)
    if err != nil {
        logrus.Fatalf("unix.Recvmsg err: %v", err)
    }
    cbuf = cbuf[:n]
    logrus.Infof("server recvmsg from client: %s", string(cbuf))

    // Load pre-compiled programs and maps into the kernel.
    btfSpec, err := btf.LoadKernelSpec()
    if err != nil {
        logrus.Fatalf("LoadKernelSpec err:%v", err)
    }
    objs := bpfObjects{}
    opts := &ebpf.CollectionOptions{
        Programs: ebpf.ProgramOptions{
            LogLevel:    ebpf.LogLevelInstruction,
            LogSize:     64 * 1024 * 1024, // 64M
            KernelTypes: btfSpec,          // 注意 btf 概念
        },
    }
    spec, err := loadBpf()
    if err != nil {
        logrus.Fatal(err)
    }
    consts := map[string]interface{}{
        "XDPACL_DEBUG": uint32(1),
        //"XDPACL_BITMAP_ARRAY_SIZE_LIMIT": uint32(getBitmapArraySizeLimit(ruleNum)),
    }
    if err = spec.RewriteConstants(consts); err != nil {
        logrus.Fatal(err)
    }
    if err := spec.LoadAndAssign(&objs, opts); err != nil {
        logrus.Fatalf("loading objects: %v", err)
    }
    defer objs.Close()

    // bpf_tail_call_static
    if err := objs.Progs.Put(uint32(0), objs.bpfPrograms.XdpAclFuncImm); err != nil {
        logrus.Error(err)
    }

    var addr uint32
    if IsLittleEndian() {
        addr = binary.LittleEndian.Uint32(net.ParseIP(INADDR_TEST).To4()) // byte[]{a,b,c,d} -> dcba
    } else {
        addr = binary.BigEndian.Uint32(net.ParseIP(INADDR_TEST).To4()) // byte[]{a,b,c,d} -> abcd
    }
    // serverIPs := bpfServerIps{
    // 	TargetIps: [4]uint32{
    // 		addr,
    // 		// binary.BigEndian.Uint32(net.ParseIP(ip1).To4()),
    // 		// binary.LittleEndian.Uint32(net.ParseIP(ip1).To4()),
    // 	},
    // }
    serverIPs := bpfServerIps{}
    serverIPs.TargetIps[0] = addr
    if err := objs.bpfMaps.Servers.Put(uint32(0), serverIPs); err != nil {
        logrus.Error(err)
    }

    endpoint := bpfEndpoint{
        Protocol: unix.IPPROTO_TCP,
        Dport:    uint16(BindPort),
    }
    action := bpfAction{
        Action: uint8(*actionArg),
    }
    if err := objs.bpfMaps.Endpoints.Put(endpoint, action); err != nil {
        logrus.Error(err)
    }

    var action1 bpfAction
    if err := objs.bpfMaps.Endpoints.Lookup(&endpoint, &action1); err != nil {
        logrus.Error(err)
    }
    logrus.Infof("%+v", action1)

    ifaceObj, err := net.InterfaceByName(*iface)
    if err != nil {
        logrus.Fatalf("loading objects: %v", err)
    }
    l, err := link.AttachXDP(link.XDPOptions{
        Program:   objs.bpfPrograms.XdpAclFunc,
        Interface: ifaceObj.Index,
        Flags:     link.XDPGenericMode,
    })
    if err != nil {
        logrus.Fatal(err)
    }
    defer l.Close()

    if _, err = unix.Write(clientFd, []byte(DATA)); err != nil {
        logrus.Fatalf("unix.Write err: %v", err)
    }
    cbuf2 := make([]byte, 1024)
    n2, _, _, _, err := unix.Recvmsg(clientServerFd, cbuf2, nil, 0)
    if err != nil {
        logrus.Fatalf("unix.Recvmsg err: %v", err)
    }
    cbuf2 = cbuf2[:n2]
    logrus.Infof("server recvmsg from client: %s", string(cbuf2))

    // Wait
    <-stopCh
}

func connectToFd(serverFd int) int {
    socketType, err := unix.GetsockoptInt(serverFd, unix.SOL_SOCKET, unix.SO_TYPE)
    if err != nil {
        logrus.Fatal(err)
    }

    //tcpSaveSyn, err := unix.GetsockoptInt(serverFd, unix.SOL_TCP, unix.TCP_SAVE_SYN)

    serverSockAddr, err := unix.Getsockname(serverFd)
    if err != nil {
        logrus.Fatal(err)
    }

    clientFd, err := unix.Socket(unix.AF_INET, socketType, 0)
    if err != nil {
        logrus.Fatal(err)
    }
    setSocketTimeout(clientFd, 5000)

    ip := net.ParseIP(INADDR_TEST)
    sa := &unix.SockaddrInet4{
        Port: BindPort + 1,
        Addr: [4]byte{},
    }
    copy(sa.Addr[:], ip)
    err = unix.Bind(clientFd, sa)
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

func startServer() int {
    serverFd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
    if err != nil {
        logrus.Fatal(err)
    }
    setSocketTimeout(serverFd, 5000)

    err = unix.SetsockoptInt(serverFd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
    if err != nil {
        logrus.Fatalf("unix.SO_REUSEADDR error: %v", err)
    }

    ip := net.ParseIP(INADDR_TEST)
    sa := &unix.SockaddrInet4{
        Port: BindPort,
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

func IsLittleEndian() bool {
    val := int32(0x1)

    return *(*byte)(unsafe.Pointer(&val)) == 1
}
