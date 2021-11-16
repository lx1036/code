package options

import (
	"k8s.io/klog/v2"
	"reflect"
	"sort"
	"testing"
)

func TestName(test *testing.T) {
	iBGPPeerCIDRs := make([]string, 0)
	iBGPPeerCIDRs2 := make([]string, 0)

	iBGPPeerCIDRs = append(iBGPPeerCIDRs, []string{"192.168", "192.169"}...)
	iBGPPeerCIDRs2 = append(iBGPPeerCIDRs2, []string{"192.169", "192.168"}...)

	if !reflect.DeepEqual([]string{"192.168.0.0", "192.169.0.0"}, []string{"192.169.0.0", "192.168.0.0"}) {
		klog.Infof("not equal")
	}

	sort.Strings(iBGPPeerCIDRs)
	sort.Strings(iBGPPeerCIDRs2)
	if reflect.DeepEqual(iBGPPeerCIDRs, iBGPPeerCIDRs2) { // "equal"
		klog.Infof("equal")
	} else {
		klog.Infof("not equal")
	}

	type PodIP struct {
		// ip is an IP address (IPv4 or IPv6) assigned to the pod
		IP string `json:"ip,omitempty" protobuf:"bytes,1,opt,name=ip"`
	}
	type PodIPs []*PodIP
	ip1 := PodIP{IP: "192.168"}
	ip2 := PodIP{IP: "192.169"}
	PodIPs1 := PodIPs{&ip1, &ip2}
	PodIPs2 := PodIPs{&ip2, &ip1}
	sort.SliceStable(PodIPs1, func(i, j int) bool {
		return PodIPs1[i].IP < PodIPs1[j].IP
	})
	sort.SliceStable(PodIPs2, func(i, j int) bool {
		return PodIPs2[i].IP < PodIPs2[j].IP
	})
	if reflect.DeepEqual(PodIPs1, PodIPs2) { // "not equal"
		klog.Infof("equal")
		klog.Info(PodIPs1, PodIPs2)
	} else {
		klog.Infof("not equal")
	}
}
