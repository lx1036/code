package nodeports

import (
	"context"
	"reflect"
	"strconv"
	"strings"
	"testing"

	framework "k8s-lx1036/k8s/scheduler/pkg/scheduler/framework/v1alpha1"

	v1 "k8s.io/api/core/v1"
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
