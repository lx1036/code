package controller

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"reflect"
	"time"

	v1 "k8s-lx1036/k8s/storage/etcd/etcd-operator/pkg/apis/etcd.k9s.io/v1"
	"k8s-lx1036/k8s/storage/etcd/etcd-operator/pkg/client/clientset/versioned"

	"github.com/google/uuid"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

const (
	ReconcileInterval = 10 * time.Second

	EtcdDefaultRequestTimeout = 5 * time.Second
	EtcdDefaultDialTimeout    = 5 * time.Second
)

// ErrLostQuorum indicates that the etcd cluster lost its quorum.
var ErrLostQuorum = errors.New("lost quorum")

type ClusterConfig struct {
	kubeClient        kubernetes.Interface
	etcdClusterClient *versioned.Clientset
}

type Cluster struct {
	etcdCluster       *v1.EtcdCluster
	etcdClusterClient *versioned.Clientset

	kubeClient kubernetes.Interface

	// pod name 就是 member name, etcd cluster 中的所有 members
	members MemberSet

	stopCh chan struct{}

	tlsConfig *tls.Config

	eventClient corev1.Event

	// INFO: 缓存 EtcdClusterStatus，有种状态机感觉，显示处于不同状态
	//  cluster.status 会和 EtcdCluster.status 保持一致
	status v1.EtcdClusterStatus
}

func NewCluster(clusterConfig *ClusterConfig, etcdCluster *v1.EtcdCluster) *Cluster {
	cluster := &Cluster{
		etcdCluster:       etcdCluster,
		etcdClusterClient: clusterConfig.etcdClusterClient,
		kubeClient:        clusterConfig.kubeClient,
		status:            *(etcdCluster.Status.DeepCopy()),
	}

	// INFO: 这里有个重要逻辑，setup() 先启动一个 etcd seed pod，然后 run() 里去根据 EtcdCluster .spec.size 去 reconcile，
	//  一个一个去创建 etcd pod。但是，setup() 的 etcd pod 是 "new"，run() reconcile 的 etcd pod 是 "existing"!!!
	go func() {
		if err := cluster.createSeedMember(); err != nil {
			klog.Errorf(fmt.Sprintf("[NewCluster]cluster failed to setup err %v", err))
			if cluster.status.Phase != v1.ClusterPhaseFailed {
				cluster.status.SetPhase(v1.ClusterPhaseFailed)
				cluster.status.SetReason(err.Error())
				if err := cluster.UpdateEtcdClusterStatus(); err != nil {
					klog.Errorf(fmt.Sprintf("[NewCluster]cluster failed to update etcd cluster status err %v", err))
				}
			}
		}

		cluster.run()
	}()

	return cluster
}

// INFO: 创建一个 seed etcd pod
func (cluster *Cluster) createSeedMember() error {
	var shouldCreateCluster bool
	switch cluster.status.Phase {
	case v1.ClusterPhaseNone:
		shouldCreateCluster = true
	case v1.ClusterPhaseCreating:

	case v1.ClusterPhaseRunning:
		shouldCreateCluster = false

	default:
		return fmt.Errorf("")
	}

	if shouldCreateCluster {
		return cluster.create()
	}

	return nil
}

func (cluster *Cluster) create() error {
	cluster.status.SetPhase(v1.ClusterPhaseCreating)
	err := cluster.UpdateEtcdClusterStatus()
	if err != nil {
		return err
	}

	return cluster.prepareSeedMember()
}

