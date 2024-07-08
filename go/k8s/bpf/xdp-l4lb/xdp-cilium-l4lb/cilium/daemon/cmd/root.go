package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cilium/cilium/pkg/datapath/maps"
	"github.com/cilium/cilium/pkg/endpoint"
	"github.com/cilium/cilium/pkg/ipcache"
	"github.com/cilium/cilium/pkg/ipmasq"
	"github.com/cilium/cilium/pkg/maps/ctmap/gc"
	"github.com/cilium/cilium/pkg/node"
	"github.com/go-openapi/loads"
	gops "github.com/google/gops/agent"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/api/v1/server"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/api/v1/server/restapi"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/datapath/iptables"
	linuxdatapath "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/datapath/linux"
	datapathOption "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/datapath/option"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/logging/logfields"
	nodeTypes "k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/node/types"
	"k8s-lx1036/k8s/bpf/xdp-l4lb/xdp-cilium-l4lb/cilium/pkg/option"
)

const (
	apiTimeout = 60 * time.Second
)

var (
	log = logging.DefaultLogger.WithField(logfields.LogSubsys, daemonSubsys)

	bootstrapTimestamp = time.Now()

	// RootCmd represents the base command when called without any subcommands
	RootCmd = &cobra.Command{
		Use:   "cilium-agent",
		Short: "Run the cilium agent",
		Run: func(cmd *cobra.Command, args []string) {
			// Open socket for using gops to get stacktraces of the agent.
			addr := fmt.Sprintf("127.0.0.1:%d", viper.GetInt(option.GopsPort))
			addrField := logrus.Fields{"address": addr}
			if err := gops.Listen(gops.Options{
				Addr:                   addr,
				ReuseSocketAddrAndPort: true,
			}); err != nil {
				log.WithError(err).WithFields(addrField).Fatal("Cannot start gops server")
			}
			log.WithFields(addrField).Info("Started gops server")

			initEnv()
			runDaemon()
		},
	}
)

func init() {
	initializeFlags()
}

func initializeFlags() {
	cobra.OnInitialize(option.InitConfig(RootCmd, "Cilium", "ciliumd"))

	flags := RootCmd.Flags()

	flags.String(option.ConfigFile, "", `Configuration file (default "$HOME/ciliumd.yaml")`)
	option.BindEnv(option.ConfigFile)

	flags.String(option.ConfigDir, "", `Configuration directory that contains a file for each option`)
	option.BindEnv(option.ConfigDir)

	flags.String(option.BPFRoot, "", "Path to BPF filesystem")
	option.BindEnv(option.BPFRoot)

	viper.BindPFlags(flags)
}

func initEnv() {

	switch option.Config.DatapathMode {
	case datapathOption.DatapathModeVeth:
		if option.Config.Tunnel == "" {
			option.Config.Tunnel = option.TunnelVXLAN
		}
	case datapathOption.DatapathModeLBOnly:
		log.Info("Running in LB-only mode")
		//
	default:
		log.WithField(logfields.DatapathMode, option.Config.DatapathMode).Fatal("Invalid datapath mode")
	}

}

