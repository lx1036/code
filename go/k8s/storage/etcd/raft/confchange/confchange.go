package confchange

import (
	"errors"
	"fmt"

	"k8s-lx1036/k8s/storage/etcd/raft/tracker"

	"go.etcd.io/etcd/raft/v3/quorum"
	pb "go.etcd.io/etcd/raft/v3/raftpb"
)

// Changer facilitates configuration changes. It exposes methods to handle
// simple and joint consensus while performing the proper validation that allows
// refusing invalid configuration changes before they affect the active
// configuration.
type Changer struct {
	Tracker   tracker.ProgressTracker
	LastIndex uint64
}

func (c Changer) Simple(ccs ...pb.ConfChangeSingle) (tracker.Config, tracker.ProgressMap, error) {
	cfg, prs, err := c.checkAndCopy()
	if err != nil {
		return tracker.Config{}, nil, err
	}
	if joint(cfg) {
		err := errors.New("can't apply simple config change in joint config")
		return tracker.Config{}, nil, err
	}

	if err := c.apply(&cfg, prs, ccs...); err != nil {
		return tracker.Config{}, nil, err
	}
	if n := symdiff(incoming(c.Tracker.Voters), incoming(cfg.Voters)); n > 1 {
		return tracker.Config{}, nil, errors.New("more than one voter changed without entering joint config")
	}

	return checkAndReturn(cfg, prs)
}

// symdiff returns the count of the symmetric difference between the sets of
// uint64s, i.e. len( (l - r) \union (r - l)).
func symdiff(l, r map[uint64]struct{}) int {
	var n int
	pairs := [][2]quorum.MajorityConfig{
		{l, r}, // count elems in l but not in r
		{r, l}, // count elems in r but not in l
	}
	for _, p := range pairs {
		for id := range p[0] {
			if _, ok := p[1][id]; !ok {
				n++
			}
		}
	}
	return n
}

func (c Changer) apply(cfg *tracker.Config, prs tracker.ProgressMap, ccs ...pb.ConfChangeSingle) error {
	for _, cc := range ccs {
		if cc.NodeID == 0 {
			// etcd replaces the NodeID with zero if it decides (downstream of
			// raft) to not apply a change, so we have to have explicit code
			// here to ignore these.
			continue
		}
		switch cc.Type {
		case pb.ConfChangeAddNode:
			c.makeVoter(cfg, prs, cc.NodeID)
		case pb.ConfChangeAddLearnerNode:
			c.makeLearner(cfg, prs, cc.NodeID)
		case pb.ConfChangeRemoveNode:
			c.remove(cfg, prs, cc.NodeID)
		case pb.ConfChangeUpdateNode:
		default:
			return fmt.Errorf("unexpected conf type %d", cc.Type)
		}
	}
	if len(incoming(cfg.Voters)) == 0 {
		return errors.New("removed all voters")
	}
	return nil
}

// makeVoter adds or promotes the given ID to be a voter in the incoming
// majority config.
func (c Changer) makeVoter(cfg *tracker.Config, prs tracker.ProgressMap, id uint64) {
	pr := prs[id]
	if pr == nil {
		c.initProgress(cfg, prs, id, false /* isLearner */)
		return
	}

	pr.IsLearner = false
	nilAwareDelete(&cfg.Learners, id)
	nilAwareDelete(&cfg.LearnersNext, id)
	incoming(cfg.Voters)[id] = struct{}{}
}

