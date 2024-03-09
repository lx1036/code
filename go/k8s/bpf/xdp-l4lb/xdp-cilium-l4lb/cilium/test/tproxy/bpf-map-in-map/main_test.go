package main

import (
    "github.com/cilium/ebpf"
    "github.com/sirupsen/logrus"
    "testing"
    "unsafe"
)

// github.com/cilium/ebpf@v0.12.3/examples/map_in_map/main.go

const BPF_F_INNER_MAP = 0x1000

// 验证没有问题: 1.可以代码创建 innerMap 和 outerMap; 2. outerMap 从 elf 文件里加载(未验证通过)

// CGO_ENABLED=0 go test -v -run ^TestMapInMap$ .
func TestMapInMap(test *testing.T) {
    logrus.SetReportCaller(true)

    innerMapSpec := &ebpf.MapSpec{
        Name:       "inner_map",
        Type:       ebpf.Hash,
        KeySize:    uint32(unsafe.Sizeof(int(0))), // sizeof(int)
        ValueSize:  uint32(unsafe.Sizeof(int(0))), // sizeof(int)
        MaxEntries: 2,

        //Contents: make([]ebpf.MapKV, 2), // 不要填充 Contents 字段，否则必须填数据

        // This flag is required for dynamically sized inner maps.
        // Added in linux 5.10.
        //Flags: BPF_F_INNER_MAP,
    }
    innerMap, err := ebpf.NewMap(innerMapSpec)
    if err != nil {
        logrus.Fatalf("inner_map: %v", err)
    }
    defer innerMap.Close()

    outerMapSpec := ebpf.MapSpec{
        Name:       "outer_map",
        Type:       ebpf.ArrayOfMaps,
        KeySize:    uint32(unsafe.Sizeof(uint32(0))), // 4 bytes for u32
        ValueSize:  uint32(unsafe.Sizeof(uint32(0))), // 4
        MaxEntries: 1,                                // We'll have 1 maps inside this map
        //Contents:   make([]ebpf.MapKV, 1),
    }
    // INFO: 0.这里需要指定 inner map
    outerMapSpec.InnerMap = innerMapSpec.Copy()
    outerMap, err := ebpf.NewMap(&outerMapSpec)
    if err != nil {
        logrus.Fatalf("outer_map: %v", err)
    }
    defer outerMap.Close()

    // INFO: 1.这里输入的是 innerMap fd,内核示例里也是使用的 map fd
    //  /root/linux-5.10.142/tools/testing/selftests/bpf/test_maps.c#L1190-L1207
    //  /root/linux-5.10.142/tools/testing/selftests/bpf/prog_tests/select_reuseport.c#L692
    // 或者也可以输入 map id
    //info1, _ := innerMap.Info()
    //info1.ID()
    err = outerMap.Put(uint32(0), uint32(innerMap.FD()))
    if err != nil {
        logrus.Fatalf("outer_map put: %v", err)
    }

    mapIter := outerMap.Iterate()
    var outerMapKey uint32
    var innerMapID ebpf.MapID
    for mapIter.Next(&outerMapKey, &innerMapID) {
        // INFO: 2.这里输出的是 innerMap id
        innerMap2, err := ebpf.NewMapFromID(innerMapID)
        if err != nil {
            logrus.Fatal(err)
        }

        innerMapInfo, err := innerMap2.Info()
        if err != nil {
            logrus.Fatal(err)
        }

        logrus.Infof("outerMapKey %d, innerMap.Info: %+v", outerMapKey, innerMapInfo)
    }

    // INFO: 2.这里输出的是 innerMap id
    var innerMapId uint32
    err = outerMap.Lookup(uint32(0), &innerMapId)
    if err != nil {
        logrus.Errorf("outer_map err: %v", err)
    }
    logrus.Infof("outer_map innerMap id: %d", innerMapId)

    select {}
}
