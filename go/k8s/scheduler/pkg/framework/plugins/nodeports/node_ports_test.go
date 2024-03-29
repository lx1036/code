package nodeports

import (
	"context"
	"reflect"
	"strconv"
	"strings"
	"testing"

	framework "k8s-lx1036/k8s/scheduler/pkg/framework"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/diff"
)

func TestNodePorts(test *testing.T) {
	tests := []struct {
		pod        *v1.Pod
		nodeInfo   *framework.NodeInfo
		name       string
		wantStatus *framework.Status
	}{
		{
			pod:      &v1.Pod{},
			nodeInfo: framework.NewNodeInfo(),
			name:     "nothing running",
		},
		{
			pod: newPod("m1", "UDP/127.0.0.1/8080"),
			nodeInfo: framework.NewNodeInfo(
				newPod("m1", "UDP/127.0.0.1/9090")),
			name: "other port",
		},
		{
			pod: newPod("m1", "UDP/127.0.0.1/8080"),
			nodeInfo: framework.NewNodeInfo(
				newPod("m1", "UDP/127.0.0.1/8080")),
			name:       "same udp port",
			wantStatus: framework.NewStatus(framework.Unschedulable, ErrReason),
		},
		{
			pod: newPod("m1", "TCP/127.0.0.1/8080"),
			nodeInfo: framework.NewNodeInfo(
				newPod("m1", "TCP/127.0.0.1/8080")),
			name:       "same tcp port",
			wantStatus: framework.NewStatus(framework.Unschedulable, ErrReason),
		},
		{
			pod: newPod("m1", "TCP/127.0.0.1/8080"),
			nodeInfo: framework.NewNodeInfo(
				newPod("m1", "TCP/127.0.0.2/8080")),
			name: "different host ip",
		},
		{
			pod: newPod("m1", "UDP/127.0.0.1/8080"),
			nodeInfo: framework.NewNodeInfo(
				newPod("m1", "TCP/127.0.0.1/8080")),
			name: "different protocol",
		},
		{
			pod: newPod("m1", "UDP/127.0.0.1/8000", "UDP/127.0.0.1/8080"),
			nodeInfo: framework.NewNodeInfo(
				newPod("m1", "UDP/127.0.0.1/8080")),
			name:       "second udp port conflict",
			wantStatus: framework.NewStatus(framework.Unschedulable, ErrReason),
		},
		{
			pod: newPod("m1", "TCP/127.0.0.1/8001", "UDP/127.0.0.1/8080"),
			nodeInfo: framework.NewNodeInfo(
				newPod("m1", "TCP/127.0.0.1/8001", "UDP/127.0.0.1/8081")),
			name:       "first tcp port conflict",
			wantStatus: framework.NewStatus(framework.Unschedulable, ErrReason),
		},
		{
			pod: newPod("m1", "TCP/0.0.0.0/8001"),
			nodeInfo: framework.NewNodeInfo(
				newPod("m1", "TCP/127.0.0.1/8001")),
			name:       "first tcp port conflict due to 0.0.0.0 hostIP",
			wantStatus: framework.NewStatus(framework.Unschedulable, ErrReason),
		},
		{
			pod: newPod("m1", "TCP/10.0.10.10/8001", "TCP/0.0.0.0/8001"),
			nodeInfo: framework.NewNodeInfo(
				newPod("m1", "TCP/127.0.0.1/8001")),
			name:       "TCP hostPort conflict due to 0.0.0.0 hostIP",
			wantStatus: framework.NewStatus(framework.Unschedulable, ErrReason),
		},
		{
			pod: newPod("m1", "TCP/127.0.0.1/8001"),
			nodeInfo: framework.NewNodeInfo(
				newPod("m1", "TCP/0.0.0.0/8001")),
			name:       "second tcp port conflict to 0.0.0.0 hostIP",
			wantStatus: framework.NewStatus(framework.Unschedulable, ErrReason),
		},
		{
			pod: newPod("m1", "UDP/127.0.0.1/8001"),
			nodeInfo: framework.NewNodeInfo(
				newPod("m1", "TCP/0.0.0.0/8001")),
			name: "second different protocol",
		},
		{
			pod: newPod("m1", "UDP/127.0.0.1/8001"),
			nodeInfo: framework.NewNodeInfo(
				newPod("m1", "TCP/0.0.0.0/8001", "UDP/0.0.0.0/8001")),
			name:       "UDP hostPort conflict due to 0.0.0.0 hostIP",
			wantStatus: framework.NewStatus(framework.Unschedulable, ErrReason),
		},
	}

	for _, fixture := range tests {
		test.Run(fixture.name, func(t *testing.T) {
			p, _ := New(nil, nil)
			cycleState := framework.NewCycleState()
			preFilterStatus := p.(framework.PreFilterPlugin).PreFilter(context.Background(), cycleState, fixture.pod)
			if !preFilterStatus.IsSuccess() {
				t.Errorf("prefilter failed with status: %v", preFilterStatus)
			}
			gotStatus := p.(framework.FilterPlugin).Filter(context.Background(), cycleState, fixture.pod, fixture.nodeInfo)
			if !reflect.DeepEqual(gotStatus, fixture.wantStatus) {
				t.Errorf("status does not match: %v, want: %v", gotStatus, fixture.wantStatus)
			}
		})
	}
}

