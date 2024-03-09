package linux

import (
	"github.com/cilium/cilium/pkg/versioncheck"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/datapath"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/datapath/linux/config"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/datapath/loader"
)

const (
	minKernelVer = "4.8.0"
	minClangVer  = "3.8.0"
	recKernelVer = "4.9.0"
	recClangVer  = "3.9.0"
)

var (
	isMinKernelVer = versioncheck.MustCompile(">=" + minKernelVer)
	isMinClangVer  = versioncheck.MustCompile(">=" + minClangVer)

	isRecKernelVer = versioncheck.MustCompile(">=" + recKernelVer)
	isRecClangVer  = versioncheck.MustCompile(">=" + recClangVer)

	// LLVM/clang version which supports `-mattr=dwarfris`
	isDwarfrisClangVer         = versioncheck.MustCompile(">=7.0.0")
	canDisableDwarfRelocations bool
)

type DatapathConfiguration struct {
	// HostDevice is the name of the device to be used to access the host.
	HostDevice string
}

type linuxDatapath struct {
	datapath.ConfigWriter
	datapath.IptablesManager
	node           datapath.NodeHandler
	nodeAddressing datapath.NodeAddressing
	config         DatapathConfiguration
	loader         *loader.Loader
	wgAgent        datapath.WireguardAgent
}

func NewDatapath(cfg DatapathConfiguration, ruleManager datapath.IptablesManager, wgAgent datapath.WireguardAgent) datapath.Datapath {
	dp := &linuxDatapath{
		ConfigWriter:    &config.HeaderfileWriter{},
		IptablesManager: ruleManager,
		nodeAddressing:  NewNodeAddressing(),
		config:          cfg,
		loader:          loader.NewLoader(canDisableDwarfRelocations),
		wgAgent:         wgAgent,
	}

	dp.node = NewNodeHandler(cfg, dp.nodeAddressing, wgAgent)
	return dp
}