func runDaemon() {
	datapathConfig := linuxdatapath.DatapathConfiguration{
		HostDevice: option.Config.HostDevice,
	}

	log.Info("Initializing daemon")

	option.Config.RunMonitorAgent = true

	if err := enableIPForwarding(); err != nil {
		log.WithError(err).Fatal("Error when enabling sysctl parameters")
	}

	iptablesManager := &iptables.IptablesManager{}
	iptablesManager.Init()

	ctx, cancel := context.WithCancel(server.ServerCtx)
	d, restoredEndpoints, err := NewDaemon(ctx, cancel,
		WithDefaultEndpointManager(ctx, endpoint.CheckHealth),
		linuxdatapath.NewDatapath(datapathConfig, iptablesManager, wgAgent))
	if err != nil {
		select {
		case <-server.ServerCtx.Done():
			log.WithError(err).Debug("Error while creating daemon")
		default:
			log.WithError(err).Fatal("Error while creating daemon")
		}
		return
	}

	log.Info("Validating configured node address ranges")
	if err := node.ValidatePostInit(); err != nil {
		log.WithError(err).Fatal("postinit failed")
	}

	log.Info("Starting connection tracking garbage collector")
	gc.Enable(option.Config.EnableIPv4, option.Config.EnableIPv6,
		restoredEndpoints.restored, d.endpointManager)

	restoreComplete := d.initRestore(restoredEndpoints)

	// INFO: add a new host endpoint for bpf_host.c
	if d.endpointManager.HostEndpointExists() {
		d.endpointManager.InitHostEndpointLabels(d.ctx)
	} else {
		log.Info("Creating host endpoint")
		if err := d.endpointManager.AddHostEndpoint(d.ctx, d, d, d.ipcache, d.l7Proxy, d.identityAllocator,
			"Create host endpoint", nodeTypes.GetName()); err != nil {
			log.WithError(err).Fatal("Unable to create host endpoint")
		}
	}

	if option.Config.EnableIPMasqAgent {
		ipmasqAgent, err := ipmasq.NewIPMasqAgent(option.Config.IPMasqAgentConfigPath)
		if err != nil {
			log.WithError(err).Fatal("Failed to create ip-masq-agent")
		}
		ipmasqAgent.Start()
	}

	if !option.Config.DryMode {
		go func() {
			if restoreComplete != nil {
				<-restoreComplete
			}
			d.dnsNameManager.CompleteBootstrap()

			ms := maps.NewMapSweeper(&EndpointMapManager{
				EndpointManager: d.endpointManager,
			})
			ms.CollectStaleMapGarbage()
			ms.RemoveDisabledMaps()

			if len(d.restoredCIDRs) > 0 {
				// Release restored CIDR identities after a grace period (default 10
				// minutes).  Any identities actually in use will still exist after
				// this.
				//
				// This grace period is needed when running on an external workload
				// where policy synchronization is not done via k8s. Also in k8s
				// case it is prudent to allow concurrent endpoint regenerations to
				// (re-)allocate the restored identities before we release them.
				time.Sleep(option.Config.IdentityRestoreGracePeriod)
				log.Debugf("Releasing reference counts for %d restored CIDR identities", len(d.restoredCIDRs))

				ipcache.ReleaseCIDRIdentitiesByCIDR(d.restoredCIDRs)
				// release the memory held by restored CIDRs
				d.restoredCIDRs = nil
			}
		}()
		d.endpointManager.Subscribe(d)
		defer d.endpointManager.Unsubscribe(d)
	}

	if option.Config.EnableHealthChecking {
		//d.initHealth()
	}

	d.startStatusCollector()

	d.startAgentHealthHTTPService()

	srv := server.NewServer(d.instantiateAPI()) // instantiateAPI() 里定义每一个 restapi 的具体 handler
	srv.EnabledListeners = []string{"unix"}
	srv.SocketPath = option.Config.SocketPath // /var/run/cilium/cilium.sock
	srv.ReadTimeout = apiTimeout
	srv.WriteTimeout = apiTimeout
	defer srv.Shutdown()

	srv.ConfigureAPI()

	err = d.SendNotification(monitorAPI.StartMessage(time.Now()))
	if err != nil {
		log.WithError(err).Warn("Failed to send agent start monitor message")
	}

	// clean up all arp PERM entries that might have previously set by
	// a Cilium instance
	if !d.datapath.Node().NodeNeighDiscoveryEnabled() {
		d.datapath.Node().NodeCleanNeighbors()
	}
	// Start periodical arping to refresh neighbor table
	if d.datapath.Node().NodeNeighDiscoveryEnabled() && option.Config.ARPPingRefreshPeriod != 0 {
		d.nodeDiscovery.Manager.StartNeighborRefresh(d.datapath.Node())
	}

	log.WithField("bootstrapTime", time.Since(bootstrapTimestamp)).
		Info("Daemon initialization completed")

	errs := make(chan error, 1)

	go func() {
		errs <- srv.Serve()
	}()

	//err = option.Config.StoreInFile(option.Config.StateDir)
	//if err != nil {
	//    log.WithError(err).Error("Unable to store Cilium's configuration")
	//}
	//
	//err = option.StoreViperInFile(option.Config.StateDir)
	//if err != nil {
	//    log.WithError(err).Error("Unable to store Viper's configuration")
	//}

	select {
	case err := <-errs:
		if err != nil {
			log.WithError(err).Fatal("Error returned from non-returning Serve() call")
		}
	}
}

// instantiateAPI() 里定义每一个 restapi 的具体 handler
func (d *Daemon) instantiateAPI() *restapi.CiliumAPIAPI {
	swaggerSpec, err := loads.Analyzed(server.SwaggerJSON, "")
	if err != nil {
		log.WithError(err).Fatal("Cannot load swagger spec")
	}

	log.Info("Initializing Cilium API")
	restAPI := restapi.NewCiliumAPIAPI(swaggerSpec)
	restAPI.Logger = log.Infof

	if option.Config.DatapathMode != datapathOption.DatapathModeLBOnly {
		// /endpoint/
		restAPI.EndpointGetEndpointHandler = NewGetEndpointHandler(d)

		// /endpoint/{id}
		restAPI.EndpointGetEndpointIDHandler = NewGetEndpointIDHandler(d)
		restAPI.EndpointPutEndpointIDHandler = NewPutEndpointIDHandler(d)
		restAPI.EndpointPatchEndpointIDHandler = NewPatchEndpointIDHandler(d)
		restAPI.EndpointDeleteEndpointIDHandler = NewDeleteEndpointIDHandler(d)

	}

	return restAPI
}

func Execute() {
	interruptCh := cleaner.registerSigHandler()
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	<-interruptCh
}
