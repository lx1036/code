package allocator

import (
	"fmt"
	"net"

	"k8s-lx1036/k8s/network/cni/vpc-cni/host-local-cluster-wide-ipam/pkg/store/kubernetes"

	current "github.com/containernetworking/cni/pkg/types/100"
	"github.com/containernetworking/plugins/pkg/ip"
)

// IPReservation is an address that has been reserved by this plugin
type IPReservation struct {
	IP          net.IP `json:"ip"`
	ContainerID string `json:"id"`
	PodRef      string `json:"podref,omitempty"`
	IsAllocated bool
}

func (ir IPReservation) String() string {
	return fmt.Sprintf("IP: %s is reserved for pod: %s", ir.IP.String(), ir.PodRef)
}

type IPAllocator struct {
	store kubernetes.Store

	r *Range

	// Our current position
	cur net.IP

	// The IP where we started iterating; if we hit this again, we're done.
	startIP net.IP
}

func NewIPAllocator(ipamConf IPAMConfig) {
	r, err := NewRange(ipamConf)
	ipAllocator := &IPAllocator{
		r:       r,
		startIP: r.RangeStart,
	}
}

func (alloc *IPAllocator) AllocateNext(containerID, ifName string) (*current.IPConfig, error) {
	var reservedIP *net.IPNet
	var gw net.IP

	// if already allocated for pod
	allocatedIPs := alloc.store.GetByID(containerID, ifName)
	for _, allocatedIP := range allocatedIPs {

	}

	for {
		reservedIP, gw = alloc.Next()
		if reservedIP == nil {
			break
		}

		reserved, err := alloc.store.Reserve(containerID, ifName, reservedIP.IP, alloc.rangeID)
		if err != nil {
			return nil, err
		}

		if reserved {
			break
		}

	}

	if reservedIP == nil {
		return nil, fmt.Errorf("no IP addresses available in range set: %s", alloc.r.String())
	}

	return &current.IPConfig{
		Address: *reservedIP,
		Gateway: gw,
	}, nil
}

// Next returns the next IP, its mask, and its gateway. Returns nil
// if the iterator has been exhausted
// 并且排除了 RangeEnd/Gateway IP
func (alloc *IPAllocator) Next() (*net.IPNet, net.IP) {
	r := alloc.r
	// If this is the first time iterating and we're not starting in the middle
	// of the range, then start at rangeStart, which is inclusive
	if alloc.cur == nil {
		alloc.cur = r.RangeStart
		alloc.startIP = alloc.cur
		if alloc.cur.Equal(r.Gateway) {
			return alloc.Next()
		}
		return &net.IPNet{IP: alloc.cur, Mask: r.Subnet.Mask}, r.Gateway
	}

	if alloc.cur.Equal(r.RangeEnd) {
		// IF we've looped back to where we started, ip is exhausted
		return nil, nil
	} else {
		alloc.cur = ip.NextIP(alloc.cur)
	}

	if alloc.startIP == nil {
		alloc.startIP = alloc.cur
	} else if alloc.cur.Equal(alloc.startIP) {
		// IF we've looped back to where we started, ip is exhausted
		return nil, nil
	}

	if alloc.cur.Equal(r.Gateway) {
		return alloc.Next()
	}

	return &net.IPNet{IP: alloc.cur, Mask: r.Subnet.Mask}, r.Gateway
}

// AssignIP assigns an IP using a range and a reserve list.
func AssignIP(ipamConf IPAMConfig, reservelist []IPReservation, containerID string,
	podRef string) (*net.IPNet, []IPReservation, error) {
	_, ipnet, _ := net.ParseCIDR(ipamConf.Range)

	r, err := NewRange(ipamConf)
	if err != nil {
		return nil, nil, err
	}

	newip, updatedreservelist, err := IterateForAssignment(r, reservelist, containerID, podRef)
	if err != nil {
		return nil, nil, err
	}

	return &net.IPNet{IP: newip, Mask: ipnet.Mask}, updatedreservelist, nil
}

// DeallocateIP assigns an IP using a range and a reserve list.
func DeallocateIP(reservelist []IPReservation, containerID string) ([]IPReservation, net.IP, error) {

}

// IterateForAssignment iterates given an IP/IPNet and a list of reserved IPs
func IterateForAssignment(r *Range, reservelist []IPReservation, containerID string, podRef string) (net.IP, []IPReservation, error) {
	var (
		firstip net.IP
		lastip  net.IP
	)

	reserved := make(map[string]bool)
	for _, r := range reservelist {
		reserved[r.IP.String()] = true
	}

	// Iterate every IP address in the range
	var assignedip net.IP
	performedassignment := false
	endip := IPAddOffset(lastip, uint64(1))
	for i := firstip; !i.Equal(endip); i = ip.NextIP(i) {
		// if already reserved, skip it
		if reserved[i.String()] {
			continue
		}

		// Ok, this one looks like we can assign it!
		performedassignment = true

		assignedip = i
		reservelist = append(reservelist, IPReservation{IP: assignedip, ContainerID: containerID, PodRef: podRef})
		break
	}

	if !performedassignment {
		return net.IP{}, reservelist, AssignmentError{firstip, lastip, ipnet}
	}

	return assignedip, reservelist, nil
}
