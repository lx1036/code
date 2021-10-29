package server

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"

	api "github.com/osrg/gobgp/api"
	"github.com/osrg/gobgp/pkg/packet/bgp"
	"k8s-lx1036/k8s/network/bgp/pkg/config"
	"k8s-lx1036/k8s/network/bgp/pkg/table"
	"strings"
)

type peer struct {
	tableId           string
	fsm               *fsm
	adjRibIn          *table.AdjRib
	policy            *table.RoutingPolicy
	localRib          *table.TableManager
	prefixLimitWarned map[bgp.RouteFamily]bool
	llgrEndChs        []chan struct{}
}

func newPeer(g *config.Global, conf *config.Neighbor, localRib *table.TableManager, policy *table.RoutingPolicy) *peer {
	peer := &peer{
		localRib:          localRib,
		policy:            policy,
		fsm:               newFSM(g, conf),
		prefixLimitWarned: make(map[bgp.RouteFamily]bool),
	}
	if peer.isRouteServerClient() {
		peer.tableId = conf.State.NeighborAddress
	} else {
		peer.tableId = table.GLOBAL_RIB_NAME
	}
	rfs, _ := config.AfiSafis(conf.AfiSafis).ToRfList()
	peer.adjRibIn = table.NewAdjRib(rfs)
	return peer
}

func (peer *peer) startFSMHandler() {
	handler := newFSMHandler(peer.fsm, peer.fsm.outgoingCh)
	peer.fsm.lock.Lock()
	peer.fsm.h = handler
	peer.fsm.lock.Unlock()
}

func (peer *peer) ID() string {
	peer.fsm.lock.RLock()
	defer peer.fsm.lock.RUnlock()

	return peer.fsm.pConf.State.NeighborAddress
}

func (s *BgpServer) addNeighbor(neighbor *config.Neighbor) error {
	addr, err := neighbor.ExtractNeighborAddress()
	if err != nil {
		return err
	}
	if _, ok := s.neighborMap[addr]; ok {
		return fmt.Errorf("can't overwrite the existing peer: %s", addr)
	}

	var pgConf *config.PeerGroup
	if neighbor.Config.PeerGroup != "" {
		pg, ok := s.peerGroupMap[neighbor.Config.PeerGroup]
		if !ok {
			return fmt.Errorf("no such peer-group: %s", neighbor.Config.PeerGroup)
		}
		pgConf = pg.Conf
	}

	log.WithFields(log.Fields{
		"Topic": "Peer",
	}).Infof("Add a peer configuration for:%s", addr)

	rib := s.globalRib
	if neighbor.RouteServer.Config.RouteServerClient {
		rib = s.rsRib
	}
	peer := newPeer(&s.bgpConfig.Global, neighbor, rib, s.policy)
	s.addIncoming(peer.fsm.incomingCh)
	s.policy.SetPeerPolicy(peer.ID(), neighbor.ApplyPolicy)
	s.neighborMap[addr] = peer
	if name := neighbor.Config.PeerGroup; name != "" {
		s.peerGroupMap[name].AddMember(*neighbor)
	}
	peer.startFSMHandler()
	//s.broadcastPeerState(peer, bgp.BGP_FSM_IDLE, nil)
	return nil
}

func (server *BgpServer) AddPeer(ctx context.Context, r *api.AddPeerRequest) error {
	if r == nil || r.Peer == nil {
		return fmt.Errorf("nil request")
	}
	return server.mgmtOperation(func() error {
		config, err := newNeighborFromAPIStruct(r.Peer)
		if err != nil {
			return err
		}
		return server.addNeighbor(config)
	}, true)
}

