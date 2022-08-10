package podgroups

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	podgroupv1 "k8s-lx1036/k8s/scheduler/pkg/apis/podgroup/v1"
	"k8s-lx1036/k8s/scheduler/pkg/client/clientset/versioned"
	podgroupInformer "k8s-lx1036/k8s/scheduler/pkg/client/informers/externalversions/podgroup/v1"
	podgroupLister "k8s-lx1036/k8s/scheduler/pkg/client/listers/podgroup/v1"
	"k8s-lx1036/k8s/scheduler/pkg/framework"

	gochache "github.com/patrickmn/go-cache"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	informerv1 "k8s.io/client-go/informers/core/v1"
	listerv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/klog/v2"
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

		permittedPG: gochache.New(3*time.Second, 3*time.Second), // TODO: 这个设计有什么用? 这个是带有过期时间的 Cache.
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

func (pgMgr *PodGroupManager) DeletePermittedPodGroup(pgFullName string) {
	pgMgr.permittedPG.Delete(pgFullName)
}

func (pgMgr *PodGroupManager) GetPodGroup(pod *corev1.Pod) (string, *podgroupv1.PodGroup) {
	pgName := GetPodGroupName(pod)
	pg, err := pgMgr.pgLister.PodGroups(pod.Namespace).Get(pgName)
	if err != nil {
		return "", nil
	}
	return fmt.Sprintf("%v/%v", pod.Namespace, pgName), pg
}

// CalculateAssignedPods returns the number of pods that has been assigned nodes: assumed or bound. 注意这里 pod 已经被 assumed or bound
// INFO: pod group 内已经 assume(即 pod.spec.nodeName=node1) 的 pod 数量总和
func (pgMgr *PodGroupManager) CalculateAssignedPods(podGroupName, namespace string) int {
	nodeInfos, err := pgMgr.snapshotSharedLister.NodeInfos().List()
	if err != nil {
		klog.ErrorS(err, "Cannot get nodeInfos from frameworkHandle")
		return 0
	}
	var count int
	for _, nodeInfo := range nodeInfos {
		for _, podInfo := range nodeInfo.Pods { // INFO: 这里基本上就是在遍历整个集群的 pods
			pod := podInfo.Pod
			if pod.Labels[podgroupv1.PodGroupLabel] == podGroupName && pod.Namespace == namespace && pod.Spec.NodeName != "" {
				count++
			}
		}
	}

	return count
}

// Permit 只要 group pods 数量大于 minNumber，就可以 permit
func (pgMgr *PodGroupManager) Permit(ctx context.Context, pod *corev1.Pod) Status {
	pgFullName, pg := pgMgr.GetPodGroup(pod)
	if pgFullName == "" {
		return PodGroupNotSpecified
	}
	if pg == nil {
		// A Pod with a podGroup name but without a PodGroup found is denied.
		return PodGroupNotFound
	}

	assigned := pgMgr.CalculateAssignedPods(pg.Name, pg.Namespace) // INFO: pod group 内已经 assume(即 pod.spec.nodeName=node1) 的 pod 数量总和
	if int32(assigned)+1 >= pg.Spec.MinMember {
		return Success
	}
	return Wait
}

func (pgMgr *PodGroupManager) ActivateSiblings(pod *corev1.Pod, state *framework.CycleState) {
	pgName := GetPodGroupName(pod)
	pods, err := pgMgr.podLister.Pods(pod.Namespace).List(
		labels.SelectorFromSet(labels.Set{podgroupv1.PodGroupLabel: GetPodGroupLabel(pod)}),
	)
	if err != nil {
		klog.ErrorS(err, "Failed to obtain pods belong to a PodGroup", "podGroup", pgName)
		return
	}
	for i := range pods {
		if pods[i].UID == pod.UID {
			pods = append(pods[:i], pods[i+1:]...)
			break
		}
	}

	// INFO: 这里用了 framework 框架的一个机制，framework 会在 schedule cycle 和 bind cycle 完成后，把 cycleState[framework.PodsToActivateKey]
	//  里的 pods 重新放回 activeQ 让该 pod 再次被调度，这针对对于本次调度失败的 pod。比如这里的那些兄弟 pods 应该被重新调度。代码见：
	//  https://github.com/kubernetes/kubernetes/blob/v1.24.3/pkg/scheduler/schedule_one.go#L88-L92
	//  https://github.com/kubernetes/kubernetes/blob/v1.24.3/pkg/scheduler/schedule_one.go#L185-L190
	//  https://github.com/kubernetes/kubernetes/blob/v1.24.3/pkg/scheduler/schedule_one.go#L271-L276
	if len(pods) != 0 {
		if c, err := state.Read(framework.PodsToActivateKey); err == nil {
			if s, ok := c.(*framework.PodsToActivate); ok {
				s.Lock()
				for _, pod := range pods {
					namespacedName := GetNamespacedName(pod)
					s.Map[namespacedName] = pod
				}
				s.Unlock()
			}
		}
	}
}

