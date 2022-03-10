package main

import (
	"context"
	"fmt"
	api "github.com/osrg/gobgp/v3/api"
	"github.com/spf13/viper"
	"k8s-lx1036/k8s/network/bgp/pkg/config"

	"os"

	"k8s-lx1036/k8s/network/bgp/pkg/server"

	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
)

// go run . -f ./route-client-conf.conf
func main() {
	stopCh := make(chan struct{})

	var opts struct {
		ConfigFile      string `short:"f" long:"config-file" description:"specifying a config file"`
		ConfigType      string `short:"t" long:"config-type" description:"specifying config type (toml, yaml, json)" default:"toml"`
		LogLevel        string `short:"l" long:"log-level" description:"specifying log level"`
		LogPlain        bool   `short:"p" long:"log-plain" description:"use plain format for logging (json by default)"`
		UseSyslog       string `short:"s" long:"syslog" description:"use syslogd"`
		Facility        string `long:"syslog-facility" description:"specify syslog facility"`
		DisableStdlog   bool   `long:"disable-stdlog" description:"disable standard logging"`
		CPUs            int    `long:"cpus" description:"specify the number of CPUs to be used"`
		GrpcHosts       string `long:"api-hosts" description:"specify the hosts that gobgpd listens on" default:":50053"`
		GracefulRestart bool   `short:"r" long:"graceful-restart" description:"flag restart-state in graceful-restart capability"`
		Dry             bool   `short:"d" long:"dry-run" description:"check configuration"`
		PProfHost       string `long:"pprof-host" description:"specify the host that gobgpd listens on for pprof" default:"localhost:6060"`
		PProfDisable    bool   `long:"pprof-disable" description:"disable pprof profiling"`
		UseSdNotify     bool   `long:"sdnotify" description:"use sd_notify protocol"`
		TLS             bool   `long:"tls" description:"enable TLS authentication for gRPC API"`
		TLSCertFile     string `long:"tls-cert-file" description:"The TLS cert file"`
		TLSKeyFile      string `long:"tls-key-file" description:"The TLS key file"`
		Version         bool   `long:"version" description:"show version number"`
	}
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}
	if len(opts.ConfigFile) == 0 {
		klog.Fatal("config-file is required")
	}

	log.SetLevel(log.DebugLevel)
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.JSONFormatter{})

	maxSize := 256 << 20 // 256MB
	grpcOpts := []grpc.ServerOption{grpc.MaxRecvMsgSize(maxSize), grpc.MaxSendMsgSize(maxSize)}
	bgpServer := server.NewBgpServer(server.GrpcListenAddress(opts.GrpcHosts), server.GrpcOption(grpcOpts)) // localhost:50051
	go bgpServer.Serve()
	defer bgpServer.StopBgp(context.Background(), &api.StopBgpRequest{})

	initialConfig, err := ReadConfigfile(opts.ConfigFile, opts.ConfigType)
	if err != nil {
		log.WithFields(log.Fields{
			"Topic": "Config",
			"Error": err,
		}).Fatalf("Can't read config file %s", opts.ConfigFile)
	}
	ctx := context.Background()
	if err := bgpServer.StartBgp(ctx, &api.StartBgpRequest{
		Global: config.NewGlobalFromConfigStruct(&initialConfig.Global),
	}); err != nil {
		log.Fatalf("failed to set global config: %s", err)
	}

	/*p := ConfigSetToRoutingPolicy(initialConfig)
	rp, err := table.NewAPIRoutingPolicyFromConfigStruct(p)
	if err != nil {
		klog.Error(err)
	} else {
		bgpServer.SetPolicies(ctx, &api.SetPoliciesRequest{
			DefinedSets: rp.DefinedSets,
			Policies:    rp.Policies,
		})
	}*/

	//assignGlobalpolicy(ctx, bgpServer, &initialConfig.Global.ApplyPolicy.Config)
	added := initialConfig.Neighbors
	if opts.GracefulRestart {
		for i, n := range added {
			if n.GracefulRestart.Config.Enabled {
				added[i].GracefulRestart.State.LocalRestarting = true
			}
		}
	}

	addNeighbors(ctx, bgpServer, added)

	<-stopCh
}

type BgpConfigSet struct {
	Global    config.Global     `mapstructure:"global"`
	Neighbors []config.Neighbor `mapstructure:"neighbors"`
	//DefinedSets       config.DefinedSets        `mapstructure:"defined-sets"`
	//PolicyDefinitions []config.PolicyDefinition `mapstructure:"policy-definitions"`
}

