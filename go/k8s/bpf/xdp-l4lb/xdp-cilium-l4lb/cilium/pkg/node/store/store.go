package store

import (
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/kvstore/store"
)

// NodeRegistrar is a wrapper around store.SharedStore.
type NodeRegistrar struct {
    *store.SharedStore

    registerStore *store.SharedStore
}
