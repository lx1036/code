package cidrset

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"net"
	"sync"
)

const (
	clusterSubnetMaxDiff = 16
)

var (
	ErrCIDRSetSubNetTooBig = fmt.Errorf("New CIDR set failed; the node CIDR size is too big")

	ErrCIDRRangeNoCIDRsRemaining = fmt.Errorf("CIDR allocation failed; there are no remaining CIDRs left to allocate in the accepted range")
)

// CidrSet manages a set of CIDR ranges from which blocks of IPs can
// be allocated from.
type CidrSet struct {
	sync.Mutex
	// clusterCIDR is the CIDR assigned to the cluster
	clusterCIDR *net.IPNet
	// clusterMaskSize is the mask size, in bits, assigned to the cluster
	// caches the mask size to avoid the penalty of calling clusterCIDR.Mask.Size()
	clusterMaskSize int
	// nodeMask is the network mask assigned to the nodes
	nodeMask net.IPMask
	// nodeMaskSize is the mask size, in bits,assigned to the nodes
	// caches the mask size to avoid the penalty of calling nodeMask.Size()
	nodeMaskSize int
	// maxCIDRs is the maximum number of CIDRs that can be allocated
	maxCIDRs int
	// allocatedCIDRs counts the number of CIDRs allocated
	allocatedCIDRs int
	// nextCandidate points to the next CIDR that should be free
	nextCandidate int
	// used is a bitmap used to track the CIDRs allocated
	used big.Int
	// label is used to identify the metrics
	label string
}

// NewCIDRSet 100.202.0.0/16, 24
func NewCIDRSet(clusterCIDR *net.IPNet, subNetMaskSize int) (*CidrSet, error) {
	clusterMask := clusterCIDR.Mask
	clusterMaskSize, bits := clusterMask.Size()

	var maxCIDRs int
	if (clusterCIDR.IP.To4() == nil) && (subNetMaskSize-clusterMaskSize > clusterSubnetMaxDiff) {
		return nil, ErrCIDRSetSubNetTooBig
	}

	maxCIDRs = 1 << uint32(subNetMaskSize-clusterMaskSize) // 2^(24-16)
	return &CidrSet{
		clusterCIDR:     clusterCIDR,
		nodeMask:        net.CIDRMask(subNetMaskSize, bits), // 256
		clusterMaskSize: clusterMaskSize,
		maxCIDRs:        maxCIDRs,
		nodeMaskSize:    subNetMaskSize,
		label:           clusterCIDR.String(),
	}, nil
}

// AllocateNext allocates the next free CIDR range. This will set the range
// as occupied and return the allocated range.
func (s *CidrSet) AllocateNext() (*net.IPNet, error) {
	s.Lock()
	defer s.Unlock()

	if s.allocatedCIDRs == s.maxCIDRs {
		return nil, ErrCIDRRangeNoCIDRsRemaining
	}
	candidate := s.nextCandidate
	var i int
	for i = 0; i < s.maxCIDRs; i++ {
		if s.used.Bit(candidate) == 0 {
			break
		}
		candidate = (candidate + 1) % s.maxCIDRs
	}

	s.nextCandidate = (candidate + 1) % s.maxCIDRs // 1 % 256
	s.used.SetBit(&s.used, candidate, 1)
	s.allocatedCIDRs++
	// Update metrics
	cidrSetAllocations.WithLabelValues(s.label).Inc()
	cidrSetAllocationTriesPerRequest.WithLabelValues(s.label).Observe(float64(i))
	cidrSetUsage.WithLabelValues(s.label).Set(float64(s.allocatedCIDRs) / float64(s.maxCIDRs))

	return s.indexToCIDRBlock(candidate), nil
}

func (s *CidrSet) indexToCIDRBlock(index int) *net.IPNet {
	var ip []byte
	switch /*v4 or v6*/ {
	case s.clusterCIDR.IP.To4() != nil:
		{
			j := uint32(index) << uint32(32-s.nodeMaskSize)
			ipInt := (binary.BigEndian.Uint32(s.clusterCIDR.IP)) | j
			ip = make([]byte, net.IPv4len)
			binary.BigEndian.PutUint32(ip, ipInt)
		}
	}

	return &net.IPNet{
		IP:   ip,
		Mask: s.nodeMask, // 256
	}
}

