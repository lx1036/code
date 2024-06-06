package nodediscovery

import (
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/cidr"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/datapath"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/lock"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging/logfields"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/mtu"
    nodemanager "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/node/manager"
    nodestore "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/node/store"
    nodeTypes "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/node/types"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/option"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/source"
    cnitypes "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/plugins/cilium-cni/types"
)

const (
    // AutoCIDR indicates that a CIDR should be allocated
    AutoCIDR            = "auto"
    nodeDiscoverySubsys = "nodediscovery"
    maxRetryCount       = 10
)

var (
    log = logging.DefaultLogger.WithField(logfields.LogSubsys, nodeDiscoverySubsys)
)

// NodeDiscovery represents a node discovery action
type NodeDiscovery struct {
    Manager               *nodemanager.Manager
    LocalConfig           datapath.LocalNodeConfiguration
    Registrar             nodestore.NodeRegistrar
    Registered            chan struct{}
    localStateInitialized chan struct{}
    NetConf               *cnitypes.NetConf
    k8sNodeGetter         k8sNodeGetter
    localNodeLock         lock.Mutex
    localNode             nodeTypes.Node
}

// NewNodeDiscovery returns a pointer to new node discovery object
func NewNodeDiscovery(manager *nodemanager.Manager, mtuConfig mtu.Configuration, netConf *cnitypes.NetConf) *NodeDiscovery {
    var auxPrefixes []*cidr.CIDR
    if option.Config.IPv4ServiceRange != AutoCIDR {
        serviceCIDR, err := cidr.ParseCIDR(option.Config.IPv4ServiceRange)
        if err != nil {
            log.WithError(err).WithField(logfields.V4Prefix, option.Config.IPv4ServiceRange).Fatal("Invalid IPv4 service prefix")
        }

        auxPrefixes = append(auxPrefixes, serviceCIDR)
    }

    return &NodeDiscovery{
        Manager: manager,
        LocalConfig: datapath.LocalNodeConfiguration{
            MtuConfig:               mtuConfig,
            UseSingleClusterRoute:   option.Config.UseSingleClusterRoute,
            EnableIPv4:              option.Config.EnableIPv4,
            EnableIPv6:              option.Config.EnableIPv6,
            EnableEncapsulation:     option.Config.Tunnel != option.TunnelDisabled,
            EnableAutoDirectRouting: option.Config.EnableAutoDirectRouting,
            EnableLocalNodeRoute:    enableLocalNodeRoute(),
            AuxiliaryPrefixes:       auxPrefixes,
            EnableIPSec:             option.Config.EnableIPSec,
            EncryptNode:             option.Config.EncryptNode,
            IPv4PodSubnets:          option.Config.IPv4PodSubnets,
            IPv6PodSubnets:          option.Config.IPv6PodSubnets,
        },
        localNode: nodeTypes.Node{
            Source: source.Local,
        },
        Registered:            make(chan struct{}),
        localStateInitialized: make(chan struct{}),
        NetConf:               netConf,
    }
}
