package allocator

import (
	"fmt"
	"net"

	fakestore "k8s-lx1036/k8s/network/cni/cni-on-vm/host-local-ipam/pkg/store/testing"

	"github.com/containernetworking/cni/pkg/types"
	current "github.com/containernetworking/cni/pkg/types/100"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/klog/v2"
)

func mockAllocator() IPAllocator {
	s := RangeSet{
		Range{
			Subnet: mustSubnet("192.168.1.0/29"),
		},
	}
	if err := s.Canonicalize(); err != nil {
		klog.Fatal(err)
	}

	store := fakestore.NewFakeStore(map[string]string{}, map[string]net.IP{})
	alloc := IPAllocator{
		rangeset: &s,
		store:    store,
		rangeID:  "rangeid",
	}

	return alloc
}

type AllocatorTestCase struct {
	subnets      []string
	ipmap        map[string]string
	expectResult string
	lastIP       string
}

func (t AllocatorTestCase) run(idx int) (*current.IPConfig, error) {
	fmt.Fprintln(GinkgoWriter, "Index:", idx)
	p := RangeSet{}
	for _, s := range t.subnets {
		subnet, err := types.ParseCIDR(s)
		if err != nil {
			return nil, err
		}
		p = append(p, Range{Subnet: types.IPNet(*subnet)})
	}

	Expect(p.Canonicalize()).To(BeNil())

	store := fakestore.NewFakeStore(t.ipmap, map[string]net.IP{"rangeid": net.ParseIP(t.lastIP)})

	alloc := IPAllocator{
		rangeset: &p,
		store:    store,
		rangeID:  "rangeid",
	}

	return alloc.AllocateNext("ID", "eth0")
}

var _ = Describe("host-local ip allocator", func() {
	Context("RangeIterator", func() {
		It("should loop correctly from the beginning", func() {
			ipAllocator := mockAllocator()
			ipIterator, err := ipAllocator.GetIter()
			if err != nil {
				klog.Fatal(err)
			}
			Expect(ipIterator.nextip()).To(Equal(net.IP{192, 168, 1, 2}))
			Expect(ipIterator.nextip()).To(Equal(net.IP{192, 168, 1, 3}))
			Expect(ipIterator.nextip()).To(Equal(net.IP{192, 168, 1, 4}))
			Expect(ipIterator.nextip()).To(Equal(net.IP{192, 168, 1, 5}))
			Expect(ipIterator.nextip()).To(Equal(net.IP{192, 168, 1, 6}))
			Expect(ipIterator.nextip()).To(BeNil())
		})

		It("should loop correctly from the end", func() {
			ipAllocator := mockAllocator()
			ipAllocator.store.Reserve("ID", "eth0", net.IP{192, 168, 1, 6}, ipAllocator.rangeID)
			ipAllocator.store.ReleaseByID("ID", "eth0")
			r, _ := ipAllocator.GetIter()
			Expect(r.nextip()).To(Equal(net.IP{192, 168, 1, 2}))
			Expect(r.nextip()).To(Equal(net.IP{192, 168, 1, 3}))
			Expect(r.nextip()).To(Equal(net.IP{192, 168, 1, 4}))
			Expect(r.nextip()).To(Equal(net.IP{192, 168, 1, 5}))
			Expect(r.nextip()).To(Equal(net.IP{192, 168, 1, 6}))
			Expect(r.nextip()).To(BeNil())
		})

		It("should loop correctly from the middle", func() {
			ipAllocator := mockAllocator()
			ipAllocator.store.Reserve("ID", "eth0", net.IP{192, 168, 1, 3}, ipAllocator.rangeID)
			ipAllocator.store.ReleaseByID("ID", "eth0")
			r, _ := ipAllocator.GetIter()
			Expect(r.nextip()).To(Equal(net.IP{192, 168, 1, 4}))
			Expect(r.nextip()).To(Equal(net.IP{192, 168, 1, 5}))
			Expect(r.nextip()).To(Equal(net.IP{192, 168, 1, 6}))
			Expect(r.nextip()).To(Equal(net.IP{192, 168, 1, 2}))
			Expect(r.nextip()).To(Equal(net.IP{192, 168, 1, 3}))
			Expect(r.nextip()).To(BeNil())
		})
	})

	Context("when has free ip", func() {
		It("should allocate ips in round robin", func() {
			testCases := []AllocatorTestCase{
				// fresh start
				{
					subnets:      []string{"10.0.0.0/29"},
					ipmap:        map[string]string{},
					expectResult: "10.0.0.2",
					lastIP:       "",
				},
				{
					subnets:      []string{"2001:db8:1::0/64"},
					ipmap:        map[string]string{},
					expectResult: "2001:db8:1::2",
					lastIP:       "",
				},
				{
					subnets:      []string{"10.0.0.0/30"},
					ipmap:        map[string]string{},
					expectResult: "10.0.0.2",
					lastIP:       "",
				},
				{
					subnets: []string{"10.0.0.0/29"},
					ipmap: map[string]string{
						"10.0.0.2": "id",
					},
					expectResult: "10.0.0.3",
					lastIP:       "",
				},
				// next ip of last reserved ip
				{
					subnets:      []string{"10.0.0.0/29"},
					ipmap:        map[string]string{},
					expectResult: "10.0.0.6",
					lastIP:       "10.0.0.5",
				},
				{
					subnets: []string{"10.0.0.0/29"},
					ipmap: map[string]string{
						"10.0.0.4": "id",
						"10.0.0.5": "id",
					},
					expectResult: "10.0.0.6",
					lastIP:       "10.0.0.3",
				},
				// round robin to the beginning
				{
					subnets: []string{"10.0.0.0/29"},
					ipmap: map[string]string{
						"10.0.0.6": "id",
					},
					expectResult: "10.0.0.2",
					lastIP:       "10.0.0.5",
				},
				// lastIP is out of range
				{
					subnets: []string{"10.0.0.0/29"},
					ipmap: map[string]string{
						"10.0.0.2": "id",
					},
					expectResult: "10.0.0.3",
					lastIP:       "10.0.0.128",
				},
				// subnet is completely full except for lastip
				// wrap around and reserve lastIP
				{
					subnets: []string{"10.0.0.0/29"},
					ipmap: map[string]string{
						"10.0.0.2": "id",
						"10.0.0.4": "id",
						"10.0.0.5": "id",
						"10.0.0.6": "id",
					},
					expectResult: "10.0.0.3",
					lastIP:       "10.0.0.3",
				},
				// allocate from multiple subnets
				{
					subnets:      []string{"10.0.0.0/30", "10.0.1.0/30"},
					expectResult: "10.0.0.2",
					ipmap:        map[string]string{},
				},
				// advance to next subnet
				{
					subnets:      []string{"10.0.0.0/30", "10.0.1.0/30"},
					lastIP:       "10.0.0.2",
					expectResult: "10.0.1.2",
					ipmap:        map[string]string{},
				},
				// Roll to start subnet
				{
					subnets:      []string{"10.0.0.0/30", "10.0.1.0/30", "10.0.2.0/30"},
					lastIP:       "10.0.2.2",
					expectResult: "10.0.0.2",
					ipmap:        map[string]string{},
				},
			}

			for idx, tc := range testCases {
				res, err := tc.run(idx)
				Expect(err).ToNot(HaveOccurred())
				Expect(res.Address.IP.String()).To(Equal(tc.expectResult))
			}
		})

	})

})

// nextip is a convenience function used for testing
func (i *RangeIterator) nextip() net.IP {
	c, _ := i.Next()
	if c == nil {
		return nil
	}

	return c.IP
}
