package demo_node_status

import (
	"context"
	goruntime "runtime"
	"time"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	v1helper "k8s.io/kubernetes/pkg/apis/core/v1/helper"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
	nodeutil "k8s.io/kubernetes/pkg/util/node"
	taintutil "k8s.io/kubernetes/pkg/util/taints"
	volutil "k8s.io/kubernetes/pkg/volume/util"
)

// registerWithAPIServer registers the node with the cluster master. It is safe
// to call multiple times, but not concurrently (kl.registrationCompleted is
// not locked).
func (kl *Kubelet) registerWithAPIServer() {
	if kl.registrationCompleted {
		return
	}
	step := 100 * time.Millisecond

	for {
		time.Sleep(step)
		step = step * 2
		if step >= 7*time.Second {
			step = 7 * time.Second
		}

		node, err := kl.initialNode(context.TODO())
		if err != nil {
			klog.Errorf("Unable to construct v1.Node object for kubelet: %v", err)
			continue
		}

		klog.Infof("Attempting to register node %s", node.Name)
		registered := kl.tryRegisterWithAPIServer(node)
		if registered {
			klog.Infof("Successfully registered node %s", node.Name)
			kl.registrationCompleted = true
			return
		}
	}
}

// 初始化一个 node 对象，包括 annotations/labels/taints, capacity/allocatable 等等，该对象将会写入 apiserver 中
func (kl *Kubelet) initialNode(ctx context.Context) (*v1.Node, error) {
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: string(kl.nodeName),
			Labels: map[string]string{
				v1.LabelHostname:      kl.hostname,
				v1.LabelOSStable:      goruntime.GOOS,
				v1.LabelArchStable:    goruntime.GOARCH,
				kubeletapis.LabelOS:   goruntime.GOOS,
				kubeletapis.LabelArch: goruntime.GOARCH,
			},
		},
		Spec: v1.NodeSpec{
			Unschedulable: !kl.registerSchedulable,
		},
	}

	nodeTaints := make([]v1.Taint, 0)
	unschedulableTaint := v1.Taint{
		Key:    v1.TaintNodeUnschedulable,
		Effect: v1.TaintEffectNoSchedule,
	}
	// Taint node with TaintNodeUnschedulable when initializing
	// node to avoid race condition; refer to #63897 for more detail.
	if node.Spec.Unschedulable &&
		!taintutil.TaintExists(nodeTaints, &unschedulableTaint) {
		nodeTaints = append(nodeTaints, unschedulableTaint)
	}
	if len(nodeTaints) > 0 {
		node.Spec.Taints = nodeTaints
	}

	// Initially, set NodeNetworkUnavailable to true.
	node.Status.Conditions = append(node.Status.Conditions, v1.NodeCondition{
		Type:               v1.NodeNetworkUnavailable,
		Status:             v1.ConditionTrue,
		Reason:             "NoRouteCreated",
		Message:            "Node created without a route",
		LastTransitionTime: metav1.NewTime(kl.clock.Now()),
	})

	kl.setNodeStatus(node)

	return node, nil
}

// setNodeStatus fills in the Status fields of the given Node, overwriting any fields that are currently set.
func (kl *Kubelet) setNodeStatus(node *v1.Node) {
	for i, f := range kl.setNodeStatusFuncs {
		klog.V(5).Infof("Setting node status at position %v", i)
		if err := f(node); err != nil {
			klog.Errorf("Failed to set some node status fields: %s", err)
		}
	}
}

// 1. 写一个 Node 对象；2. 如果 apiserver 中已经存在，去更新相关 annotations/labels，尤其是 extended-resource/huge-page allocatable/capacity
func (kl *Kubelet) tryRegisterWithAPIServer(node *v1.Node) bool {
	// 1. 写一个 Node 对象
	_, err := kl.kubeClient.CoreV1().Nodes().Create(context.TODO(), node, metav1.CreateOptions{})
	if err == nil {
		return true
	}

	if !apierrors.IsAlreadyExists(err) {
		klog.Errorf("Unable to register node %q with API server: %v", kl.nodeName, err)
		return false
	}

	existingNode, err := kl.kubeClient.CoreV1().Nodes().Get(context.TODO(), string(kl.nodeName), metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Unable to register node %q with API server: error getting existing node: %v", kl.nodeName, err)
		return false
	}
	if existingNode == nil {
		klog.Errorf("Unable to register node %q with API server: no node instance returned", kl.nodeName)
		return false
	}

	originalNode := existingNode.DeepCopy()
	klog.Infof("Node %s was previously registered", kl.nodeName)

	// 2. node 之前注册过，reconcile 下 volumes.kubernetes.io/controller-managed-attach-detach annotation
	requiresUpdate := kl.reconcileCMADAnnotationWithExistingNode(node, existingNode)
	requiresUpdate = kl.updateDefaultLabels(node, existingNode) || requiresUpdate
	requiresUpdate = kl.reconcileExtendedResource(node, existingNode) || requiresUpdate
	requiresUpdate = kl.reconcileHugePageResource(node, existingNode) || requiresUpdate
	if requiresUpdate {
		if _, _, err := nodeutil.PatchNodeStatus(kl.kubeClient.CoreV1(), types.NodeName(kl.nodeName), originalNode, existingNode); err != nil {
			klog.Errorf("Unable to reconcile node %q with API server: error updating node: %v", kl.nodeName, err)
			return false
		}
	}

	return true
}

