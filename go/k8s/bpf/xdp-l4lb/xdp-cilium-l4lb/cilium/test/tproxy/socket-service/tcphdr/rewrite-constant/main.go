package main

import (
    "github.com/cilium/ebpf"
    "github.com/cilium/ebpf/link"
    "github.com/sirupsen/logrus"
    "net"
    "os"
    "os/signal"
    "syscall"
    "time"
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

    time.Sleep(time.Second)

    for name, mapSpec := range spec.Maps {
        logrus.Infof("map name: %s, mapSpec: %+v", name, *mapSpec)
        //if name == ".bss" { // 有两个 .bss map
        m, err := ebpf.NewMap(mapSpec)
        if err != nil {
            logrus.Errorf("err: %v", err)
            continue
        }

        mapInfo, err := m.Info()
        if err != nil {
            logrus.Error(err)
        }
        logrus.Infof("map.Info: %+v", mapInfo)

        var val uint32
        err = m.Lookup(uint32(0), &val)
        if err != nil {
            logrus.Errorf("err: %v", err)
        }
        logrus.Infof("val: %d", val)

        //iterator := m.Iterate()
        //var key uint32
        //var value uint32
        //for iterator.Next(&key, &value) {
        //    logrus.Infof("key: %v, value: %v", key, value)
        //}

        //for _, varSecinfo := range mapSpec.Value.(*btf.Datasec).Vars {
        //    //logrus.Infof("%+v", varSecinfo.Type.(*btf.Var))
        //    v := varSecinfo.Type.(*btf.Var)
        //    logrus.Infof("%+v", v.Type)
        //    u32 := v.Type.(*btf.Volatile).Type.(*btf.Typedef)
        //    logrus.Infof("%+v", u32.Type.(*btf.Int))
        //}
        //}
    }

    // TODO: 无法访问 bpf 里 inherit_cb_flags 变量值
    //typ, err := spec.Types.AnyTypeByName("inherit_cb_flags")
    //logrus.Infof("%+v", typ)
    //logrus.Infof("%+v", typ.(*btf.Var).Type.(*btf.Volatile).Type.(*btf.Typedef).Type.(*btf.Int))

    // Wait
    <-stopCh
}

// 调试:
// cat /sys/kernel/debug/tracing/trace_pipe
// bpf_trace_printk: flags: 10, max_delack_ms:10, rand:10
