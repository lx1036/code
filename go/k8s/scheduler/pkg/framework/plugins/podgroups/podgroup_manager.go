package podgroups

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
	"sync"
	"time"

	podgroupv1 "k8s-lx1036/k8s/scheduler/pkg/apis/podgroup/v1"
	"k8s-lx1036/k8s/scheduler/pkg/client/clientset/versioned"
	podgroupInformer "k8s-lx1036/k8s/scheduler/pkg/client/informers/externalversions/podgroup/v1"
	podgroupLister "k8s-lx1036/k8s/scheduler/pkg/client/listers/podgroup/v1"
	"k8s-lx1036/k8s/scheduler/pkg/framework"

	gochache "github.com/patrickmn/go-cache"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	informerv1 "k8s.io/client-go/informers/core/v1"
	listerv1 "k8s.io/client-go/listers/core/v1"
)

type Status string

const (
	// PodGroupNotSpecified denotes no PodGroup is specified in the Pod spec.
	PodGroupNotSpecified Status = "PodGroup not specified"
	// PodGroupNotFound denotes the specified PodGroup in the Pod spec is not found in API server.
	PodGroupNotFound Status = "PodGroup not found"
	Success          Status = "Success"
	Wait             Status = "Wait"
)

type PodGroupManager struct {
	sync.RWMutex

	// pgClient is a podGroup client
	pgClient versioned.Interface
	// snapshotSharedLister is pod shared list
	snapshotSharedLister framework.SharedLister
	// scheduleTimeout is the default timeout for podgroup scheduling.
	// If podgroup's scheduleTimeoutSeconds is set, it will be used.
	scheduleTimeout time.Duration
	// permittedPG stores the podgroup name which has passed the pre resource check.
	permittedPG *gochache.Cache
	pgLister    podgroupLister.PodGroupLister
	// podLister is pod lister
	podLister listerv1.PodLister
	// reserveResourcePercentage is the reserved resource for the max finished group, range (0,100]
	reserveResourcePercentage int32
}

func NewPodGroupManager(pgClient versioned.Interface, snapshotSharedLister framework.SharedLister, scheduleTimeout time.Duration,
	pgInformer podgroupInformer.PodGroupInformer, podInformer informerv1.PodInformer) *PodGroupManager {
	return &PodGroupManager{
		pgClient:             pgClient,
		snapshotSharedLister: snapshotSharedLister,
		scheduleTimeout:      scheduleTimeout,
		pgLister:             pgInformer.Lister(),
		podLister:            podInformer.Lister(),

		permittedPG: gochache.New(3*time.Second, 3*time.Second), // TODO: 这个设计有什么用?
	}
}

func (pgMgr *PodGroupManager) GetCreationTimestamp(pod *corev1.Pod, ts time.Time) time.Time {
	pgName := GetPodGroupLabel(pod)
	if len(pgName) == 0 {
		return ts
	}

	pg, err := pgMgr.pgLister.PodGroups(pod.Namespace).Get(pgName)
	if err != nil {
		return ts
	}
	return pg.CreationTimestamp.Time
}

// PreFilter
// 1. pod 不属于 pod-group，过滤掉
// 2. pods 不满足 pod-group MinMember 或者 MinResources，过滤掉
func (pgMgr *PodGroupManager) PreFilter(ctx context.Context, pod *corev1.Pod) error {
	pgFullName, pg := pgMgr.GetPodGroup(pod)
	if pg == nil {
		return nil
	}

	pods, err := pgMgr.podLister.Pods(pod.Namespace).List(
		labels.SelectorFromSet(labels.Set{podgroupv1.PodGroupLabel: GetPodGroupLabel(pod)}),
	)
	if err != nil {
		return fmt.Errorf("podLister list pods failed: %v", err)
	}
	if len(pods) < int(pg.Spec.MinMember) {
		return fmt.Errorf("pre-filter pod %v cannot find enough sibling pods, "+
			"current pods number: %v, minMember of group: %v", pod.Name, len(pods), pg.Spec.MinMember)
	}

	if pg.Spec.MinResources == nil {
		return nil
	}

	if _, ok := pgMgr.permittedPG.Get(pgFullName); ok {
		return nil
	}

	nodes, err := pgMgr.snapshotSharedLister.NodeInfos().List()
	if err != nil {
		return err
	}

	minResources := pg.Spec.MinResources.DeepCopy()
	podQuantity := resource.NewQuantity(int64(pg.Spec.MinMember), resource.DecimalSI)
	minResources[corev1.ResourcePods] = *podQuantity
	err = CheckClusterResource(nodes, minResources, pgFullName)
	if err != nil {
		klog.ErrorS(err, "Failed to PreFilter", "podGroup", klog.KObj(pg))
		return err
	}

	pgMgr.permittedPG.Add(pgFullName, pgFullName, pgMgr.scheduleTimeout)
	return nil
}

