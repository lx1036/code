package bpf

import (
	"fmt"

	xdp_socket "k8s-lx1036/k8s/bpf/xdp-l4/xdp/xdp-socket"

	"github.com/cilium/ebpf"
)

// go generate requires appropriate linux headers in included (-I) paths.
// See accompanying Makefile + Dockerfile to make updates.

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc "$CLANG" -strip "$STRIP" -makebase "$MAKEDIR" ipproto protocol_filter.c -- -mcpu=v2 -nostdinc -Wall -Werror -Wno-compare-distinct-pointer-types -I./include

// NewIPProtoProgram returns an new eBPF that directs packets of the given ip protocol to XDP sockets
func NewIPProtoProgram(protocol uint32, options *ebpf.CollectionOptions) (*xdp_socket.Program, error) {
	spec, err := loadIpproto()
	if err != nil {
		return nil, err
	}

	if protocol <= 255 {
		if err := spec.RewriteConstants(map[string]interface{}{"PROTO": uint8(protocol)}); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("protocol must be between 0 and 255")
	}
	var program ipprotoObjects
	if err := spec.LoadAndAssign(&program, options); err != nil {
		return nil, err
	}

	p := &xdp_socket.Program{Program: program.XdpSockProg, Queues: program.QidconfMap, Sockets: program.XsksMap}
	return p, nil
}
