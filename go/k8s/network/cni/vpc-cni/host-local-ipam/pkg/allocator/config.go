package allocator

import (
	"encoding/json"
	"fmt"
	"github.com/containernetworking/plugins/pkg/ip"
	"k8s.io/klog/v2"
	"net"
	"strings"

	"github.com/containernetworking/cni/pkg/types"
)

const (
	IPAMType = "host-local"
)

type Range struct {
	Subnet     types.IPNet `json:"subnet"`
	RangeStart net.IP      `json:"rangeStart,omitempty"` // The first ip, inclusive
	RangeEnd   net.IP      `json:"rangeEnd,omitempty"`   // The last ip, inclusive
	Gateway    net.IP      `json:"gateway,omitempty"`
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

func (r *Range) Overlaps(r1 *Range) bool {
	if len(r.RangeStart) != len(r1.RangeStart) {
		return false
	}

	return r.Contains(r1.RangeStart) ||
		r.Contains(r1.RangeEnd) ||
		r1.Contains(r.RangeStart) ||
		r1.Contains(r.RangeEnd)
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

func (r *Range) RangeFor(addr net.IP) (*Range, error) {
	if err := canonicalizeIP(&addr); err != nil {
		return nil, err
	}

	if !r.Contains(addr) {
		return nil, fmt.Errorf("%s not in range set %s", addr.String(), r.String())
	}

	return r, nil
}

func (r *Range) String() string {
	return fmt.Sprintf("%s-%s", r.RangeStart.String(), r.RangeEnd.String())
}

type RangeSet []Range

func (s *RangeSet) Canonicalize() error {
	if len(*s) == 0 {
		return fmt.Errorf("empty range set")
	}

	fam := 0
	for i := range *s {
		if err := (*s)[i].Canonicalize(); err != nil {
			return err
		}
		if i == 0 {
			fam = len((*s)[i].RangeStart)
		} else {
			if fam != len((*s)[i].RangeStart) {
				return fmt.Errorf("mixed address families")
			}
		}
	}

	// Make sure none of the ranges in the set overlap
	l := len(*s)
	for i, r1 := range (*s)[:l-1] {
		for _, r2 := range (*s)[i+1:] {
			if r1.Overlaps(&r2) {
				return fmt.Errorf("subnets %s and %s overlap", r1.String(), r2.String())
			}
		}
	}

	return nil
}

func (s *RangeSet) RangeFor(addr net.IP) (*Range, error) {
	if err := canonicalizeIP(&addr); err != nil {
		return nil, err
	}

	for _, r := range *s {
		if r.Contains(addr) {
			return &r, nil
		}
	}

	return nil, fmt.Errorf("%s not in range set %s", addr.String(), s.String())
}

func (s *RangeSet) Contains(addr net.IP) bool {
	r, _ := s.RangeFor(addr)
	return r != nil
}

func (s *RangeSet) String() string {
	out := []string{}
	for _, r := range *s {
		out = append(out, r.String())
	}

	return strings.Join(out, ",")
}

type IPAMArgs struct {
	IPs []*ip.IP `json:"ips"`
}
type IPAMConfig struct {
	Name       string
	Type       string         `json:"type"`
	Routes     []*types.Route `json:"routes"`
	DataDir    string         `json:"dataDir"`
	ResolvConf string         `json:"resolvConf"` // /etc/resolv.conf
	Ranges     RangeSet       `json:"ranges"`
	IPArgs     []net.IP       `json:"-"` // Requested IPs from CNI_ARGS, args and capabilities
}
type Net struct {
	Name          string      `json:"name"`
	CNIVersion    string      `json:"cniVersion"`
	IPAM          *IPAMConfig `json:"ipam"`
	RuntimeConfig struct {
		// The capability arg
		IPRanges RangeSet `json:"ipRanges,omitempty"`
		IPs      []*ip.IP `json:"ips,omitempty"`
	} `json:"runtimeConfig,omitempty"`
	Args *struct {
		A *IPAMArgs `json:"cni"`
	} `json:"args"`
}
type IPAMEnvArgs struct {
	types.CommonArgs
	IP ip.IP `json:"ip,omitempty"`
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

	if envArgs != "" {
		e := IPAMEnvArgs{}
		err := types.LoadArgs(envArgs, &e)
		if err != nil {
			return nil, "", err
		}

		if e.IP.ToIP() != nil {
			n.IPAM.IPArgs = []net.IP{e.IP.ToIP()}
		}
	}
	// parse custom IPs from CNI args in network config
	if n.Args != nil && n.Args.A != nil && len(n.Args.A.IPs) != 0 {
		for _, i := range n.Args.A.IPs {
			n.IPAM.IPArgs = append(n.IPAM.IPArgs, i.ToIP())
		}
	}
	// parse custom IPs from runtime configuration
	if len(n.RuntimeConfig.IPs) > 0 {
		for _, i := range n.RuntimeConfig.IPs {
			n.IPAM.IPArgs = append(n.IPAM.IPArgs, i.ToIP())
		}
	}
	for idx := range n.IPAM.IPArgs {
		if err := canonicalizeIP(&n.IPAM.IPArgs[idx]); err != nil {
			return nil, "", fmt.Errorf("cannot understand ip: %v", err)
		}
	}

	// If a range is supplied as a runtime config, prepend it to the Ranges
	if len(n.RuntimeConfig.IPRanges) > 0 {
		n.IPAM.Ranges = append(n.RuntimeConfig.IPRanges, n.IPAM.Ranges...)
	}
	if len(n.IPAM.Ranges) == 0 {
		return nil, "", fmt.Errorf("no IP ranges specified")
	}

	// Validate all ranges
	for i := range n.IPAM.Ranges {
		if err := n.IPAM.Ranges[i].Canonicalize(); err != nil { // INFO: 注意这里用的是 n.IPAM.Ranges[i]
			return nil, "", fmt.Errorf("invalid range set %d: %s", i, err)
		}
	}

	// Check for overlaps
	l := len(n.IPAM.Ranges)
	for i, p1 := range n.IPAM.Ranges[:l-1] {
		for j, p2 := range n.IPAM.Ranges[i+1:] {
			if p1.Overlaps(&p2) {
				return nil, "", fmt.Errorf("range set %d overlaps with %d", i, (i + j + 1))
			}
		}
	}

	// Copy net name into IPAM so not to drag Net struct around
	n.IPAM.Name = n.Name

	return n.IPAM, n.CNIVersion, nil
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
func lastIP(subnet types.IPNet) net.IP {
	var end net.IP
	for i := 0; i < len(subnet.IP); i++ {
		end = append(end, subnet.IP[i]|^subnet.Mask[i])
	}
	if subnet.IP.To4() != nil {
		end[3]--
	}

	return end
}

func mustSubnet(s string) types.IPNet {
	n, err := types.ParseCIDR(s)
	if err != nil {
		klog.Fatal(err)
	}
	canonicalizeIP(&n.IP)
	return types.IPNet(*n)
}
