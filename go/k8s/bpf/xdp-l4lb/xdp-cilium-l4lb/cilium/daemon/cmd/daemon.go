package cmd

import (
	"context"
	"github.com/cilium/cilium/api/v1/models"
	"github.com/cilium/cilium/pkg/clustermesh"
	"github.com/cilium/cilium/pkg/counter"
	linuxrouting "github.com/cilium/cilium/pkg/datapath/linux/routing"
	"github.com/cilium/cilium/pkg/egressgateway"
	"github.com/cilium/cilium/pkg/eventqueue"
	"github.com/cilium/cilium/pkg/fqdn"
	"github.com/cilium/cilium/pkg/hubble/observer"
	"github.com/cilium/cilium/pkg/ipam"
	"github.com/cilium/cilium/pkg/ipcache"
	"github.com/cilium/cilium/pkg/logging/logfields"
	"github.com/cilium/cilium/pkg/maps/ctmap"
	"github.com/cilium/cilium/pkg/maps/eppolicymap"
	"github.com/cilium/cilium/pkg/maps/policymap"
	"github.com/cilium/cilium/pkg/mtu"
	"github.com/cilium/cilium/pkg/policy"
	"github.com/cilium/cilium/pkg/rate"
	"github.com/cilium/cilium/pkg/recorder"
	"github.com/cilium/cilium/pkg/redirectpolicy"
	"github.com/cilium/cilium/pkg/status"
	"github.com/cilium/cilium/pkg/trigger"
	cnitypes "github.com/containernetworking/cni/pkg/types"
	log "github.com/sirupsen/logrus"
	"os"

	"github.com/cilium/cilium/pkg/lock"
	"golang.org/x/sync/semaphore"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/datapath"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/maps/lbmap"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/option"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/service"
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

	ctmap.InitMapInfo(option.Config.CTMapEntriesGlobalTCP, option.Config.CTMapEntriesGlobalAny,
		option.Config.EnableIPv4, option.Config.EnableIPv6, option.Config.EnableNodePort)
	policymap.InitMapInfo(option.Config.PolicyMapEntries)
	lbmap.Init(lbmap.InitParams{
		IPv4: option.Config.EnableIPv4,
		IPv6: option.Config.EnableIPv6,

		MaxSockRevNatMapEntries: option.Config.SockRevNatEntries,
		MaxEntries:              option.Config.LBMapEntries,
	})

	d := Daemon{
		ctx:               ctx,
		cancel:            cancel,
		prefixLengths:     createPrefixLengthCounter(),
		buildEndpointSem:  semaphore.NewWeighted(int64(numWorkerThreads())),
		compilationMutex:  new(lock.RWMutex),
		netConf:           netConf,
		mtuConfig:         mtuConfig,
		datapath:          dp,
		nodeDiscovery:     nd,
		endpointCreations: newEndpointCreationManager(),
		apiLimiterSet:     apiLimiterSet,
	}

	d.svc = service.NewService(&d)

	// Open or create BPF maps.
	bootstrapStats.mapsInit.Start()
	err = d.initMaps()
	bootstrapStats.mapsInit.EndError(err)
	if err != nil {
		log.WithError(err).Error("Error while opening/creating BPF maps")
		return nil, nil, err
	}
	// Upsert restored CIDRs after the new ipcache has been opened above
	if len(restoredCIDRidentities) > 0 {
		ipcache.UpsertGeneratedIdentities(restoredCIDRidentities, nil)
	}

	// option.Config.RestoreState=true, restore from 从已有的 bpf maps
	if option.Config.RestoreState && !option.Config.DryMode {
		bootstrapStats.restore.Start()
		d.svc.RestoreServices()
		bootstrapStats.restore.End(true)
	}

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
