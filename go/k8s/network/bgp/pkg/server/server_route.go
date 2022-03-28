package server

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	api "github.com/osrg/gobgp/v3/api"
	log "github.com/sirupsen/logrus"
	"k8s-lx1036/k8s/network/bgp/pkg/packet/bgp"
	"k8s-lx1036/k8s/network/bgp/pkg/table"
	"net"
)

func (s *BgpServer) ListPath(ctx context.Context, r *api.ListPathRequest, fn func(*api.Destination)) error {
	if r == nil {
		return fmt.Errorf("nil request")
	}
	var tbl *table.Table
	var v map[*table.Path]*table.Validation
	var filtered map[string]*table.Path

	f := func() []*table.LookupPrefix {
		l := make([]*table.LookupPrefix, 0, len(r.Prefixes))
		for _, p := range r.Prefixes {
			l = append(l, &table.LookupPrefix{
				Prefix:       p.Prefix,
				LookupOption: table.LookupOption(p.LookupOption),
			})
		}
		return l
	}

	in := false
	family := bgp.RouteFamily(0)
	if r.Family != nil {
		family = bgp.AfiSafiToRouteFamily(uint16(r.Family.Afi), uint8(r.Family.Safi))
	}
	var err error
	switch r.TableType {
	case api.TableType_LOCAL, api.TableType_GLOBAL:
		tbl, v, err = s.getRib(r.Name, family, f())
	case api.TableType_ADJ_IN:
		in = true
		fallthrough
	case api.TableType_ADJ_OUT:
		tbl, filtered, v, err = s.getAdjRib(r.Name, family, in, r.EnableFiltered, f())
	default:
		return fmt.Errorf("unsupported resource type: %v", r.TableType)
	}

	if err != nil {
		return err
	}

	err = func() error {
		for _, dst := range tbl.GetDestinations() {
			d := api.Destination{
				Prefix: dst.GetNlri().String(),
				Paths:  make([]*api.Path, 0, len(dst.GetAllKnownPathList())),
			}
			knownPathList := dst.GetAllKnownPathList()
			for i, path := range knownPathList {
				p := toPathApi(path, getValidation(v, path), r.EnableNlriBinary, r.EnableAttributeBinary)
				if !table.SelectionOptions.DisableBestPathSelection {
					if i == 0 {
						switch r.TableType {
						case api.TableType_LOCAL, api.TableType_GLOBAL:
							p.Best = true
						}
					} else if s.bgpConfig.Global.UseMultiplePaths.Config.Enabled && path.Equal(knownPathList[i-1]) {
						p.Best = true
					}
				}
				d.Paths = append(d.Paths, p)
				if r.EnableFiltered {
					if _, ok := filtered[path.GetNlri().String()]; ok {
						p.Filtered = true
					}
				}
			}

			select {
			case <-ctx.Done():
				return nil
			default:
				fn(&d)
			}
		}
		return nil
	}()
	return err
}

func (s *BgpServer) addPathList(vrfId string, pathList []*table.Path) error {
	err := s.fixupApiPath(vrfId, pathList)
	if err == nil {
		s.propagateUpdate(nil, pathList)
	}
	return err
}

func (s *BgpServer) addPathStream(vrfId string, pathList []*table.Path) error {
	return s.addPathList(vrfId, pathList)
}

func (s *BgpServer) AddPath(ctx context.Context, r *api.AddPathRequest) (*api.AddPathResponse, error) {
	if r == nil || r.Path == nil {
		return nil, fmt.Errorf("nil request")
	}
	var uuidBytes []byte

	if err := s.active(); err != nil {
		return nil, err
	}

	path, err := api2Path(r.TableType, r.Path, false)
	if err != nil {
		return nil, err
	}
	err = s.addPathList(r.VrfId, []*table.Path{path})
	if err != nil {
		return nil, err
	}
	if id, err := uuid.NewRandom(); err == nil {
		s.uuidMap[pathTokey(path)] = id
		uuidBytes, _ = id.MarshalBinary()
	}

	return &api.AddPathResponse{Uuid: uuidBytes}, err
}

