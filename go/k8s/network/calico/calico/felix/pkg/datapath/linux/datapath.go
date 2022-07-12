package linux

import (
	"k8s-lx1036/k8s/network/calico/calico/felix/pkg/bpf"
)

type Config struct {
}

type LinuxDatapath struct {
}

func NewLinuxDatapath(config Config) *LinuxDatapath {

	bpfMapContext := bpf.CreateBPFMapContext(config.BPFMapSizeIPSets, config.BPFMapSizeNATFrontend,
		config.BPFMapSizeNATBackend, config.BPFMapSizeNATAffinity, config.BPFMapSizeRoute,
		config.BPFMapSizeConntrack, config.BPFMapRepin)
	err := bpf.CreateBPFMaps(bpfMapContext)

}
