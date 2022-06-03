package allocator

import (
	"fmt"
	"github.com/containernetworking/plugins/pkg/ip"
	"k8s.io/klog/v2"
	"net"
	"os"
	"strconv"

	current "github.com/containernetworking/cni/pkg/types/100"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/host-local-ipam/pkg/store"
)

type IPAllocator struct {
	r       *Range
	store   store.Store
	rangeID string // Used for tracking last reserved ip
}

func NewIPAllocator(r *Range, store store.Store, id int) *IPAllocator {
	return &IPAllocator{
		r:       r,
		store:   store,
		rangeID: strconv.Itoa(id),
	}
}

func (alloc *IPAllocator) AllocateNext(containerID, ifName string) (*current.IPConfig, error) {
	alloc.store.Lock()
	defer alloc.store.Unlock()

	var reservedIP *net.IPNet
	var gw net.IP

	allocatedIPs := alloc.store.GetByID(containerID, ifName)
	for _, allocatedIP := range allocatedIPs {
		// check whether the existing IP belong to this range set
		if _, err := alloc.r.RangeFor(allocatedIP); err == nil {
			return nil, fmt.Errorf("%s has been allocated to %s, duplicate allocation is not allowed",
				allocatedIP.String(), containerID)
		}

		iter, err := alloc.GetIter()
		if err != nil {
			return nil, err
		}
		for {
			reservedIP, gw = iter.Next()
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
	}

	if reservedIP == nil {
		return nil, fmt.Errorf("no IP addresses available in range set: %s", alloc.r.String())
	}

	return &current.IPConfig{
		Address: *reservedIP,
		Gateway: gw,
	}, nil
}

func (alloc *IPAllocator) AllocateIP(containerID, ifName string, requestedIP net.IP) (*current.IPConfig, error) {
	if err := canonicalizeIP(&requestedIP); err != nil {
		return nil, err
	}

	r, err := alloc.r.RangeFor(requestedIP)
	if err != nil {
		return nil, err
	}

	if requestedIP.Equal(r.Gateway) {
		return nil, fmt.Errorf("requested ip %s is subnet's gateway", requestedIP.String())
	}

	reserved, err := alloc.store.Reserve(containerID, ifName, requestedIP, alloc.rangeID)
	if err != nil {
		return nil, err
	}
	if !reserved {
		return nil, fmt.Errorf("requested IP address %s is not available in range set %s",
			requestedIP, alloc.r.String())
	}

	return &current.IPConfig{
		Address: net.IPNet{IP: requestedIP, Mask: r.Subnet.Mask},
		Gateway: r.Gateway,
	}, nil
}

func (alloc *IPAllocator) Release(containerID, ifName string) error {
	alloc.store.Lock()
	defer alloc.store.Unlock()

	return alloc.store.ReleaseByID(containerID, ifName)
}

type RangeIterator struct {
	r *Range

	// The current range id
	rangeIdx int

	// Our current position
	cur net.IP

	// The IP where we started iterating; if we hit this again, we're done.
	startIP net.IP
}

// GetIter encapsulates the strategy for this allocator.
// We use a round-robin strategy, attempting to evenly use the whole set.
// More specifically, a crash-looping container will not see the same IP until
// the entire range has been run through.
// We may wish to consider avoiding recently-released IPs in the future.
func (alloc *IPAllocator) GetIter() (*RangeIterator, error) {
	iter := RangeIterator{
		r: alloc.r,
	}

	// Round-robin by trying to allocate from the last reserved IP + 1
	startFromLastReservedIP := false
	// We might get a last reserved IP that is wrong if the range indexes changed.
	// This is not critical, we just lose round-robin this one time.
	lastReservedIP, err := alloc.store.LastReservedIP(alloc.rangeID)
	if err != nil && !os.IsNotExist(err) {
		klog.Info(fmt.Errorf("error retrieving last reserved ip: %v", err))
	} else if lastReservedIP != nil {
		startFromLastReservedIP = alloc.r.Contains(lastReservedIP)
	}

	// Find the range in the set with this IP
	if startFromLastReservedIP {
		if alloc.r.Contains(lastReservedIP) {
			iter.rangeIdx = 0
			// We advance the cursor on every Next(), so the first call
			// to next() will return lastReservedIP + 1
			iter.cur = lastReservedIP
		}
	} else { // 首次分配 ip 从 0 开始
		iter.rangeIdx = 0
		iter.startIP = alloc.r.RangeStart
	}

	return &iter, nil
}

// Next returns the next IP, its mask, and its gateway. Returns nil
// if the iterator has been exhausted
func (i *RangeIterator) Next() (*net.IPNet, net.IP) {
	// If this is the first time iterating and we're not starting in the middle
	// of the range, then start at rangeStart, which is inclusive
	if i.cur == nil {
		i.cur = i.r.RangeStart
		i.startIP = i.cur
		if i.cur.Equal(i.r.Gateway) {
			return i.Next()
		}
		return &net.IPNet{IP: i.cur, Mask: i.r.Subnet.Mask}, i.r.Gateway
	}

	i.cur = ip.NextIP(i.cur)
	if i.cur.Equal(i.r.RangeEnd) {
		return nil, nil // ip is exhausted
	}

	if i.startIP == nil {
		i.startIP = i.cur
	} else if i.cur.Equal(i.startIP) {
		// IF we've looped back to where we started, give up
		return nil, nil
	}

	if i.cur.Equal(i.r.Gateway) {
		return i.Next()
	}

	return &net.IPNet{IP: i.cur, Mask: i.r.Subnet.Mask}, i.r.Gateway
}
