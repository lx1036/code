package bpf

import (
	"log"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
)

// 这里 xdp_basic.c 不能和 xdp_basic.go 在同一个文件夹，否则报错 "C source files not allowed when not using cgo or SWIG"

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc "$CLANG" -strip "$STRIP" -makebase "$MAKEDIR" xdp_basic ./include/xdp_basic.c -- -mcpu=v2 -nostdinc -Wall -Werror -Wno-compare-distinct-pointer-types -I./include

type XdpObjects struct {
	XdpDropFunc *ebpf.Program `ebpf:"xdp_drop_func"`
	XdpPassFunc *ebpf.Program `ebpf:"xdp_pass_func"`

	Link link.Link
}

// LoadAndAttachXdp Load pre-compiled programs into the kernel.
func LoadAndAttachXdp(iIndex int) (*XdpObjects, error) {
	// Load pre-compiled programs into the kernel
	objs := xdp_basicObjects{}
	if err := loadXdp_basicObjects(&objs, nil); err != nil {
		return nil, err
	}

	xdpObjects := &XdpObjects{
		XdpDropFunc: objs.XdpDropFunc,
	}

	// Attach the program into interface
	l, err := link.AttachXDP(link.XDPOptions{
		Program:   objs.XdpDropFunc, // 挂载 xdp_drop_func xdp
		Interface: iIndex,
	})
	if err != nil {
		log.Fatalf("could not attach XDP program: %s", err)
	}
	//defer l.Close() // 这里 close 会卸载 xdp 程序
	xdpObjects.Link = l

	return xdpObjects, nil
}

func (objs *XdpObjects) Close() error {
	var err error
	if objs.XdpPassFunc != nil {
		err = objs.XdpPassFunc.Close()
	}
	if objs.XdpDropFunc != nil {
		err = objs.XdpDropFunc.Close()
	}
	if objs.Link != nil {
		err = objs.Link.Close()
	}

	return err
}