// UpdateEtcdClusterStatus INFO: 这里更新 etcdCluster status phase字段值，先 api-server 获取最新的再去 update，防止 conflict error
func (cluster *Cluster) UpdateEtcdClusterStatus() error {
	if reflect.DeepEqual(cluster.etcdCluster.Status, cluster.status) {
		return nil
	}

	newCluster, err := cluster.etcdClusterClient.EtcdV1().EtcdClusters(cluster.etcdCluster.Namespace).Get(context.TODO(),
		cluster.etcdCluster.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	newCluster.Status = cluster.status
	newCluster, err = cluster.etcdClusterClient.EtcdV1().EtcdClusters(cluster.etcdCluster.Namespace).Update(context.TODO(),
		newCluster, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	cluster.etcdCluster = newCluster
	return nil
}

// INFO: 首先创建一个 seed etcd pod，然后再 reconcile size 依次创建 etcd pod
func (cluster *Cluster) prepareSeedMember() error {
	return cluster.startSeedMember()
}

func (cluster *Cluster) startSeedMember() error {
	member := &Member{
		Name:      UniqueMemberName(cluster.etcdCluster.Name),
		Namespace: cluster.etcdCluster.Namespace,
		//SecurePeer:   cluster.isSecurePeer(),
		//SecureClient: cluster.isSecureClient(),
	}
	memberSet := NewMemberSet(member)
	if err := cluster.createPod(memberSet, member, "new"); err != nil {
		return fmt.Errorf("failed to create seed member (%s): %v", member.Name, err)
	}

	// INFO: cluster.members 先缓存一个 "new" state member
	cluster.members = memberSet

	return nil
}

func (cluster *Cluster) isPodPVEnabled() bool {
	if podPolicy := cluster.etcdCluster.Spec.Pod; podPolicy != nil {
		return podPolicy.PersistentVolumeClaimSpec != nil
	}

	return false
}

// INFO: 创建一个 etcd pod
func (cluster *Cluster) createPod(members MemberSet, member *Member, state string) error {
	pod := NewEtcdPod(member, members.PeerURLPairs(), cluster.etcdCluster.Name, state, uuid.New().String(),
		cluster.etcdCluster.Spec, cluster.etcdCluster.AsOwner())

	if cluster.isPodPVEnabled() {
		pvc := NewEtcdPodPVC(member, *cluster.etcdCluster.Spec.Pod.PersistentVolumeClaimSpec, cluster.etcdCluster.Name, cluster.etcdCluster.Namespace, cluster.etcdCluster.AsOwner())
		_, err := cluster.kubeClient.CoreV1().PersistentVolumeClaims(cluster.etcdCluster.Namespace).Create(context.TODO(), pvc, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("[createPod]failed to create PVC for member (%s): %v", member.Name, err))
		}
		AddEtcdVolumeToPod(pod, pvc)
	} else {
		AddEtcdVolumeToPod(pod, nil)
	}

	_, err := cluster.kubeClient.CoreV1().Pods(cluster.etcdCluster.Namespace).Create(context.TODO(), pod,
		metav1.CreateOptions{})

	return err
}

// INFO: 不断去reconcile，后续 etcd pod 以 existing 方式加入 cluster 中
func (cluster *Cluster) run() {
	var rerr error
	for {
		select {
		case <-cluster.stopCh:
			return
		//case event := <-cluster.eventCh: // 监听 update 事件

		case <-time.After(ReconcileInterval):
			running, pending, err := cluster.pollPods()
			if err != nil {
				klog.Errorf(fmt.Sprintf("[run]fail to poll pods: %v", err))
				continue
			}
			// 如果有 pending pod，说明第一个 etcd pod 还没起来
			if len(pending) > 0 {
				// Pod startup might take long, e.g. pulling image. It would deterministically become running or succeeded/failed later.
				klog.Infof("[run]skip reconciliation: running (%v), pending (%v)", GetPodNames(running), GetPodNames(pending))
				continue
			}
			if len(running) == 0 { // etcd pod 是一个一个起来的
				// TODO: how to handle this case?
				klog.Warningf("[run]all etcd pods are dead.")
				break
			}

			rerr = cluster.reconcile(running)
			if rerr != nil {
				klog.Errorf(fmt.Sprintf("[run]failed to reconcile err %v", rerr))
				break
			}
		}
	}
}

func (cluster *Cluster) pollPods() (running, pending []*corev1.Pod, err error) {
	podList, err := cluster.kubeClient.CoreV1().Pods(cluster.etcdCluster.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(GetLabelsForEtcdPod(cluster.etcdCluster.Name)).String(),
	})
	if err != nil {
		return nil, nil, err
	}

	for _, pod := range podList.Items {
		// Avoid polling deleted pods. k8s issue where deleted pods would sometimes show the status Pending
		// See https://github.com/coreos/etcd-operator/issues/1693
		if pod.DeletionTimestamp != nil {
			continue
		}
		if len(pod.OwnerReferences) == 0 {
			continue
		}
		if pod.OwnerReferences[0].UID != cluster.etcdCluster.UID {
			klog.Warningf(fmt.Sprintf("[pollPods]ignore pod %s/%s, owner %s is not %s", pod.Namespace, pod.Name,
				pod.OwnerReferences[0].UID, cluster.etcdCluster.UID))
			continue
		}
		switch pod.Status.Phase {
		case corev1.PodRunning:
			running = append(running, &pod)
		case corev1.PodPending:
			running = append(running, &pod)
		}
	}

	return running, pending, nil
}

