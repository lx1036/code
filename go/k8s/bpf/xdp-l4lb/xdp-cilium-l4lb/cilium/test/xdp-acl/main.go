package main

import (
	"encoding/binary"
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

// go generate .
// CGO_ENABLED=0 go run .
func main() {
	logrus.SetReportCaller(true)

	stopper := make(chan os.Signal, 1)
	signal.Notify(stopper, os.Interrupt, syscall.SIGTERM)

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

	ip1 := "172.16.10.3"
	var addr uint32
	if IsLittleEndian() {
		addr = binary.LittleEndian.Uint32(net.ParseIP(ip1).To4()) // byte[]{a,b,c,d} -> dcba
	} else {
		addr = binary.BigEndian.Uint32(net.ParseIP(ip1).To4()) // byte[]{a,b,c,d} -> abcd
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
		Dport:    uint16(9090),
	}
	action := bpfAction{
		Action: uint8(0),
	}
	if err := objs.bpfMaps.Endpoints.Put(endpoint, action); err != nil {
		logrus.Error(err)
	}

	var action1 bpfAction
	if err := objs.bpfMaps.Endpoints.Lookup(&endpoint, &action1); err != nil {
		logrus.Error(err)
	}
	logrus.Infof("%+v", action1)

	ifaceObj, err := net.InterfaceByName("eth0")
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

	// Wait
	<-stopper
}

func IsLittleEndian() bool {
	var val int32 = 0x1

	return *(*byte)(unsafe.Pointer(&val)) == 1
}
