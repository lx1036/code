package endpointmanager

import (
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/endpoint"
)

func (mgr *EndpointManager) GetHostEndpoint() *endpoint.Endpoint {
    mgr.mutex.RLock()
    defer mgr.mutex.RUnlock()
    for _, ep := range mgr.endpoints {
        if ep.IsHost() {
            return ep
        }
    }
    return nil
}

// HostEndpointExists returns true if the host endpoint exists.
func (mgr *EndpointManager) HostEndpointExists() bool {
    return mgr.GetHostEndpoint() != nil
}
