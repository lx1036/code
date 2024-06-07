package cmd

import (
    "context"
    "fmt"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/k8s"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/source"
    "net"
    "os"
    "sync"

    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/api/v1/models"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/bpf"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/bpf/maps/sockmap"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/bpf/sockops"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/clustermesh"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/controller"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/counter"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/datapath"
    linuxdatapath "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/datapath/linux"
    linuxrouting "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/datapath/linux/routing"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/defaults"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/egressgateway"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/endpointmanager"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/eventqueue"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/fqdn"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/hubble/observer"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/identity"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/ipam"
    ipamOption "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/ipam/option"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/ipcache"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/k8s/watchers"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/lock"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging/logfields"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/maps/ctmap"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/maps/eppolicymap"
    ipcachemap "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/maps/ipcache"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/maps/lbmap"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/maps/policymap"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/metrics"
    monitoragent "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/monitor/agent"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/mtu"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/node"
    nodemanager "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/node/manager"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/nodediscovery"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/option"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/policy"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/rate"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/recorder"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/redirectpolicy"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/service"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/status"
    "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/trigger"
    cnitypes "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/plugins/cilium-cni/types"
    proxy "k8s-lx1036/k8s/network/calico/calico/felix/pkg/bpf/kube-proxy"
    "k8s-lx1036/k8s/network/loadbalancer/metallb/pkg/speaker"

    "github.com/cilium/ebpf/rlimit"
    "golang.org/x/sync/semaphore"
)

var (
    bootstrapStats = bootstrapStatistics{}
)

const (
    // ConfigModifyQueueSize is the size of the event queue for serializing
    // configuration updates to the daemon
    ConfigModifyQueueSize = 10
)

type Daemon struct {
    ctx              context.Context
    cancel           context.CancelFunc
    buildEndpointSem *semaphore.Weighted
    l7Proxy          *proxy.Proxy
    svc              *service.Service
    rec              *recorder.Recorder
    policy           *policy.Repository
    preFilter        datapath.PreFilter

    statusCollectMutex lock.RWMutex
    statusResponse     models.StatusResponse
    statusCollector    *status.Collector

    monitorAgent *monitoragent.Agent
    ciliumHealth *health.CiliumHealth

    // dnsNameManager tracks which api.FQDNSelector are present in policy which
    // apply to locally running endpoints.
    dnsNameManager *fqdn.NameManager

    // Used to synchronize generation of daemon's BPF programs and endpoint BPF
    // programs.
    compilationMutex *lock.RWMutex

    // prefixLengths tracks a mapping from CIDR prefix length to the count
    // of rules that refer to that prefix length.
    prefixLengths *counter.PrefixLengthCounter

    clustermesh *clustermesh.ClusterMesh

    mtuConfig     mtu.Configuration
    policyTrigger *trigger.Trigger

    datapathRegenTrigger *trigger.Trigger

    // datapath is the underlying datapath implementation to use to
    // implement all aspects of an agent
    datapath datapath.Datapath

    // nodeDiscovery defines the node discovery logic of the agent
    nodeDiscovery *nodediscovery.NodeDiscovery

    deviceManager *linuxdatapath.DeviceManager

    controllers *controller.Manager

    ipcache *ipcache.IPCache

    // ipam is the IP address manager of the agent
    ipam *ipam.IPAM

    netConf *cnitypes.NetConf

    endpointManager *endpointmanager.EndpointManager

    identityAllocator CachingIdentityAllocator

    k8sWatcher *watchers.K8sWatcher

    // healthEndpointRouting is the information required to set up the health
    // endpoint's routing in ENI or Azure IPAM mode
    healthEndpointRouting *linuxrouting.RoutingInfo

    hubbleObserver *observer.LocalObserverServer

    // k8sCachesSynced is closed when all essential Kubernetes caches have
    // been fully synchronized
    k8sCachesSynced <-chan struct{}

    // endpointCreations is a map of all currently ongoing endpoint
    // creation events
    endpointCreations *endpointCreationManager

    redirectPolicyManager *redirectpolicy.Manager

    bgpSpeaker *speaker.Speaker

    egressGatewayManager *egressgateway.Manager

    apiLimiterSet *rate.APILimiterSet

    // event queue for serializing configuration updates to the daemon.
    configModifyQueue *eventqueue.EventQueue

    // CIDRs for which identities were restored during bootstrap
    restoredCIDRs []*net.IPNet
}

