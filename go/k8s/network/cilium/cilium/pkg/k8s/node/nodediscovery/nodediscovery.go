package nodediscovery

import (
	"github.com/cilium/cilium/pkg/option"
	"k8s-lx1036/k8s/network/cilium/cilium/pkg/datapath"

	"github.com/cilium/cilium/pkg/mtu"
	cnitypes "github.com/containernetworking/cni/pkg/types"
)

// NodeDiscovery represents a node discovery action
type NodeDiscovery struct {
	LocalConfig datapath.LocalNodeConfiguration
}

func NewNodeDiscovery(manager *nodemanager.Manager, mtuConfig mtu.Configuration,
	netConf *cnitypes.NetConf) *NodeDiscovery {
	var auxPrefixes []*cidr.CIDR

	return &NodeDiscovery{
		LocalConfig: datapath.LocalNodeConfiguration{
			MtuConfig:               mtuConfig,
			AuxiliaryPrefixes:       auxPrefixes,
			EnableIPv4:              option.Config.EnableIPv4,
			EnableIPv6:              option.Config.EnableIPv6,
			UseSingleClusterRoute:   option.Config.UseSingleClusterRoute,
			EnableEncapsulation:     option.Config.Tunnel != option.TunnelDisabled, // 是不是走 overlay tunnel 模式
			EnableAutoDirectRouting: option.Config.EnableAutoDirectRouting,
			EnableLocalNodeRoute:    enableLocalNodeRoute(),
			EnableIPSec:             option.Config.EnableIPSec,
			EncryptNode:             option.Config.EncryptNode,
			IPv4PodSubnets:          option.Config.IPv4PodSubnets,
			IPv6PodSubnets:          option.Config.IPv6PodSubnets,
		},
	}
}

func enableLocalNodeRoute() bool {
	return option.Config.EnableLocalNodeRoute && option.Config.IPAM != ipamOption.IPAMENI
}
