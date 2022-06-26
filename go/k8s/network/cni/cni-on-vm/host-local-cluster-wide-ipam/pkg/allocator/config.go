package allocator

import (
	"encoding/json"
	"fmt"
	"github.com/containernetworking/plugins/pkg/ip"
	"net"
	"time"

	cnitypes "github.com/containernetworking/cni/pkg/types"
)

const (
	IPAMType = "host-local-cluster-wide"

	AddTimeLimit = 2 * time.Minute

	// Allocate operation identifier
	Allocate = 0
	// Deallocate operation identifier
	Deallocate = 1
)

type Net struct {
	Name       string      `json:"name"`
	CNIVersion string      `json:"cniVersion"`
	IPAM       *IPAMConfig `json:"ipam"`
}

type IPAMConfig struct {
	Name       string
	Type       string            `json:"type"`
	Routes     []*cnitypes.Route `json:"routes"`
	ResolvConf string            `json:"resolvConf"` // /etc/resolv.conf

	Range      string `json:"range"`
	RangeStart net.IP `json:"range_start,omitempty"`
	RangeEnd   net.IP `json:"range_end,omitempty"`
	Gateway    net.IP `json:"gateway"`
}

func LoadIPAMConfig(bytes []byte, envArgs string) (*IPAMConfig, string, error) {
	n := Net{}
	if err := json.Unmarshal(bytes, &n); err != nil {
		return nil, "", err
	}

	if n.IPAM == nil {
		return nil, "", fmt.Errorf("IPAM config missing 'ipam' key")
	} else if n.IPAM.Type != IPAMType {
		return nil, "", fmt.Errorf(fmt.Sprintf("ipam type %s is not valid", n.IPAM.Type))
	}

	// Copy net name into IPAM so not to drag Net struct around
	n.IPAM.Name = n.Name

	return n.IPAM, n.CNIVersion, nil
}

type Range struct {
	Subnet     net.IPNet `json:"subnet"`
	RangeStart net.IP    `json:"rangeStart,omitempty"` // The first ip, inclusive
	RangeEnd   net.IP    `json:"rangeEnd,omitempty"`   // The last ip, inclusive
	Gateway    net.IP    `json:"gateway,omitempty"`
}

func NewRange(ipamConf IPAMConfig) (*Range, error) {
	subnet, err := mustSubnet(ipamConf.Range)
	if err != nil {
		return nil, err
	}
	r := &Range{
		Subnet:     subnet,
		RangeStart: ipamConf.RangeStart,
		RangeEnd:   ipamConf.RangeEnd,
		Gateway:    ipamConf.Gateway,
	}

	if err = r.Canonicalize(); err != nil {
		return nil, err
	}

	return r, nil
}

// Canonicalize takes a given range and ensures that all information is consistent,
// filling out Start, End, and Gateway with sane values if missing
func (r *Range) Canonicalize() error {
	if err := canonicalizeIP(&r.Subnet.IP); err != nil {
		return err
	}

	// Can't create an allocator for a network with no addresses, eg
	// a /32 or /31
	ones, masklen := r.Subnet.Mask.Size() // 24, 32
	if ones > masklen-2 {
		return fmt.Errorf("network %s too small to allocate from", (*net.IPNet)(&r.Subnet).String())
	}
	if len(r.Subnet.IP) != len(r.Subnet.Mask) {
		return fmt.Errorf("IPNet IP and Mask version mismatch")
	}

	// Ensure Subnet IP is the network address, not some other address
	networkIP := r.Subnet.IP.Mask(r.Subnet.Mask)
	if !r.Subnet.IP.Equal(networkIP) {
		return fmt.Errorf("network has host bits set. For a subnet mask of length %d the network address is %s", ones, networkIP.String())
	}

	// If the gateway is nil, claim 1
	if r.Gateway == nil {
		r.Gateway = ip.NextIP(r.Subnet.IP) // 10.1.2.0 + 1 = 10.1.2.1
	} else {
		if err := canonicalizeIP(&r.Gateway); err != nil {
			return err
		}
	}

	if r.RangeStart != nil {
		if err := canonicalizeIP(&r.RangeStart); err != nil {
			return err
		}

		if !r.Contains(r.RangeStart) {
			return fmt.Errorf("RangeStart %s not in network %s", r.RangeStart.String(), (*net.IPNet)(&r.Subnet).String())
		}
	} else {
		r.RangeStart = ip.NextIP(r.Subnet.IP)
	}

	if r.RangeEnd != nil {
		if err := canonicalizeIP(&r.RangeEnd); err != nil {
			return err
		}

		if !r.Contains(r.RangeEnd) {
			return fmt.Errorf("RangeEnd %s not in network %s", r.RangeEnd.String(), (*net.IPNet)(&r.Subnet).String())
		}
	} else {
		r.RangeEnd = lastIP(r.Subnet)
	}

	return nil
}

func (r *Range) Contains(addr net.IP) bool {
	if err := canonicalizeIP(&addr); err != nil {
		return false
	}

	subnet := (net.IPNet)(r.Subnet)

	// Not the same address family
	if len(addr) != len(r.Subnet.IP) {
		return false
	}

	// Not in network
	if !subnet.Contains(addr) {
		return false
	}

	// We ignore nils here so we can use this function as we initialize the range.
	if r.RangeStart != nil {
		// Before the range start
		if ip.Cmp(addr, r.RangeStart) < 0 {
			return false
		}
	}

	if r.RangeEnd != nil {
		if ip.Cmp(addr, r.RangeEnd) > 0 {
			// After the  range end
			return false
		}
	}

	return true
}

func (r *Range) String() string {
	return fmt.Sprintf("%s-%s", r.RangeStart.String(), r.RangeEnd.String())
}

// canonicalizeIP makes sure a provided ip is in standard form
func canonicalizeIP(ip *net.IP) error {
	if ip.To4() != nil {
		*ip = ip.To4()
		return nil
	} else if ip.To16() != nil {
		*ip = ip.To16()
		return nil
	}
	return fmt.Errorf("IP %s not v4 nor v6", *ip)
}

// Determine the last IP of a subnet, excluding the broadcast if IPv4
func lastIP(subnet net.IPNet) net.IP {
	var end net.IP
	for i := 0; i < len(subnet.IP); i++ {
		end = append(end, subnet.IP[i]|^subnet.Mask[i])
	}
	if subnet.IP.To4() != nil {
		end[3]--
	}

	return end
}

func mustSubnet(s string) (types.IPNet, error) {
	n, err := types.ParseCIDR(s)
	if err != nil {
		return nil, err
	}
	canonicalizeIP(&n.IP)
	return types.IPNet(*n), nil
}
