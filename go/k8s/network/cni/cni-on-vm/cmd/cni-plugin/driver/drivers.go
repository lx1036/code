package driver

import (
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/utils/types"

	"github.com/containernetworking/plugins/pkg/ns"
)

// NetnsDriver to config container netns interface and routes
type NetnsDriver interface {
	Setup(cfg *types.SetupConfig, netNS ns.NetNS) error

	Teardown(cfg *types.TeardownCfg, netNS ns.NetNS) error

	Check(cfg *CheckConfig) error
}

type RecordPodEvent func(msg string)

type CheckConfig struct {
	RecordPodEvent

	NetNS ns.NetNS

	HostVETHName    string
	ContainerIFName string

	ContainerIPNet *types.IPNetSet
	HostIPSet      *types.IPNetSet
	GatewayIP      *types.IPSet

	ENIIndex int32 // phy device
	TrunkENI bool
	MTU      int
}
