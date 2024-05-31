package loader

import (
    "context"
    "github.com/sirupsen/logrus"
    "github.com/vishvananda/netlink"
    "net"
    "path"
    "sync"

    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/datapath"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/datapath/linux/route"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/datapath/loader/metrics"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging/logfields"
)

var (
    log = logging.DefaultLogger.WithField(logfields.LogSubsys, Subsystem)
)

type Loader struct {
    once sync.Once

    // templateCache is the cache of pre-compiled datapaths.
    templateCache *objectCache

    canDisableDwarfRelocations bool
}

func (l *Loader) CallsMapPath(id uint16) string {
    //TODO implement me
    panic("implement me")
}

func (l *Loader) CustomCallsMapPath(id uint16) string {
    //TODO implement me
    panic("implement me")
}

// CompileAndLoad compiles the BPF datapath programs for the specified endpoint
// and loads it onto the interface associated with the endpoint.
//
// Expects the caller to have created the directory at the path ep.StateDir().
func (l *Loader) CompileAndLoad(ctx context.Context, ep datapath.Endpoint, stats *metrics.SpanStat) error {
    if ep == nil {
        log.Fatalf("LoadBPF() doesn't support non-endpoint load")
    }

    dirs := directoryInfo{
        Library: option.Config.BpfDir,
        Runtime: option.Config.StateDir,
        State:   ep.StateDir(),
        Output:  ep.StateDir(),
    }
    return l.compileAndLoad(ctx, ep, &dirs, stats)
}

func (l *Loader) compileAndLoad(ctx context.Context, ep datapath.Endpoint, dirs *directoryInfo, stats *metrics.SpanStat) error {
    stats.BpfCompilation.Start()
    // clang xxx| llc xxx
    err := compileDatapath(ctx, dirs, ep.IsHost(), ep.Logger(Subsystem))
    stats.BpfCompilation.End(err == nil)
    if err != nil {
        return err
    }

    stats.BpfLoadProg.Start()
    err = l.reloadDatapath(ctx, ep, dirs)
    stats.BpfLoadProg.End(err == nil)
    return err
}

// 直接使用 `ip` 和 `tc` 命令 load bpf 程序到内核，和 attach 到对应的网卡
func (l *Loader) reloadDatapath(ctx context.Context, ep datapath.Endpoint, dirs *directoryInfo) error {
    // Replace the current program
    objPath := path.Join(dirs.Output, endpointObj)

    if ep.IsHost() {
        objPath = path.Join(dirs.Output, hostEndpointObj)
        if err := l.reloadHostDatapath(ctx, ep, objPath); err != nil {
            return err
        }
    } else {
        // `ip -force link set dev eth0 xdp obj xxx sec xxx`
        // `tc filter replace dev eth0 ingress/egress prio xxx handle 1 bpf da obj xxx sec xxx`
        finalize, err := replaceDatapath(ctx, ep.InterfaceName(), objPath, symbolFromEndpoint, dirIngress, false, "")
        if err != nil {
            scopedLog := ep.Logger(Subsystem).WithFields(logrus.Fields{
                logfields.Path: objPath,
                logfields.Veth: ep.InterfaceName(),
            })
            // Don't log an error here if the context was canceled or timed out;
            // this log message should only represent failures with respect to
            // loading the program.
            if ctx.Err() == nil {
                scopedLog.WithError(err).Warn("JoinEP: Failed to load program")
            }
            return err
        }
        defer finalize()

        if ep.RequireEgressProg() {
            finalize, err := replaceDatapath(ctx, ep.InterfaceName(), objPath, symbolToEndpoint, dirEgress, false, "")
            if err != nil {
                scopedLog := ep.Logger(Subsystem).WithFields(logrus.Fields{
                    logfields.Path: objPath,
                    logfields.Veth: ep.InterfaceName(),
                })
                // Don't log an error here if the context was canceled or timed out;
                // this log message should only represent failures with respect to
                // loading the program.
                if ctx.Err() == nil {
                    scopedLog.WithError(err).Warn("JoinEP: Failed to load program")
                }
                return err
            }
            defer finalize()
        } else {
            err := RemoveTCFilters(ep.InterfaceName(), netlink.HANDLE_MIN_EGRESS)
            if err != nil {
                log.WithField("device", ep.InterfaceName()).Error(err)
            }
        }
    }

    if ep.RequireEndpointRoute() {
        scopedLog := ep.Logger(Subsystem).WithFields(logrus.Fields{
            logfields.Veth: ep.InterfaceName(),
        })
        if ip := ep.IPv4Address(); ip.IsSet() {
            if err := upsertEndpointRoute(ep, *ip.EndpointPrefix()); err != nil {
                scopedLog.WithError(err).Warn("Failed to upsert route")
            }
        }
    }

    return nil
}

// CompileOrLoad loads the BPF datapath programs for the specified endpoint.
//
// In contrast with CompileAndLoad(), it attempts to find a pre-compiled
// template datapath object to use, to avoid a costly compile operation.
// Only if there is no existing template that has the same configuration
// parameters as the specified endpoint, this function will compile a new
// template for this configuration.
//
// This function will block if the cache does not contain an entry for the
// same EndpointConfiguration and multiple goroutines attempt to concurrently
// CompileOrLoad with the same configuration parameters. When the first
// goroutine completes compilation of the template, all other CompileOrLoad
// invocations will be released.
func (l *Loader) CompileOrLoad(ctx context.Context, ep datapath.Endpoint, stats *metrics.SpanStat) error {
    //TODO implement me
    panic("implement me")
}

func (l *Loader) ReloadDatapath(ctx context.Context, ep datapath.Endpoint, stats *metrics.SpanStat) error {
    dirs := directoryInfo{
        Library: option.Config.BpfDir,
        Runtime: option.Config.StateDir,
        State:   ep.StateDir(),
        Output:  ep.StateDir(),
    }
    stats.BpfLoadProg.Start()
    err = l.reloadDatapath(ctx, ep, &dirs)
    stats.BpfLoadProg.End(err == nil)
    return err
}

func (l *Loader) EndpointHash(cfg datapath.EndpointConfiguration) (string, error) {
    //TODO implement me
    panic("implement me")
}

func (l *Loader) Unload(ep datapath.Endpoint) {
    //TODO implement me
    panic("implement me")
}

func (l *Loader) Reinitialize(ctx context.Context, o interface{}, deviceMTU int, iptMgr datapath.IptablesManager, p interface{}) error {
    //TODO implement me
    panic("implement me")
}

// NewLoader returns a new loader.
func NewLoader(canDisableDwarfRelocations bool) *Loader {
    return &Loader{
        canDisableDwarfRelocations: canDisableDwarfRelocations,
    }
}

func upsertEndpointRoute(ep datapath.Endpoint, ip net.IPNet) error {
    endpointRoute := route.Route{
        Prefix: ip,
        Device: ep.InterfaceName(),
        Scope:  netlink.SCOPE_LINK,
    }

    return route.Upsert(endpointRoute)
}
