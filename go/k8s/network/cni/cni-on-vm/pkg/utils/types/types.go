package types

import (
	"fmt"
	cniTypes "github.com/containernetworking/cni/pkg/types"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/rpc"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/utils"
	"net"
	"strings"
	"time"
)

const (
	ResourceTypeENIIP = "eniIP"
)

type DataPath int

const (
	VPCRoute DataPath = iota
	PolicyRoute
	IPVlan
	ExclusiveENI
	Vlan
)

// NetworkResource interface of network resources
type NetworkResource interface {
	GetResourceID() string
	GetType() string
	ToResItems() []ResourceItem
}

type IPNetSet struct {
	IPv4 *net.IPNet
	IPv6 *net.IPNet
}

func (i *IPNetSet) String() string {
	var result []string
	if i.IPv4 != nil {
		result = append(result, i.IPv4.String())
	}
	if i.IPv6 != nil {
		result = append(result, i.IPv6.String())
	}
	return strings.Join(result, ",")
}

func ToIPNetSet(ip *rpc.IPSet) (*IPNetSet, error) {
	if ip == nil {
		return nil, fmt.Errorf("ip is nil")
	}
	ipNetSet := &IPNetSet{}
	var err error
	if ip.IPv4 != "" {
		_, ipNetSet.IPv4, err = net.ParseCIDR(ip.IPv4)
		if err != nil {
			return nil, err
		}
	}
	if ip.IPv6 != "" {
		_, ipNetSet.IPv6, err = net.ParseCIDR(ip.IPv6)
		if err != nil {
			return nil, err
		}
	}
	return ipNetSet, nil
}

// IPSet is the type hole both ipv4 and ipv6 net.IP
type IPSet struct {
	IPv4 net.IP
	IPv6 net.IP
}

type ENI struct {
	ID               string
	MAC              string
	MaxIPs           int
	SecurityGroupIDs []string

	Trunk bool

	PrimaryIP IPSet
	GatewayIP IPSet

	VSwitchCIDR IPNetSet

	VSwitchID string
}

// GetResourceID return mac address of eni
func (e *ENI) GetResourceID() string {
	return e.MAC
}

type ENIIP struct {
	ENI   *ENI
	IPSet IPSet
}

// GetResourceID return mac address of eni and secondary ip address
func (e *ENIIP) GetResourceID() string {
	return fmt.Sprintf("%s.%s", e.ENI.GetResourceID(), e.IPSet.String())
}

func (e *ENIIP) GetType() string {
	return ResourceTypeENIIP
}

func (e *ENIIP) ToResItems() []ResourceItem {
	return []ResourceItem{
		{
			Type:   e.GetType(),
			ID:     e.GetResourceID(),
			ENIID:  e.ENI.ID,
			ENIMAC: e.ENI.MAC,
		},
	}
}

func IPv6(ip net.IP) bool {
	return ip.To4() == nil
}

func BuildIPNet(ip, subnet *rpc.IPSet) (*IPNetSet, error) {
	ipnet := &IPNetSet{}
	if ip == nil || subnet == nil {
		return ipnet, nil
	}
	exec := func(ip, subnet string) (*net.IPNet, error) {
		i, err := utils.ToIP(ip)
		if err != nil {
			return nil, err
		}
		_, sub, err := net.ParseCIDR(subnet)
		if err != nil {
			return nil, err
		}
		sub.IP = i
		return sub, nil
	}
	var err error
	if ip.IPv4 != "" && subnet.IPv4 != "" {
		ipnet.IPv4, err = exec(ip.IPv4, subnet.IPv4)
		if err != nil {
			return nil, err
		}
	}
	if ip.IPv6 != "" && subnet.IPv6 != "" {
		ipnet.IPv6, err = exec(ip.IPv6, subnet.IPv6)
		if err != nil {
			return nil, err
		}
	}
	return ipnet, nil
}

