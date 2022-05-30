package kube_proxy

import (
	"k8s.io/klog/v2"
	"reflect"
	"sort"
	"testing"
)

const (
	FlagPersistent = 0x1
	FlagHashed     = 0x2
)

func TestProxy(test *testing.T) {
	args := []string{"-m", "comment"}
	klog.Info(args[:0])

	//flags := FlagPersistent + FlagHashed
	flags := FlagHashed
	flags |= FlagPersistent // flags 不管是不是两个都有，都要加上 FlagPersistent
	klog.Info(flags)

	type svc struct {
		name []string
	}
	s := svc{name: []string{"test1"}}
	type svcInfo struct {
		name []string
	}
	info := &svcInfo{
		name: s.name,
	}
	info.name = []string{"test2"}
	klog.Info(s.name) // test1
}

type endpointInfo struct {
	ip      string
	port    int
	isLocal bool
}
type endpointInfoMap map[string][]endpointInfo

func (ep endpointInfoMap) equal(other endpointInfoMap) bool {
	if len(ep) != len(other) {
		return false
	}

	for epID, infos := range ep {
		otherInfos, ok := other[epID]
		if !ok || len(otherInfos) != len(infos) {
			return false
		}
		sort.SliceStable(infos, func(i, j int) bool {
			return infos[i].port < infos[j].port
		})
		sort.SliceStable(otherInfos, func(i, j int) bool {
			return otherInfos[i].port < otherInfos[j].port
		})
		if !reflect.DeepEqual(infos, otherInfos) {
			return false
		}
	}

	return true
}
func TestEqual(test *testing.T) {
	a1 := []endpointInfo{
		{
			ip:      "127.0.0.1",
			port:    80,
			isLocal: true,
		},
		{
			ip:      "127.0.0.2",
			port:    81,
			isLocal: false,
		},
	}
	a2 := []endpointInfo{
		{
			ip:      "127.0.0.2",
			port:    81,
			isLocal: false,
		},
		{
			ip:      "127.0.0.1",
			port:    80,
			isLocal: true,
		},
	}

	a := make(endpointInfoMap)
	a["a1"] = a1

	b := make(endpointInfoMap)
	b["a1"] = a2

	if a.equal(b) {
		klog.Info("success")
	}
}
