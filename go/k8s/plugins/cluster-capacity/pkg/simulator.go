package pkg

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	corev1Informers "k8s.io/client-go/informers/core/v1"
	policyinformers "k8s.io/client-go/informers/policy/v1beta1"
	storageinformers "k8s.io/client-go/informers/storage/v1"
	clientset "k8s.io/client-go/kubernetes"
	fakeClientset "k8s.io/client-go/kubernetes/fake"
	schedulerConfig "k8s.io/kubernetes/cmd/kube-scheduler/app/config"
)

type ClusterCapacity struct {
	// new cluster capacity
	externalkubeclient            clientset.Interface
	nodeInformer                  corev1Informers.NodeInformer
	podInformer                   corev1Informers.PodInformer
	pvInformer                    corev1Informers.PersistentVolumeInformer
	pvcInformer                   corev1Informers.PersistentVolumeClaimInformer
	replicationControllerInformer corev1Informers.ReplicationControllerInformer
	replicaSetInformer            appsinformers.ReplicaSetInformer
	statefulSetInformer           appsinformers.StatefulSetInformer
	serviceInformer               corev1Informers.ServiceInformer
	pdbInformer                   policyinformers.PodDisruptionBudgetInformer
	storageClassInformer          storageinformers.StorageClassInformer
	csiNodeInformer               storageinformers.CSINodeInformer

	// pod to schedule
	simulatedPod *corev1.Pod

	informerFactory informers.SharedInformerFactory
	// analysis limitation
	informerStopChannel chan struct{}

	report *ClusterCapacityReview
	status Status
}
type Status struct {
	Pods       []*corev1.Pod
	StopReason string
}

// sync k8s resources into fake kubeclient
func (clusterCapacity *ClusterCapacity) SyncWithClient(client clientset.Interface) error {
	// sync nodes into fake kubeclient
	nodeList, err := client.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("unable to list nodes: %v", err)
	}
	for _, item := range nodeList.Items {
		if _, err := clusterCapacity.externalkubeclient.CoreV1().Nodes().Create(&item); err != nil {
			return fmt.Errorf("unable to copy node: %v", err)
		}
	}

	// sync pods into fake kubeclient
	podList, err := client.CoreV1().Pods(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, item := range podList.Items {
		if _, err := clusterCapacity.externalkubeclient.CoreV1().Pods(item.Namespace).Create(&item); err != nil {
			return fmt.Errorf("unable to copy pod: %v", err)
		}
	}

	// sync services into fake kubeclient
	serviceList, err := client.CoreV1().Services(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("unable to list services: %v", err)
	}
	for _, item := range serviceList.Items {
		if _, err := clusterCapacity.externalkubeclient.CoreV1().Services(item.Namespace).Create(&item); err != nil {
			return fmt.Errorf("unable to copy service: %v", err)
		}
	}

	// sync pvc into fake kubeclient
	pvcItems, err := client.CoreV1().PersistentVolumeClaims(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("unable to list pvcs: %v", err)
	}
	for _, item := range pvcItems.Items {
		if _, err := clusterCapacity.externalkubeclient.CoreV1().PersistentVolumeClaims(item.Namespace).Create(&item); err != nil {
			return fmt.Errorf("unable to copy pvc: %v", err)
		}
	}

	// sync ReplicationController into fake kubeclient
	rcItems, err := client.CoreV1().ReplicationControllers(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("unable to list RCs: %v", err)
	}
	for _, item := range rcItems.Items {
		if _, err := clusterCapacity.externalkubeclient.CoreV1().ReplicationControllers(item.Namespace).Create(&item); err != nil {
			return fmt.Errorf("unable to copy RC: %v", err)
		}
	}

	// sync PodDisruptionBudget into fake kubeclient
	pdbItems, err := client.PolicyV1beta1().PodDisruptionBudgets(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("unable to list PDBs: %v", err)
	}
	for _, item := range pdbItems.Items {
		if _, err := clusterCapacity.externalkubeclient.PolicyV1beta1().PodDisruptionBudgets(item.Namespace).Create(&item); err != nil {
			return fmt.Errorf("unable to copy PDB: %v", err)
		}
	}

	// sync ReplicaSets into fake kubeclient
	replicaSetItems, err := client.AppsV1().ReplicaSets(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("unable to list replicas sets: %v", err)
	}
	for _, item := range replicaSetItems.Items {
		if _, err := clusterCapacity.externalkubeclient.AppsV1().ReplicaSets(item.Namespace).Create(&item); err != nil {
			return fmt.Errorf("unable to copy replica set: %v", err)
		}
	}

	// sync StatefulSets into fake kubeclient
	statefulSetItems, err := client.AppsV1().StatefulSets(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("unable to list stateful sets: %v", err)
	}
	for _, item := range statefulSetItems.Items {
		if _, err := clusterCapacity.externalkubeclient.AppsV1().StatefulSets(item.Namespace).Create(&item); err != nil {
			return fmt.Errorf("unable to copy stateful set: %v", err)
		}
	}

	// sync StorageClasses into fake kubeclient
	storageClassesItems, err := client.StorageV1().StorageClasses().List(metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("unable to list storage classes: %v", err)
	}
	for _, item := range storageClassesItems.Items {
		if _, err := clusterCapacity.externalkubeclient.StorageV1().StorageClasses().Create(&item); err != nil {
			return fmt.Errorf("unable to copy storage class: %v", err)
		}
	}

	return nil
}
func (clusterCapacity *ClusterCapacity) Run() error {
	clusterCapacity.informerFactory.Start(clusterCapacity.informerStopChannel)
	clusterCapacity.informerFactory.WaitForCacheSync(clusterCapacity.informerStopChannel)

}
func (clusterCapacity *ClusterCapacity) Report() *ClusterCapacityReview {
	var pods []*corev1.Pod
	if clusterCapacity.report == nil {
		pods = append(pods, clusterCapacity.simulatedPod)
		clusterCapacity.report = GetReport(pods, clusterCapacity.status)
	}

	return clusterCapacity.report
}

func New(kubeSchedulerConfig *schedulerConfig.CompletedConfig, simulatedPod *corev1.Pod) (*ClusterCapacity, error) {
	client := fakeClientset.NewSimpleClientset()
	sharedInformerFactory := informers.NewSharedInformerFactory(client, 0)

	clusterCapacity := &ClusterCapacity{
		externalkubeclient:            client,
		nodeInformer:                  sharedInformerFactory.Core().V1().Nodes(),
		podInformer:                   sharedInformerFactory.Core().V1().Pods(),
		pvInformer:                    sharedInformerFactory.Core().V1().PersistentVolumes(),
		pvcInformer:                   sharedInformerFactory.Core().V1().PersistentVolumeClaims(),
		replicationControllerInformer: sharedInformerFactory.Core().V1().ReplicationControllers(),
		replicaSetInformer:            sharedInformerFactory.Apps().V1().ReplicaSets(),
		statefulSetInformer:           sharedInformerFactory.Apps().V1().StatefulSets(),
		serviceInformer:               sharedInformerFactory.Core().V1().Services(),
		pdbInformer:                   sharedInformerFactory.Policy().V1beta1().PodDisruptionBudgets(),
		storageClassInformer:          sharedInformerFactory.Storage().V1().StorageClasses(),
		simulatedPod:                  simulatedPod,

		informerFactory:     sharedInformerFactory,
		informerStopChannel: make(chan struct{}),
	}

	return clusterCapacity, nil
}