// makeLearner makes the given ID a learner or stages it to be a learner once
// an active joint configuration is exited.
//
// The former happens when the peer is not a part of the outgoing config, in
// which case we either add a new learner or demote a voter in the incoming
// config.
//
// The latter case occurs when the configuration is joint and the peer is a
// voter in the outgoing config. In that case, we do not want to add the peer
// as a learner because then we'd have to track a peer as a voter and learner
// simultaneously. Instead, we add the learner to LearnersNext, so that it will
// be added to Learners the moment the outgoing config is removed by
// LeaveJoint().
func (c Changer) makeLearner(cfg *tracker.Config, prs tracker.ProgressMap, id uint64) {
	pr := prs[id]
	if pr == nil {
		c.initProgress(cfg, prs, id, true /* isLearner */)
		return
	}
	if pr.IsLearner {
		return
	}
	// Remove any existing voter in the incoming config...
	c.remove(cfg, prs, id)
	// ... but save the Progress.
	prs[id] = pr
	// Use LearnersNext if we can't add the learner to Learners directly, i.e.
	// if the peer is still tracked as a voter in the outgoing config. It will
	// be turned into a learner in LeaveJoint().
	//
	// Otherwise, add a regular learner right away.
	if _, onRight := outgoing(cfg.Voters)[id]; onRight {
		nilAwareAdd(&cfg.LearnersNext, id)
	} else {
		pr.IsLearner = true
		nilAwareAdd(&cfg.Learners, id)
	}
}

// remove this peer as a voter or learner from the incoming config.
func (c Changer) remove(cfg *tracker.Config, prs tracker.ProgressMap, id uint64) {
	if _, ok := prs[id]; !ok {
		return
	}

	delete(incoming(cfg.Voters), id)
	nilAwareDelete(&cfg.Learners, id)
	nilAwareDelete(&cfg.LearnersNext, id)

	// If the peer is still a voter in the outgoing config, keep the Progress.
	if _, onRight := outgoing(cfg.Voters)[id]; !onRight {
		delete(prs, id)
	}
}

// initProgress initializes a new progress for the given node or learner.
func (c Changer) initProgress(cfg *tracker.Config, prs tracker.ProgressMap, id uint64, isLearner bool) {
	if !isLearner {
		incoming(cfg.Voters)[id] = struct{}{}
	} else {
		nilAwareAdd(&cfg.Learners, id)
	}
	prs[id] = &tracker.Progress{
		// Initializing the Progress with the last index means that the follower
		// can be probed (with the last index).
		Next:      c.LastIndex,
		Match:     0,
		Inflights: tracker.NewInflights(c.Tracker.MaxInflight),
		IsLearner: isLearner,
		// When a node is first added, we should mark it as recently active.
		// Otherwise, CheckQuorum may cause us to step down if it is invoked
		// before the added node has had a chance to communicate with us.
		RecentActive: true,
	}
}

// checkAndCopy copies the tracker's config and progress map (deeply enough for
// the purposes of the Changer) and returns those copies. It returns an error
// if checkInvariants does.
func (c Changer) checkAndCopy() (tracker.Config, tracker.ProgressMap, error) {
	cfg := c.Tracker.Config.Clone()
	prs := tracker.ProgressMap{}

	for id, pr := range c.Tracker.Progress {
		// A shallow copy is enough because we only mutate the Learner field.
		ppr := *pr
		prs[id] = &ppr
	}

	return checkAndReturn(cfg, prs)
}

// EnterJoint [1]: https://github.com/ongardie/dissertation/blob/master/online-trim.pdf
func (c Changer) EnterJoint(autoLeave bool, ccs ...pb.ConfChangeSingle) (tracker.Config, tracker.ProgressMap, error) {
	cfg, prs, err := c.checkAndCopy()
	if err != nil {
		return tracker.Config{}, nil, err
	}
	if joint(cfg) {
		err := errors.New("config is already joint")
		return tracker.Config{}, nil, err
	}
	if len(incoming(cfg.Voters)) == 0 {
		// We allow adding nodes to an empty config for convenience (testing and
		// bootstrap), but you can't enter a joint state.
		err := errors.New("can't make a zero-voter config joint")
		return tracker.Config{}, nil, err
	}
	// Clear the outgoing config.
	*outgoingPtr(&cfg.Voters) = quorum.MajorityConfig{}
	// Copy incoming to outgoing.
	for id := range incoming(cfg.Voters) {
		outgoing(cfg.Voters)[id] = struct{}{}
	}

	if err := c.apply(&cfg, prs, ccs...); err != nil {
		return tracker.Config{}, nil, err
	}
	cfg.AutoLeave = autoLeave
	return checkAndReturn(cfg, prs)
}

