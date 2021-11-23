package config

import (
	"encoding/binary"
	"fmt"
	"net"
	"reflect"
	"strconv"

	"github.com/spf13/viper"
	"k8s-lx1036/k8s/network/bgp/pkg/packet/bgp"
)

const (
	DEFAULT_HOLDTIME                  = 90
	DEFAULT_IDLE_HOLDTIME_AFTER_RESET = 30
	DEFAULT_CONNECT_RETRY             = 120
)

var forcedOverwrittenConfig = []string{
	"neighbor.config.peer-as",
	"neighbor.timers.config.minimum-advertisement-interval",
}

var configuredFields map[string]interface{}

func RegisterConfiguredFields(addr string, n interface{}) {
	if configuredFields == nil {
		configuredFields = make(map[string]interface{})
	}
	configuredFields[addr] = n
}

func defaultAfiSafi(typ AfiSafiType, enable bool) AfiSafi {
	return AfiSafi{
		Config: AfiSafiConfig{
			AfiSafiName: typ,
			Enabled:     enable,
		},
		State: AfiSafiState{
			AfiSafiName: typ,
			Family:      bgp.AddressFamilyValueMap[string(typ)],
		},
	}
}

func SetDefaultNeighborConfigValues(n *Neighbor, pg *PeerGroup, g *Global) error {
	// Determines this function is called against the same Neighbor struct,
	// and if already called, returns immediately.
	if n.State.LocalAs != 0 {
		return nil
	}

	return SetDefaultNeighborConfigValuesWithViper(nil, n, g)
}