// INFO: 依次开始创建剩余的 etcd pod，注意这里是 --initial-cluster-state=existing
// reconcile reconciles cluster current state to desired state specified by spec.
// - it tries to reconcile the cluster to desired size.
// - if the cluster needs for upgrade, it tries to upgrade old member one by one.
func (cluster *Cluster) reconcile(pods []*corev1.Pod) error {
	clusterSpec := cluster.etcdCluster.Spec
	runningMemberSet := podsToMemberSet(pods, cluster.isSecureClient())
	// INFO: (1) ; (2)resize
	if !runningMemberSet.IsEqual(cluster.members) || cluster.members.Size() != clusterSpec.Size {
		return cluster.reconcileMembers(runningMemberSet)
	}

	return nil
}

// reconcileMembers reconciles
// - running pods on k8s and cluster membership
// - cluster membership and expected size of etcd cluster
// Steps:
// 1. Remove all pods from running set that does not belong to member set.
// 2. L consist of remaining pods of running
// 3. If L = members, the current state matches the membership state. END.
// 4. If len(L) < len(members)/2 + 1, return quorum lost error.
// 5. Add one missing member. END.
// INFO:
//
//	(1) 先与 running diff，删除 unknown etcd member
//	(2) 再去 resize 到期望节点数量
func (cluster *Cluster) reconcileMembers(running MemberSet) error {
	unknownMembers := running.Diff(cluster.members)
	if unknownMembers.Size() > 0 {
		klog.Infof("removing unexpected pods: %v", unknownMembers)
		for _, m := range unknownMembers {
			if err := cluster.removePod(m.Name); err != nil {
				return err
			}
		}
	}

	// INFO: 减去 unknown member，开始依次加入 "existing" member
	remaining := running.Diff(unknownMembers)
	if remaining.Size() == cluster.members.Size() {
		return cluster.resize()
	}
	if remaining.Size() < cluster.members.Size()/2+1 {
		return ErrLostQuorum
	}

	klog.Infof("removing one dead member")
	// remove dead members that doesn't have any running pods before doing resizing.
	return cluster.removeDeadMember(cluster.members.Diff(remaining).PickOne())
}

// INFO: add new member/remove member
func (cluster *Cluster) resize() error {
	if cluster.members.Size() == cluster.etcdCluster.Spec.Size {
		return nil
	}

	if cluster.members.Size() < cluster.etcdCluster.Spec.Size {
		return cluster.addOneMember()
	}

	return cluster.removeOneMember()
}

func (cluster *Cluster) isSecurePeer() bool {
	return cluster.etcdCluster.Spec.TLS.IsSecurePeer()
}

func (cluster *Cluster) isSecureClient() bool {
	return cluster.etcdCluster.Spec.TLS.IsSecureClient()
}

func (cluster *Cluster) newMember() *Member {
	member := &Member{
		Name:         UniqueMemberName(cluster.etcdCluster.Name),
		Namespace:    cluster.etcdCluster.Namespace,
		SecurePeer:   cluster.isSecurePeer(),
		SecureClient: cluster.isSecureClient(),
	}

	return member
}

