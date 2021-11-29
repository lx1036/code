package server

import (
	"context"
	"fmt"
	api "github.com/osrg/gobgp/api"
	log "github.com/sirupsen/logrus"
	"k8s-lx1036/k8s/network/bgp/pkg/config"
	"k8s-lx1036/k8s/network/bgp/pkg/table"
	"strconv"
)

func (s *BgpServer) ListDefinedSet(ctx context.Context, r *api.ListDefinedSetRequest, fn func(*api.DefinedSet)) error {
	if r == nil {
		return fmt.Errorf("nil request")
	}

	cd, err := s.policy.GetDefinedSet(table.DefinedType(r.DefinedType), r.Name)
	if err != nil {
		return err
	}
	exec := func(d *api.DefinedSet) bool {
		select {
		case <-ctx.Done():
			return true
		default:
			fn(d)
		}
		return false
	}

	for _, cs := range cd.PrefixSets {
		ad := &api.DefinedSet{
			DefinedType: api.DefinedType_PREFIX,
			Name:        cs.PrefixSetName,
			Prefixes: func() []*api.Prefix {
				l := make([]*api.Prefix, 0, len(cs.PrefixList))
				for _, p := range cs.PrefixList {
					elems := _regexpPrefixMaskLengthRange.FindStringSubmatch(p.MasklengthRange)
					min, _ := strconv.ParseUint(elems[1], 10, 32)
					max, _ := strconv.ParseUint(elems[2], 10, 32)

					l = append(l, &api.Prefix{IpPrefix: p.IpPrefix, MaskLengthMin: uint32(min), MaskLengthMax: uint32(max)})
				}
				return l
			}(),
		}
		if exec(ad) {
			return nil
		}
	}
	for _, cs := range cd.NeighborSets {
		ad := &api.DefinedSet{
			DefinedType: api.DefinedType_NEIGHBOR,
			Name:        cs.NeighborSetName,
			List:        cs.NeighborInfoList,
		}
		if exec(ad) {
			return nil
		}
	}
	for _, cs := range cd.BgpDefinedSets.CommunitySets {
		ad := &api.DefinedSet{
			DefinedType: api.DefinedType_COMMUNITY,
			Name:        cs.CommunitySetName,
			List:        cs.CommunityList,
		}
		if exec(ad) {
			return nil
		}
	}
	for _, cs := range cd.BgpDefinedSets.ExtCommunitySets {
		ad := &api.DefinedSet{
			DefinedType: api.DefinedType_EXT_COMMUNITY,
			Name:        cs.ExtCommunitySetName,
			List:        cs.ExtCommunityList,
		}
		if exec(ad) {
			return nil
		}
	}
	for _, cs := range cd.BgpDefinedSets.LargeCommunitySets {
		ad := &api.DefinedSet{
			DefinedType: api.DefinedType_LARGE_COMMUNITY,
			Name:        cs.LargeCommunitySetName,
			List:        cs.LargeCommunityList,
		}
		if exec(ad) {
			return nil
		}
	}
	for _, cs := range cd.BgpDefinedSets.AsPathSets {
		ad := &api.DefinedSet{
			DefinedType: api.DefinedType_AS_PATH,
			Name:        cs.AsPathSetName,
			List:        cs.AsPathList,
		}
		if exec(ad) {
			return nil
		}
	}
	return nil
}

func (s *BgpServer) AddDefinedSet(ctx context.Context, r *api.AddDefinedSetRequest) error {
	if r == nil || r.DefinedSet == nil {
		return fmt.Errorf("nil request")
	}
	set, err := newDefinedSetFromApiStruct(r.DefinedSet)
	if err != nil {
		return err
	}
	return s.policy.AddDefinedSet(set)
}

func (s *BgpServer) DeleteDefinedSet(ctx context.Context, r *api.DeleteDefinedSetRequest) error {
	if r == nil || r.DefinedSet == nil {
		return fmt.Errorf("nil request")
	}
	set, err := newDefinedSetFromApiStruct(r.DefinedSet)
	if err != nil {
		return err
	}
	return s.policy.DeleteDefinedSet(set, r.All)
}

func (s *BgpServer) ListStatement(ctx context.Context, r *api.ListStatementRequest, fn func(*api.Statement)) error {
	if r == nil {
		return fmt.Errorf("nil request")
	}
	statement := s.policy.GetStatement(r.Name)
	statements := make([]*api.Statement, 0, len(statement))
	for _, st := range statement {
		statements = append(statements, toStatementApi(st))
	}
	for _, st := range statements {
		select {
		case <-ctx.Done():
			return nil
		default:
			fn(st)
		}
	}
	return nil
}