// volumes.kubernetes.io/controller-managed-attach-detach: "true"
func (kl *Kubelet) reconcileCMADAnnotationWithExistingNode(node, existingNode *v1.Node) bool {
	var (
		existingCMAAnnotation    = existingNode.Annotations[volutil.ControllerManagedAttachAnnotation]
		newCMAAnnotation, newSet = node.Annotations[volutil.ControllerManagedAttachAnnotation]
	)

	if newCMAAnnotation == existingCMAAnnotation {
		return false
	}

	if !newSet {
		klog.Info("Controller attach-detach setting changed to false; updating existing Node")
		delete(existingNode.Annotations, volutil.ControllerManagedAttachAnnotation)
	} else {
		klog.Info("Controller attach-detach setting changed to true; updating existing Node")
		if existingNode.Annotations == nil {
			existingNode.Annotations = make(map[string]string)
		}
		existingNode.Annotations[volutil.ControllerManagedAttachAnnotation] = newCMAAnnotation
	}

	return true
}

// updateDefaultLabels will set the default labels on the node
func (kl *Kubelet) updateDefaultLabels(initialNode, existingNode *v1.Node) bool {
	defaultLabels := []string{
		v1.LabelHostname,
		v1.LabelZoneFailureDomainStable,
		v1.LabelZoneRegionStable,
		v1.LabelZoneFailureDomain,
		v1.LabelZoneRegion,
		v1.LabelInstanceTypeStable,
		v1.LabelInstanceType,
		v1.LabelOSStable,
		v1.LabelArchStable,
		v1.LabelWindowsBuild,
		kubeletapis.LabelOS,
		kubeletapis.LabelArch,
	}

	needsUpdate := false
	if existingNode.Labels == nil {
		existingNode.Labels = make(map[string]string)
	}
	//Set default labels but make sure to not set labels with empty values
	for _, label := range defaultLabels {
		if _, hasInitialValue := initialNode.Labels[label]; !hasInitialValue {
			continue
		}

		if existingNode.Labels[label] != initialNode.Labels[label] {
			existingNode.Labels[label] = initialNode.Labels[label]
			needsUpdate = true
		}

		if existingNode.Labels[label] == "" {
			delete(existingNode.Labels, label)
		}
	}

	return needsUpdate
}

// 更新下 initialNode 的 capacity/allocatable
func updateDefaultResources(initialNode, existingNode *v1.Node) bool {
	requiresUpdate := false
	if existingNode.Status.Capacity == nil {
		if initialNode.Status.Capacity != nil {
			// INFO: 这里是 deep copy 一个新对象而不是引用赋值
			existingNode.Status.Capacity = initialNode.Status.Capacity.DeepCopy()
			requiresUpdate = true
		} else {
			existingNode.Status.Capacity = make(map[v1.ResourceName]resource.Quantity)
		}
	}

	if existingNode.Status.Allocatable == nil {
		if initialNode.Status.Allocatable != nil {
			existingNode.Status.Allocatable = initialNode.Status.Allocatable.DeepCopy()
			requiresUpdate = true
		} else {
			existingNode.Status.Allocatable = make(map[v1.ResourceName]resource.Quantity)
		}
	}
	return requiresUpdate
}

// Zeros out extended resource capacity during reconciliation.
func (kl *Kubelet) reconcileExtendedResource(initialNode, node *v1.Node) bool {
	requiresUpdate := updateDefaultResources(initialNode, node)
	// Check with the device manager to see if node has been recreated, in which case extended resources should be zeroed until they are available
	if kl.containerManager.ShouldResetExtendedResourceCapacity() {
		for k := range node.Status.Capacity {
			if v1helper.IsExtendedResourceName(k) {
				klog.Infof("Zero out resource %s capacity in existing node.", k)
				node.Status.Capacity[k] = *resource.NewQuantity(int64(0), resource.DecimalSI)
				node.Status.Allocatable[k] = *resource.NewQuantity(int64(0), resource.DecimalSI)
				requiresUpdate = true
			}
		}
	}
	return requiresUpdate
}

// 更新 huge page capacity，同时考虑了如果有不支持的 huge page resource
func (kl *Kubelet) reconcileHugePageResource(initialNode, existingNode *v1.Node) bool {
	requiresUpdate := updateDefaultResources(initialNode, existingNode)
	supportedHugePageResources := sets.String{}

	for resourceName := range initialNode.Status.Capacity {
		if !v1helper.IsHugePageResourceName(resourceName) {
			continue
		}
		supportedHugePageResources.Insert(string(resourceName))

		initialCapacity := initialNode.Status.Capacity[resourceName]
		initialAllocatable := initialNode.Status.Allocatable[resourceName]

		capacity, resourceIsSupported := existingNode.Status.Capacity[resourceName]
		allocatable := existingNode.Status.Allocatable[resourceName]

		// existingNode hugepage capacity < initialNode hugepage capacity
		if !resourceIsSupported || capacity.Cmp(initialCapacity) != 0 {
			existingNode.Status.Capacity[resourceName] = initialCapacity.DeepCopy()
			requiresUpdate = true
		}

		// Add or update allocatable if it the size was previously unsupported or has changed
		if !resourceIsSupported || allocatable.Cmp(initialAllocatable) != 0 {
			existingNode.Status.Allocatable[resourceName] = initialAllocatable.DeepCopy()
			requiresUpdate = true
		}
	}

	for resourceName := range existingNode.Status.Capacity {
		if !v1helper.IsHugePageResourceName(resourceName) {
			continue
		}

		// If huge page size no longer is supported, we remove it from the node
		if !supportedHugePageResources.Has(string(resourceName)) {
			delete(existingNode.Status.Capacity, resourceName)
			delete(existingNode.Status.Allocatable, resourceName)
			klog.Infof("Removing now unsupported huge page resource named: %s", resourceName)
			requiresUpdate = true
		}
	}

	return requiresUpdate
}
