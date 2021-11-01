package main

import (
	"context"

	"k8s-lx1036/k8s/network/bgp/pkg/config"
	"k8s-lx1036/k8s/network/bgp/pkg/server"

	api "github.com/osrg/gobgp/api"
	"github.com/osrg/gobgp/pkg/packet/bgp"
	"github.com/spf13/viper"
)

type BgpConfigSet struct {
	Global     config.Global      `mapstructure:"global"`
	Neighbors  []config.Neighbor  `mapstructure:"neighbors"`
	PeerGroups []config.PeerGroup `mapstructure:"peer-groups"`
	BmpServers []config.BmpServer `mapstructure:"bmp-servers"`
	MrtDump    []config.Mrt       `mapstructure:"mrt-dump"`
	//Zebra             config.Zebra              `mapstructure:"zebra"`
	//Collector         config.Collector          `mapstructure:"collector"`
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
