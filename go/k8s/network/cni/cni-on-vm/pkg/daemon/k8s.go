package daemon

import (
	"context"
	"fmt"

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
)

func podInfoKey(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

type K8sService struct {
	client kubernetes.Interface

	storage storage.DiskStorage
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
	}

	return k8sService
}

func (k8sService *K8sService) GetPod(namespace, name string) (*types.PodInfo, error) {

	pod, err := k8sService.client.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) { // fetch from local boltdb
			key := podInfoKey(namespace, name)
			obj, err := k8sService.storage.Get(key)
			if err == nil {
				item := obj.(*storage.Item)
				return item.Pod, nil
			}

			if err != storage.ErrNotFound {
				return nil, err
			}
		}

		return nil, err
	}

}
