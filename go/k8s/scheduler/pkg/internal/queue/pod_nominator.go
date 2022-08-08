package queue

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sync"

	"k8s-lx1036/k8s/scheduler/pkg/framework"

	"k8s.io/apimachinery/pkg/types"
	listersv1 "k8s.io/client-go/listers/core/v1"
)

// PodNominator 主要存储抢占的 pod
type PodNominator struct {
	sync.RWMutex

	podLister listersv1.PodLister

	nominatedPods map[string][]*framework.PodInfo

	nominatedPodToNode map[types.UID]string
}

func NewPodNominator(podLister listersv1.PodLister) *PodNominator {
	return &PodNominator{
		podLister:          podLister,
		nominatedPods:      make(map[string][]*framework.PodInfo),
		nominatedPodToNode: make(map[types.UID]string),
	}
}

// AddNominatedPod INFO: @see Scheduler.handleSchedulingFailure()
func (nominator *PodNominator) AddNominatedPod(pod *framework.PodInfo, nominatingInfo *framework.NominatingInfo) {
	nominator.Lock()
	defer nominator.Unlock()
	nominator.add(pod, nominatingInfo)
}

func (nominator *PodNominator) add(podInfo *framework.PodInfo, nominatingInfo *framework.NominatingInfo) {
	nominator.delete(podInfo.Pod)

	var nodeName string
	if nominatingInfo.Mode() == framework.ModeOverride {
		nodeName = nominatingInfo.NominatedNodeName
	} else if nominatingInfo.Mode() == framework.ModeNoop {
		if podInfo.Pod.Status.NominatedNodeName == "" {
			return
		}
		nodeName = podInfo.Pod.Status.NominatedNodeName
	}

	if nominator.podLister != nil {
		// If the pod was removed or if it was already scheduled, don't nominate it.
		updatedPod, err := nominator.podLister.Pods(podInfo.Pod.Namespace).Get(podInfo.Pod.Name)
		if err != nil {
			klog.V(4).InfoS("Pod doesn't exist in podLister, aborted adding it to the nominator", "pod", klog.KObj(podInfo.Pod))
			return
		}
		if updatedPod.Spec.NodeName != "" {
			klog.V(4).InfoS("Pod is already scheduled to a node, aborted adding it to the nominator", "pod",
				klog.KObj(podInfo.Pod), "node", updatedPod.Spec.NodeName)
			return
		}
	}

	nominator.nominatedPodToNode[podInfo.Pod.UID] = nodeName
	for _, npi := range nominator.nominatedPods[nodeName] {
		if npi.Pod.UID == podInfo.Pod.UID {
			klog.V(4).InfoS("Pod already exists in the nominator", "pod", klog.KObj(npi.Pod))
			return
		}
	}
	nominator.nominatedPods[nodeName] = append(nominator.nominatedPods[nodeName], podInfo)
}
func (nominator *PodNominator) delete(pod *corev1.Pod) {
	nnn, ok := nominator.nominatedPodToNode[pod.UID]
	if !ok {
		return
	}
	for i, np := range nominator.nominatedPods[nnn] {
		if np.Pod.UID == pod.UID {
			nominator.nominatedPods[nnn] = append(nominator.nominatedPods[nnn][:i], nominator.nominatedPods[nnn][i+1:]...)
			if len(nominator.nominatedPods[nnn]) == 0 {
				delete(nominator.nominatedPods, nnn)
			}
			break
		}
	}
	delete(nominator.nominatedPodToNode, pod.UID)
}

func (nominator *PodNominator) DeleteNominatedPodIfExists(pod *corev1.Pod) {
	panic("implement me")
}

func (nominator *PodNominator) UpdateNominatedPod(oldPod, newPod *corev1.Pod) {
	panic("implement me")
}

func (nominator *PodNominator) NominatedPodsForNode(nodeName string) []*framework.PodInfo {
	nominator.RLock() // 只读锁，性能高
	defer nominator.RUnlock()

	pods := make([]*framework.PodInfo, len(nominator.nominatedPods[nodeName]))
	for i := 0; i < len(pods); i++ {
		pods[i] = nominator.nominatedPods[nodeName][i].DeepCopy()
	}

	return pods
}