func SetDefaultNeighborConfigValuesWithViper(v *viper.Viper, n *Neighbor, g *Global) error {
	if n == nil {
		return fmt.Errorf("neighbor config is nil")
	}
	if g == nil {
		return fmt.Errorf("global config is nil")
	}

	if v == nil {
		v = viper.New()
	}

	if n.Config.LocalAs == 0 {
		n.Config.LocalAs = g.Config.As
		if !g.Confederation.Config.Enabled || n.IsConfederation(g) {
			n.Config.LocalAs = g.Config.As
		} else {
			n.Config.LocalAs = g.Confederation.Config.Identifier
		}
	}
	n.State.LocalAs = n.Config.LocalAs

	if n.Config.PeerAs != n.Config.LocalAs {
		n.Config.PeerType = PEER_TYPE_EXTERNAL
		n.State.PeerType = PEER_TYPE_EXTERNAL
		n.State.RemovePrivateAs = n.Config.RemovePrivateAs
		n.AsPathOptions.State.ReplacePeerAs = n.AsPathOptions.Config.ReplacePeerAs
	} else {
		n.Config.PeerType = PEER_TYPE_INTERNAL
		n.State.PeerType = PEER_TYPE_INTERNAL
		if string(n.Config.RemovePrivateAs) != "" {
			return fmt.Errorf("can't set remove-private-as for iBGP peer")
		}
		if n.AsPathOptions.Config.ReplacePeerAs {
			return fmt.Errorf("can't set replace-peer-as for iBGP peer")
		}
	}

	if n.State.NeighborAddress == "" {
		n.State.NeighborAddress = n.Config.NeighborAddress
	}

	n.State.PeerAs = n.Config.PeerAs
	n.AsPathOptions.State.AllowOwnAs = n.AsPathOptions.Config.AllowOwnAs

	if n.Transport.Config.LocalAddress == "" {
		if n.State.NeighborAddress == "" {
			return fmt.Errorf("no neighbor address/interface specified")
		}
		ipAddr, err := net.ResolveIPAddr("ip", n.State.NeighborAddress)
		if err != nil {
			return err
		}
		localAddress := "0.0.0.0"
		if ipAddr.IP.To4() == nil {
			localAddress = "::"
			if ipAddr.Zone != "" {
				localAddress, err = getIPv6LinkLocalAddress(ipAddr.Zone)
				if err != nil {
					return err
				}
			}
		}
		n.Transport.Config.LocalAddress = localAddress
	}

	if len(n.AfiSafis) == 0 {
		if n.Config.NeighborInterface != "" {
			n.AfiSafis = []AfiSafi{
				defaultAfiSafi(AFI_SAFI_TYPE_IPV4_UNICAST, true),
				defaultAfiSafi(AFI_SAFI_TYPE_IPV6_UNICAST, true),
			}
		} else if ipAddr, err := net.ResolveIPAddr("ip", n.State.NeighborAddress); err != nil {
			return fmt.Errorf("invalid neighbor address: %s", n.State.NeighborAddress)
		} else if ipAddr.IP.To4() != nil {
			n.AfiSafis = []AfiSafi{defaultAfiSafi(AFI_SAFI_TYPE_IPV4_UNICAST, true)}
		} else {
			n.AfiSafis = []AfiSafi{defaultAfiSafi(AFI_SAFI_TYPE_IPV6_UNICAST, true)}
		}
		for i := range n.AfiSafis {
			n.AfiSafis[i].AddPaths.Config.Receive = n.AddPaths.Config.Receive
			n.AfiSafis[i].AddPaths.State.Receive = n.AddPaths.Config.Receive
			n.AfiSafis[i].AddPaths.Config.SendMax = n.AddPaths.Config.SendMax
			n.AfiSafis[i].AddPaths.State.SendMax = n.AddPaths.Config.SendMax
		}
	} else {
		afs, err := extractArray(v.Get("neighbor.afi-safis"))
		if err != nil {
			return err
		}
		for i := range n.AfiSafis {
			vv := viper.New()
			if len(afs) > i {
				vv.Set("afi-safi", afs[i])
			}
			rf, err := bgp.GetRouteFamily(string(n.AfiSafis[i].Config.AfiSafiName))
			if err != nil {
				return err
			}
			n.AfiSafis[i].State.Family = rf
			n.AfiSafis[i].State.AfiSafiName = n.AfiSafis[i].Config.AfiSafiName
			if !vv.IsSet("afi-safi.config.enabled") {
				n.AfiSafis[i].Config.Enabled = true
			}
			n.AfiSafis[i].MpGracefulRestart.State.Enabled = n.AfiSafis[i].MpGracefulRestart.Config.Enabled
			if !vv.IsSet("afi-safi.add-paths.config.receive") {
				if n.AddPaths.Config.Receive {
					n.AfiSafis[i].AddPaths.Config.Receive = n.AddPaths.Config.Receive
				}
			}
			n.AfiSafis[i].AddPaths.State.Receive = n.AfiSafis[i].AddPaths.Config.Receive
			if !vv.IsSet("afi-safi.add-paths.config.send-max") {
				if n.AddPaths.Config.SendMax != 0 {
					n.AfiSafis[i].AddPaths.Config.SendMax = n.AddPaths.Config.SendMax
				}
			}
			n.AfiSafis[i].AddPaths.State.SendMax = n.AfiSafis[i].AddPaths.Config.SendMax
		}
	}

	n.State.Description = n.Config.Description
	n.State.AdminDown = n.Config.AdminDown

	if n.GracefulRestart.Config.Enabled {
		if !v.IsSet("neighbor.graceful-restart.config.restart-time") && n.GracefulRestart.Config.RestartTime == 0 {
			// RFC 4724 4. Operation
			// A suggested default for the Restart Time is a value less than or
			// equal to the HOLDTIME carried in the OPEN.
			n.GracefulRestart.Config.RestartTime = uint16(n.Timers.Config.HoldTime)
		}
		if !v.IsSet("neighbor.graceful-restart.config.deferral-time") && n.GracefulRestart.Config.DeferralTime == 0 {
			// RFC 4724 4.1. Procedures for the Restarting Speaker
			// The value of this timer should be large
			// enough, so as to provide all the peers of the Restarting Speaker with
			// enough time to send all the routes to the Restarting Speaker
			n.GracefulRestart.Config.DeferralTime = uint16(360)
		}
	}

	if n.EbgpMultihop.Config.Enabled {
		if n.TtlSecurity.Config.Enabled {
			return fmt.Errorf("ebgp-multihop and ttl-security are mututally exclusive")
		}
		if n.EbgpMultihop.Config.MultihopTtl == 0 {
			n.EbgpMultihop.Config.MultihopTtl = 255
		}
	} else if n.TtlSecurity.Config.Enabled {
		if n.TtlSecurity.Config.TtlMin == 0 {
			n.TtlSecurity.Config.TtlMin = 255
		}
	}

	if n.RouteReflector.Config.RouteReflectorClient {
		if n.RouteReflector.Config.RouteReflectorClusterId == "" {
			n.RouteReflector.State.RouteReflectorClusterId = RrClusterIdType(g.Config.RouterId)
		} else {
			id := string(n.RouteReflector.Config.RouteReflectorClusterId)
			if ip := net.ParseIP(id).To4(); ip != nil {
				n.RouteReflector.State.RouteReflectorClusterId = n.RouteReflector.Config.RouteReflectorClusterId
			} else if num, err := strconv.ParseUint(id, 10, 32); err == nil {
				ip = make(net.IP, 4)
				binary.BigEndian.PutUint32(ip, uint32(num))
				n.RouteReflector.State.RouteReflectorClusterId = RrClusterIdType(ip.String())
			} else {
				return fmt.Errorf("route-reflector-cluster-id should be specified as IPv4 address or 32-bit unsigned integer")
			}
		}
	}

	return nil
}

func SetDefaultPolicyConfigValuesWithViper(v *viper.Viper, p *PolicyDefinition) error {
	stmts, err := extractArray(v.Get("policy.statements"))
	if err != nil {
		return err
	}
	for i := range p.Statements {
		vv := viper.New()
		if len(stmts) > i {
			vv.Set("statement", stmts[i])
		}
		if !vv.IsSet("statement.actions.route-disposition") {
			p.Statements[i].Actions.RouteDisposition = ROUTE_DISPOSITION_NONE
		}
	}
	return nil
}

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
		g.Config.LocalAddressList = []string{"0.0.0.0", "::"}
	}
	return nil
}

func overwriteConfig(c, pg interface{}, tagPrefix string, v *viper.Viper) {
	nValue := reflect.Indirect(reflect.ValueOf(c))
	nType := reflect.Indirect(nValue).Type()
	pgValue := reflect.Indirect(reflect.ValueOf(pg))
	pgType := reflect.Indirect(pgValue).Type()

	for i := 0; i < pgType.NumField(); i++ {
		field := pgType.Field(i).Name
		tag := tagPrefix + "." + nType.Field(i).Tag.Get("mapstructure")
		if func() bool {
			for _, t := range forcedOverwrittenConfig {
				if t == tag {
					return true
				}
			}
			return false
		}() || !v.IsSet(tag) {
			nValue.FieldByName(field).Set(pgValue.FieldByName(field))
		}
	}
}
