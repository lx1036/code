package cidrset

import (
	"fmt"
	"net"
	"testing"

	"k8s.io/klog/v2"
)

func TestCIDRSetFullyAllocated(t *testing.T) {
	fixtures := []struct {
		clusterCIDRStr string
		subNetMaskSize int
		expectedCIDR   string
		description    string
	}{
		{
			clusterCIDRStr: "127.123.1.0/24",
			subNetMaskSize: 26,
			expectedCIDR:   "127.123.1.0/26", // 127.123.1.0/26, 127.123.1.64/26, 127.123.1.128/26, 127.123.1.192/26
			description:    "Fully allocated CIDR with IPv4",
		},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.description, func(t *testing.T) {
			ip, clusterCIDR, _ := net.ParseCIDR(fixture.clusterCIDRStr)
			klog.Infof(fmt.Sprintf("ip:%s, clusterCIDR:%s", ip.String(), clusterCIDR.String()))
			cidrSet, err := NewCIDRSet(clusterCIDR, fixture.subNetMaskSize)
			if err != nil {
				t.Fatalf("unexpected error: %v for %v", err, fixture.description)
			}

			// INFO: Allocate
			cidr, err := cidrSet.AllocateNext() // 127.123.1.0/26
			if err != nil {
				t.Fatalf("unexpected error: %v for %v", err, fixture.description)
			}
			if cidr.String() != fixture.expectedCIDR {
				t.Fatalf("unexpected allocated cidr: %v, expecting %v for %v",
					cidr.String(), fixture.expectedCIDR, fixture.description)
			}

			cidr1, err := cidrSet.AllocateNext() // 127.123.1.64/26
			if err != nil {
				t.Fatalf("expected error because of fully-allocated range for %v", fixture.description)
			}
			if cidr1.String() != "127.123.1.64/26" {
				t.Fatalf("unexpected allocated cidr: %v, expecting 127.123.1.0/24 for %v",
					cidr1.String(), fixture.description)
			}

			cidr, err = cidrSet.AllocateNext() // 127.123.1.128/26
			if err != nil {
				t.Fatalf("expected error because of fully-allocated range for %v", fixture.description)
			}
			if cidr.String() != "127.123.1.128/26" {
				t.Fatalf("unexpected allocated cidr: %v, expecting 127.123.1.0/24 for %v",
					cidr.String(), fixture.description)
			}

			// INFO: Release
			err = cidrSet.Release(cidr1) // release "127.123.1.64/26"
			if err != nil {
				t.Fatalf(fmt.Sprintf("release cidr %s err:%v", cidr1.String(), err))
			}
			cidr, err = cidrSet.AllocateNext() // 127.123.1.192/26，注意这里是继续 next allocate 127.123.1.192/26，而不是 127.123.1.64/26
			if err != nil {
				t.Fatalf("expected error because of fully-allocated range for %v", fixture.description)
			}
			if cidr.String() != "127.123.1.192/26" {
				t.Fatalf("unexpected allocated cidr: %v, expecting 127.123.1.0/24 for %v",
					cidr.String(), fixture.description)
			}
			cidr, err = cidrSet.AllocateNext() // 127.123.1.64/26, 一次循环回来，allocate 127.123.1.64/26
			if err != nil {
				t.Fatalf("expected error because of fully-allocated range for %v", fixture.description)
			}
			if cidr.String() != "127.123.1.64/26" {
				t.Fatalf("unexpected allocated cidr: %v, expecting 127.123.1.0/24 for %v",
					cidr.String(), fixture.description)
			}
		})
	}
}

func TestOccupy(t *testing.T) {
	fixtures := []struct {
		clusterCIDRStr string
		subNetMaskSize int
		expectedCIDR   string
		description    string
	}{
		{
			clusterCIDRStr: "127.123.1.0/24",
			subNetMaskSize: 26,
			expectedCIDR:   "127.123.1.0/26", // 127.123.1.0/26, 127.123.1.64/26, 127.123.1.128/26, 127.123.1.192/26
			description:    "Fully allocated CIDR with IPv4",
		},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.description, func(t *testing.T) {
			ip, clusterCIDR, _ := net.ParseCIDR(fixture.clusterCIDRStr)
			klog.Infof(fmt.Sprintf("ip:%s, clusterCIDR:%s", ip.String(), clusterCIDR.String()))
			cidrSet, err := NewCIDRSet(clusterCIDR, fixture.subNetMaskSize)
			if err != nil {
				t.Fatalf("unexpected error: %v for %v", err, fixture.description)
			}

			_, ipnet, _ := net.ParseCIDR("127.123.1.32/27") // INFO: 这里居然可以 Occupy "127.123.1.32/27"，不符合预期
			if cidrSet.InRange(ipnet) {
				klog.Infof(fmt.Sprintf("%s is in range", ipnet.String()))
			}
			if err = cidrSet.Occupy(ipnet); err != nil {
				t.Fatal(err)
			}

			cidr, err := cidrSet.AllocateNext()
			if err != nil {
				t.Fatal(err)
			}
			klog.Infof(fmt.Sprintf("%s", cidr.String()))
		})
	}
}