func (s *BgpServer) AddStatement(ctx context.Context, r *api.AddStatementRequest) error {
	if r == nil || r.Statement == nil {
		return fmt.Errorf("nil request")
	}
	st, err := newStatementFromApiStruct(r.Statement)
	if err != nil {
		return err
	}
	return s.policy.AddStatement(st)
}

func (s *BgpServer) DeleteStatement(ctx context.Context, r *api.DeleteStatementRequest) error {
	if r == nil || r.Statement == nil {
		return fmt.Errorf("nil request")
	}
	st, err := newStatementFromApiStruct(r.Statement)
	if err == nil {
		err = s.policy.DeleteStatement(st, r.All)
	}
	return err
}

func (s *BgpServer) ListPolicy(ctx context.Context, r *api.ListPolicyRequest, fn func(*api.Policy)) error {
	if r == nil {
		return fmt.Errorf("nil request")
	}
	pl := s.policy.GetPolicy(r.Name)
	l := make([]*api.Policy, 0, len(pl))
	for _, p := range pl {
		l = append(l, table.ToPolicyApi(p))
	}
	for _, p := range l {
		select {
		case <-ctx.Done():
			return nil
		default:
			fn(p)
		}
	}
	return nil
}

func (s *BgpServer) AddPolicy(ctx context.Context, r *api.AddPolicyRequest) error {
	if r == nil || r.Policy == nil {
		return fmt.Errorf("nil request")
	}
	p, err := newPolicyFromApiStruct(r.Policy)
	if err == nil {
		err = s.policy.AddPolicy(p, r.ReferExistingStatements)
	}
	return err
}

func (s *BgpServer) DeletePolicy(ctx context.Context, r *api.DeletePolicyRequest) error {
	if r == nil || r.Policy == nil {
		return fmt.Errorf("nil request")
	}
	p, err := newPolicyFromApiStruct(r.Policy)
	if err != nil {
		return err
	}

	l := make([]string, 0, len(s.neighborMap)+1)
	for _, peer := range s.neighborMap {
		l = append(l, peer.ID())
	}
	l = append(l, table.GLOBAL_RIB_NAME)

	return s.policy.DeletePolicy(p, r.All, r.PreserveStatements, l)
}

func (s *BgpServer) toPolicyInfo(name string, dir api.PolicyDirection) (string, table.PolicyDirection, error) {
	if name == "" {
		return "", table.POLICY_DIRECTION_NONE, fmt.Errorf("empty table name")
	}

	if name == table.GLOBAL_RIB_NAME {
		name = table.GLOBAL_RIB_NAME
	} else {
		peer, ok := s.neighborMap[name]
		if !ok {
			return "", table.POLICY_DIRECTION_NONE, fmt.Errorf("not found peer %s", name)
		}
		if !peer.isRouteServerClient() {
			return "", table.POLICY_DIRECTION_NONE, fmt.Errorf("non-rs-client peer %s doesn't have per peer policy", name)
		}
		name = peer.ID()
	}
	switch dir {
	case api.PolicyDirection_IMPORT:
		return name, table.POLICY_DIRECTION_IMPORT, nil
	case api.PolicyDirection_EXPORT:
		return name, table.POLICY_DIRECTION_EXPORT, nil
	}
	return "", table.POLICY_DIRECTION_NONE, fmt.Errorf("invalid policy type")
}

func (s *BgpServer) ListPolicyAssignment(ctx context.Context, r *api.ListPolicyAssignmentRequest, fn func(*api.PolicyAssignment)) error {
	if r == nil {
		return fmt.Errorf("nil request")
	}
	var a []*api.PolicyAssignment
	names := make([]string, 0, len(s.neighborMap)+1)
	if r.Name == "" {
		names = append(names, table.GLOBAL_RIB_NAME)
		for name, peer := range s.neighborMap {
			if peer.isRouteServerClient() {
				names = append(names, name)
			}
		}
	} else {
		names = append(names, r.Name)
	}
	dirs := make([]api.PolicyDirection, 0, 2)
	if r.Direction == api.PolicyDirection_UNKNOWN {
		dirs = []api.PolicyDirection{api.PolicyDirection_EXPORT, api.PolicyDirection_IMPORT}
	} else {
		dirs = append(dirs, r.Direction)
	}

	a = make([]*api.PolicyAssignment, 0, len(names))
	for _, name := range names {
		for _, dir := range dirs {
			id, dir, err := s.toPolicyInfo(name, dir)
			if err != nil {
				return err
			}
			rt, policies, err := s.policy.GetPolicyAssignment(id, dir)
			if err != nil {
				return err
			}
			t := &table.PolicyAssignment{
				Name:     name,
				Type:     dir,
				Default:  rt,
				Policies: policies,
			}
			a = append(a, table.NewAPIPolicyAssignmentFromTableStruct(t))
		}
	}

	for _, p := range a {
		select {
		case <-ctx.Done():
			return nil
		default:
			fn(p)
		}
	}
	return nil
}