func NewDaemon(ctx context.Context, cancel context.CancelFunc, epMgr *endpointmanager.EndpointManager,
    dp datapath.Datapath) (*Daemon, *endpointRestoreState, error) {
    var (
        err           error
        netConf       *cnitypes.NetConf
        configuredMTU = option.Config.MTU
    )

    // INFO: 1. bootstrap daemon
    bootstrapStats.daemonInit.Start()

    // Validate the daemon-specific global options.
    if err := option.Config.Validate(); err != nil {
        return nil, nil, fmt.Errorf("invalid daemon configuration: %s", err)
    }

    if option.Config.ReadCNIConfiguration != "" {
        netConf, err = cnitypes.ReadNetConf(option.Config.ReadCNIConfiguration)
        if err != nil {
            log.WithError(err).Error("Unable to read CNI configuration")
            return nil, nil, fmt.Errorf("unable to read CNI configuration: %w", err)
        }

        if netConf.MTU != 0 {
            configuredMTU = netConf.MTU
            log.WithField("mtu", configuredMTU).Info("Overwriting MTU based on CNI configuration")
        }
    }

    apiLimiterSet, err := rate.NewAPILimiterSet(option.Config.APIRateLimit, apiRateLimitDefaults, &apiRateLimitingMetrics{})
    if err != nil {
        log.WithError(err).Error("unable to configure API rate limiting")
        return nil, nil, fmt.Errorf("unable to configure API rate limiting: %w", err)
    }

    // Do the partial kube-proxy replacement initialization before creating BPF
    // maps. Otherwise, some maps might not be created (e.g. session affinity).
    // finishKubeProxyReplacementInit(), which is called later after the device
    // detection, might disable BPF NodePort and friends. But this is fine, as
    // the feature does not influence the decision which BPF maps should be
    // created.
    isKubeProxyReplacementStrict := true

    // INFO: init bpf ct/policy/lb maps
    ctmap.InitMapInfo(option.Config.CTMapEntriesGlobalTCP, option.Config.CTMapEntriesGlobalAny,
        option.Config.EnableIPv4, option.Config.EnableIPv6, option.Config.EnableNodePort)
    policymap.InitMapInfo(option.Config.PolicyMapEntries)
    lbmapInitParams := lbmap.InitParams{
        IPv4:                     option.Config.EnableIPv4,
        IPv6:                     option.Config.EnableIPv6,
        MaxSockRevNatMapEntries:  option.Config.SockRevNatEntries,
        ServiceMapMaxEntries:     option.Config.LBMapEntries,
        BackEndMapMaxEntries:     option.Config.LBMapEntries,
        RevNatMapMaxEntries:      option.Config.LBMapEntries,
        AffinityMapMaxEntries:    option.Config.LBMapEntries,
        SourceRangeMapMaxEntries: option.Config.LBMapEntries,
        MaglevMapMaxEntries:      option.Config.LBMapEntries,
    }
    if option.Config.LBServiceMapEntries > 0 {
        lbmapInitParams.ServiceMapMaxEntries = option.Config.LBServiceMapEntries
    }
    if option.Config.LBBackendMapEntries > 0 {
        lbmapInitParams.BackEndMapMaxEntries = option.Config.LBBackendMapEntries
    }
    if option.Config.LBRevNatEntries > 0 {
        lbmapInitParams.RevNatMapMaxEntries = option.Config.LBRevNatEntries
    }
    if option.Config.LBAffinityMapEntries > 0 {
        lbmapInitParams.AffinityMapMaxEntries = option.Config.LBAffinityMapEntries
    }
    if option.Config.LBSourceRangeMapEntries > 0 {
        lbmapInitParams.SourceRangeMapMaxEntries = option.Config.LBSourceRangeMapEntries
    }
    if option.Config.LBMaglevMapEntries > 0 {
        lbmapInitParams.MaglevMapMaxEntries = option.Config.LBMaglevMapEntries
    }
    lbmap.Init(lbmapInitParams)

    if option.Config.DryMode == false {
        if err := rlimit.RemoveMemlock(); err != nil {
            log.WithError(err).Error("unable to set memory resource limits")
            return nil, nil, fmt.Errorf("unable to set memory resource limits: %w", err)
        }
    }

    authKeySize := 0
    externalIP := node.GetIPv4()
    var mtuConfig mtu.Configuration
    // ExternalIP could be nil but we are covering that case inside NewConfiguration
    mtuConfig = mtu.NewConfiguration(
        authKeySize,
        option.Config.EnableIPSec,
        option.Config.TunnelExists(),
        option.Config.EnableWireguard,
        configuredMTU,
        externalIP,
    )
    nodeMgr, err := nodemanager.NewManager("all", dp.Node(), option.Config, nil, nil)
    if err != nil {
        return nil, nil, err
    }
    identity.IterateReservedIdentities(func(_ identity.NumericIdentity, _ *identity.Identity) {
        metrics.Identity.WithLabelValues(identity.ReservedIdentityType).Inc()
    })
    if option.Config.EnableWellKnownIdentities {
        // Must be done before calling policy.NewPolicyRepository() below.
        num := identity.InitWellKnownIdentities(option.Config)
        metrics.Identity.WithLabelValues(identity.WellKnownIdentityType).Add(float64(num))
    }
    nodeDiscovery := nodediscovery.NewNodeDiscovery(nodeMgr, mtuConfig, netConf)

    devMgr, err := linuxdatapath.NewDeviceManager()
    if err != nil {
        return nil, nil, err
    }

    d := Daemon{
        ctx:               ctx,
        cancel:            cancel,
        prefixLengths:     createPrefixLengthCounter(),
        buildEndpointSem:  semaphore.NewWeighted(int64(numWorkerThreads())),
        compilationMutex:  new(lock.RWMutex),
        netConf:           netConf,
        mtuConfig:         mtuConfig,
        datapath:          dp,
        deviceManager:     devMgr,
        nodeDiscovery:     nodeDiscovery,
        endpointCreations: newEndpointCreationManager(),
        apiLimiterSet:     apiLimiterSet,
        controllers:       controller.NewManager(),
    }

    if option.Config.RunMonitorAgent {
        d.monitorAgent = monitoragent.NewAgent(ctx)
    }
    // cilium-agent config reload 功能
    d.configModifyQueue = eventqueue.NewEventQueueBuffered("config-modify-queue", ConfigModifyQueueSize)
    d.configModifyQueue.Run()

    /**
     * INFO: 1.1 Restore
     */
    // Collect old CIDR identities
    var oldNumericIdentities []identity.NumericIdentity
    var oldIngressIPs []*net.IPNet
    if option.Config.RestoreState && !option.Config.DryMode {
        if err := ipcachemap.IPCache.DumpWithCallback(func(key bpf.MapKey, value bpf.MapValue) {
            k := key.(*ipcachemap.Key)
            v := value.(*ipcachemap.RemoteEndpointInfo)
            nid := identity.NumericIdentity(v.SecurityIdentity)
            if nid.HasLocalScope() {
                d.restoredCIDRs = append(d.restoredCIDRs, k.IPNet())
                oldNumericIdentities = append(oldNumericIdentities, nid)
            } else if nid == identity.ReservedIdentityIngress && v.TunnelEndpoint.IsZero() {
                oldIngressIPs = append(oldIngressIPs, k.IPNet())
                ip := k.IPNet().IP
                if ip.To4() != nil {
                    node.SetIngressIPv4(ip)
                } else {
                    node.SetIngressIPv6(ip)
                }
            }
        }); err != nil && !os.IsNotExist(err) {
            log.WithError(err).Debug("Error dumping ipcache")
        }
        // DumpWithCallback() leaves the ipcache map open, must close before opened for
        // parallel mode in Daemon.initMaps()
        // 注意这里需要 Close() bpf map
        ipcachemap.IPCache.Close()
    }
    // Propagate identity allocator down to packages which themselves do not
    // have types to which we can add an allocator member.
    //
    // **NOTE** The global identity allocator is not yet initialized here; that
    // happens below vie InitIdentityAllocator(). Only the local identity allocator
    // is initialized here.
    //
    // TODO: convert these package level variables to types for easier unit
    // testing in the future.
    d.identityAllocator = NewCachingIdentityAllocator(&d)
    if err := d.initPolicy(epMgr); err != nil {
        return nil, nil, fmt.Errorf("error while initializing policy subsystem: %w", err)
    }
    d.ipcache = ipcache.NewIPCache(&ipcache.Configuration{
        IdentityAllocator: d.identityAllocator,
        PolicyHandler:     d.policy.GetSelectorCache(),
        DatapathHandler:   epMgr,
    })
    // Preallocate IDs for old CIDRs. This must be done before any Identity allocations are
    // possible so that the old IDs are still available. That is why we do this ASAP after the
    // new (userspace) ipcache is created above.
    //
    // CIDRs were dumped from the old ipcache, they are re-allocated here, hopefully with the
    // same numeric IDs as before, but the restored identities are to be upsterted to the new
    // (datapath) ipcache after it has been initialized below. This is accomplished by passing
    // 'restoredCIDRidentities' to AllocateCIDRs() and then calling
    // UpsertGeneratedIdentities(restoredCIDRidentities) after initMaps() below.
    restoredCIDRidentities := make(map[string]*identity.Identity)
    if len(d.restoredCIDRs) > 0 {
        log.Infof("Restoring %d old CIDR identities", len(d.restoredCIDRs))
        _, err = d.ipcache.AllocateCIDRs(d.restoredCIDRs, oldNumericIdentities, restoredCIDRidentities)
        if err != nil {
            log.WithError(err).Error("Error allocating old CIDR identities")
        }
        // Log a warning for the first CIDR identity than could not be restored with the
        // same numeric identity as before the restart. This can only happen if we have
        // re-introduced bugs into this agent bootstrap order, so we want to surface this.
        for i, prefix := range d.restoredCIDRs {
            id, exists := restoredCIDRidentities[prefix.String()]
            if !exists || id.ID != oldNumericIdentities[i] {
                log.WithField(logfields.Identity, oldNumericIdentities[i]).Warn("Could not restore all CIDR identities")
                break
            }
        }
    }
    nodeMgr = nodeMgr.WithIPCache(d.ipcache)
    nodeMgr = nodeMgr.WithSelectorCacheUpdater(d.policy.GetSelectorCache()) // must be after initPolicy
    nodeMgr = nodeMgr.WithPolicyTriggerer(epMgr)                            // must be after initPolicy
    d.endpointManager = epMgr
    d.endpointManager.InitMetrics()

    d.svc = service.NewService(&d)
    if option.Config.EnableIPv4EgressGateway {
        d.egressGatewayManager = egressgateway.NewEgressGatewayManager(&d, d.identityAllocator)
    }

    d.k8sWatcher = watchers.NewK8sWatcher(
        d.endpointManager,
        d.nodeDiscovery,
        &d,
        d.policy,
        d.svc,
        d.datapath,
        d.redirectPolicyManager,
        d.bgpSpeaker,
        d.egressGatewayManager,
        d.l7Proxy,
        option.Config,
        d.ipcache,
    )
    nodeDiscovery.RegisterK8sNodeGetter(d.k8sWatcher)
    d.ipcache.RegisterK8sSyncedChecker(&d)
    d.k8sWatcher.RegisterNodeSubscriber(d.endpointManager)
    if option.Config.EnableServiceTopology {
        d.k8sWatcher.RegisterNodeSubscriber(&d.k8sWatcher.K8sSvcCache)
    }
    // watchers.NewCiliumNodeUpdater needs to be registered *after* d.endpointManager
    d.k8sWatcher.RegisterNodeSubscriber(watchers.NewCiliumNodeUpdater(d.nodeDiscovery))
    d.redirectPolicyManager.RegisterSvcCache(&d.k8sWatcher.K8sSvcCache)
    d.redirectPolicyManager.RegisterGetStores(d.k8sWatcher)

    bootstrapStats.daemonInit.End(true)

    // Stop all endpoints (its goroutines) on exit.
    cleaner.cleanupFuncs.Add(func() {
        log.Info("Waiting for all endpoints' go routines to be stopped.")
        var wg sync.WaitGroup

        eps := d.endpointManager.GetEndpoints()
        wg.Add(len(eps))

        for _, ep := range eps {
            go func(ep *endpoint.Endpoint) {
                ep.Stop()
                wg.Done()
            }(ep)
        }

        wg.Wait()
        log.Info("All endpoints' goroutines stopped.")
    })

    // INFO: 1.2 Open or create BPF maps
    bootstrapStats.mapsInit.Start()
    err = d.initMaps()
    bootstrapStats.mapsInit.EndError(err)
    if err != nil {
        log.WithError(err).Error("Error while opening/creating BPF maps")
        return nil, nil, err
    }
    // Upsert restored CIDRs after the new ipcache has been opened above
    if len(restoredCIDRidentities) > 0 {
        d.ipcache.UpsertGeneratedIdentities(restoredCIDRidentities, nil)
    }
    // Upsert restored local Ingress IPs
    restoredIngressIPs := []string{}
    for _, ingressIP := range oldIngressIPs {
        _, err := d.ipcache.Upsert(ingressIP.String(), nil, 0, nil, ipcache.Identity{
            ID:     identity.ReservedIdentityIngress,
            Source: source.Restored,
        })
        if err == nil {
            restoredIngressIPs = append(restoredIngressIPs, ingressIP.String())
        } else {
            log.WithError(err).Warning("could not restore Ingress IP, a new one will be allocated")
        }
    }
    if len(restoredIngressIPs) > 0 {
        log.WithField(logfields.Ingress, restoredIngressIPs).Info("Restored ingress IPs")
    }
    // Read the service IDs of existing services from the BPF map and
    // reserve them. This must be done *before* connecting to the
    // Kubernetes apiserver and serving the API to ensure service IDs are
    // not changing across restarts or that a new service could accidentally
    // use an existing service ID.
    // Also, create missing v2 services from the corresponding legacy ones.
    if option.Config.RestoreState && !option.Config.DryMode {
        bootstrapStats.restore.Start()
        if err := d.svc.RestoreServices(); err != nil {
            log.WithError(err).Warn("Failed to restore services from BPF maps")
        }
        bootstrapStats.restore.End(true)
    }

    d.k8sWatcher.RunK8sServiceHandler()

    // fetch old endpoints before k8s is configured.
    bootstrapStats.restore.Start()
    restoredEndpoints, err := d.fetchOldEndpoints(option.Config.StateDir)
    if err != nil {
        log.WithError(err).Error("Unable to read existing endpoints")
    }
    bootstrapStats.restore.End(true)

    // bootstrap k8s init
    if k8s.IsEnabled() {
        bootstrapStats.k8sInit.Start()
        // Errors are handled inside WaitForCRDsToRegister. It will fatal on a
        // context deadline or if the context has been cancelled, the context's
        // error will be returned. Otherwise, it succeeded.
        if err := d.k8sWatcher.WaitForCRDsToRegister(d.ctx); err != nil {
            return nil, restoredEndpoints, err
        }

        // Launch the K8s node watcher so we can start receiving node events.
        // Launching the k8s node watcher at this stage will prevent all agents
        // from performing Gets directly into kube-apiserver to get the most up
        // to date version of the k8s node. This allows for better scalability
        // in large clusters.
        d.k8sWatcher.NodesInit(k8s.Client())

        if option.Config.IPAM == ipamOption.IPAMClusterPool || option.Config.IPAM == ipamOption.IPAMClusterPoolV2 {
            // Create the CiliumNode custom resource. This call will block until
            // the custom resource has been created
            d.nodeDiscovery.UpdateCiliumNodeResource()
        }

        if err := k8s.WaitForNodeInformation(d.ctx, d.k8sWatcher); err != nil {
            log.WithError(err).Error("unable to connect to get node spec from apiserver")
            return nil, nil, fmt.Errorf("unable to connect to get node spec from apiserver: %w", err)
        }

        // Kubernetes demands that the localhost can always reach local
        // pods. Therefore unless the AllowLocalhost policy is set to a
        // specific mode, always allow localhost to reach local
        // endpoints.
        if option.Config.AllowLocalhost == option.AllowLocalhostAuto {
            option.Config.AllowLocalhost = option.AllowLocalhostAlways
            log.Info("k8s mode: Allowing localhost to reach local endpoints")
        }

        bootstrapStats.k8sInit.End(true)
    }
    // The kube-proxy replacement and host-fw devices detection should happen after
    // establishing a connection to kube-apiserver, but before starting a k8s watcher.
    // This is because the device detection requires self (Cilium)Node object,
    // and the k8s service watcher depends on option.Config.EnableNodePort flag
    // which can be modified after the device detection.
    if _, err := d.deviceManager.Detect(); err != nil {
        if d.deviceManager.AreDevicesRequired() {
            // Fail hard if devices are required to function.
            return nil, nil, fmt.Errorf("failed to detect devices: %w", err)
        }
        log.WithError(err).Warn("failed to detect devices, disabling BPF NodePort")
        disableNodePort()
    }
    if err := finishKubeProxyReplacementInit(isKubeProxyReplacementStrict); err != nil {
        log.WithError(err).Error("failed to finalise LB initialization")
        return nil, nil, fmt.Errorf("failed to finalise LB initialization: %w", err)
    }

    // INFO: bpf debug. We can only attach the monitor agent once cilium_event has been set up.
    if option.Config.RunMonitorAgent {
        err = d.monitorAgent.AttachToEventsMap(defaults.MonitorBufferPages)
        if err != nil {
            log.WithError(err).Error("encountered error configuring run monitor agent")
            return nil, nil, fmt.Errorf("encountered error configuring run monitor agent: %w", err)
        }

        if option.Config.EnableMonitor {
            err = monitoragent.ServeMonitorAPI(d.monitorAgent)
            if err != nil {
                log.WithError(err).Error("encountered error configuring run monitor agent")
                return nil, nil, fmt.Errorf("encountered error configuring run monitor agent: %w", err)
            }
        }
    }

    return &d, restoredEndpoints, nil
}