func (s *CidrSet) Release(cidr *net.IPNet) error {
	begin, end, err := s.getBeginingAndEndIndices(cidr)
	if err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()
	for i := begin; i <= end; i++ {
		// Only change the counters if we change the bit to prevent double counting.
		if s.used.Bit(i) != 0 {
			s.used.SetBit(&s.used, i, 0)
			s.allocatedCIDRs--
			cidrSetReleases.WithLabelValues(s.label).Inc()
		}
	}

	cidrSetUsage.WithLabelValues(s.label).Set(float64(s.allocatedCIDRs) / float64(s.maxCIDRs))
	return nil
}

func (s *CidrSet) getBeginingAndEndIndices(cidr *net.IPNet) (begin, end int, err error) {
	if cidr == nil {
		return -1, -1, fmt.Errorf("error getting indices for cluster cidr %v, cidr is nil", s.clusterCIDR)
	}

	begin, end = 0, s.maxCIDRs-1
	maskSize, _ := cidr.Mask.Size()
	var ipSize int
	if !s.clusterCIDR.Contains(cidr.IP.Mask(s.clusterCIDR.Mask)) && !cidr.Contains(s.clusterCIDR.IP.Mask(cidr.Mask)) {
		return -1, -1, fmt.Errorf("cidr %v is out the range of cluster cidr %v", cidr, s.clusterCIDR)
	}

	if s.clusterMaskSize < maskSize { // 16 < 24
		ipSize = net.IPv4len
		if cidr.IP.To4() == nil {
			ipSize = net.IPv6len
		}
		begin, err = s.getIndexForCIDR(&net.IPNet{
			IP:   cidr.IP.Mask(s.nodeMask),
			Mask: s.nodeMask, // 256
		})
		if err != nil {
			return -1, -1, err
		}
		ip := make([]byte, ipSize)
		if cidr.IP.To4() != nil {
			ipInt := binary.BigEndian.Uint32(cidr.IP) | (^binary.BigEndian.Uint32(cidr.Mask))
			binary.BigEndian.PutUint32(ip, ipInt)
		}
		end, err = s.getIndexForCIDR(&net.IPNet{
			IP:   net.IP(ip).Mask(s.nodeMask),
			Mask: s.nodeMask,
		})
		if err != nil {
			return -1, -1, err
		}
	}

	return begin, end, nil
}

func (s *CidrSet) getIndexForCIDR(cidr *net.IPNet) (int, error) {
	ip := cidr.IP
	if ip.To4() != nil {
		cidrIndex := (binary.BigEndian.Uint32(s.clusterCIDR.IP) ^ binary.BigEndian.Uint32(ip.To4())) >> uint32(32-s.nodeMaskSize)
		if cidrIndex >= uint32(s.maxCIDRs) {
			return 0, fmt.Errorf("CIDR: %v/%v is out of the range of CIDR allocator", ip, s.nodeMaskSize)
		}
		return int(cidrIndex), nil
	}

	return 0, fmt.Errorf("invalid IP: %v", ip)
}

func (s *CidrSet) Occupy(cidr *net.IPNet) error {
	begin, end, err := s.getBeginingAndEndIndices(cidr)
	if err != nil {
		return err
	}

	// INFO: 和 Release() 类似
	s.Lock()
	defer s.Unlock()
	for i := begin; i <= end; i++ {
		// Only change the counters if we change the bit to prevent double counting.
		if s.used.Bit(i) == 0 {
			s.used.SetBit(&s.used, i, 1)
			s.allocatedCIDRs++
			cidrSetAllocations.WithLabelValues(s.label).Inc()
		}
	}

	cidrSetUsage.WithLabelValues(s.label).Set(float64(s.allocatedCIDRs) / float64(s.maxCIDRs))
	return nil
}
