//go:build linux

package main

import (
    "flag"
    "github.com/cilium/ebpf/link"
    "github.com/cilium/ebpf/rlimit"
    "github.com/sirupsen/logrus"
    "golang.org/x/sys/unix"
    "net"
    "os"
    "os/signal"
    "syscall"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -tags "linux" bpf tcp_syncookie.c -- -I.

/**
tcpdump 抓包:
ip netns exec ns1 bash
tcpdump -i lo -nneevv -A -w syncookie.pcap
*/

// go generate .
// go run .
// bpftool prog list
func main() {
    logrus.SetReportCaller(true)

    xdp := flag.Bool("xdp", true, "for xdp")
    iface := flag.String("iface", "lo", "the interface to attach this program to")
    flag.Parse()

    stopCh := make(chan os.Signal, 1)
    signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

    // Allow the current process to lock memory for eBPF resources.
    if err := rlimit.RemoveMemlock(); err != nil {
        logrus.Fatalf("RemoveMemlock %v", err)
    }

    // Load pre-compiled programs and maps into the kernel.
    objs := bpfObjects{}
    if err := loadBpfObjects(&objs, nil); err != nil {
        logrus.Fatalf("loading objects: %v", err)
    }
    defer objs.Close()

    ifaceObj, err := net.InterfaceByName(*iface)
    if err != nil {
        logrus.Fatalf("lookup network iface %+v: %s", iface, err)
    }
    logrus.Infof("interface index: %d", ifaceObj.Index)
    l, err := link.AttachXDP(link.XDPOptions{
        Program:   objs.bpfPrograms.CheckSyncookieXdp, // 挂载 xdp_drop_func xdp
        Interface: ifaceObj.Index,
        Flags:     link.XDPGenericMode, // ecs eth0 貌似不支持 XDPDriverMode
    })
    if err != nil {
        logrus.Fatalf("could not attach XDP program: %s", err)
    }
    defer l.Close()

    serverFd := startServer()
    defer unix.Close(serverFd)

    key, cookie := uint32(0), uint32(0)
    if err := objs.bpfMaps.Results.Put(uint32(key), uint32(cookie)); err != nil {
        logrus.Fatal(err)
    }

    keyGen, valueGen := uint32(1), uint32(0)
    if err := objs.bpfMaps.Results.Put(uint32(keyGen), uint32(valueGen)); err != nil {
        logrus.Fatal(err)
    }

    keyMss, valueMss := uint32(2), uint32(0)
    if err := objs.bpfMaps.Results.Put(uint32(keyMss), uint32(valueMss)); err != nil {
        logrus.Fatal(err)
    }

    // connect server
    clientFd := connectToFd(serverFd)
    defer unix.Close(clientFd)

    _, _, err = unix.Accept(serverFd)
    if err != nil {
        logrus.Fatal(err)
    }

    if err = objs.bpfMaps.Results.Lookup(uint32(key), &cookie); err != nil {
        logrus.Fatal(err)
    }
    if cookie == 0 {
        logrus.Errorf("err: cookie is 0")
    } else {
        logrus.Infof("cookie is %d", cookie)
    }

    if err = objs.bpfMaps.Results.Lookup(uint32(keyGen), &valueGen); err != nil {
        logrus.Fatal(err)
    }
    if *xdp && valueGen == 0 {
        // SYN packets do not get passed through generic XDP, skip the rest of the test.
        logrus.Errorf("Skipping XDP cookie check")
        //return
    } else {
        logrus.Infof("valueGen is %d", valueGen)
    }
    if cookie != valueGen {
        logrus.Errorf("BPF generated cookie does not match kernel one")
    }

    if err = objs.bpfMaps.Results.Lookup(uint32(keyMss), &valueMss); err != nil {
        logrus.Fatal(err)
    }
    if valueMss < 536 {
        logrus.Errorf("Unexpected MSS retrieved: %d", valueMss)
    } else {
        logrus.Infof("MSS is %d", valueMss)
    }
}

func startServer() int {
    fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
    if err != nil {
        logrus.Fatal(err)
    }

    ip := net.ParseIP("127.0.0.1")
    sa := &unix.SockaddrInet4{
        Port: 8000,
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

func connectToFd(serverFd int) int {
    serverSockAddr, err := unix.Getsockname(serverFd)
    if err != nil {
        logrus.Fatal(err)
    }

    clientFd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
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
