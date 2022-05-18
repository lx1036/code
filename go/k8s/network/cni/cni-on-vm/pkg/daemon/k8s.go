package daemon

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/util/bandwidth"
	"strconv"
	"time"

	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/utils/storage"
	"k8s-lx1036/k8s/network/cni/cni-on-vm/pkg/utils/types"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"k8s.io/client-go/tools/clientcmd"
)

const (
	podNetworkTypeENIMultiIP = "ENIMultiIP"

	dbPath = "/var/lib/cni/pod.db"
	dbName = "pods"

	defaultStickTimeForSts = 5 * time.Minute

	conditionTrue = "true"

	podWithEip      = "k8s.aliyun.com/pod-with-eip"
	podEipBandwidth = "k8s.aliyun.com/eip-bandwidth"

	AnnotationPrefix = "k8s.aliyun.com/"

	// PodIPReservation whether pod's IP will be reserved for a reuse
	PodIPReservation = AnnotationPrefix + "pod-ip-reservation"
)

func podInfoKey(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

type K8sService struct {
	client kubernetes.Interface

	storage storage.DiskStorage

	mode string
}

func newK8sServiceOrDie(kubeconfig string, daemonMode string) *K8sService {
	k8sRestConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		klog.Fatal(err)
	}
	client, err := kubernetes.NewForConfig(k8sRestConfig)
	if err != nil {
		klog.Fatal(err)
	}

	s, err := storage.NewDiskStorage(dbName, dbPath)
	if err != nil {
		klog.Fatal(err)
	}

	k8sService := &K8sService{
		client:  client,
		storage: s,

		mode: daemonMode,
	}

	return k8sService
}

// GetPod get pod from apiserver or local boltdb when deleted pod
func (k8sService *K8sService) GetPod(namespace, name string) (*types.PodInfo, error) {
	pod, err := k8sService.client.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) { // fetch from local boltdb when deleted pod
			key := podInfoKey(namespace, name)
			if obj, err := k8sService.storage.Get(key); err == nil {
				return obj.Pod, nil
			}
		}

		return nil, err
	}

	podInfo := convertPod(k8sService.mode, pod)
	item := &storage.Item{
		Pod: podInfo,
	}
	if err = k8sService.storage.Put(podInfoKey(podInfo.Namespace, podInfo.Name), item); err != nil {
		return nil, err
	}

	return podInfo, nil
}

func convertPod(daemonMode string, pod *corev1.Pod) *types.PodInfo {
	podInfo := &types.PodInfo{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		PodIPs:    types.IPSet{},
		PodUID:    string(pod.UID),
	}
	podInfo.PodNetworkType = podNetworkType(daemonMode, pod)

	for _, str := range pod.Status.PodIPs {
		podInfo.PodIPs.SetIP(str.IP)
	}
	podInfo.PodIPs.SetIP(pod.Status.PodIP)

	// ingress/egress
	podAnnotation := pod.GetAnnotations()
	if ingress, egress, err := bandwidth.ExtractPodBandwidthResources(podAnnotation); err == nil {
		podInfo.TcIngress = uint64(ingress.Value())
		podInfo.TcEgress = uint64(egress.Value())
	}

	if eipAnnotation, ok := podAnnotation[podWithEip]; ok && eipAnnotation == conditionTrue {
		podInfo.EipInfo.PodEip = true
		podInfo.EipInfo.PodEipBandWidth = 5
		//podInfo.EipInfo.PodEipChargeType = types.PayByTraffic
	}
	if eipAnnotation, ok := podAnnotation[podEipBandwidth]; ok {
		if eipBandwidth, err := strconv.Atoi(eipAnnotation); err == nil {
			podInfo.EipInfo.PodEipBandWidth = eipBandwidth
		}
	}

	// determine whether pod's IP will stick 5 minutes for a reuse, priorities as below,
	// 1. pod has a positive pod-ip-reservation annotation
	// 2. pod is owned by a known stateful workload
	if podIPReservation, _ := strconv.ParseBool(pod.Annotations[PodIPReservation]); podIPReservation {
		podInfo.IPStickTime = defaultStickTimeForSts
	}

	return podInfo
}

func podNetworkType(daemonMode string, pod *corev1.Pod) string {
	switch daemonMode {
	case daemonModeENIMultiIP:
		return podNetworkTypeENIMultiIP

	default:
		return podNetworkTypeENIMultiIP
	}
}
