package main

import (
	"k8s-lx1036/k8s/network/bgp/pkg/config"
	"k8s-lx1036/k8s/network/bgp/pkg/server"

	api "github.com/osrg/gobgp/api"
	"github.com/osrg/gobgp/pkg/packet/bgp"
	"github.com/spf13/viper"
)

type BgpConfigSet struct {
	Global            config.Global             `mapstructure:"global"`
	Neighbors         []config.Neighbor         `mapstructure:"neighbors"`
	PeerGroups        []config.PeerGroup        `mapstructure:"peer-groups"`
	RpkiServers       []config.RpkiServer       `mapstructure:"rpki-servers"`
	BmpServers        []config.BmpServer        `mapstructure:"bmp-servers"`
	Vrfs              []config.Vrf              `mapstructure:"vrfs"`
	MrtDump           []config.Mrt              `mapstructure:"mrt-dump"`
	Zebra             config.Zebra              `mapstructure:"zebra"`
	Collector         config.Collector          `mapstructure:"collector"`
	DefinedSets       config.DefinedSets        `mapstructure:"defined-sets"`
	PolicyDefinitions []config.PolicyDefinition `mapstructure:"policy-definitions"`
	DynamicNeighbors  []config.DynamicNeighbor  `mapstructure:"dynamic-neighbors"`
}

func ReadConfigfile(path, format string) (*BgpConfigSet, error) {
	// Update config file type, if detectable
	format = detectConfigFileType(path, format)

	bgpConfig := &BgpConfigSet{}
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType(format)
	var err error
	if err = v.ReadInConfig(); err != nil {
		return nil, err
	}
	if err = v.UnmarshalExact(bgpConfig); err != nil {
		return nil, err
	}
	/*if err = setDefaultConfigValuesWithViper(v, bgpConfig); err != nil {
		return nil, err
	}*/

	return bgpConfig, nil
}

func InitialConfig(ctx context.Context, bgpServer *server.BgpServer, newConfig *BgpConfigSet, isGracefulRestart bool) (*BgpConfigSet, error) {
	if err := bgpServer.StartBgp(ctx, &api.StartBgpRequest{
		Global: config.NewGlobalFromConfigStruct(&newConfig.Global),
	}); err != nil {
		log.Fatalf("failed to set global config: %s", err)
	}

	if newConfig.Zebra.Config.Enabled {
		tps := newConfig.Zebra.Config.RedistributeRouteTypeList
		l := make([]string, 0, len(tps))
		for _, t := range tps {
			l = append(l, string(t))
		}
		if err := bgpServer.EnableZebra(ctx, &api.EnableZebraRequest{
			Url:                  newConfig.Zebra.Config.Url,
			RouteTypes:           l,
			Version:              uint32(newConfig.Zebra.Config.Version),
			NexthopTriggerEnable: newConfig.Zebra.Config.NexthopTriggerEnable,
			NexthopTriggerDelay:  uint32(newConfig.Zebra.Config.NexthopTriggerDelay),
			MplsLabelRangeSize:   uint32(newConfig.Zebra.Config.MplsLabelRangeSize),
			SoftwareName:         newConfig.Zebra.Config.SoftwareName,
		}); err != nil {
			log.Fatalf("failed to set zebra config: %s", err)
		}
	}

	if len(newConfig.Collector.Config.Url) > 0 {
		log.Fatal("collector feature is not supported")
	}

	for _, c := range newConfig.RpkiServers {
		if err := bgpServer.AddRpki(ctx, &api.AddRpkiRequest{
			Address:  c.Config.Address,
			Port:     c.Config.Port,
			Lifetime: c.Config.RecordLifetime,
		}); err != nil {
			log.Fatalf("failed to set rpki config: %s", err)
		}
	}
	for _, c := range newConfig.BmpServers {
		if err := bgpServer.AddBmp(ctx, &api.AddBmpRequest{
			Address:           c.Config.Address,
			Port:              c.Config.Port,
			SysName:           c.Config.SysName,
			SysDescr:          c.Config.SysDescr,
			Policy:            api.AddBmpRequest_MonitoringPolicy(c.Config.RouteMonitoringPolicy.ToInt()),
			StatisticsTimeout: int32(c.Config.StatisticsTimeout),
		}); err != nil {
			log.Fatalf("failed to set bmp config: %s", err)
		}
	}
	for _, vrf := range newConfig.Vrfs {
		rd, err := bgp.ParseRouteDistinguisher(vrf.Config.Rd)
		if err != nil {
			log.Fatalf("failed to load vrf rd config: %s", err)
		}

		importRtList, err := marshalRouteTargets(vrf.Config.ImportRtList)
		if err != nil {
			log.Fatalf("failed to load vrf import rt config: %s", err)
		}
		exportRtList, err := marshalRouteTargets(vrf.Config.ExportRtList)
		if err != nil {
			log.Fatalf("failed to load vrf export rt config: %s", err)
		}

		if err := bgpServer.AddVrf(ctx, &api.AddVrfRequest{
			Vrf: &api.Vrf{
				Name:     vrf.Config.Name,
				Rd:       apiutil.MarshalRD(rd),
				Id:       uint32(vrf.Config.Id),
				ImportRt: importRtList,
				ExportRt: exportRtList,
			},
		}); err != nil {
			log.Fatalf("failed to set vrf config: %s", err)
		}
	}
	for _, c := range newConfig.MrtDump {
		if len(c.Config.FileName) == 0 {
			continue
		}
		if err := bgpServer.EnableMrt(ctx, &api.EnableMrtRequest{
			DumpType:         int32(c.Config.DumpType.ToInt()),
			Filename:         c.Config.FileName,
			DumpInterval:     c.Config.DumpInterval,
			RotationInterval: c.Config.RotationInterval,
		}); err != nil {
			log.Fatalf("failed to set mrt config: %s", err)
		}
	}
	p := config.ConfigSetToRoutingPolicy(newConfig)
	rp, err := table.NewAPIRoutingPolicyFromConfigStruct(p)
	if err != nil {
		log.Warn(err)
	} else {
		bgpServer.SetPolicies(ctx, &api.SetPoliciesRequest{
			DefinedSets: rp.DefinedSets,
			Policies:    rp.Policies,
		})
	}

	assignGlobalpolicy(ctx, bgpServer, &newConfig.Global.ApplyPolicy.Config)

	added := newConfig.Neighbors
	addedPg := newConfig.PeerGroups
	if isGracefulRestart {
		for i, n := range added {
			if n.GracefulRestart.Config.Enabled {
				added[i].GracefulRestart.State.LocalRestarting = true
			}
		}
	}

	//addPeerGroups(ctx, bgpServer, addedPg)
	//addDynamicNeighbors(ctx, bgpServer, newConfig.DynamicNeighbors)
	addNeighbors(ctx, bgpServer, added)
	return newConfig, nil
}
