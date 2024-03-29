//go:build linux

package main

import (
    "bytes"
    "encoding/binary"
    "errors"
    "github.com/cilium/ebpf"
    "github.com/cilium/ebpf/link"
    "github.com/cilium/ebpf/ringbuf"
    "github.com/cilium/ebpf/rlimit"
    "golang.org/x/sys/unix"
    "log"
    "net"
    "os"
    "os/signal"
    "path/filepath"
    "syscall"
)

// https://github.com/cilium/ebpf/blob/main/examples/tcprtt_sockops/main.go

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -tags "linux" -type rtt_event bpf tcprtt_sockops.c -- -I.

/*
This program demonstrates attaching an eBPF program to a cgroupv2 path and using sockops to process TCP socket events.
It prints the IPs/ports/RTT information every time TCP sockets update their internal RTT value.
It supports only IPv4 for this example.
*/

// go generate .
// 验证没问题：go run .
/*
2023/12/12 23:55:40 eBPF program loaded and attached on cgroup /sys/fs/cgroup
2023/12/12 23:55:40 Src addr        Port   -> Dest addr       Port   RTT (ms)
2023/12/12 23:55:49 172.16.10.2     35276  -> 100.103.0.45    80     0
2023/12/12 23:55:52 172.16.10.2     49984  -> 100.100.167.126 443    31
2023/12/12 23:55:52 172.16.10.2     49984  -> 100.100.167.126 443    31
2023/12/12 23:55:52 172.16.10.2     49984  -> 100.100.167.126 443    32
2023/12/12 23:55:52 172.16.10.2     49984  -> 100.100.167.126 443    32
2023/12/12 23:55:52 172.16.10.2     49984  -> 100.100.167.126 443    32
2023/12/12 23:55:52 172.16.10.2     49984  -> 100.100.167.126 443    32
2023/12/12 23:55:52 172.16.10.2     49984  -> 100.100.167.126 443    31
2023/12/12 23:55:52 172.16.10.2     49984  -> 100.100.167.126 443    31
2023/12/12 23:55:52 172.16.10.2     49984  -> 100.100.167.126 443    31
2023/12/12 23:55:52 172.16.10.2     49984  -> 100.100.167.126 443    31
2023/12/12 23:55:52 172.16.10.2     49984  -> 100.100.167.126 443    31
2023/12/12 23:55:52 172.16.10.2     49984  -> 100.100.167.126 443    31
2023/12/12 23:55:53 172.16.10.2     49984  -> 100.100.167.126 443    31
2023/12/12 23:55:53 172.16.10.2     49984  -> 100.100.167.126 443    31
2023/12/12 23:55:53 172.16.10.2     49984  -> 100.100.167.126 443    31
*/

// bpftool prog list -p
// bpftool map list -p
func main() {
    stopCh := make(chan os.Signal, 1)
    signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

    // Allow the current process to lock memory for eBPF resources.
    if err := rlimit.RemoveMemlock(); err != nil {
        log.Fatal(err)
    }

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
    link, err := link.AttachCgroup(link.CgroupOptions{
        Path:    cgroupPath,
        Program: objs.bpfPrograms.BpfSockopsCb,
        Attach:  ebpf.AttachCGroupSockOps,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer link.Close()

    log.Printf("eBPF program loaded and attached on cgroup %s\n", cgroupPath)

    rd, err := ringbuf.NewReader(objs.bpfMaps.RttEvents)
    if err != nil {
        log.Fatalf("opening ringbuf reader: %s", err)
    }
    defer rd.Close()

    log.Printf("%-15s %-6s -> %-15s %-6s %-6s", "Src addr", "Port", "Dest addr", "Port", "RTT (ms)")

    go readLoop(rd)

    // Wait
    <-stopCh
}

func readLoop(rd *ringbuf.Reader) {
    // bpfRttEvent is generated by bpf2go.
    var event bpfRttEvent
    for {
        record, err := rd.Read()
        if err != nil {
            if errors.Is(err, ringbuf.ErrClosed) {
                log.Println("received signal, exiting..")
                return
            }
            log.Printf("reading from reader: %s", err)
            continue
        }

        // Parse the ringbuf event entry into a bpfRttEvent structure.
        if err := binary.Read(bytes.NewBuffer(record.RawSample), NativeEndian, &event); err != nil {
            log.Printf("parsing ringbuf event: %s", err)
            continue
        }

        log.Printf("%-15s %-6d -> %-15s %-6d %-6d",
            intToIP(event.Saddr),
            event.Sport,
            intToIP(event.Daddr),
            event.Dport,
            event.Srtt,
        )
    }
}

// intToIP converts IPv4 number to net.IP
func intToIP(ipNum uint32) net.IP {
    ip := make(net.IP, 4)
    binary.BigEndian.PutUint32(ip, ipNum)
    return ip
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