//
// [1]: https://github.com/ongardie/dissertation/blob/master/online-trim.pdf
func (c Changer) LeaveJoint() (tracker.Config, tracker.ProgressMap, error) {
	cfg, prs, err := c.checkAndCopy()
	if err != nil {
		return tracker.Config{}, nil, err
	}
	if !joint(cfg) {
		err := errors.New("can't leave a non-joint config")
		return tracker.Config{}, nil, err
	}
	if len(outgoing(cfg.Voters)) == 0 {
		err := fmt.Errorf("configuration is not joint: %v", cfg)
		return tracker.Config{}, nil, err
	}
	for id := range cfg.LearnersNext {
		nilAwareAdd(&cfg.Learners, id)
		prs[id].IsLearner = true
	}
	cfg.LearnersNext = nil

	for id := range outgoing(cfg.Voters) {
		_, isVoter := incoming(cfg.Voters)[id]
		_, isLearner := cfg.Learners[id]

		if !isVoter && !isLearner {
			delete(prs, id)
		}
	}
	*outgoingPtr(&cfg.Voters) = nil
	cfg.AutoLeave = false

	return checkAndReturn(cfg, prs)
}

// nilAwareAdd populates a map entry, creating the map if necessary.
func nilAwareAdd(m *map[uint64]struct{}, id uint64) {
	if *m == nil {
		*m = map[uint64]struct{}{}
	}
	(*m)[id] = struct{}{}
}

// nilAwareDelete deletes from a map, nil'ing the map itself if it is empty after.
func nilAwareDelete(m *map[uint64]struct{}, id uint64) {
	if *m == nil {
		return
	}
	delete(*m, id)
	if len(*m) == 0 {
		*m = nil
	}
}

// checkAndReturn calls checkInvariants on the input and returns either the
// resulting error or the input.
func checkAndReturn(cfg tracker.Config, prs tracker.ProgressMap) (tracker.Config, tracker.ProgressMap, error) {
	if err := checkInvariants(cfg, prs); err != nil {
		return tracker.Config{}, tracker.ProgressMap{}, err
	}
	return cfg, prs, nil
}

// checkInvariants makes sure that the config and progress are compatible with
// each other. This is used to check both what the Changer is initialized with,
// as well as what it returns.
func checkInvariants(cfg tracker.Config, prs tracker.ProgressMap) error {
	// NB: intentionally allow the empty config. In production we'll never see a
	// non-empty config (we prevent it from being created) but we will need to
	// be able to *create* an initial config, for example during bootstrap (or
	// during tests). Instead of having to hand-code this, we allow
	// transitioning from an empty config into any other legal and non-empty
	// config.
	for _, ids := range []map[uint64]struct{}{
		cfg.Voters.IDs(),
		cfg.Learners,
		cfg.LearnersNext,
	} {
		for id := range ids {
			if _, ok := prs[id]; !ok {
				return fmt.Errorf("no progress for %d", id)
			}
		}
	}

	// Any staged learner was staged because it could not be directly added due
	// to a conflicting voter in the outgoing config.
	for id := range cfg.LearnersNext {
		if _, ok := outgoing(cfg.Voters)[id]; !ok {
			return fmt.Errorf("%d is in LearnersNext, but not Voters[1]", id)
		}
		if prs[id].IsLearner {
			return fmt.Errorf("%d is in LearnersNext, but is already marked as learner", id)
		}
	}
	// Conversely Learners and Voters doesn't intersect at all.
	for id := range cfg.Learners {
		if _, ok := outgoing(cfg.Voters)[id]; ok {
			return fmt.Errorf("%d is in Learners and Voters[1]", id)
		}
		if _, ok := incoming(cfg.Voters)[id]; ok {
			return fmt.Errorf("%d is in Learners and Voters[0]", id)
		}
		if !prs[id].IsLearner {
			return fmt.Errorf("%d is in Learners, but is not marked as learner", id)
		}
	}

	if !joint(cfg) {
		// We enforce that empty maps are nil instead of zero.
		if outgoing(cfg.Voters) != nil {
			return fmt.Errorf("cfg.Voters[1] must be nil when not joint")
		}
		if cfg.LearnersNext != nil {
			return fmt.Errorf("cfg.LearnersNext must be nil when not joint")
		}
		if cfg.AutoLeave {
			return fmt.Errorf("AutoLeave must be false when not joint")
		}
	}

	return nil
}