// INFO: 这里先后顺序很重要，先往etcd里写数据，再去起一个etcd实例
//
//	(1)使用 etcdctl cli 来 add member，这样可以先更新下 etcd 数据；
//	(2)然后创建 etcd pod
func (cluster *Cluster) addOneMember() error {
	cfg := clientv3.Config{
		Endpoints:   cluster.members.ClientURLs(),
		DialTimeout: EtcdDefaultDialTimeout,
		TLS:         cluster.tlsConfig,
	}
	etcdClient, err := clientv3.New(cfg)
	if err != nil {
		return fmt.Errorf("[addOneMember]add one member failed: creating etcd client failed %v", err)
	}
	defer etcdClient.Close()

	// INFO: 添加一个新的 member，这里重点是 --initial-cluster-state=existing
	ctx, cancel := context.WithTimeout(context.TODO(), EtcdDefaultRequestTimeout)
	defer cancel()
	newMember := cluster.newMember()
	response, err := etcdClient.MemberAdd(ctx, []string{newMember.PeerURL()})
	if err != nil {
		return fmt.Errorf("[addOneMember]fail to add new member (%s): %v", newMember.Name, err)
	}
	newMember.ID = response.Member.ID
	cluster.members.Add(newMember)
	if err := cluster.createPod(cluster.members, newMember, "existing"); err != nil {
		return fmt.Errorf("[addOneMember]fail to create member's pod (%s): %v", newMember.Name, err)
	}

	_, err = cluster.kubeClient.CoreV1().Events(cluster.etcdCluster.Namespace).Create(context.TODO(),
		NewMemberAddEvent(newMember.Name, cluster.etcdCluster), metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("[addOneMember]failed to create new member add event: %v", err)
	}

	return nil
}

func (cluster *Cluster) removeDeadMember(member *Member) error {
	_, err := cluster.kubeClient.CoreV1().Events(cluster.etcdCluster.Namespace).Create(context.TODO(),
		ReplacingDeadMemberEvent(member.Name, cluster.etcdCluster), metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("[removeDeadMember]failed to create new member add event: %v", err)
	}

	return cluster.removeMember(member)
}

func (cluster *Cluster) removeOneMember() error {
	return cluster.removeMember(cluster.members.PickOne())
}

// INFO: 先使用 etcdClient 从 etcd 中删除 member 数据，再去删除 etcd pod
func (cluster *Cluster) removeMember(member *Member) error {
	cfg := clientv3.Config{
		Endpoints:   cluster.members.ClientURLs(),
		DialTimeout: EtcdDefaultDialTimeout,
		TLS:         cluster.tlsConfig,
	}
	etcdClient, err := clientv3.New(cfg)
	if err != nil {
		return fmt.Errorf("[addOneMember]add one member failed: creating etcd client failed %v", err)
	}
	defer etcdClient.Close()

	// INFO: 添加一个新的 member，这里重点是 --initial-cluster-state=existing
	ctx, cancel := context.WithTimeout(context.TODO(), EtcdDefaultRequestTimeout)
	defer cancel()
	_, err = etcdClient.MemberRemove(ctx, member.ID)
	if err != nil {
		switch err {
		case rpctypes.ErrMemberNotFound:
			klog.Infof(fmt.Sprintf("etcd member (%v) has been removed", member.Name))
		default:
			return err
		}
	}
	cluster.members.Remove(member.Name)
	if err = cluster.removePod(member.Name); err != nil {
		return err
	}

	_, err = cluster.kubeClient.CoreV1().Events(cluster.etcdCluster.Namespace).Create(context.TODO(),
		MemberRemoveEvent(member.Name, cluster.etcdCluster), metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("[addOneMember]failed to create new member add event: %v", err)
	}

	return nil
}

// INFO: 删除 etcd pod
func (cluster *Cluster) removePod(name string) error {
	// INFO: 这里删除 etcd pod 时，加个优雅删除还是比较好的
	gracePeriodSeconds := int64(5)
	err := cluster.kubeClient.CoreV1().Pods(cluster.etcdCluster.Namespace).Delete(context.TODO(), name, metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	return nil
}
