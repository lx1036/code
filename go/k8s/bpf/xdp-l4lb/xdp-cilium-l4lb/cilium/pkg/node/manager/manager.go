package manager

import (
    "net"

    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/controller"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/datapath"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/ipcache"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/labels"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/lock"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/metrics"

    "github.com/prometheus/client_golang/prometheus"
)

// IPCache is the set of interactions the node manager performs with the ipcache
type IPCache interface {
    Upsert(ip string, hostIP net.IP, hostKey uint8, k8sMeta *ipcache.K8sMetadata, newIdentity ipcache.Identity) (bool, error)
    Delete(IP string, source source.Source) bool
    TriggerLabelInjection(source source.Source)
    UpsertMetadata(string, labels.Labels)
}

// Manager is the entity that manages a collection of nodes
type Manager struct {
    // mutex is the lock protecting access to the nodes map. The mutex must
    // be held for any access of the nodes map.
    //
    // The manager mutex works together with the entry mutex in the
    // following way to minimize the duration the manager mutex is held:
    //
    // 1. Acquire manager mutex to safely access nodes map and to retrieve
    //    node entry.
    // 2. Acquire mutex of the entry while the manager mutex is still held.
    //    This guarantees that no change to the entry has happened.
    // 3. Release of the manager mutex to unblock changes or reads to other
    //    node entries.
    // 4. Change of entry data or performing of datapath interactions
    // 5. Release of the entry mutex
    //
    // If both the nodeEntry.mutex and Manager.mutex must be held, then the
    // Manager.mutex must *always* be acquired first.
    mutex lock.RWMutex

    // nodes is the list of nodes. Access must be protected via mutex.
    nodes map[nodeTypes.Identity]*nodeEntry

    // nodeHandlersMu protects the nodeHandlers map against concurrent access.
    nodeHandlersMu lock.RWMutex
    // nodeHandlers has a slice containing all node handlers subscribed to node
    // events.
    nodeHandlers map[datapath.NodeHandler]struct{}

    // closeChan is closed when the manager is closed
    closeChan chan struct{}

    // name is the name of the manager. It must be unique and feasibility
    // to be used a prometheus metric name.
    name string

    // metricEventsReceived is the prometheus metric to track the number of
    // node events received
    metricEventsReceived *prometheus.CounterVec

    // metricNumNodes is the prometheus metric to track the number of nodes
    // being managed
    metricNumNodes prometheus.Gauge

    // metricDatapathValidations is the prometheus metric to track the
    // number of datapath node validation calls
    metricDatapathValidations prometheus.Counter

    // conf is the configuration of the caller passed in via NewManager.
    // This field is immutable after NewManager()
    conf Configuration

    // ipcache is the set operations performed against the ipcache
    ipcache IPCache

    // controllerManager manages the controllers that are launched within the
    // Manager.
    controllerManager *controller.Manager

    // selectorCacheUpdater updates the identities inside the selector cache.
    selectorCacheUpdater selectorCacheUpdater

    // policyTriggerer triggers policy updates (recalculations).
    policyTriggerer policyTriggerer
}

// Subscribe subscribes the given node handler to node events.
func (m *Manager) Subscribe(nh datapath.NodeHandler) {
    m.nodeHandlersMu.Lock()
    m.nodeHandlers[nh] = struct{}{}
    m.nodeHandlersMu.Unlock()
    // Add all nodes already received by the manager.
    m.mutex.RLock()
    for _, v := range m.nodes {
        v.mutex.Lock()
        nh.NodeAdd(v.node)
        v.mutex.Unlock()
    }
    m.mutex.RUnlock()
}

func NewManager(name string, dp datapath.NodeHandler, c Configuration, sc selectorCacheUpdater, pt policyTriggerer) (*Manager, error) {
    m := &Manager{
        name:                 name,
        nodes:                map[nodeTypes.Identity]*nodeEntry{},
        conf:                 c,
        controllerManager:    controller.NewManager(),
        selectorCacheUpdater: sc,
        policyTriggerer:      pt,
        nodeHandlers:         map[datapath.NodeHandler]struct{}{},
        closeChan:            make(chan struct{}),
    }
    m.Subscribe(dp)

    m.metricEventsReceived = prometheus.NewCounterVec(prometheus.CounterOpts{
        Namespace: metrics.Namespace,
        Subsystem: "nodes",
        Name:      name + "_events_received_total",
        Help:      "Number of node events received",
    }, []string{"event_type", "source"})

    m.metricNumNodes = prometheus.NewGauge(prometheus.GaugeOpts{
        Namespace: metrics.Namespace,
        Subsystem: "nodes",
        Name:      name + "_num",
        Help:      "Number of nodes managed",
    })

    m.metricDatapathValidations = prometheus.NewCounter(prometheus.CounterOpts{
        Namespace: metrics.Namespace,
        Subsystem: "nodes",
        Name:      name + "_datapath_validations_total",
        Help:      "Number of validation calls to implement the datapath implementation of a node",
    })

    err := metrics.RegisterList([]prometheus.Collector{m.metricDatapathValidations, m.metricEventsReceived, m.metricNumNodes})
    if err != nil {
        return nil, err
    }

    go m.backgroundSync()

    return m, nil
}