// PostBind updates a PodGroup's status.
// TODO: move this logic to PodGroup's controller.
func (pgMgr *PodGroupManager) PostBind(ctx context.Context, pod *corev1.Pod, nodeName string) {
	pgFullName, pg := pgMgr.GetPodGroup(pod)
	if pgFullName == "" || pg == nil {
		return
	}

	pgCopy := pg.DeepCopy()
	pgCopy.Status.Scheduled++
	if pgCopy.Status.Scheduled >= pgCopy.Spec.MinMember {
		pgCopy.Status.Phase = podgroupv1.PodGroupScheduled
	} else {
		pgCopy.Status.Phase = podgroupv1.PodGroupScheduling
		if pgCopy.Status.ScheduleStartTime.IsZero() {
			pgCopy.Status.ScheduleStartTime = metav1.Time{Time: time.Now()}
		}
	}
	if pgCopy.Status.Phase != pg.Status.Phase {
		pg, err := pgMgr.pgLister.PodGroups(pgCopy.Namespace).Get(pgCopy.Name) // 获取最新的 pg
		if err != nil {
			klog.ErrorS(err, "Failed to get PodGroup", "podGroup", klog.KObj(pgCopy))
			return
		}
		patch, err := CreateMergePatch(pg, pgCopy)
		if err != nil {
			klog.ErrorS(err, "Failed to create merge patch", "podGroup", klog.KObj(pg), "podGroup", klog.KObj(pgCopy))
			return
		}
		if len(patch) == 0 {
			return
		}
		_, err = pgMgr.pgClient.PodGroupV1().PodGroups(pg.Namespace).Patch(context.TODO(), pg.Name,
			types.MergePatchType, patch, metav1.PatchOptions{})
		if err != nil {
			klog.ErrorS(err, "Failed to patch", "podGroup", klog.KObj(pg))
			return
		}
	}
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
		if GetPodGroupName(podInfo.Pod) != desiredPodGroupName {
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

func GetPodGroupLabel(pod *corev1.Pod) string {
	return pod.Labels[podgroupv1.PodGroupLabel] // "pod-group.PodGroup=test1"
}

// GetPodGroupName get namespaced group name from pod annotations
func GetPodGroupName(pod *corev1.Pod) string {
	pgName := GetPodGroupLabel(pod)
	if len(pgName) == 0 {
		pgName = fmt.Sprintf("%v/%v", pod.Namespace, pod.Name)
	}
	return pgName
}

const DefaultWaitTime = 60 * time.Second

func GetWaitTimeDuration(pg *podgroupv1.PodGroup, scheduleTimeout time.Duration) time.Duration {
	if pg != nil && pg.Spec.ScheduleTimeoutSeconds != nil {
		return time.Duration(*pg.Spec.ScheduleTimeoutSeconds) * time.Second
	}
	if scheduleTimeout != 0 {
		return scheduleTimeout
	}
	return DefaultWaitTime
}

func CreateMergePatch(original, new interface{}) ([]byte, error) {
	pvByte, err := json.Marshal(original)
	if err != nil {
		return nil, err
	}
	cloneByte, err := json.Marshal(new)
	if err != nil {
		return nil, err
	}
	patch, err := strategicpatch.CreateTwoWayMergePatch(pvByte, cloneByte, original)
	if err != nil {
		return nil, err
	}
	return patch, nil
}
