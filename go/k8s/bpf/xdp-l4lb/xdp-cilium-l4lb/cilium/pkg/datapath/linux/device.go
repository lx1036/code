package linux

import (
    "github.com/vishvananda/netlink"
    "github.com/vishvananda/netns"

    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/lock"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/option"
)

type deviceFilter []string

type DeviceManager struct {
    lock.Mutex
    devices map[string]struct{}
    filter  deviceFilter
    handle  *netlink.Handle
    netns   netns.NsHandle
}

func NewDeviceManager() (*DeviceManager, error) {
    return NewDeviceManagerAt(netns.None())
}

func NewDeviceManagerAt(netns netns.NsHandle) (*DeviceManager, error) {
    handle, err := netlink.NewHandleAt(netns)
    if err != nil {
        return nil, err
    }
    return &DeviceManager{
        devices: make(map[string]struct{}),
        filter:  deviceFilter(option.Config.GetDevices()),
        handle:  handle,
        netns:   netns,
    }, err
}
