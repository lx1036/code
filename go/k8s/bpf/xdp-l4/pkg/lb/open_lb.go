package lb

import (
	"fmt"
	"net"

	"k8s-lx1036/k8s/bpf/xdp-l4/pkg/rpc"

	"github.com/cilium/ebpf"
	"github.com/sirupsen/logrus"
)

type AddressType int

const (
	INVALID AddressType = iota
	HOST
	NETWORK
)

const (
	V6DADDR = uint8(1)
)

type OpenLbConfig struct {
	disableForwarding bool
	testing           bool

	maxReals   uint32
	maxVips    uint32
	chRingSize uint32

	// mac address of default gateway: {0x00, 0x00, 0xDE, 0xAD, 0xBE, 0xAF}
	// ip route | grep default
	// ip n show | grep 10.0.2.2
	defaultMac         [6]uint8
	mainInterfaceIndex uint32
	mainInterface      string // eth0
	hcInterfaceIndex   uint32
	hcInterface        string // eth0
	v4TunInterface     string // ipip0
	v6TunInterface     string

	enableHc           bool
	tunnelBasedHCEncap bool
}

var defaultOpenLbConfig = &OpenLbConfig{
	disableForwarding: false,
	testing:           false,
	maxReals:          4096,
	maxVips:           512,
	chRingSize:        DefaultChRingSize,
}

type OpenLbFeatures struct {
	srcRouting                bool
	inlineDecap               bool
	introspection             bool
	gueEncap                  bool
	directHealthchecking      bool
	localDeliveryOptimization bool
	flowDebug                 bool
}

type OpenLb struct {
	rpc.UnimplementedOpenLbServiceServer

	config *OpenLbConfig

	stats    OpenLbStats
	features OpenLbFeatures

	vipNums    []uint32
	realNums   []uint32
	vips       map[VipKey]Vip
	numToReals map[uint32]IPAddress
	reals      map[*IPAddress]*RealMeta

	/**
	 * Callback to be notified when a real is added or deleted
	 */
	realsIdCallback RealsIdCallback

	// "vip_map" bpf map
	vipMap *ebpf.Map
	// "reals" bpf map
	realsMap    *ebpf.Map
	statsMap    *ebpf.Map
	ctlArrayMap *ebpf.Map
	chRingsMap  *ebpf.Map

	ctlValues map[int]CtlValue
}

func NewOpenLb() (*OpenLb, error) {
	lb := &OpenLb{}

	if !lb.config.testing {
		var ctlValue CtlValue
		for index, value := range lb.config.defaultMac {
			ctlValue.mac[index] = value
		}

		lb.ctlValues[MacAddrPos] = ctlValue

		if lb.config.enableHc {
			ifindex := lb.config.hcInterfaceIndex
			if ifindex == 0 {
				ifindex = getInterfaceIndex(lb.config.hcInterface)
				if ifindex == 0 {
					return nil, fmt.Errorf("can't resolve ifindex for healthcheck interface %s", lb.config.mainInterface)
				}
			}

			ctlValue.ifindex = ifindex
			lb.ctlValues[HcIntfPos] = ctlValue
			if lb.config.tunnelBasedHCEncap {
				ifindex = getInterfaceIndex(lb.config.v4TunInterface)
				if ifindex == 0 {
					return nil, fmt.Errorf("can't resolve ifindex for v4tunel interface %s", lb.config.v4TunInterface)
				}
				ctlValue.ifindex = ifindex
				lb.ctlValues[Ipv4TunPos] = ctlValue

				ifindex = getInterfaceIndex(lb.config.v6TunInterface)
				if ifindex == 0 {
					return nil, fmt.Errorf("can't resolve ifindex for v6tunel interface %s", lb.config.v6TunInterface)
				}
				ctlValue.ifindex = ifindex
				lb.ctlValues[Ipv6TunPos] = ctlValue
			}
		}

		ifindex := lb.config.mainInterfaceIndex
		if ifindex == 0 {
			// attempt to resolve interface name to the interface index
			ifindex = getInterfaceIndex(lb.config.mainInterface)
			if ifindex == 0 {
				return nil, fmt.Errorf("can't resolve ifindex for main interface %s", lb.config.mainInterface)
			}
		}
		ctlValue.ifindex = ifindex
		lb.ctlValues[MainIntfPos] = ctlValue
	}

	return lb, nil
}

func (lb *OpenLb) validateAddress(address string, allowNetAddr bool) AddressType {
	if ip := net.ParseIP(address); ip == nil {
		if allowNetAddr && (lb.features.srcRouting || lb.config.testing) {
			// TODO
			return NETWORK
		}

		lb.stats.bpfFailedCalls++
		logrus.Errorf("Invalid address: %s", address)
		return INVALID
	}

	return HOST
}

func (lb *OpenLb) LoadBpfProgs() error {

	// add values to main prog ctl_array
	if !lb.config.disableForwarding {
		ctlKey := MacAddrPos
		ctlValue := lb.ctlValues[ctlKey]
		if err := lb.ctlArrayMap.Update(&ctlKey, &ctlValue, ebpf.UpdateAny); err != nil {
			return fmt.Errorf("can't update ctl array for main program, error: %v", err)
		}
	}

}

func getInterfaceIndex(mainInterface string) uint32 {

}