func (s *BgpServer) DeletePath(ctx context.Context, r *api.DeletePathRequest) error {
	if r == nil {
		return fmt.Errorf("nil request")
	}

	if err := s.active(); err != nil {
		return err
	}

	deletePathList := make([]*table.Path, 0)

	pathList, err := func() ([]*table.Path, error) {
		if r.Path != nil {
			path, err := api2Path(r.TableType, r.Path, true)
			return []*table.Path{path}, err
		}
		return []*table.Path{}, nil
	}()
	if err != nil {
		return err
	}

	if len(r.Uuid) > 0 {
		// Delete locally generated path which has the given UUID
		path := func() *table.Path {
			id, _ := uuid.FromBytes(r.Uuid)
			for k, v := range s.uuidMap {
				if v == id {
					for _, path := range s.globalRib.GetPathList(table.GLOBAL_RIB_NAME, 0, s.globalRib.GetRFlist()) {
						if path.IsLocal() && k == pathTokey(path) {
							delete(s.uuidMap, k)
							return path
						}
					}
				}
			}
			return nil
		}()
		if path == nil {
			return fmt.Errorf("can't find a specified path")
		}
		deletePathList = append(deletePathList, path.Clone(true))
	} else if len(pathList) == 0 {
		// Delete all locally generated paths
		families := s.globalRib.GetRFlist()
		if r.Family != nil {
			families = []bgp.RouteFamily{bgp.AfiSafiToRouteFamily(uint16(r.Family.Afi), uint8(r.Family.Safi))}

		}
		for _, path := range s.globalRib.GetPathList(table.GLOBAL_RIB_NAME, 0, families) {
			if path.IsLocal() {
				deletePathList = append(deletePathList, path.Clone(true))
			}
		}
		s.uuidMap = make(map[string]uuid.UUID)
	} else {
		if err := s.fixupApiPath(r.VrfId, pathList); err != nil {
			return err
		}
		deletePathList = pathList
		for _, p := range deletePathList {
			delete(s.uuidMap, pathTokey(p))
		}
	}
	s.propagateUpdate(nil, deletePathList)
	return nil
}

func (s *BgpServer) updatePath(vrfId string, pathList []*table.Path) error {
	if err := s.active(); err != nil {
		return err
	}

	if err := s.fixupApiPath(vrfId, pathList); err != nil {
		return err
	}
	s.propagateUpdate(nil, pathList)
	return nil
}

// INFO: 应用 route policy，判断是否需要 export route
func (s *BgpServer) processOutgoingPaths(peer *peer, paths, olds []*table.Path) []*table.Path {
	if !needToAdvertise(peer) {
		return nil
	}

	outgoing := make([]*table.Path, 0, len(paths))

	for idx, path := range paths {
		var old *table.Path
		if olds != nil {
			old = olds[idx]
		}
		if p := s.filterpath(peer, path, old); p != nil {
			outgoing = append(outgoing, p)
		}
	}
	return outgoing
}

func (s *BgpServer) fixupApiPath(vrfId string, pathList []*table.Path) error {
	for _, path := range pathList {
		if !path.IsWithdraw {
			if _, err := path.GetOrigin(); err != nil {
				return err
			}
		}

		if vrfId != "" {
			vrf := s.globalRib.Vrfs[vrfId]
			if vrf == nil {
				return fmt.Errorf("vrf %s not found", vrfId)
			}
			if err := vrf.ToGlobalPath(path); err != nil {
				return err
			}
		}

		// Address Family specific Handling
		switch nlri := path.GetNlri().(type) {
		case *bgp.EVPNNLRI:
			switch r := nlri.RouteTypeData.(type) {
			case *bgp.EVPNMacIPAdvertisementRoute:
				// MAC Mobility Extended Community
				paths := s.globalRib.GetBestPathList(table.GLOBAL_RIB_NAME, 0, []bgp.RouteFamily{bgp.RF_EVPN})
				if m := getMacMobilityExtendedCommunity(r.ETag, r.MacAddress, paths); m != nil {
					pm := getMacMobilityExtendedCommunity(r.ETag, r.MacAddress, []*table.Path{path})
					if pm == nil {
						path.SetExtCommunities([]bgp.ExtendedCommunityInterface{m}, false)
					} else if pm != nil && pm.Sequence < m.Sequence {
						return fmt.Errorf("invalid MAC mobility sequence number")
					}
				}
			case *bgp.EVPNEthernetSegmentRoute:
				// RFC7432: BGP MPLS-Based Ethernet VPN
				// 7.6. ES-Import Route Target
				// The value is derived automatically for the ESI Types 1, 2,
				// and 3, by encoding the high-order 6-octet portion of the 9-octet ESI
				// Value, which corresponds to a MAC address, in the ES-Import Route
				// Target.
				// Note: If the given path already has the ES-Import Route Target,
				// skips deriving a new one.
				found := false
				for _, extComm := range path.GetExtCommunities() {
					if _, found = extComm.(*bgp.ESImportRouteTarget); found {
						break
					}
				}
				if !found {
					switch r.ESI.Type {
					case bgp.ESI_LACP, bgp.ESI_MSTP, bgp.ESI_MAC:
						mac := net.HardwareAddr(r.ESI.Value[0:6])
						rt := &bgp.ESImportRouteTarget{ESImport: mac}
						path.SetExtCommunities([]bgp.ExtendedCommunityInterface{rt}, false)
					}
				}
			}
		}
	}
	return nil
}

