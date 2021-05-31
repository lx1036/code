package nodeunschedulable

import (
	"context"
	"reflect"
	"testing"

	framework "k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/v1alpha1"

	v1 "k8s.io/api/core/v1"
)

func TestNodeUnschedulable(t *testing.T) {
	testCases := []struct {
		name       string
		pod        *v1.Pod
		node       *v1.Node
		wantStatus *framework.Status
	}{
		{
			name: "Does not schedule pod to unschedulable node (node.Spec.Unschedulable==true)",
			pod:  &v1.Pod{},
			node: &v1.Node{
				Spec: v1.NodeSpec{
					Unschedulable: true,
				},
			},
			wantStatus: framework.NewStatus(framework.UnschedulableAndUnresolvable, ErrReasonUnschedulable),
		},
		{
			name: "Schedule pod to normal node",
			pod:  &v1.Pod{},
			node: &v1.Node{
				Spec: v1.NodeSpec{
					Unschedulable: false,
				},
			},
		},
		{
			name: "Schedule pod with toleration to unschedulable node (node.Spec.Unschedulable==true)",
			pod: &v1.Pod{
				Spec: v1.PodSpec{
					Tolerations: []v1.Toleration{
						{
							Key:    v1.TaintNodeUnschedulable,
							Effect: v1.TaintEffectNoSchedule,
						},
					},
				},
			},
			node: &v1.Node{
				Spec: v1.NodeSpec{
					Unschedulable: true,
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			nodeInfo := framework.NewNodeInfo()
			nodeInfo.SetNode(test.node)

			p, _ := New(nil, nil)
			gotStatus := p.(framework.FilterPlugin).Filter(context.Background(), nil, test.pod, nodeInfo)
			if !reflect.DeepEqual(gotStatus, test.wantStatus) {
				t.Errorf("status does not match: %v, want: %v", gotStatus, test.wantStatus)
			}
		})
	}
}