func TestPreFilterDisabled(t *testing.T) {
	pod := &v1.Pod{}
	nodeInfo := framework.NewNodeInfo()
	node := v1.Node{}
	nodeInfo.SetNode(&node)
	p, _ := New(nil, nil)
	cycleState := framework.NewCycleState()
	gotStatus := p.(framework.FilterPlugin).Filter(context.Background(), cycleState, pod, nodeInfo)
	wantStatus := framework.NewStatus(framework.Error, `error reading "PreFilterNodePorts" from cycleState: not found`)
	if !reflect.DeepEqual(gotStatus, wantStatus) {
		t.Errorf("status does not match: %v, want: %v", gotStatus, wantStatus)
	}
}

func TestGetContainerPorts(t *testing.T) {
	tests := []struct {
		pod1     *v1.Pod
		pod2     *v1.Pod
		expected []*v1.ContainerPort
	}{
		{
			pod1: &v1.Pod{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 8001,
									Protocol:      v1.ProtocolTCP,
								},
								{
									ContainerPort: 8002,
									Protocol:      v1.ProtocolTCP,
								},
							},
						},
						{
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 8003,
									Protocol:      v1.ProtocolTCP,
								},
								{
									ContainerPort: 8004,
									Protocol:      v1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
			pod2: &v1.Pod{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 8011,
									Protocol:      v1.ProtocolTCP,
								},
								{
									ContainerPort: 8012,
									Protocol:      v1.ProtocolTCP,
								},
							},
						},
						{
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 8013,
									Protocol:      v1.ProtocolTCP,
								},
								{
									ContainerPort: 8014,
									Protocol:      v1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
			expected: []*v1.ContainerPort{
				{
					ContainerPort: 8001,
					Protocol:      v1.ProtocolTCP,
				},
				{
					ContainerPort: 8002,
					Protocol:      v1.ProtocolTCP,
				},
				{
					ContainerPort: 8003,
					Protocol:      v1.ProtocolTCP,
				},
				{
					ContainerPort: 8004,
					Protocol:      v1.ProtocolTCP,
				},
				{
					ContainerPort: 8011,
					Protocol:      v1.ProtocolTCP,
				},
				{
					ContainerPort: 8012,
					Protocol:      v1.ProtocolTCP,
				},
				{
					ContainerPort: 8013,
					Protocol:      v1.ProtocolTCP,
				},
				{
					ContainerPort: 8014,
					Protocol:      v1.ProtocolTCP,
				},
			},
		},
	}

	for _, test := range tests {
		result := getContainerPorts(test.pod1, test.pod2)
		if !reflect.DeepEqual(test.expected, result) {
			t.Errorf("Got different result than expected.\nDifference detected on:\n%s", diff.ObjectGoPrintSideBySide(test.expected, result))
		}
	}
}

func newPod(host string, hostPortInfos ...string) *v1.Pod {
	networkPorts := []v1.ContainerPort{}
	for _, portInfo := range hostPortInfos {
		splited := strings.Split(portInfo, "/")
		hostPort, _ := strconv.Atoi(splited[2])

		networkPorts = append(networkPorts, v1.ContainerPort{
			HostIP:   splited[1],
			HostPort: int32(hostPort),
			Protocol: v1.Protocol(splited[0]),
		})
	}
	return &v1.Pod{
		Spec: v1.PodSpec{
			NodeName: host,
			Containers: []v1.Container{
				{
					Ports: networkPorts,
				},
			},
		},
	}
}
