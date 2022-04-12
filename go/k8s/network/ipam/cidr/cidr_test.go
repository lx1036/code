package cidr

import (
	"fmt"
	"k8s.io/klog/v2"
	"net"
	"testing"
)

func TestIPRange(test *testing.T) {
	c1, _ := ParseCIDR("100.217.144.0/20")
	start, end := c1.IPRange()
	fmt.Println(start, end, c1.Gateway()) // 100.217.144.0 100.217.159.255 100.217.144.1

	_, ipnet1, _ := net.ParseCIDR("100.217.144.0/20")
	fmt.Println(ipnet1.Mask.String())
	ones, bits := ipnet1.Mask.Size()
	fmt.Println(ones, bits, len(ipnet1.Mask))
}

func TestSubNetting(t *testing.T) {
	c1, _ := ParseCIDR("100.217.144.0/20")
	cs1, _ := c1.SubNetting(SUBNETTING_METHOD_SUBNET_NUM, 2)
	fmt.Println(c1.CIDR(), "按子网数量划分:")
	for _, c := range cs1 {
		fmt.Println(c.CIDR())
	}
	/* 两个
	100.217.144.0/21
	100.217.152.0/21
	*/

	/*c2, _ := ParseCIDR("2001:db8::/64")
	cs2, _ := c2.SubNetting(SUBNETTING_METHOD_SUBNET_NUM, 4)
	fmt.Println(c2.CIDR(), "按子网数量划分:")
	for _, c := range cs2 {
		fmt.Println(c.CIDR())
	}*/

	c3, _ := ParseCIDR("127.123.1.0/24")                    // 我要划分成 /26 pod cidr
	cs3, _ := c3.SubNetting(SUBNETTING_METHOD_HOST_NUM, 64) // 这里64是 2^(32-26)
	fmt.Println(c3.CIDR(), "按主机数量划分:")
	for _, c := range cs3 {
		fmt.Println(c.CIDR())
	}

	/* 32 个
	100.217.152.0/26
	100.217.152.64/26
	100.217.152.128/26
	100.217.152.192/26
	100.217.153.0/26
	100.217.153.64/26
	100.217.153.128/26
	100.217.153.192/26
	100.217.154.0/26
	100.217.154.64/26
	100.217.154.128/26
	100.217.154.192/26
	100.217.155.0/26
	100.217.155.64/26
	100.217.155.128/26
	100.217.155.192/26
	100.217.156.0/26
	100.217.156.64/26
	100.217.156.128/26
	100.217.156.192/26
	100.217.157.0/26
	100.217.157.64/26
	100.217.157.128/26
	100.217.157.192/26
	100.217.158.0/26
	100.217.158.64/26
	100.217.158.128/26
	100.217.158.192/26
	100.217.159.0/26
	100.217.159.64/26
	100.217.159.128/26
	100.217.159.192/26
	*/
}

func TestForEach(test *testing.T) {
	c1, _ := ParseCIDR("100.216.137.0/25")
	c1.ForEachIP(func(ip string) error {
		fmt.Println(ip)
		return nil
	})

	_, ipnet, _ := net.ParseCIDR("100.216.137.0/25")
	_, ipnet2, _ := net.ParseCIDR("100.216.137.0/25")
	if ipnet.String() == ipnet2.String() {
		klog.Info("equal")
	}
	ones, bit := ipnet.Mask.Size()
	klog.Info(ones, bit)
	klog.Info(ipnet.Mask.String())
}
