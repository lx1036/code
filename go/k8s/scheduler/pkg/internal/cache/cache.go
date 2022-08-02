package cache

import (
	"sync"
	"time"

	"k8s-lx1036/k8s/scheduler/pkg/framework"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

type podState struct {
	pod *v1.Pod
	// Used by assumedPod to determinate expiration.
	deadline *time.Time
	// Used to block cache from expiring assumedPod if binding still runs
	bindingFinished bool
}

// nodeInfoListItem holds a NodeInfo pointer and acts as an item in a doubly
// linked list. When a NodeInfo is updated, it goes to the head of the list.
// The items closer to the head are the most recently updated items.
type nodeInfoListItem struct {
	info *framework.NodeInfo
	next *nodeInfoListItem
	prev *nodeInfoListItem
}

type imageState struct {
	// Size of the image
	size int64
	// A set of node names for nodes having this image present
	nodes sets.String
}

// Dump is a dump of the cache state.
type Dump struct {
	AssumedPods map[string]bool
	Nodes       map[string]*framework.NodeInfo
}

// INFO: Cache 设计原理 https://github.com/jindezgm/k8s-src-analysis/blob/master/kube-scheduler/Cache.md
// INFO: 缓存了pod和node信息，同时缓存了调度结果，见属性 podStates

type Cache struct {
	stop   <-chan struct{}
	ttl    time.Duration
	period time.Duration

	// schedulerCache会被多个goroutine读写，所以需要读写锁
	mu sync.RWMutex
	// a set of assumed pod keys.
	// The key could further be used to get an entry in podStates.
	assumedPods map[string]bool
	// a map from pod key to podState.
	podStates map[string]*podState
	nodes     map[string]*nodeInfoListItem
	// headNode points to the most recently updated NodeInfo in "nodes". It is the
	// head of the linked list.
	headNode *nodeInfoListItem
	nodeTree *nodeTree
	// A map from image name to its imageState.
	imageStates map[string]*imageState
}

// New returns a Cache implementation.
// It automatically starts a go routine that manages expiration of assumed pods.
// "ttl" is how long the assumed pod will get expired.
// "stop" is the channel that would close the background goroutine.
var (
	cleanAssumedPeriod = 1 * time.Second
)

func newSchedulerCache(ttl, period time.Duration, stop <-chan struct{}) *Cache {
	return &Cache{
		ttl:    ttl,
		period: period,
		stop:   stop,

		nodes:       make(map[string]*nodeInfoListItem),
		nodeTree:    newNodeTree(nil),
		assumedPods: make(map[string]bool),
		podStates:   make(map[string]*podState),
		imageStates: make(map[string]*imageState),
	}
}

func New(ttl time.Duration, stop <-chan struct{}) *Cache {
	cache := newSchedulerCache(ttl, cleanAssumedPeriod, stop)
	cache.run()

	return cache
}

func (cache *Cache) run() {
	go wait.Until(cache.cleanupExpiredAssumedPods, cache.period, cache.stop)
}

func (cache *Cache) PodCount() (int, error) {
	panic("implement me")
}

func (cache *Cache) AssumePod(pod *v1.Pod) error {
	key, err := framework.GetPodKey(pod)
	if err != nil {
		return err
	}

}

func (cache *Cache) FinishBinding(pod *v1.Pod) error {
	panic("implement me")
}

func (cache *Cache) ForgetPod(pod *v1.Pod) error {
	panic("implement me")
}

func (cache *Cache) AddPod(pod *v1.Pod) error {

}

func (cache *Cache) UpdatePod(oldPod, newPod *v1.Pod) error {
	panic("implement me")
}

func (cache *Cache) RemovePod(pod *v1.Pod) error {
	panic("implement me")
}

func (cache *Cache) GetPod(pod *v1.Pod) (*v1.Pod, error) {
	panic("implement me")
}

func (cache *Cache) IsAssumedPod(pod *v1.Pod) (bool, error) {
	key, err := framework.GetPodKey(pod)
	if err != nil {
		return false, err
	}

	cache.mu.RLock()
	defer cache.mu.RUnlock()

	return cache.assumedPods.Has(key), nil
}

func (cache *Cache) AddNode(node *v1.Node) *framework.NodeInfo {
	panic("implement me")
}

func (cache *Cache) UpdateNode(oldNode, newNode *v1.Node) error {
	panic("implement me")
}

func (cache *Cache) RemoveNode(node *v1.Node) error {
	panic("implement me")
}

func (cache *Cache) UpdateSnapshot(nodeSnapshot *Snapshot) error {
	panic("implement me")
}

func (cache *Cache) Dump() *Dump {
	panic("implement me")
}

func (cache *Cache) cleanupExpiredAssumedPods() {
	cache.cleanupAssumedPods(time.Now())
}

// cleanupAssumedPods exists for making test deterministic by taking time as input argument.
// It also reports metrics on the cache size for nodes, pods, and assumed pods.
func (cache *Cache) cleanupAssumedPods(now time.Time) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	//defer cache.updateMetrics()

	// The size of assumedPods should be small
	for key := range cache.assumedPods {
		ps, ok := cache.podStates[key]
		if !ok {
			klog.Fatal("Key found in assumed set but not in podStates. Potentially a logical error.")
		}
		if !ps.bindingFinished {
			klog.V(5).Infof("Couldn't expire cache for pod %v/%v. Binding is still in progress.",
				ps.pod.Namespace, ps.pod.Name)
			continue
		}
		if now.After(*ps.deadline) {
			klog.Warningf("Pod %s/%s expired", ps.pod.Namespace, ps.pod.Name)
			if err := cache.expirePod(key, ps); err != nil {
				klog.Errorf("ExpirePod failed for %s: %v", key, err)
			}
		}
	}
}