func newNeighborFromAPIStruct(a *api.Peer) (*config.Neighbor, error) {
	pconf := &config.Neighbor{}
	if a.Conf != nil {
		pconf.Config.PeerAs = a.Conf.PeerAs
		pconf.Config.LocalAs = a.Conf.LocalAs
		pconf.Config.AuthPassword = a.Conf.AuthPassword
		pconf.Config.RouteFlapDamping = a.Conf.RouteFlapDamping
		pconf.Config.Description = a.Conf.Description
		pconf.Config.PeerGroup = a.Conf.PeerGroup
		pconf.Config.PeerType = config.IntToPeerTypeMap[int(a.Conf.PeerType)]
		pconf.Config.NeighborAddress = a.Conf.NeighborAddress
		pconf.Config.AdminDown = a.Conf.AdminDown
		pconf.Config.NeighborInterface = a.Conf.NeighborInterface
		pconf.Config.Vrf = a.Conf.Vrf
		pconf.AsPathOptions.Config.AllowOwnAs = uint8(a.Conf.AllowOwnAs)
		pconf.AsPathOptions.Config.ReplacePeerAs = a.Conf.ReplacePeerAs

		switch a.Conf.RemovePrivateAs {
		case api.PeerConf_ALL:
			pconf.Config.RemovePrivateAs = config.REMOVE_PRIVATE_AS_OPTION_ALL
		case api.PeerConf_REPLACE:
			pconf.Config.RemovePrivateAs = config.REMOVE_PRIVATE_AS_OPTION_REPLACE
		}

		if a.State != nil {
			localCaps, err := apiutil.UnmarshalCapabilities(a.State.LocalCap)
			if err != nil {
				return nil, err
			}
			remoteCaps, err := apiutil.UnmarshalCapabilities(a.State.RemoteCap)
			if err != nil {
				return nil, err
			}
			pconf.State.LocalCapabilityList = localCaps
			pconf.State.RemoteCapabilityList = remoteCaps

			pconf.State.RemoteRouterId = a.State.RouterId
		}

		for _, af := range a.AfiSafis {
			afiSafi := config.AfiSafi{}
			readMpGracefulRestartFromAPIStruct(&afiSafi.MpGracefulRestart, af.MpGracefulRestart)
			readAfiSafiConfigFromAPIStruct(&afiSafi.Config, af.Config)
			readAfiSafiStateFromAPIStruct(&afiSafi.State, af.Config)
			readApplyPolicyFromAPIStruct(&afiSafi.ApplyPolicy, af.ApplyPolicy)
			readRouteSelectionOptionsFromAPIStruct(&afiSafi.RouteSelectionOptions, af.RouteSelectionOptions)
			readUseMultiplePathsFromAPIStruct(&afiSafi.UseMultiplePaths, af.UseMultiplePaths)
			readPrefixLimitFromAPIStruct(&afiSafi.PrefixLimit, af.PrefixLimits)
			readRouteTargetMembershipFromAPIStruct(&afiSafi.RouteTargetMembership, af.RouteTargetMembership)
			readLongLivedGracefulRestartFromAPIStruct(&afiSafi.LongLivedGracefulRestart, af.LongLivedGracefulRestart)
			readAddPathsFromAPIStruct(&afiSafi.AddPaths, af.AddPaths)
			pconf.AfiSafis = append(pconf.AfiSafis, afiSafi)
		}
	}

	if a.Timers != nil {
		if a.Timers.Config != nil {
			pconf.Timers.Config.ConnectRetry = float64(a.Timers.Config.ConnectRetry)
			pconf.Timers.Config.HoldTime = float64(a.Timers.Config.HoldTime)
			pconf.Timers.Config.KeepaliveInterval = float64(a.Timers.Config.KeepaliveInterval)
			pconf.Timers.Config.MinimumAdvertisementInterval = float64(a.Timers.Config.MinimumAdvertisementInterval)
		}
		if a.Timers.State != nil {
			pconf.Timers.State.KeepaliveInterval = float64(a.Timers.State.KeepaliveInterval)
			pconf.Timers.State.NegotiatedHoldTime = float64(a.Timers.State.NegotiatedHoldTime)
		}
	}
	if a.RouteReflector != nil {
		pconf.RouteReflector.Config.RouteReflectorClusterId = config.RrClusterIdType(a.RouteReflector.RouteReflectorClusterId)
		pconf.RouteReflector.Config.RouteReflectorClient = a.RouteReflector.RouteReflectorClient
	}
	if a.RouteServer != nil {
		pconf.RouteServer.Config.RouteServerClient = a.RouteServer.RouteServerClient
		pconf.RouteServer.Config.SecondaryRoute = a.RouteServer.SecondaryRoute
	}
	if a.GracefulRestart != nil {
		pconf.GracefulRestart.Config.Enabled = a.GracefulRestart.Enabled
		pconf.GracefulRestart.Config.RestartTime = uint16(a.GracefulRestart.RestartTime)
		pconf.GracefulRestart.Config.HelperOnly = a.GracefulRestart.HelperOnly
		pconf.GracefulRestart.Config.DeferralTime = uint16(a.GracefulRestart.DeferralTime)
		pconf.GracefulRestart.Config.NotificationEnabled = a.GracefulRestart.NotificationEnabled
		pconf.GracefulRestart.Config.LongLivedEnabled = a.GracefulRestart.LonglivedEnabled
		pconf.GracefulRestart.State.LocalRestarting = a.GracefulRestart.LocalRestarting
	}
	readApplyPolicyFromAPIStruct(&pconf.ApplyPolicy, a.ApplyPolicy)
	if a.Transport != nil {
		pconf.Transport.Config.LocalAddress = a.Transport.LocalAddress
		pconf.Transport.Config.PassiveMode = a.Transport.PassiveMode
		pconf.Transport.Config.RemotePort = uint16(a.Transport.RemotePort)
		pconf.Transport.Config.BindInterface = a.Transport.BindInterface
	}
	if a.EbgpMultihop != nil {
		pconf.EbgpMultihop.Config.Enabled = a.EbgpMultihop.Enabled
		pconf.EbgpMultihop.Config.MultihopTtl = uint8(a.EbgpMultihop.MultihopTtl)
	}
	if a.TtlSecurity != nil {
		pconf.TtlSecurity.Config.Enabled = a.TtlSecurity.Enabled
		pconf.TtlSecurity.Config.TtlMin = uint8(a.TtlSecurity.TtlMin)
	}
	if a.State != nil {
		pconf.State.SessionState = config.SessionState(strings.ToUpper(string(a.State.SessionState)))
		pconf.State.AdminState = config.IntToAdminStateMap[int(a.State.AdminState)]

		pconf.State.PeerAs = a.State.PeerAs
		pconf.State.PeerType = config.IntToPeerTypeMap[int(a.State.PeerType)]
		pconf.State.NeighborAddress = a.State.NeighborAddress

		if a.State.Messages != nil {
			if a.State.Messages.Sent != nil {
				pconf.State.Messages.Sent.Update = a.State.Messages.Sent.Update
				pconf.State.Messages.Sent.Notification = a.State.Messages.Sent.Notification
				pconf.State.Messages.Sent.Open = a.State.Messages.Sent.Open
				pconf.State.Messages.Sent.Refresh = a.State.Messages.Sent.Refresh
				pconf.State.Messages.Sent.Keepalive = a.State.Messages.Sent.Keepalive
				pconf.State.Messages.Sent.Discarded = a.State.Messages.Sent.Discarded
				pconf.State.Messages.Sent.Total = a.State.Messages.Sent.Total
			}
			if a.State.Messages.Received != nil {
				pconf.State.Messages.Received.Update = a.State.Messages.Received.Update
				pconf.State.Messages.Received.Open = a.State.Messages.Received.Open
				pconf.State.Messages.Received.Refresh = a.State.Messages.Received.Refresh
				pconf.State.Messages.Received.Keepalive = a.State.Messages.Received.Keepalive
				pconf.State.Messages.Received.Discarded = a.State.Messages.Received.Discarded
				pconf.State.Messages.Received.Total = a.State.Messages.Received.Total
			}
		}
	}
	return pconf, nil
}
