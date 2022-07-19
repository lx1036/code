package subnet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"reflect"
	"time"

	"k8s-lx1036/k8s/network/cni/flannel/pkg/ip"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/klog/v2"
)

var (
	ErrUnimplemented = errors.New("unimplemented")
)

type LeaseAttrs struct {
	// INFO: 这里的 PublicIP 一般都是 nodeIP，vxlan 封包时会把 nodeIP 放到 header 里，打通容器网络和机器网络为一个平面
	PublicIP      ip.IP4
	BackendType   string          `json:",omitempty"`
	BackendData   json.RawMessage `json:",omitempty"`
	BackendV6Data json.RawMessage `json:",omitempty"`
}

type Lease struct {
	EnableIPv4 bool
	Attrs      LeaseAttrs
	Subnet     ip.IP4Net
	Expiration time.Time

	Asof int64
}

func (l *Lease) Key() string {
	return MakeSubnetKey(l.Subnet)
}

func MakeSubnetKey(sn ip.IP4Net) string {
	return sn.StringSep(".", "-")
}

type LeaseWatchResult struct {
	// Either Events or Snapshot will be set.  If Events is empty, it means
	// the cursor was out of range and Snapshot contains the current list
	// of items, even if empty.
	Events   []Event     `json:"events"`
	Snapshot []Lease     `json:"snapshot"`
	Cursor   interface{} `json:"cursor"`
}

// AcquireLease INFO: 这个函数的重点是 subnet 其实来自于 newNode.Spec.PodCIDR, 会用来配置 vxlan 网卡 IP，同时该 cidr 得在 pod cidr 里 "Network": "10.244.0.0/16"
func (subnetMgr *kubeSubnetManager) AcquireLease(ctx context.Context, attrs *LeaseAttrs) (*Lease, error) {
	oldNode, err := subnetMgr.nodeStore.Get(subnetMgr.nodeName)
	if err != nil {
		return nil, err
	}
	newNode := oldNode.DeepCopy()
	if newNode.Spec.PodCIDR == "" {
		return nil, fmt.Errorf("node %q pod cidr not assigned", subnetMgr.nodeName)
	}

	data, err := attrs.BackendData.MarshalJSON()
	if err != nil {
		return nil, err
	}

	_, cidr, err := net.ParseCIDR(newNode.Spec.PodCIDR)
	if err != nil {
		return nil, err
	}

	// patch node annotation if needed
	if newNode.Annotations == nil {
		newNode.Annotations = make(map[string]string)
	}
	newNode.Annotations[subnetMgr.annotations.BackendType] = attrs.BackendType
	newNode.Annotations[subnetMgr.annotations.BackendData] = string(data)
	newNode.Annotations[subnetMgr.annotations.BackendPublicIP] = attrs.PublicIP.String()
	newNode.Annotations[subnetMgr.annotations.SubnetKubeManaged] = "true"
	if !reflect.DeepEqual(newNode.Annotations, oldNode.Annotations) {
		oldData, err := json.Marshal(oldNode)
		if err != nil {
			return nil, err
		}
		newData, err := json.Marshal(newNode)
		if err != nil {
			return nil, err
		}
		patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, corev1.Node{})
		if err != nil {
			return nil, fmt.Errorf("failed to create patch for node %q: %v", subnetMgr.nodeName, err)
		}
		_, err = subnetMgr.kubeClient.CoreV1().Nodes().Patch(ctx, subnetMgr.nodeName, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
		if err != nil {
			return nil, err
		}
	}

	// patch node NetworkUnavailable condition
	if subnetMgr.setNodeNetworkUnavailable {
		klog.Infof(fmt.Sprintf("patching node NodeNetworkUnavailable"))
		if err = subnetMgr.patchNodeNetworkUnavailable(ctx); err != nil {
			klog.Errorf(fmt.Sprintf("unable to set NodeNetworkUnavailable err: %v", err))
		}
	}

	lease := &Lease{
		EnableIPv4: true,
		Attrs:      *attrs,
		Expiration: time.Now().Add(24 * time.Hour),
	}
	if cidr != nil && subnetMgr.enableIPv4 {
		if !containsCIDR(subnetMgr.subnetConf.Network.ToIPNet(), cidr) {
			return nil, fmt.Errorf("subnet %q specified in the flannel net config doesn't contain %q PodCIDR of the %q node",
				subnetMgr.subnetConf.Network, cidr, subnetMgr.nodeName)
		}

		lease.Subnet = ip.FromIPNet(cidr)
	}

	return lease, nil
}

func (subnetMgr *kubeSubnetManager) RenewLease(ctx context.Context, lease *Lease) error {
	return ErrUnimplemented
}

func (subnetMgr *kubeSubnetManager) WatchLease(ctx context.Context, sn ip.IP4Net, cursor interface{}) (LeaseWatchResult, error) {
	return LeaseWatchResult{}, ErrUnimplemented
}

func (subnetMgr *kubeSubnetManager) WatchLeases(ctx context.Context) (interface{}, error) {
	select {
	case event := <-subnetMgr.events:
		return LeaseWatchResult{
			Events: []Event{event},
		}, nil
	case <-ctx.Done():
		return LeaseWatchResult{}, context.Canceled
	}
}

type leaseWatcher struct {
	ownLease *Lease  // 当前 node 的 subnet cidr
	leases   []Lease // 除了当前 node 的其余 nodes 的 subnet
}

func (watcher *leaseWatcher) update(events []Event) []Event {
	var batch []Event
	for _, event := range events {
		if watcher.ownLease != nil && event.Lease.EnableIPv4 &&
			event.Lease.Subnet.Equal(watcher.ownLease.Subnet) {
			continue
		}

		switch event.Type {
		case EventAdded:
			batch = append(batch, watcher.add(&event.Lease))
		case EventRemoved:
			batch = append(batch, watcher.remove(&event.Lease))
		}
	}

	return batch
}

func (watcher *leaseWatcher) add(lease *Lease) Event {
	for i, l := range watcher.leases {
		if l.EnableIPv4 && l.Subnet.Equal(lease.Subnet) {
			watcher.leases[i] = *lease
			return Event{EventAdded, watcher.leases[i]}
		}
	}

	watcher.leases = append(watcher.leases, *lease)
	return Event{EventAdded, *lease}
}

func (watcher *leaseWatcher) remove(lease *Lease) Event {
	for i, l := range watcher.leases {
		if l.EnableIPv4 && l.Subnet.Equal(lease.Subnet) {
			watcher.leases = append(watcher.leases[:i], watcher.leases[i+1:]...) // delete i
			return Event{EventRemoved, l}
		}
	}

	return Event{EventRemoved, *lease}
}

func WatchLeases(ctx context.Context, manager Manager, ownLease *Lease, receiver chan []Event) {
	// INFO: 内存存储 map[node]subnet 映射关系
	lw := &leaseWatcher{
		ownLease: ownLease,
	}
	for {
		result, err := manager.WatchLeases(ctx)
		if err != nil {
			// 锁死则关闭 channel
			if err == context.Canceled || err == context.DeadlineExceeded {
				close(receiver)
				return
			}

			time.Sleep(time.Second) // 并发竞争，继续 watch
			continue
		}

		var batch []Event
		if len(result.Events) > 0 {
			batch = lw.update(result.Events)
		}

		if len(batch) > 0 {
			receiver <- batch
		}
	}
}
