package config

import "github.com/osrg/gobgp/pkg/packet/bgp"

func SetDefaultGlobalConfigValues(g *Global) error {
	if len(g.AfiSafis) == 0 {
		g.AfiSafis = []AfiSafi{}
		for k := range AfiSafiTypeToIntMap {
			g.AfiSafis = append(g.AfiSafis, defaultAfiSafi(k, true))
		}
	}

	if g.Config.Port == 0 {
		g.Config.Port = bgp.BGP_PORT
	}

	if len(g.Config.LocalAddressList) == 0 {
		g.Config.LocalAddressList = []string{"0.0.0.0"} // only ipv4
	}
	return nil
}
