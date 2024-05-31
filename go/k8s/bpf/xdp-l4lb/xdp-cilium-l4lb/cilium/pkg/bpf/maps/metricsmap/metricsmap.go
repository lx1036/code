package metricsmap

import (
	"unsafe"

	"k8s-lx1036/k8s/network/cilium/cilium/pkg/bpf"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/common"
)

const (
	// MapName for metrics map.
	MapName = "cilium_metrics"
	// MaxEntries is the maximum number of keys that can be present in the
	// Metrics Map.
	//
	// Currently max. 2 bits of the Key.Dir member are used (unknown,
	// ingress or egress). Thus we can reduce from the theoretical max. size
	// of 2**16 (2 uint8) to 2**10 (1 uint8 + 2 bits).
	MaxEntries = 1024
)

var (
	Metrics *bpf.Map

	possibleCpus int
)

func init() {
	possibleCpus = common.GetNumPossibleCPUs() // 比如总共 24 个核

	vs := make(Values, possibleCpus)

	// Metrics is a mapping of all packet drops and forwards associated with
	// the node on ingress/egress direction
	Metrics = bpf.NewPerCPUHashMap(
		MapName,
		&Key{},
		int(unsafe.Sizeof(Key{})),
		&vs,
		int(unsafe.Sizeof(Value{})),
		possibleCpus,
		MaxEntries,
		0, 0,
		bpf.ConvertKeyValue,
	)
}

// Key must be in sync with struct metrics_key in <bpf/lib/common.h>
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
type Key struct {
	Reason   uint8      `align:"reason"`
	Dir      uint8      `align:"dir"`
	Reserved pad3uint16 `align:"reserved"`
}

// Value must be in sync with struct metrics_value in <bpf/lib/common.h>
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapValue
type Value struct {
	Count uint64 `align:"count"`
	Bytes uint64 `align:"bytes"`
}

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapValue
// Values is a slice of Values
type Values []Value