func pathTokey(path *table.Path) string {
	return fmt.Sprintf("%d:%s", path.GetNlri().PathIdentifier(), path.GetNlri().String())
}

func dstsToPaths(id string, as uint32, dsts []*table.Update) ([]*table.Path, []*table.Path, [][]*table.Path) {
	bestList := make([]*table.Path, 0, len(dsts))
	oldList := make([]*table.Path, 0, len(dsts))
	mpathList := make([][]*table.Path, 0, len(dsts))

	for _, dst := range dsts {
		best, old, mpath := dst.GetChanges(id, as, false)
		bestList = append(bestList, best)
		oldList = append(oldList, old)
		if mpath != nil {
			mpathList = append(mpathList, mpath)
		}
	}
	return bestList, oldList, mpathList
}

func (s *BgpServer) sendSecondaryRoutes(peer *peer, newPath *table.Path, dsts []*table.Update) []*table.Path {
	if !needToAdvertise(peer) {
		return nil
	}
	pl := make([]*table.Path, 0, len(dsts))

	f := func(path, old *table.Path) *table.Path {
		path, options, stop := s.prePolicyFilterpath(peer, path, old)
		if stop {
			return nil
		}
		options.Validate = s.roaTable.Validate
		path = peer.policy.ApplyPolicy(peer.TableID(), table.POLICY_DIRECTION_EXPORT, path, options)
		if path != nil {
			return s.postFilterpath(peer, path)
		}
		return nil
	}

	for _, dst := range dsts {
		old := func() *table.Path {
			for _, old := range dst.OldKnownPathList {
				o := f(old, nil)
				if o != nil {
					return o
				}
			}
			return nil
		}()
		path := func() *table.Path {
			for _, known := range dst.KnownPathList {
				path := f(known, old)
				if path != nil {
					return path
				}
			}
			return nil
		}()
		if path != nil {
			pl = append(pl, path)
		} else if old != nil {
			pl = append(pl, old.Clone(true))
		}
	}
	return pl
}

func (s *BgpServer) handleRouteRefresh(peer *peer, e *fsmMsg) []*table.Path {
	m := e.MsgData.(*bgp.BGPMessage)
	rr := m.Body.(*bgp.BGPRouteRefresh)
	rf := bgp.AfiSafiToRouteFamily(rr.AFI, rr.SAFI)

	peer.fsm.lock.RLock()
	_, ok := peer.fsm.rfMap[rf]
	peer.fsm.lock.RUnlock()
	if !ok {
		log.WithFields(log.Fields{
			"Topic": "Peer",
			"Key":   peer.ID(),
			"Data":  rf,
		}).Warn("Route family isn't supported")
		return nil
	}

	peer.fsm.lock.RLock()
	_, ok = peer.fsm.capMap[bgp.BGP_CAP_ROUTE_REFRESH]
	peer.fsm.lock.RUnlock()
	if !ok {
		log.WithFields(log.Fields{
			"Topic": "Peer",
			"Key":   peer.ID(),
		}).Warn("ROUTE_REFRESH received but the capability wasn't advertised")
		return nil
	}
	rfList := []bgp.RouteFamily{rf}
	accepted, _ := s.getBestFromLocal(peer, rfList)
	return accepted
}