func incoming(voters quorum.JointConfig) quorum.MajorityConfig {
	return voters[0]
}
func outgoing(voters quorum.JointConfig) quorum.MajorityConfig {
	return voters[1]
}
func joint(cfg tracker.Config) bool {
	return len(outgoing(cfg.Voters)) > 0
}

func outgoingPtr(voters *quorum.JointConfig) *quorum.MajorityConfig {
	return &voters[1]
}

// toConfChangeSingle translates a conf state into 1) a slice of operations creating
// first the config that will become the outgoing one, and then the incoming one, and
// 2) another slice that, when applied to the config resulted from 1), represents the
// ConfState.
func toConfChangeSingle(cs pb.ConfState) (out []pb.ConfChangeSingle, in []pb.ConfChangeSingle) {
	for _, id := range cs.VotersOutgoing {
		// If there are outgoing voters, first add them one by one so that the
		// (non-joint) config has them all.
		out = append(out, pb.ConfChangeSingle{
			Type:   pb.ConfChangeAddNode,
			NodeID: id,
		})
	}

	// First, we'll remove all of the outgoing voters.
	for _, id := range cs.VotersOutgoing {
		in = append(in, pb.ConfChangeSingle{
			Type:   pb.ConfChangeRemoveNode,
			NodeID: id,
		})
	}
	// Then we'll add the incoming voters and learners.
	for _, id := range cs.Voters {
		in = append(in, pb.ConfChangeSingle{
			Type:   pb.ConfChangeAddNode,
			NodeID: id,
		})
	}
	for _, id := range cs.Learners {
		in = append(in, pb.ConfChangeSingle{
			Type:   pb.ConfChangeAddLearnerNode,
			NodeID: id,
		})
	}
	// Same for LearnersNext; these are nodes we want to be learners but which
	// are currently voters in the outgoing config.
	for _, id := range cs.LearnersNext {
		in = append(in, pb.ConfChangeSingle{
			Type:   pb.ConfChangeAddLearnerNode,
			NodeID: id,
		})
	}

	return out, in
}

func Restore(changer Changer, confState pb.ConfState) (tracker.Config, tracker.ProgressMap, error) {
	outgoing, incoming := toConfChangeSingle(confState)

	var ops []func(Changer) (tracker.Config, tracker.ProgressMap, error)
	if len(outgoing) == 0 {
		// No outgoing config, so just apply the incoming changes one by one.
		for _, cc := range incoming {
			c := cc // loop-local copy
			ops = append(ops, func(changer Changer) (tracker.Config, tracker.ProgressMap, error) {
				return changer.Simple(c)
			})
		}
	} else {
		for _, cc := range outgoing {
			cc := cc // loop-local copy
			ops = append(ops, func(changer Changer) (tracker.Config, tracker.ProgressMap, error) {
				return changer.Simple(cc)
			})
		}

		ops = append(ops, func(changer Changer) (tracker.Config, tracker.ProgressMap, error) {
			return changer.EnterJoint(confState.AutoLeave, incoming...)
		})
	}

	return chain(changer, ops...)
}

func chain(chg Changer, ops ...func(Changer) (tracker.Config, tracker.ProgressMap, error)) (tracker.Config, tracker.ProgressMap, error) {
	for _, op := range ops {
		cfg, prs, err := op(chg)
		if err != nil {
			return tracker.Config{}, nil, err
		}
		chg.Tracker.Config = cfg
		chg.Tracker.Progress = prs
	}

	return chg.Tracker.Config, chg.Tracker.Progress, nil
}
