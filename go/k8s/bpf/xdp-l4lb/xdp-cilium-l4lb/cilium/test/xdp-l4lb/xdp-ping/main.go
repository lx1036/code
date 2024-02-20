package main

import (
    "encoding/binary"
    "flag"
    "github.com/cilium/ebpf/link"
    "github.com/cilium/ebpf/rlimit"
    "github.com/sirupsen/logrus"
    "net"
    "os"
    "os/exec"
    "os/signal"
    "syscall"
    "time"
)

// /root/linux-5.10.142/tools/testing/selftests/bpf/xdping.c

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -tags "linux" -type pinginfo bpf xdp_ping.c -- -I.

// go generate .
// CGO_ENABLED=0 go run .
func main() {
    logrus.SetReportCaller(true)

    iface := flag.String("iface", "veth1", "interface name")
    mode := flag.String("mode", "driver", "interface name")
    flag.Parse()

    stopCh := make(chan os.Signal, 1)
    signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

    // Allow the current process to lock memory for eBPF resources.
    if err := rlimit.RemoveMemlock(); err != nil {
        logrus.Fatal(err)
    }

    device, err := net.InterfaceByName(*iface)
    if err != nil {
        logrus.Fatal(err)
    }

    // Load pre-compiled programs and maps into the kernel.
    objs := bpfObjects{}
    if err := loadBpfObjects(&objs, nil); err != nil {
        logrus.Fatalf("loading objects: %v", err)
    }
    defer objs.Close()

    xdpMode := link.XDPDriverMode
    if *mode == "skb" {
        xdpMode = link.XDPGenericMode
    }
    l, err := link.AttachXDP(link.XDPOptions{
        Program:   objs.bpfPrograms.XdpingClient,
        Interface: device.Index,
        Flags:     xdpMode,
    })
    if err != nil {
        logrus.Fatal(err)
    }
    defer l.Close()
    info, err := l.Info()
    if err != nil {
        logrus.Fatal(err)
    }
    logrus.Infof("ProgramID: %d", info.Program)

    key := IPToUint32()
    pingInfo := bpfPinginfo{
        Seq:   5,
        Count: 6,
    }
    err = objs.bpfMaps.PingMap.Put(key, pingInfo)
    if err != nil {
        logrus.Fatal(err)
    }

    logrus.Infof("success to attach")

    /* Start xdping-ing from last regular ping reply, e.g. for a count
     * of 10 ICMP requests, we start xdping-ing using reply with seq number
     * 10.  The reason the last "real" ping RTT is much higher is that
     * the ping program sees the ICMP reply associated with the last
     * XDP-generated packet, so ping doesn't get a reply until XDP is done.
     */
    /* We need to wait for XDP setup to complete. */
    time.Sleep(3)
    err = exec.Command("ping", "-c", "3", "-I", *iface, "10.1.1.100").Run()
    if err != nil {
        logrus.Fatal(err)
    }

    // get stats
    pingInfo2 := bpfPinginfo{}
    err = objs.bpfMaps.PingMap.Lookup(key, &pingInfo2)
    if err != nil {
        logrus.Fatal(err)
    }

    logrus.Infof("pingInfo: %+v", pingInfo2)

    // Wait
    <-stopCh
}

func IPToUint32() uint32 {
    ip := "10.1.1.100"
    netip := net.ParseIP(ip)
    ip4 := netip.To4()
    // 本地容器 ebpf-for-mac 是 LittleEndian
    return binary.LittleEndian.Uint32(ip4) // 1677787402
    //return binary.BigEndian.Uint32(ip4) // 167838052

    //ui32 := uint32(ip4[0])<<24 | uint32(ip4[1])<<16 | uint32(ip4[2])<<8 | uint32(ip4[3])
    //fmt.Println(ui32) // 167838052
    //return ui32
}
