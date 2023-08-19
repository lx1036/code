package internal

import (
	"encoding/binary"
	"fmt"
	"github.com/cilium/ebpf"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc "$CLANG" -strip "$STRIP" -makebase "$MAKEDIR" dispatcher ../ebpf/socket_dispatch.c -- -mcpu=v2 -nostdinc -Wall -Werror -I../ebpf/include

type Dispatcher struct {
}

func loadPatchedDispatcher(to interface{}, opts *ebpf.CollectionOptions) (*ebpf.CollectionSpec, error) {
	spec, err := loadDispatcher()
	if err != nil {
		return nil, err
	}

	var specs dispatcherSpecs
	if err := spec.Assign(&specs); err != nil {
		return nil, err
	}

	// 确保 destinations 和 destination_metrics 最大条目与 sockets 一致
	maxSockets := specs.Sockets.MaxEntries
	for _, m := range []*ebpf.MapSpec{
		specs.Destinations,
		specs.DestinationMetrics,
	} {
		if m.MaxEntries != maxSockets {
			return nil, fmt.Errorf("map %q has %d max entries instead of %d", m.Name, m.MaxEntries, maxSockets)
		}
	}

	// c 里没有定义，在 go 里定义了
	specs.Destinations.KeySize = uint32(binary.Size(destinationKey{}))
	specs.Destinations.ValueSize = uint32(binary.Size(destinationAlloc{}))

	if to != nil {
		return spec, spec.LoadAndAssign(to, opts)
	}

	return spec, nil
}