func (d *Daemon) init() error {
    globalsDir := option.Config.GetGlobalsDir()
    if err := os.MkdirAll(globalsDir, defaults.RuntimePathRights); err != nil {
        log.WithError(err).WithField(logfields.Path, globalsDir).Fatal("Could not create runtime directory")
    }

    if err := os.Chdir(option.Config.StateDir); err != nil {
        log.WithError(err).WithField(logfields.Path, option.Config.StateDir).Fatal("Could not change to runtime directory")
    }

    // Remove any old sockops and re-enable with _new_ programs if flag is set
    sockops.SockmapDisable()
    sockops.SkmsgDisable()

    if !option.Config.DryMode {
        //bandwidth.InitBandwidthManager()
        if err := d.createNodeConfigHeaderfile(); err != nil {
            return err
        }

        if option.Config.SockopsEnable {
            eppolicymap.CreateEPPolicyMap()
            if err := sockops.SockmapEnable(); err != nil {
                log.WithError(err).Error("Failed to enable Sockmap")
            } else if err := sockops.SkmsgEnable(); err != nil {
                log.WithError(err).Error("Failed to enable Sockmsg")
            } else {
                sockmap.SockmapCreate()
            }
        }

        if err := d.Datapath().Loader().Reinitialize(d.ctx, d, d.mtuConfig.GetDeviceMTU(), d.Datapath(), d.l7Proxy); err != nil {
            return err
        }
    }

    return nil
}
