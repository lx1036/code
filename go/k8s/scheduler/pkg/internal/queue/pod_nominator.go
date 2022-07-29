package queue

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sync"

	"k8s-lx1036/k8s/scheduler/pkg/framework"

	"k8s.io/apimachinery/pkg/types"
	listersv1 "k8s.io/client-go/listers/core/v1"
)

// 主要存储抢占的 pod
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

func (nominator *PodNominator) AddNominatedPod(pod *corev1.Pod, nodeName string) {
	nominator.Lock()
	nominator.add(pod, nodeName)
	nominator.Unlock()
}
func (nominator *PodNominator) add(pod *corev1.Pod, nodeName string) {
	// always delete the pod if it already exist, to ensure we never store more than
	// one instance of the pod.
	nominator.delete(pod)

	nnn := nodeName
	if len(nnn) == 0 {
		nnn = pod.Status.NominatedNodeName
		if len(nnn) == 0 {
			return
		}
	}
	nominator.nominatedPodToNode[pod.UID] = nnn
	for _, np := range nominator.nominatedPods[nnn] {
		if np.UID == pod.UID {
			klog.V(4).Infof("Pod %v/%v already exists in the nominated map!", pod.Namespace, pod.Name)
			return
		}
	}
	nominator.nominatedPods[nnn] = append(nominator.nominatedPods[nnn], pod)
}
func (nominator *PodNominator) delete(p *corev1.Pod) {
	nnn, ok := nominator.nominatedPodToNode[p.UID]
	if !ok {
		return
	}
	for i, np := range nominator.nominatedPods[nnn] {
		if np.UID == p.UID {
			nominator.nominatedPods[nnn] = append(nominator.nominatedPods[nnn][:i], nominator.nominatedPods[nnn][i+1:]...)
			if len(nominator.nominatedPods[nnn]) == 0 {
				delete(nominator.nominatedPods, nnn)
			}
			break
		}
	}
	delete(nominator.nominatedPodToNode, p.UID)
}

func (nominator *PodNominator) DeleteNominatedPodIfExists(pod *corev1.Pod) {
	panic("implement me")
}

func (nominator *PodNominator) UpdateNominatedPod(oldPod, newPod *corev1.Pod) {
	panic("implement me")
}

func (nominator *PodNominator) NominatedPodsForNode(nodeName string) []*corev1.Pod {
	panic("implement me")
}