func (cache *Cache) expirePod(key string, ps *podState) error {
	if err := cache.removePod(ps.pod); err != nil {
		return err
	}
	delete(cache.assumedPods, key)
	delete(cache.podStates, key)
	return nil
}

// Assumes that lock is already acquired.
// Removes a pod from the cached node info. If the node information was already
// removed and there are no more pods left in the node, cleans up the node from
// the cache.
func (cache *Cache) removePod(pod *v1.Pod) error {
	n, ok := cache.nodes[pod.Spec.NodeName]
	if !ok {
		klog.Errorf("node %v not found when trying to remove pod %v", pod.Spec.NodeName, pod.Name)
		return nil
	}
	if err := n.info.RemovePod(pod); err != nil {
		return err
	}
	if len(n.info.Pods) == 0 && n.info.Node() == nil {
		cache.removeNodeInfoFromList(pod.Spec.NodeName)
	} else {
		cache.moveNodeInfoToHead(pod.Spec.NodeName)
	}

	return nil
}

// removeNodeInfoFromList removes a NodeInfo from the "cache.nodes" doubly
// linked list.
// We assume cache lock is already acquired.
func (cache *Cache) removeNodeInfoFromList(name string) {
	ni, ok := cache.nodes[name]
	if !ok {
		klog.Errorf("No NodeInfo with name %v found in the cache", name)
		return
	}

	if ni.prev != nil {
		ni.prev.next = ni.next
	}
	if ni.next != nil {
		ni.next.prev = ni.prev
	}
	// if the removed item was at the head, we must update the head.
	if ni == cache.headNode {
		cache.headNode = ni.next
	}

	delete(cache.nodes, name)
}

// moveNodeInfoToHead moves a NodeInfo to the head of "cache.nodes" doubly
// linked list. The head is the most recently updated NodeInfo.
// We assume cache lock is already acquired.
func (cache *Cache) moveNodeInfoToHead(name string) {
	ni, ok := cache.nodes[name]
	if !ok {
		klog.Errorf("No NodeInfo with name %v found in the cache", name)
		return
	}
	// if the node info list item is already at the head, we are done.
	if ni == cache.headNode {
		return
	}

	if ni.prev != nil {
		ni.prev.next = ni.next
	}
	if ni.next != nil {
		ni.next.prev = ni.prev
	}
	if cache.headNode != nil {
		cache.headNode.prev = ni
	}
	ni.next = cache.headNode
	ni.prev = nil
	cache.headNode = ni
}
