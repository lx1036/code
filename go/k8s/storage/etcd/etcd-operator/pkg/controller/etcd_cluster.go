package controller

import (
	"context"
	"fmt"
	"time"

	v1 "k8s-lx1036/k8s/storage/etcd/etcd-operator/pkg/apis/etcd.k9s.io/v1"

	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

const (
	ReconcileInterval = 10 * time.Second
)

type ClusterConfig struct {
	kubeClient kubernetes.Interface
}

type Cluster struct {
	etcdCluster *v1.EtcdCluster

	kubeClient kubernetes.Interface

	// pod name 就是 member name, etcd cluster 中的所有 members
	members MemberSet

	stopCh chan struct{}
}

func (cluster *Cluster) setup() error {
	var shouldCreateCluster bool
	switch cluster.etcdCluster.Status.Phase {
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

	return cluster.prepareSeedMember()
}

func (cluster *Cluster) prepareSeedMember() error {

	return cluster.startSeedMember()
}

func (cluster *Cluster) startSeedMember() error {

	member := &Member{
		Name:      cluster.etcdCluster.Name,
		Namespace: cluster.etcdCluster.Namespace,
		//SecurePeer:   cluster.isSecurePeer(),
		//SecureClient: cluster.isSecureClient(),
	}
	memberSet := NewMemberSet(member)
	if err := cluster.createPod(memberSet, member, "new"); err != nil {
		return fmt.Errorf("failed to create seed member (%s): %v", member.Name, err)
	}

	cluster.members = memberSet

	return nil
}

func (cluster *Cluster) createPod(members MemberSet, member *Member, state string) error {
	pod := NewEtcdPod(member, members.PeerURLPairs(), cluster.etcdCluster.Name, state, uuid.New().String(),
		cluster.etcdCluster.Spec, cluster.etcdCluster.AsOwner())

	_, err := cluster.kubeClient.CoreV1().Pods(cluster.etcdCluster.Namespace).Create(context.TODO(), pod,
		metav1.CreateOptions{})
	return err

}

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
			if len(pending) > 0 {
				// Pod startup might take long, e.g. pulling image. It would deterministically become running or succeeded/failed later.
				klog.Infof("[run]skip reconciliation: running (%v), pending (%v)", GetPodNames(running), GetPodNames(pending))
				continue
			}
			if len(running) == 0 {
				// TODO: how to handle this case?
				klog.Warningf("[run]all etcd pods are dead.")
				break
			}

			rerr = cluster.reconcile(running)
			if rerr != nil {
				klog.Errorf(fmt.Sprintf("[run]failed to reconcile err %v", err))
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

// reconcile reconciles cluster current state to desired state specified by spec.
// - it tries to reconcile the cluster to desired size.
// - if the cluster needs for upgrade, it tries to upgrade old member one by one.
func (cluster *Cluster) reconcile(pods []*corev1.Pod) error {

	return nil
}

func NewCluster(clusterConfig *ClusterConfig, etcdCluster *v1.EtcdCluster) *Cluster {
	cluster := &Cluster{
		etcdCluster: etcdCluster,
		kubeClient:  clusterConfig.kubeClient,
	}

	go func() {
		if err := cluster.setup(); err != nil {
			klog.Errorf(fmt.Sprintf("[NewCluster]cluster failed to setup err %v", err))
		}

		cluster.run()
	}()

	return cluster
}