func filterpath(peer *peer, path, old *table.Path) *table.Path {
	if path == nil {
		return nil
	}

	peer.fsm.lock.RLock()
	_, ok := peer.fsm.rfMap[path.GetRouteFamily()]
	peer.fsm.lock.RUnlock()
	if !ok {
		return nil
	}

	//RFC4684 Constrained Route Distribution
	peer.fsm.lock.RLock()
	_, y := peer.fsm.rfMap[bgp.RF_RTC_UC]
	peer.fsm.lock.RUnlock()
	if y && path.GetRouteFamily() != bgp.RF_RTC_UC {
		ignore := true
		for _, ext := range path.GetExtCommunities() {
			for _, p := range peer.adjRibIn.PathList([]bgp.RouteFamily{bgp.RF_RTC_UC}, true) {
				rt := p.GetNlri().(*bgp.RouteTargetMembershipNLRI).RouteTarget
				// Note: nil RT means the default route target
				if rt == nil || ext.String() == rt.String() {
					ignore = false
					break
				}
			}
			if !ignore {
				break
			}
		}
		if ignore {
			log.WithFields(log.Fields{
				"Topic": "Peer",
				"Key":   peer.ID(),
				"Data":  path,
			}).Debug("Filtered by Route Target Constraint, ignore")
			return nil
		}
	}

	//iBGP handling
	if peer.isIBGPPeer() {
		ignore := false
		if !path.IsLocal() {
			ignore = true
			info := path.GetSource()
			//if the path comes from eBGP peer
			if info.AS != peer.AS() {
				ignore = false
			}
			if info.RouteReflectorClient {
				ignore = false
			}
			if peer.isRouteReflectorClient() {
				// RFC4456 8. Avoiding Routing Information Loops
				// If the local CLUSTER_ID is found in the CLUSTER_LIST,
				// the advertisement received SHOULD be ignored.
				for _, clusterID := range path.GetClusterList() {
					peer.fsm.lock.RLock()
					rrClusterID := peer.fsm.peerInfo.RouteReflectorClusterID
					peer.fsm.lock.RUnlock()
					if clusterID.Equal(rrClusterID) {
						log.WithFields(log.Fields{
							"Topic":     "Peer",
							"Key":       peer.ID(),
							"ClusterID": clusterID,
							"Data":      path,
						}).Debug("cluster list path attribute has local cluster id, ignore")
						return nil
					}
				}
				ignore = false
			}
		}

		if ignore {
			if !path.IsWithdraw && old != nil {
				oldSource := old.GetSource()
				if old.IsLocal() || oldSource.Address.String() != peer.ID() && oldSource.AS != peer.AS() {
					// In this case, we suppose this peer has the same prefix
					// received from another iBGP peer.
					// So we withdraw the old best which was injected locally
					// (from CLI or gRPC for example) in order to avoid the
					// old best left on peers.
					// Also, we withdraw the eBGP route which is the old best.
					// When we got the new best from iBGP, we don't advertise
					// the new best and need to withdraw the old best.
					return old.Clone(true)
				}
			}
			log.WithFields(log.Fields{
				"Topic": "Peer",
				"Key":   peer.ID(),
				"Data":  path,
			}).Debug("From same AS, ignore.")
			return nil
		}
	}

	if path = peer.filterPathFromSourcePeer(path, old); path == nil {
		return nil
	}

	if !peer.isRouteServerClient() && isASLoop(peer, path) {
		return nil
	}
	return path
}

