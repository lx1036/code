package nodelabel

import (
	"context"
	"fmt"

	"k8s-lx1036/k8s/scheduler/pkg/apis/config"
	"k8s-lx1036/k8s/scheduler/pkg/apis/config/validation"
	framework "k8s-lx1036/k8s/scheduler/pkg/framework"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const Name = "NodeLabel"

const (
	// ErrReasonPresenceViolated is used for CheckNodeLabelPresence predicate error.
	ErrReasonPresenceViolated = "node(s) didn't have the requested labels"
)

// NodeLabel checks whether a pod can fit based on the node labels which match a filter that it requests.
type NodeLabel struct {
	handle framework.FrameworkHandle
	args   config.NodeLabelArgs
}

func (pl *NodeLabel) Name() string {
	return Name
}

// Filter invoked at the filter extension point.
// It checks whether all of the specified labels exists on a node or not, regardless of their value
//
// Consider the cases where the nodes are placed in regions/zones/racks and these are identified by labels
// In some cases, it is required that only nodes that are part of ANY of the defined regions/zones/racks be selected
//
// Alternately, eliminating nodes that have a certain label, regardless of value, is also useful
// A node may have a label with "retiring" as key and the date as the value
// and it may be desirable to avoid scheduling new pods on this node.
func (pl *NodeLabel) Filter(ctx context.Context, _ *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	node := nodeInfo.Node()
	if node == nil {
		return framework.NewStatus(framework.Error, "node not found")
	}

	nodeLabels := labels.Set(node.Labels)
	check := func(labels []string, presence bool) bool {
		for _, label := range labels {
			exists := nodeLabels.Has(label)
			if (exists && !presence) || (!exists && presence) {
				return false
			}
		}

		return true
	}

	if check(pl.args.PresentLabels, true) && check(pl.args.AbsentLabels, false) {
		return nil
	}

	return framework.NewStatus(framework.UnschedulableAndUnresolvable, ErrReasonPresenceViolated)
}

// ScoreExtensions of the Score plugin.
func (pl *NodeLabel) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

func (pl *NodeLabel) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	nodeInfo, err := pl.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	if err != nil || nodeInfo.Node() == nil {
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("getting node %q from Snapshot: %v, node is nil: %v",
			nodeName, err, nodeInfo.Node() == nil))
	}

	node := nodeInfo.Node()
	score := int64(0)
	for _, label := range pl.args.PresentLabelsPreference {
		if labels.Set(node.Labels).Has(label) {
			score += framework.MaxNodeScore
		}
	}
	for _, label := range pl.args.AbsentLabelsPreference {
		if !labels.Set(node.Labels).Has(label) {
			score += framework.MaxNodeScore
		}
	}

	// INFO: "NodeLabel" plugin 这里的 score hook 计算方式
	// Take average score for each label to ensure the score doesn't exceed MaxNodeScore.
	score /= int64(len(pl.args.PresentLabelsPreference) + len(pl.args.AbsentLabelsPreference))

	return score, nil
}

// New initializes a new plugin and returns it.
func New(plArgs runtime.Object, handle framework.FrameworkHandle) (framework.Plugin, error) {
	args, err := getArgs(plArgs)
	if err != nil {
		return nil, err
	}

	if err := validation.ValidateNodeLabelArgs(args); err != nil {
		return nil, err
	}

	return &NodeLabel{
		handle: handle,
		args:   args,
	}, nil
}

func getArgs(obj runtime.Object) (config.NodeLabelArgs, error) {
	ptr, ok := obj.(*config.NodeLabelArgs)
	if !ok {
		return config.NodeLabelArgs{}, fmt.Errorf("want args to be of type NodeLabelArgs, got %T", obj)
	}
	return *ptr, nil
}
