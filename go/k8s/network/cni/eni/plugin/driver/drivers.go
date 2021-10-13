package driver

import (
	"net"

	eniTypes "k8s-lx1036/k8s/network/cni/eni/types"

	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/plugins/pkg/ns"
)

// NetnsDriver to config container netns interface and routes
type NetnsDriver interface {
	Setup(cfg *SetupConfig, netNS ns.NetNS) error

	Teardown(cfg *TeardownCfg, netNS ns.NetNS) error

	Check(cfg *CheckConfig) error
}

type SetupConfig struct {
	HostVETHName string

	ContainerIfName string
	ContainerIPNet  *eniTypes.IPNetSet
	GatewayIP       *eniTypes.IPSet
	MTU             int
	ENIIndex        int
	TrunkENI        bool

	// add extra route in container
	ExtraRoutes []types.Route

	ServiceCIDR *eniTypes.IPNetSet
	HostIPSet   *eniTypes.IPNetSet
	// ipvlan
	HostStackCIDRs []*net.IPNet

	Ingress uint64
	Egress  uint64
}

type TeardownCfg struct {
	HostVETHName string

	ContainerIfName string
	ContainerIPNet  *eniTypes.IPNetSet
}

type RecordPodEvent func(msg string)

type CheckConfig struct {
	RecordPodEvent

	NetNS ns.NetNS

	HostVETHName    string
	ContainerIFName string

	ContainerIPNet *eniTypes.IPNetSet
	HostIPSet      *eniTypes.IPNetSet
	GatewayIP      *eniTypes.IPSet

	ENIIndex int32 // phy device
	TrunkENI bool
	MTU      int
}
