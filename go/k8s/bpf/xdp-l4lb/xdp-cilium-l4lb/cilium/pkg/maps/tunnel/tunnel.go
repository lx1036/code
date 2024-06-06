package tunnel

import (
    "net"
    "unsafe"

    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/bpf"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging/logfields"

    "github.com/sirupsen/logrus"
)

/**
 * 10.244.1.251 nginx pod 在 minikube-m02 node 上, 10.244.0.148 在 minikube node 上
 * root@minikube:/home/cilium# cilium bpf ipcache get 10.244.1.251 -D
 * 10.244.1.251 maps to identity identity=120746 encryptkey=0 tunnelendpoint=192.168.49.3
 * root@minikube:/home/cilium# cilium bpf ipcache get 10.244.0.148 -D
 * 10.244.0.148 maps to identity identity=120746 encryptkey=0 tunnelendpoint=0.0.0.0
 */

const (
    MapName = "cilium_tunnel_map"

    // MaxEntries is the maximum entries in the tunnel endpoint map
    MaxEntries = 65536

    fieldEndpoint = "endpoint"
    fieldPrefix   = "prefix"
    fieldKey      = "key"
)

var (
    log = logging.DefaultLogger.WithField(logfields.LogSubsys, "map-tunnel")

    // TunnelMap represents the BPF map for tunnels
    TunnelMap = NewTunnelMap(MapName)
)

// Map implements tunnel connectivity configuration in the BPF datapath.
type Map struct {
    *bpf.Map
}

// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapKey
// +k8s:deepcopy-gen:interfaces=github.com/cilium/cilium/pkg/bpf.MapValue
type TunnelEndpoint struct {
    bpf.EndpointKey
}

func (v *TunnelEndpoint) DeepCopyMapKey() bpf.MapKey {
    //TODO implement me
    panic("implement me")
}

func (v *TunnelEndpoint) DeepCopyMapValue() bpf.MapValue {
    //TODO implement me
    panic("implement me")
}

func (v TunnelEndpoint) NewValue() bpf.MapValue {
    return &TunnelEndpoint{}
}

func newTunnelEndpoint(ip net.IP) *TunnelEndpoint {
    return &TunnelEndpoint{
        EndpointKey: bpf.NewEndpointKey(ip),
    }
}

// NewTunnelMap returns a new tunnel map with the specified name.
func NewTunnelMap(name string) *Map {
    return &Map{
        Map: bpf.NewMap(MapName,
            bpf.MapTypeHash,
            &TunnelEndpoint{},
            int(unsafe.Sizeof(TunnelEndpoint{})),
            &TunnelEndpoint{},
            int(unsafe.Sizeof(TunnelEndpoint{})),
            MaxEntries,
            0, 0,
            bpf.ConvertKeyValue,
        ).WithCache().WithPressureMetric().WithNonPersistent(),
    }
}

// SetTunnelEndpoint adds/replaces a prefix => tunnel-endpoint mapping
func (m *Map) SetTunnelEndpoint(encryptKey uint8, prefix, endpoint net.IP) error {
    key, val := newTunnelEndpoint(prefix), newTunnelEndpoint(endpoint)
    val.EndpointKey.Key = encryptKey
    log.WithFields(logrus.Fields{
        fieldPrefix:   prefix,
        fieldEndpoint: endpoint,
        fieldKey:      encryptKey,
    }).Debug("Updating tunnel map entry")

    return TunnelMap.Update(key, val)
}

// GetTunnelEndpoint removes a prefix => tunnel-endpoint mapping
func (m *Map) GetTunnelEndpoint(prefix net.IP) (net.IP, error) {
    val, err := TunnelMap.Lookup(newTunnelEndpoint(prefix))
    if err != nil {
        return net.IP{}, err
    }

    return val.(*TunnelEndpoint).ToIP(), nil
}

// DeleteTunnelEndpoint removes a prefix => tunnel-endpoint mapping
func (m *Map) DeleteTunnelEndpoint(prefix net.IP) error {
    log.WithField(fieldPrefix, prefix).Debug("Deleting tunnel map entry")
    return TunnelMap.Delete(newTunnelEndpoint(prefix))
}

// SilentDeleteTunnelEndpoint removes a prefix => tunnel-endpoint mapping.
// If the prefix is not found no error is returned.
func (m *Map) SilentDeleteTunnelEndpoint(prefix net.IP) error {
    log.WithField(fieldPrefix, prefix).Debug("Silently deleting tunnel map entry")
    _, err := TunnelMap.SilentDelete(newTunnelEndpoint(prefix))
    return err
}