func ReadConfigfile(path, format string) (*BgpConfigSet, error) {
	config := &BgpConfigSet{}
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType(format)
	var err error
	if err = v.ReadInConfig(); err != nil {
		return nil, err
	}
	if err = v.UnmarshalExact(config); err != nil {
		return nil, err
	}
	if err = setDefaultConfigValuesWithViper(v, config); err != nil {
		return nil, err
	}
	return config, nil
}

func setDefaultConfigValuesWithViper(v *viper.Viper, b *BgpConfigSet) error {
	if v == nil {
		v = viper.New()
	}

	if err := config.SetDefaultGlobalConfigValues(&b.Global); err != nil {
		return err
	}

	list, err := extractArray(v.Get("neighbors"))
	if err != nil {
		return err
	}

	for idx, n := range b.Neighbors {
		vv := viper.New()
		if len(list) > idx {
			vv.Set("neighbor", list[idx])
		}

		if err := config.SetDefaultNeighborConfigValuesWithViper(vv, &n, &b.Global); err != nil {
			return err
		}
		b.Neighbors[idx] = n
	}

	list, err = extractArray(v.Get("policy-definitions"))
	if err != nil {
		return err
	}

	/*for idx, p := range b.PolicyDefinitions {
		vv := viper.New()
		if len(list) > idx {
			vv.Set("policy", list[idx])
		}
		if err := config.SetDefaultPolicyConfigValuesWithViper(vv, &p); err != nil {
			return err
		}
		b.PolicyDefinitions[idx] = p
	}*/

	return nil
}

func extractArray(intf interface{}) ([]interface{}, error) {
	if intf != nil {
		list, ok := intf.([]interface{})
		if ok {
			return list, nil
		}
		l, ok := intf.([]map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid configuration: neither []interface{} nor []map[string]interface{}")
		}
		list = make([]interface{}, 0, len(l))
		for _, m := range l {
			list = append(list, m)
		}
		return list, nil
	}
	return nil, nil
}

/*func ConfigSetToRoutingPolicy(c *BgpConfigSet) *config.RoutingPolicy {
	return &config.RoutingPolicy{
		DefinedSets:       c.DefinedSets,
		PolicyDefinitions: c.PolicyDefinitions,
	}
}*/

/*func assignGlobalpolicy(ctx context.Context, bgpServer *server.BgpServer, a *config.ApplyPolicyConfig) {
	toDefaultTable := func(r config.DefaultPolicyType) table.RouteType {
		var def table.RouteType
		switch r {
		case config.DEFAULT_POLICY_TYPE_ACCEPT_ROUTE:
			def = table.ROUTE_TYPE_ACCEPT
		case config.DEFAULT_POLICY_TYPE_REJECT_ROUTE:
			def = table.ROUTE_TYPE_REJECT
		}
		return def
	}
	toPolicies := func(r []string) []*table.Policy {
		p := make([]*table.Policy, 0, len(r))
		for _, n := range r {
			p = append(p, &table.Policy{
				Name: n,
			})
		}
		return p
	}

	def := toDefaultTable(a.DefaultImportPolicy)
	ps := toPolicies(a.ImportPolicyList)
	bgpServer.SetPolicyAssignment(ctx, &api.SetPolicyAssignmentRequest{
		Assignment: table.NewAPIPolicyAssignmentFromTableStruct(&table.PolicyAssignment{
			Name:     table.GLOBAL_RIB_NAME,
			Type:     table.POLICY_DIRECTION_IMPORT,
			Policies: ps,
			Default:  def,
		}),
	})

	def = toDefaultTable(a.DefaultExportPolicy)
	ps = toPolicies(a.ExportPolicyList)
	bgpServer.SetPolicyAssignment(ctx, &api.SetPolicyAssignmentRequest{
		Assignment: table.NewAPIPolicyAssignmentFromTableStruct(&table.PolicyAssignment{
			Name:     table.GLOBAL_RIB_NAME,
			Type:     table.POLICY_DIRECTION_EXPORT,
			Policies: ps,
			Default:  def,
		}),
	})

}*/

func addNeighbors(ctx context.Context, bgpServer *server.BgpServer, added []config.Neighbor) {
	for _, p := range added {
		log.Infof("Peer %v is added", p.State.NeighborAddress)
		if err := bgpServer.AddPeer(ctx, &api.AddPeerRequest{
			Peer: config.NewPeerFromConfigStruct(&p),
		}); err != nil {
			log.Warn(err)
		}
	}
}
