package main

import (
	"os"
	"time"

	"k8s-lx1036/k8s/storage/csi/external-provisioner/pkg/capacity"
	"k8s-lx1036/k8s/storage/csi/external-provisioner/pkg/capacity/topology"
	ctrl "k8s-lx1036/k8s/storage/csi/external-provisioner/pkg/controller"
	"k8s-lx1036/k8s/storage/csi/external-provisioner/pkg/owner"

	"google.golang.org/grpc"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	listersv1 "k8s.io/client-go/listers/core/v1"
	storagelistersv1 "k8s.io/client-go/listers/storage/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

func getCapacityController(
	grpcClient *grpc.ClientConn,
	rateLimiter workqueue.RateLimiter,
	provisionerName string,
	config *rest.Config,
	factory informers.SharedInformerFactory,
	clientset *kubernetes.Clientset,
	nodeDeployment *ctrl.NodeDeployment) (*capacity.Controller, informers.SharedInformerFactory) {
	podName := os.Getenv("POD_NAME")
	namespace := os.Getenv("POD_NAMESPACE")
	if podName == "" || namespace == "" {
		klog.Fatalf("need POD_NAMESPACE/POD_NAME env variables, have only POD_NAMESPACE=%q and POD_NAME=%q", namespace, podName)
	}
	controller, err := owner.Lookup(config, namespace, podName,
		schema.GroupVersionKind{
			Group:   "",
			Version: "v1",
			Kind:    "Pod",
		}, *capacityOwnerrefLevel)
	if err != nil {
		klog.Fatalf("look up owner(s) of pod: %v", err)
	}
	klog.Infof("using %s/%s %s as owner of CSIStorageCapacity objects", controller.APIVersion, controller.Kind, controller.Name)

	var topologyInformer topology.Informer
	if nodeDeployment == nil {
		topologyInformer = topology.NewNodeTopology(
			provisionerName,
			clientset,
			factory.Core().V1().Nodes(),
			factory.Storage().V1().CSINodes(),
			workqueue.NewNamedRateLimitingQueue(rateLimiter, "csitopology"),
		)
	} else {
		var segment topology.Segment
		if nodeDeployment.NodeInfo.AccessibleTopology != nil {
			for key, value := range nodeDeployment.NodeInfo.AccessibleTopology.Segments {
				segment = append(segment, topology.SegmentEntry{Key: key, Value: value})
			}
		}
		klog.Infof("producing CSIStorageCapacity objects with fixed topology segment %s", segment)
		topologyInformer = topology.NewFixedNodeTopology(&segment)
	}

	// We only need objects from our own namespace. The normal factory would give
	// us an informer for the entire cluster.
	factoryForNamespace := informers.NewSharedInformerFactoryWithOptions(clientset,
		ctrl.ResyncPeriodOfCsiNodeInformer,
		informers.WithNamespace(namespace),
	)

	capacityController := capacity.NewCentralCapacityController(
		csi.NewControllerClient(grpcClient),
		provisionerName,
		clientset,
		// TODO: metrics for the queue?!
		workqueue.NewNamedRateLimitingQueue(rateLimiter, "csistoragecapacity"),
		*controller,
		namespace,
		topologyInformer,
		factory.Storage().V1().StorageClasses(),
		factoryForNamespace.Storage().V1alpha1().CSIStorageCapacities(),
		*capacityPollInterval,
		*capacityImmediateBinding,
	)

	return capacityController, factoryForNamespace
}

func getNodeLister(
	nodeDeployment *ctrl.NodeDeployment,
	factory informers.SharedInformerFactory,
	clientset *kubernetes.Clientset,
	provisionerName string) (listersv1.NodeLister, storagelistersv1.CSINodeLister) {
	// topology
	var nodeLister listersv1.NodeLister
	var csiNodeLister storagelistersv1.CSINodeLister
	if nodeDeployment != nil {
		// Avoid watching in favor of fake, static objects. This is particularly relevant for
		// Node objects, which can generate significant traffic.
		csiNode := &storagev1.CSINode{
			ObjectMeta: metav1.ObjectMeta{
				Name: nodeDeployment.NodeName,
			},
			Spec: storagev1.CSINodeSpec{
				Drivers: []storagev1.CSINodeDriver{
					{
						Name:   provisionerName,
						NodeID: nodeDeployment.NodeInfo.NodeId,
					},
				},
			},
		}
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: nodeDeployment.NodeName,
			},
		}
		if nodeDeployment.NodeInfo.AccessibleTopology != nil {
			for key := range nodeDeployment.NodeInfo.AccessibleTopology.Segments {
				csiNode.Spec.Drivers[0].TopologyKeys = append(csiNode.Spec.Drivers[0].TopologyKeys, key)
			}
			node.Labels = nodeDeployment.NodeInfo.AccessibleTopology.Segments
		}
		klog.Infof("using local topology with Node = %+v and CSINode = %+v", node, csiNode)

		// We make those fake objects available to the topology code via informers which
		// never change.
		stoppedFactory := informers.NewSharedInformerFactory(clientset, 1000*time.Hour)
		csiNodes := stoppedFactory.Storage().V1().CSINodes()
		nodes := stoppedFactory.Core().V1().Nodes()
		csiNodes.Informer().GetStore().Add(csiNode)
		nodes.Informer().GetStore().Add(node)
		csiNodeLister = csiNodes.Lister()
		nodeLister = nodes.Lister()

	} else {
		csiNodeLister = factory.Storage().V1().CSINodes().Lister()
		nodeLister = factory.Core().V1().Nodes().Lister()
	}

	return nodeLister, csiNodeLister
}
