package main

import (
    "github.com/cilium/ebpf"
    "github.com/cilium/ebpf/link"
    "github.com/sirupsen/logrus"
    "net"
    "os"
    "os/signal"
    "syscall"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -type bpf_test_option bpf test_rewrite_const.c -- -I.

// 验证可以 RewriteConstants 一个 struct const

// go generate .
// CGO_ENABLED=0 go run .
func main() {
    logrus.SetReportCaller(true)

    stopCh := make(chan os.Signal, 1)
    signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

    objs := bpfObjects{}
    opts := &ebpf.CollectionOptions{
        Programs: ebpf.ProgramOptions{
            LogLevel: ebpf.LogLevelInstruction,
            LogSize:  64 * 1024 * 1024, // 64M
        },
    }
    spec, err := loadBpf()
    if err != nil {
        logrus.Fatal(err)
    }
    consts := map[string]interface{}{
        "passive_synack_out": bpfBpfTestOption{
            Flags:       10,
            MaxDelackMs: 10,
            Rand:        10,
        },
    }
    if err = spec.RewriteConstants(consts); err != nil {
        logrus.Fatal(err)
    }
    if err := spec.LoadAndAssign(&objs, opts); err != nil {
        logrus.Fatalf("loading objects: %v", err)
    }
    defer objs.Close()

    ifaceObj, err := net.InterfaceByName("lo")
    if err != nil {
        logrus.Fatalf("loading objects: %v", err)
    }
    l, err := link.AttachXDP(link.XDPOptions{
        Program:   objs.bpfPrograms.TestRewriteConst,
        Interface: ifaceObj.Index,
        Flags:     link.XDPGenericMode,
    })
    if err != nil {
        logrus.Fatal(err)
    }
    defer l.Close()

    // Wait
    <-stopCh
}

// 调试:
// cat /sys/kernel/debug/tracing/trace_pipe
// bpf_trace_printk: flags: 10, max_delack_ms:10, rand:10
