package demo_node_status

import (
	"context"
	goruntime "runtime"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
	taintutil "k8s.io/kubernetes/pkg/util/taints"
)

// registerWithAPIServer registers the node with the cluster master. It is safe
// to call multiple times, but not concurrently (kl.registrationCompleted is
// not locked).
func (kl *Kubelet) registerWithAPIServer() {
	if kl.registrationCompleted {
		return
	}
	step := 100 * time.Millisecond

	for {
		time.Sleep(step)
		step = step * 2
		if step >= 7*time.Second {
			step = 7 * time.Second
		}

		node, err := kl.initialNode(context.TODO())
		if err != nil {
			klog.Errorf("Unable to construct v1.Node object for kubelet: %v", err)
			continue
		}

		klog.Infof("Attempting to register node %s", node.Name)
		registered := kl.tryRegisterWithAPIServer(node)
		if registered {
			klog.Infof("Successfully registered node %s", node.Name)
			kl.registrationCompleted = true
			return
		}
	}
}

// initialNode constructs the initial v1.Node for this Kubelet, incorporating node
// labels, information from the cloud provider, and Kubelet configuration.
func (kl *Kubelet) initialNode(ctx context.Context) (*v1.Node, error) {
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: string(kl.nodeName),
			Labels: map[string]string{
				v1.LabelHostname:      kl.hostname,
				v1.LabelOSStable:      goruntime.GOOS,
				v1.LabelArchStable:    goruntime.GOARCH,
				kubeletapis.LabelOS:   goruntime.GOOS,
				kubeletapis.LabelArch: goruntime.GOARCH,
			},
		},
		Spec: v1.NodeSpec{
			Unschedulable: !kl.registerSchedulable,
		},
	}

	nodeTaints := make([]v1.Taint, 0)
	unschedulableTaint := v1.Taint{
		Key:    v1.TaintNodeUnschedulable,
		Effect: v1.TaintEffectNoSchedule,
	}
	// Taint node with TaintNodeUnschedulable when initializing
	// node to avoid race condition; refer to #63897 for more detail.
	if node.Spec.Unschedulable &&
		!taintutil.TaintExists(nodeTaints, &unschedulableTaint) {
		nodeTaints = append(nodeTaints, unschedulableTaint)
	}
	if len(nodeTaints) > 0 {
		node.Spec.Taints = nodeTaints
	}

	// Initially, set NodeNetworkUnavailable to true.
	node.Status.Conditions = append(node.Status.Conditions, v1.NodeCondition{
		Type:               v1.NodeNetworkUnavailable,
		Status:             v1.ConditionTrue,
		Reason:             "NoRouteCreated",
		Message:            "Node created without a route",
		LastTransitionTime: metav1.NewTime(kl.clock.Now()),
	})

	kl.setNodeStatus(node)

	return node, nil
}

// setNodeStatus fills in the Status fields of the given Node, overwriting any fields that are currently set.
func (kl *Kubelet) setNodeStatus(node *v1.Node) {
	for i, f := range kl.setNodeStatusFuncs {
		klog.V(5).Infof("Setting node status at position %v", i)
		if err := f(node); err != nil {
			klog.Errorf("Failed to set some node status fields: %s", err)
		}
	}
}