func (s *BgpServer) prePolicyFilterpath(peer *peer, path, old *table.Path) (*table.Path, *table.PolicyOptions, bool) {
	// Special handling for RTM NLRI.
	if path != nil && path.GetRouteFamily() == bgp.RF_RTC_UC && !path.IsWithdraw {
		// If the given "path" is locally generated and the same with "old", we
		// assumes "path" was already sent before. This assumption avoids the
		// infinite UPDATE loop between Route Reflector and its clients.
		if path.IsLocal() && path.Equal(old) {
			peer.fsm.lock.RLock()
			log.WithFields(log.Fields{
				"Topic": "Peer",
				"Key":   peer.fsm.pConf.State.NeighborAddress,
				"Path":  path,
			}).Debug("given rtm nlri is already sent, skipping to advertise")
			peer.fsm.lock.RUnlock()
			return nil, nil, true
		}

		if old != nil && old.IsLocal() {
			// We assumes VRF with the specific RT is deleted.
			path = old.Clone(true)
		} else if peer.isRouteReflectorClient() {
			// We need to send the path even if the peer is originator of the
			// path in order to signal that the client should distribute route
			// with the given RT.
		} else {
			// We send a path even if it is not the best path. See comments in
			// (*Destination) GetChanges().
			dst := peer.localRib.GetDestination(path)
			path = nil
			for _, p := range dst.GetKnownPathList(peer.TableID(), peer.AS()) {
				srcPeer := p.GetSource()
				if peer.ID() != srcPeer.Address.String() {
					if srcPeer.RouteReflectorClient {
						// The path from a RR client is preferred than others
						// for the case that RR and non RR client peering
						// (e.g., peering of different RR clusters).
						path = p
						break
					} else if path == nil {
						path = p
					}
				}
			}
		}
	}

	// only allow vpnv4 and vpnv6 paths to be advertised to VRFed neighbors.
	// also check we can import this path using table.CanImportToVrf()
	// if we can, make it local path by calling (*Path).ToLocal()
	peer.fsm.lock.RLock()
	peerVrf := peer.fsm.pConf.Config.Vrf
	peer.fsm.lock.RUnlock()
	if path != nil && peerVrf != "" {
		if f := path.GetRouteFamily(); f != bgp.RF_IPv4_VPN && f != bgp.RF_IPv6_VPN && f != bgp.RF_FS_IPv4_VPN && f != bgp.RF_FS_IPv6_VPN {
			return nil, nil, true
		}
		vrf := peer.localRib.Vrfs[peerVrf]
		if table.CanImportToVrf(vrf, path) {
			path = path.ToLocal()
		} else {
			return nil, nil, true
		}
	}

	// replace-peer-as handling
	peer.fsm.lock.RLock()
	if path != nil && !path.IsWithdraw && peer.fsm.pConf.AsPathOptions.State.ReplacePeerAs {
		path = path.ReplaceAS(peer.fsm.pConf.Config.LocalAs, peer.fsm.pConf.Config.PeerAs)
	}
	peer.fsm.lock.RUnlock()

	if path = filterpath(peer, path, old); path == nil {
		return nil, nil, true
	}

	peer.fsm.lock.RLock()
	options := &table.PolicyOptions{
		Info:       peer.fsm.peerInfo,
		OldNextHop: path.GetNexthop(),
	}
	path = table.UpdatePathAttrs(peer.fsm.gConf, peer.fsm.pConf, peer.fsm.peerInfo, path)
	peer.fsm.lock.RUnlock()

	return path, options, false
}

func (s *BgpServer) postFilterpath(peer *peer, path *table.Path) *table.Path {
	// draft-uttaro-idr-bgp-persistence-02
	// 4.3.  Processing LLGR_STALE Routes
	//
	// The route SHOULD NOT be advertised to any neighbor from which the
	// Long-lived Graceful Restart Capability has not been received.  The
	// exception is described in the Optional Partial Deployment
	// Procedure section (Section 4.7).  Note that this requirement
	// implies that such routes should be withdrawn from any such neighbor.
	if path != nil && !path.IsWithdraw && !peer.isLLGREnabledFamily(path.GetRouteFamily()) && path.IsLLGRStale() {
		// we send unnecessary withdrawn even if we didn't
		// sent the route.
		path = path.Clone(true)
	}

	// remove local-pref attribute
	// we should do this after applying export policy since policy may
	// set local-preference
	if path != nil && !peer.isIBGPPeer() && !peer.isRouteServerClient() {
		path.RemoveLocalPref()
	}
	return path
}

// INFO: 应用 route policy，判断是否需要 export route
func (s *BgpServer) filterpath(peer *peer, path, old *table.Path) *table.Path {
	path, options, stop := s.prePolicyFilterpath(peer, path, old)
	if stop {
		return path
	}
	options.Validate = s.roaTable.Validate
	path = peer.policy.ApplyPolicy(peer.TableID(), table.POLICY_DIRECTION_EXPORT, path, options)
	// When 'path' is filtered (path == nil), check 'old' has been sent to this peer.
	// If it has, send withdrawal to the peer.
	if path == nil && old != nil {
		o := peer.policy.ApplyPolicy(peer.TableID(), table.POLICY_DIRECTION_EXPORT, old, options)
		if o != nil {
			path = old.Clone(true)
		}
	}

	return s.postFilterpath(peer, path)
}