func (pgMgr *PodGroupManager) GetPodGroup(pod *corev1.Pod) (string, *podgroupv1.PodGroup) {
	pgName := GetPodGroupLabel(pod)
	if len(pgName) == 0 {
		return "", nil
	}

	pg, err := pgMgr.pgLister.PodGroups(pod.Namespace).Get(pgName)
	if err != nil {
		return fmt.Sprintf("%v/%v", pod.Namespace, pgName), nil
	}
	return fmt.Sprintf("%v/%v", pod.Namespace, pgName), pg
}

// CalculateAssignedPods returns the number of pods that has been assigned nodes: assumed or bound. 注意这里 pod 已经被 assumed or bound
func (pgMgr *PodGroupManager) CalculateAssignedPods(podGroupName, namespace string) int {
	nodeInfos, err := pgMgr.snapshotSharedLister.NodeInfos().List()
	if err != nil {
		klog.ErrorS(err, "Cannot get nodeInfos from frameworkHandle")
		return 0
	}
	var count int
	for _, nodeInfo := range nodeInfos {
		for _, podInfo := range nodeInfo.Pods {
			pod := podInfo.Pod
			if pod.Labels[podgroupv1.PodGroupLabel] == podGroupName && pod.Namespace == namespace && pod.Spec.NodeName != "" {
				count++
			}
		}
	}

	return count
}

func (pgMgr *PodGroupManager) Permit(ctx context.Context, pod *corev1.Pod) Status {

}

func GetPodGroupLabel(pod *corev1.Pod) string {
	return pod.Labels[podgroupv1.PodGroupLabel] // "pod-group.PodGroup=test1"
}

// CheckClusterResource 检查 cpu/memory/pods 资源是否满足 resourceRequest
func CheckClusterResource(nodeList []*framework.NodeInfo, resourceRequest corev1.ResourceList, desiredPodGroupName string) error {
	for _, info := range nodeList {
		if info == nil || info.Node() == nil {
			continue
		}

		nodeResource := getNodeLeftResource(info, desiredPodGroupName).ResourceList()
		for name, quant := range resourceRequest {
			quant.Sub(nodeResource[name])
			if quant.Sign() <= 0 {
				delete(resourceRequest, name)
				continue
			}
			resourceRequest[name] = quant
		}
		if len(resourceRequest) == 0 {
			return nil
		}
	}
	return fmt.Errorf("resource gap: %v", resourceRequest)
}

// 计算该 node 上所有 cpu/memory/pods/storage 资源，但是加回去 test1 pod-group 里的 pods
func getNodeLeftResource(info *framework.NodeInfo, desiredPodGroupName string) *framework.Resource {
	// 减去属于当前 test1 pod-group 的 pods
	nodeClone := info.Clone()
	for _, podInfo := range info.Pods {
		if podInfo == nil || podInfo.Pod == nil {
			continue
		}
		if GetPodGroupFullName(podInfo.Pod) != desiredPodGroupName {
			continue
		}
		nodeClone.RemovePod(podInfo.Pod)
	}

	leftResource := framework.Resource{
		ScalarResources: make(map[corev1.ResourceName]int64),
	}
	allocatable := nodeClone.Allocatable
	requested := nodeClone.Requested
	leftResource.AllowedPodNumber = allocatable.AllowedPodNumber - len(nodeClone.Pods) // 减去已经在该 node 上的不是 test1 pod-group 里的 pods 之外的所有 pods
	leftResource.MilliCPU = allocatable.MilliCPU - requested.MilliCPU
	leftResource.Memory = allocatable.Memory - requested.Memory
	leftResource.EphemeralStorage = allocatable.EphemeralStorage - requested.EphemeralStorage
	for k, allocatableEx := range allocatable.ScalarResources {
		requestEx, ok := requested.ScalarResources[k]
		if !ok {
			leftResource.ScalarResources[k] = allocatableEx
		} else {
			leftResource.ScalarResources[k] = allocatableEx - requestEx
		}
	}

	return &leftResource
}

// GetPodGroupFullName get namespaced group name from pod annotations
func GetPodGroupFullName(pod *corev1.Pod) string {
	pgName := GetPodGroupLabel(pod)
	if len(pgName) == 0 {
		return ""
	}
	return fmt.Sprintf("%v/%v", pod.Namespace, pgName)
}