func (s *BgpServer) AddPolicyAssignment(ctx context.Context, r *api.AddPolicyAssignmentRequest) error {
	if r == nil || r.Assignment == nil {
		return fmt.Errorf("nil request")
	}
	id, dir, err := s.toPolicyInfo(r.Assignment.Name, r.Assignment.Direction)
	if err != nil {
		return err
	}

	return s.policy.AddPolicyAssignment(id, dir, toPolicyDefinition(r.Assignment.Policies), defaultRouteType(r.Assignment.DefaultAction))
}

func (s *BgpServer) DeletePolicyAssignment(ctx context.Context, r *api.DeletePolicyAssignmentRequest) error {
	if r == nil || r.Assignment == nil {
		return fmt.Errorf("nil request")
	}
	id, dir, err := s.toPolicyInfo(r.Assignment.Name, r.Assignment.Direction)
	if err != nil {
		return err
	}
	return s.policy.DeletePolicyAssignment(id, dir, toPolicyDefinition(r.Assignment.Policies), r.All)
}

func (s *BgpServer) SetPolicyAssignment(ctx context.Context, r *api.SetPolicyAssignmentRequest) error {
	if r == nil || r.Assignment == nil {
		return fmt.Errorf("nil request")
	}
	id, dir, err := s.toPolicyInfo(r.Assignment.Name, r.Assignment.Direction)
	if err != nil {
		return err
	}
	return s.policy.SetPolicyAssignment(id, dir, toPolicyDefinition(r.Assignment.Policies), defaultRouteType(r.Assignment.DefaultAction))
}

func (s *BgpServer) SetPolicies(ctx context.Context, r *api.SetPoliciesRequest) error {
	if r == nil {
		return fmt.Errorf("nil request")
	}
	rp, err := newRoutingPolicyFromApiStruct(r)
	if err != nil {
		return err
	}

	getConfig := func(id string) (*config.ApplyPolicy, error) {
		f := func(id string, dir table.PolicyDirection) (config.DefaultPolicyType, []string, error) {
			rt, policies, err := s.policy.GetPolicyAssignment(id, dir)
			if err != nil {
				return config.DEFAULT_POLICY_TYPE_REJECT_ROUTE, nil, err
			}
			names := make([]string, 0, len(policies))
			for _, p := range policies {
				names = append(names, p.Name)
			}
			t := config.DEFAULT_POLICY_TYPE_ACCEPT_ROUTE
			if rt == table.ROUTE_TYPE_REJECT {
				t = config.DEFAULT_POLICY_TYPE_REJECT_ROUTE
			}
			return t, names, nil
		}

		c := &config.ApplyPolicy{}
		rt, policies, err := f(id, table.POLICY_DIRECTION_IMPORT)
		if err != nil {
			return nil, err
		}
		c.Config.ImportPolicyList = policies
		c.Config.DefaultImportPolicy = rt
		rt, policies, err = f(id, table.POLICY_DIRECTION_EXPORT)
		if err != nil {
			return nil, err
		}
		c.Config.ExportPolicyList = policies
		c.Config.DefaultExportPolicy = rt
		return c, nil
	}

	ap := make(map[string]config.ApplyPolicy, len(s.neighborMap)+1)
	a, err := getConfig(table.GLOBAL_RIB_NAME)
	if err != nil {
		return err
	}
	ap[table.GLOBAL_RIB_NAME] = *a
	for _, peer := range s.neighborMap {
		peer.fsm.lock.RLock()
		log.WithFields(log.Fields{
			"Topic": "Peer",
			"Key":   peer.fsm.pConf.State.NeighborAddress,
		}).Info("call set policy")
		peer.fsm.lock.RUnlock()
		a, err := getConfig(peer.ID())
		if err != nil {
			return err
		}
		ap[peer.ID()] = *a
	}

	return s.policy.Reset(rp, ap)
}
