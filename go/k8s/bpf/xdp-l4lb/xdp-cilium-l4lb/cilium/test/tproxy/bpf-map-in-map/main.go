package main

import (
    "github.com/cilium/ebpf"
    "github.com/cilium/ebpf/rlimit"
    "github.com/sirupsen/logrus"
    "os"
    "os/signal"
    "syscall"
    "unsafe"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go bpf test_map_in_map.c -- -I.

// github.com/cilium/ebpf@v0.12.3/examples/map_in_map/main.go
// /root/linux-5.10.142/tools/testing/selftests/bpf/test_maps.c

const (
    PinPath = "/sys/fs/bpf/map_in_map"
)

func init() {
    // Allow the current process to lock memory for eBPF resources.
    if err := rlimit.RemoveMemlock(); err != nil {
        logrus.Fatal(err)
    }
}

// INFO:
//  方式一：通过在 bpf c 里指定一个 inner map template，然后在 userspace 里去 put 替换掉 inner map，验证通过
//  方式二：无需在 bpf c 里指定一个 inner map template，在 userspace 里指定 inner map 然后再 load into kernel, 但是一直报错，没有验证通过

// map_in_map 验证通过
// go generate .
// CGO_ENABLED=0 go run .
func main() {
    logrus.SetReportCaller(true)

    stopCh := make(chan os.Signal, 1)
    signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

    // 1.create inner map
    innerMapSpec := &ebpf.MapSpec{
        Name: "inner_map",
        Type: ebpf.Hash,
        // KeySize:    uint32(unsafe.Sizeof(int(0))), // sizeof(int)
        KeySize: uint32(unsafe.Sizeof(uint32(0))), // sizeof(int)
        // ValueSize:  uint32(unsafe.Sizeof(int(0))), // sizeof(int)
        ValueSize:  uint32(unsafe.Sizeof(uint32(0))), // sizeof(int)
        MaxEntries: 2,
    }
    //innerMap, err := ebpf.NewMapWithOptions(innerMapSpec, ebpf.MapOptions{
    //    PinPath: fmt.Sprintf("%s/%s", PinPath, "inner_map"),
    //})
    innerMap, err := ebpf.NewMap(innerMapSpec)
    if err != nil {
        logrus.Fatalf("NewMapWithOptions err: %v", err)
    }
    defer innerMap.Close()

    //2. outer map 从 elf 文件里加载
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
    spec, err := loadBpf()
    if err != nil {
        logrus.Errorf("loadBpf err: %v", err)
        return
    }
    // spec.Maps["mim_array"].InnerMap = innerMapSpec.Copy()
    // spec.Maps["mim_array"].Contents = make([]ebpf.MapKV, spec.Maps["mim_array"].MaxEntries)
    // spec.Maps["mim_array"].Contents[0] = ebpf.MapKV{Key: 0, Value: innerMap}
    // spec.Maps["mim_hash"].InnerMap = innerMapSpec.Copy()
    // spec.Maps["mim_hash"].Contents = make([]ebpf.MapKV, spec.Maps["mim_hash"].MaxEntries)
    // spec.Maps["mim_hash"].Contents[0] = ebpf.MapKV{Key: 0, Value: innerMap}
    err = spec.LoadAndAssign(&objs, opts)
    if err != nil {
        logrus.Errorf("LoadAndAssign err: %v", err)
        return
    }
    defer objs.Close()

    // INFO: 1.这里输入的是 innerMap fd

    err = objs.bpfMaps.MimArray.Put(uint32(0), uint32(innerMap.FD()))
    if err != nil {
        logrus.Errorf("MimArray Put err: %v", err)
    }

    // INFO: 注意，这里直接使用 int(0) 传参报错 "marshal key: binary.Write: invalid type int", 如果 map key=int
    key1 := int(0)
    // err = objs.bpfMaps.MimHash.Put(unsafe.Pointer(&key1), uint32(innerMap.FD()))
    err = objs.bpfMaps.MimHash.Put(uint32(0), uint32(innerMap.FD()))
    if err != nil {
        logrus.Errorf("MimHash Put err: %v", err)
    }

    var innerMapId uint32
    err = objs.bpfMaps.MimArray.Lookup(uint32(0), &innerMapId)
    if err != nil {
        logrus.Errorf("MimArray Put err: %v", err)
    }
    logrus.Infof("MimArray innerMap id: %d", innerMapId)

    var innerMapId2 uint32
    err = objs.bpfMaps.MimHash.Lookup(unsafe.Pointer(&key1), &innerMapId2)
    if err != nil {
        logrus.Errorf("MimHash Put err: %v", err)
    }
    logrus.Infof("MimHash innerMap id: %d", innerMapId2)

    <-stopCh
}
