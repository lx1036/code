package types

import (
	"fmt"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/rpc"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/utils"
	"net"
	"strings"
	"time"
)

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