func ToIPSet(ip *rpc.IPSet) (*IPSet, error) {
	if ip == nil {
		return nil, fmt.Errorf("ip is nil")
	}
	ipSet := &IPSet{}
	var err error
	if ip.IPv4 != "" {
		ipSet.IPv4, err = utils.ToIP(ip.IPv4)
		if err != nil {
			return nil, err
		}
	}
	if ip.IPv6 != "" {
		ipSet.IPv6, err = utils.ToIP(ip.IPv6)
		if err != nil {
			return nil, err
		}
	}
	return ipSet, nil
}

type ResourceManagerInitItem struct {
	item    ResourceItem
	PodInfo *PodInfo
}

type ResourceItem struct {
	Type         string        `json:"type"`
	ID           string        `json:"id"`
	ExtraEipInfo *ExtraEipInfo `json:"extra_eip_info"`

	ENIID  string `json:"eni_id"`
	ENIMAC string `json:"eni_mac"`
	IPv4   string `json:"ipv4"`
	IPv6   string `json:"ipv6"`
}

type PodInfo struct {
	//K8sPod *v1.Pod
	Name           string
	Namespace      string
	TcIngress      uint64
	TcEgress       uint64
	PodNetworkType string
	PodIP          string // used for eip and mip
	PodIPs         IPSet  // used for eip and mip
	SandboxExited  bool
	EipInfo        PodEipInfo
	IPStickTime    time.Duration
	PodENI         bool
	PodUID         string
}

func PodInfoKey(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

type SetupConfig struct {
	HostVethIfName string
	HostIPSet      *IPNetSet

	ContainerIfName string
	ContainerIPNet  *IPNetSet
	GatewayIP       *IPSet
	MTU             int
	ENIIndex        int
	ENIGatewayIP    *IPSet

	// disable create peer for exclusiveENI
	DisableCreatePeer bool

	// StripVlan or use vlan
	StripVlan bool
	Vid       int

	DefaultRoute bool
	MultiNetwork bool

	// add extra route in container
	ExtraRoutes []cniTypes.Route

	ServiceCIDR *IPNetSet
	// ipvlan
	HostStackCIDRs []*net.IPNet

	Ingress uint64
	Egress  uint64
}

type TeardownCfg struct {
	ContainerIfName string

	ContainerIPNet *IPNetSet

	ServiceCIDR *IPNetSet
}

// CNIConf is the cni network config
type CNIConf struct {
	cniTypes.NetConf

	// HostVethPrefix is the veth for container prefix on host
	HostVethPrefix string `json:"veth_prefix"`

	// eniIPVirtualType is the ipvlan for container
	ENIIPVirtualType string `json:"eniip_virtual_type"`

	// HostStackCIDRs is a list of CIDRs, all traffic targeting these CIDRs will be redirected to host network stack
	HostStackCIDRs []string `json:"host_stack_cidrs"`

	DisableHostPeer bool `yaml:"disable_host_peer" json:"disable_host_peer"` // disable create peer for host and container. This will also disable ability for service

	VlanStripType VlanStripType `yaml:"vlan_strip_type" json:"vlan_strip_type"` // used in multi ip mode, how datapath config vlan

	// MTU is container and ENI network interface MTU
	MTU int `json:"mtu"`

	// RuntimeConfig represents the options to be passed in by the runtime.
	//RuntimeConfig cni.RuntimeConfig `json:"runtimeConfig"`

	// Debug
	Debug bool `json:"debug"`
}

func (n *CNIConf) IPVlan() bool {
	return strings.ToLower(n.ENIIPVirtualType) == "ipvlan"
}

// VlanStripType how datapath handle vlan
type VlanStripType string

// how datapath handle vlan
const (
	VlanStripTypeFilter = "filter"
	VlanStripTypeVlan   = "vlan"
)

type IPFamily struct {
	IPv4 bool
	IPv6 bool
}
