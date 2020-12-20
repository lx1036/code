package controller

import (
	"fmt"
	"sync"
	"time"

	"k8s-lx1036/k8s/concepts/components/controller-manager/namespace-controller/pkg/debug"

	log "github.com/sirupsen/logrus"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/metadata"
)

// operation is used for caching if an operation is supported on a dynamic client.
type operation string

const (
	operationDeleteCollection operation = "deletecollection"
	operationList             operation = "list"
	// assume a default estimate for finalizers to complete when found on items pending deletion.
	finalizerEstimateSeconds int64 = int64(15)
)

// operationKey is an entry in a cache.
type operationKey struct {
	operation operation
	gvr       schema.GroupVersionResource
}

type operationNotSupportedCache struct {
	lock sync.RWMutex
	m    map[operationKey]bool
}

func (o *operationNotSupportedCache) setNotSupported(key operationKey) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.m[key] = true
}

// isSupported returns true if the operation is supported
func (o *operationNotSupportedCache) isSupported(key operationKey) bool {
	o.lock.RLock()
	defer o.lock.RUnlock()
	return !o.m[key]
}

type NamespacedResourcesDeleterInterface interface {
	Delete(namespaceName string) error
}

type namespacedResourcesDeleter struct {
	discoverResourcesFn func() ([]*metav1.APIResourceList, error)

	opCache *operationNotSupportedCache

	clientset *kubernetes.Clientset

	// The finalizer token that should be removed from the namespace
	// when all resources in that namespace have been deleted.
	finalizerToken v1.FinalizerName

	// Dynamic client to list and delete all namespaced resources.
	metadataClient metadata.Interface
}

func NewNamespacedResourcesDeleter(
	discoverResourcesFn func() ([]*metav1.APIResourceList, error),
	clientset *kubernetes.Clientset,
	finalizerToken v1.FinalizerName,
	metadataClient metadata.Interface,
) NamespacedResourcesDeleterInterface {

	deleter := &namespacedResourcesDeleter{
		discoverResourcesFn: discoverResourcesFn,
		clientset:           clientset,
		finalizerToken:      finalizerToken,
		metadataClient:      metadataClient,
	}

	deleter.initOpCache()

	return deleter
}

func (deleter *namespacedResourcesDeleter) initOpCache() {
	resources, err := deleter.discoverResourcesFn()
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to get all supported resources from server: %v", err))
	}
	if len(resources) == 0 {
		log.Fatalf("Unable to get any supported resources from server: %v", err)
	}

	debug.LogAPIResourceList(resources)

	// 过滤出verb只有list和deletecollection的resource
	var deletableGroupVersionResources []schema.GroupVersionResource
	for _, resource := range resources {
		gv, err := schema.ParseGroupVersion(resource.GroupVersion)
		if err != nil {
			log.Errorf("Failed to parse GroupVersion %q, skipping: %v", resource.GroupVersion, err)
			continue
		}

		for _, apiResource := range resource.APIResources {
			gvr := schema.GroupVersionResource{Group: gv.Group, Version: gv.Version, Resource: apiResource.Name}
			verbs := sets.NewString([]string(apiResource.Verbs)...)

			if !verbs.Has("delete") {
				log.Infof("Skipping resource %v because it cannot be deleted.", gvr)
			}

			// verb没有list和deletecollection的resource，cache到opCache对象
			for _, op := range []operation{operationList, operationDeleteCollection} {
				if !verbs.Has(string(op)) {
					deleter.opCache.setNotSupported(operationKey{operation: op, gvr: gvr})
				}
			}

			deletableGroupVersionResources = append(deletableGroupVersionResources, gvr)
		}
	}
}

// updateNamespaceFunc is a function that makes an update to a namespace
type updateNamespaceFunc func(namespace *v1.Namespace) (*v1.Namespace, error)

// retryOnConflictError retries the specified fn if there was a conflict error
// it will return an error if the UID for an object changes across retry operations.
func (deleter *namespacedResourcesDeleter) retryOnConflictError(namespace *v1.Namespace, fn updateNamespaceFunc) (result *v1.Namespace, err error) {
	latestNamespace := namespace
	for {
		result, err = fn(latestNamespace)
		if err == nil {
			return result, nil
		}
		if !errors.IsConflict(err) {
			return nil, err
		}
		prevNamespace := latestNamespace
		latestNamespace, err = deleter.clientset.CoreV1().Namespaces().Get(latestNamespace.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		if prevNamespace.UID != latestNamespace.UID {
			return nil, fmt.Errorf("namespace uid has changed across retries")
		}
	}
}

// updateNamespaceStatusFunc will verify that the status of the namespace is correct
// 如果Status.Phase不是terminating，就更新下Phase字段值
func (deleter *namespacedResourcesDeleter) updateNamespaceStatusFunc(namespace *v1.Namespace) (*v1.Namespace, error) {
	if namespace.DeletionTimestamp.IsZero() || namespace.Status.Phase == v1.NamespaceTerminating {
		return namespace, nil
	}
	newNamespace := namespace.DeepCopy()
	newNamespace.Status.Phase = v1.NamespaceTerminating

	return deleter.clientset.CoreV1().Namespaces().UpdateStatus(newNamespace)
}

// finalized returns true if the namespace.Spec.Finalizers is an empty list
func finalized(namespace *v1.Namespace) bool {
	return len(namespace.Spec.Finalizers) == 0
}

// finalizeNamespace removes the specified finalizerToken and finalizes the namespace
// 从spec.finalizers字段值中，从slice里只是删除特定的finalizerToken
func (deleter *namespacedResourcesDeleter) finalizeNamespace(namespace *v1.Namespace) (*v1.Namespace, error) {

	finalizerSet := sets.NewString()
	for i := range namespace.Spec.Finalizers {
		if namespace.Spec.Finalizers[i] != deleter.finalizerToken {
			finalizerSet.Insert(string(namespace.Spec.Finalizers[i]))
		}
	}

	namespaceFinalize := v1.Namespace{}
	namespaceFinalize.ObjectMeta = namespace.ObjectMeta
	namespaceFinalize.Spec = namespace.Spec
	namespaceFinalize.Spec.Finalizers = make([]v1.FinalizerName, 0, len(finalizerSet))
	for _, value := range finalizerSet.List() {
		namespaceFinalize.Spec.Finalizers = append(namespaceFinalize.Spec.Finalizers, v1.FinalizerName(value))
	}

	// 调用api更新spec.finalizers字段
	namespace, err := deleter.clientset.CoreV1().Namespaces().Finalize(&namespaceFinalize)
	if err != nil {
		// it was removed already, so life is good
		if errors.IsNotFound(err) {
			return namespace, nil
		}
	}
	return namespace, err
}

// 根据GroupVersionResource来批量删除所有资源
// deleteCollection is a helper function that will delete the collection of resources
// it returns true if the operation was supported on the server.
// it returns an error if the operation was supported on the server but was unable to complete.
func (deleter *namespacedResourcesDeleter) deleteCollection(gvr schema.GroupVersionResource, namespace string) (bool, error) {
	log.Infof("namespace controller - deleteCollection - namespace: %s, gvr: %v", namespace, gvr)

	key := operationKey{operation: operationDeleteCollection, gvr: gvr}
	if !deleter.opCache.isSupported(key) {
		log.Infof("namespace controller - deleteCollection ignored since not supported - namespace: %s, gvr: %v", namespace, gvr)
		return false, nil
	}

	// namespace controller does not want the garbage collector to insert the orphan finalizer since it calls
	// resource deletions generically.  it will ensure all resources in the namespace are purged prior to releasing
	// namespace itself.
	background := metav1.DeletePropagationBackground
	opts := metav1.DeleteOptions{PropagationPolicy: &background}
	err := deleter.metadataClient.Resource(gvr).Namespace(namespace).DeleteCollection(&opts, metav1.ListOptions{})
	if err == nil {
		return true, nil
	}

	// this is strange, but we need to special case for both MethodNotSupported and NotFound errors
	// TODO: https://github.com/kubernetes/kubernetes/issues/22413
	// we have a resource returned in the discovery API that supports no top-level verbs:
	//  /apis/extensions/v1beta1/namespaces/default/replicationcontrollers
	// when working with this resource type, we will get a literal not found error rather than expected method not supported
	if errors.IsMethodNotSupported(err) || errors.IsNotFound(err) {
		log.Infof("namespace controller - deleteCollection not supported - namespace: %s, gvr: %v", namespace, gvr)
		return false, nil
	}

	log.Infof("namespace controller - deleteCollection unexpected error - namespace: %s, gvr: %v, error: %v", namespace, gvr, err)
	return true, err
}

// listCollection will list the items in the specified namespace
// it returns the following:
//  the list of items in the collection (if found)
//  a boolean if the operation is supported
//  an error if the operation is supported but could not be completed.
func (deleter *namespacedResourcesDeleter) listCollection(gvr schema.GroupVersionResource, namespace string) (*metav1.PartialObjectMetadataList, bool, error) {
	log.Infof("namespace controller - listCollection - namespace: %s, gvr: %v", namespace, gvr)

	key := operationKey{operation: operationList, gvr: gvr}
	if !deleter.opCache.isSupported(key) {
		log.Infof("namespace controller - listCollection ignored since not supported - namespace: %s, gvr: %v", namespace, gvr)
		return nil, false, nil
	}

	partialList, err := deleter.metadataClient.Resource(gvr).Namespace(namespace).List(metav1.ListOptions{})
	if err == nil {
		return partialList, true, nil
	}

	if errors.IsMethodNotSupported(err) || errors.IsNotFound(err) {
		log.Infof("namespace controller - listCollection not supported - namespace: %s, gvr: %v", namespace, gvr)
		return nil, false, nil
	}

	return nil, true, err
}

// deleteEachItem is a helper function that will list the collection of resources and delete each item 1 by 1.
func (deleter *namespacedResourcesDeleter) deleteEachItem(gvr schema.GroupVersionResource, namespace string) error {
	log.Infof("namespace controller - deleteEachItem - namespace: %s, gvr: %v", namespace, gvr)

	unstructuredList, listSupported, err := deleter.listCollection(gvr, namespace)
	if err != nil {
		return err
	}
	if !listSupported {
		return nil
	}
	for _, item := range unstructuredList.Items {
		background := metav1.DeletePropagationBackground
		opts := metav1.DeleteOptions{PropagationPolicy: &background}
		err = deleter.metadataClient.Resource(gvr).Namespace(namespace).Delete(item.GetName(), &opts)
		if err != nil && !errors.IsNotFound(err) && !errors.IsMethodNotSupported(err) {
			return err
		}
	}

	return nil
}

func (deleter *namespacedResourcesDeleter) estimateGracefulTerminationForPods(namespace string) (int64, error) {
	log.Infof("namespace controller - estimateGracefulTerminationForPods - namespace %s", namespace)

	podList, err := deleter.clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return 0, err
	}

	// 找出所有pod.Spec.TerminationGracePeriodSeconds最大值
	estimate := int64(0)
	for _, pod := range podList.Items {
		// filter out terminal pods
		if v1.PodSucceeded == pod.Status.Phase || v1.PodFailed == pod.Status.Phase {
			continue
		}
		if pod.Spec.TerminationGracePeriodSeconds != nil {
			grace := *pod.Spec.TerminationGracePeriodSeconds
			if grace > estimate {
				estimate = grace
			}
		}
	}

	return estimate, nil
}

// 评估下graceful termination该namespace下所有pod资源所需要的时间
func (deleter *namespacedResourcesDeleter) estimateGracefulTermination(gvr schema.GroupVersionResource, namespace string, namespaceDeletedAt metav1.Time) (int64, error) {
	groupResource := gvr.GroupResource()
	log.Infof("namespace controller - estimateGracefulTermination - group %s, resource: %s", groupResource.Group, groupResource.Resource)

	estimate := int64(0)
	var err error
	switch groupResource {
	case schema.GroupResource{Group: "", Resource: "pods"}:
		estimate, err = deleter.estimateGracefulTerminationForPods(namespace)
	}
	if err != nil {
		return 0, err
	}

	// determine if the estimate is greater than the deletion timestamp
	duration := time.Since(namespaceDeletedAt.Time)
	allowedEstimate := time.Duration(estimate) * time.Second
	if duration >= allowedEstimate {
		estimate = int64(0)
	}
	return estimate, nil
}

type gvrDeletionMetadata struct {
	// finalizerEstimateSeconds is an estimate of how much longer to wait.  zero means that no estimate has made and does not
	// mean that all content has been removed.
	finalizerEstimateSeconds int64
	// numRemaining is how many instances of the gvr remain
	numRemaining int
	// finalizersToNumRemaining maps finalizers to how many resources are stuck on them
	finalizersToNumRemaining map[string]int
}

func (deleter *namespacedResourcesDeleter) deleteAllContentForGroupVersionResource(gvr schema.GroupVersionResource, namespace string, namespaceDeletedAt metav1.Time) (gvrDeletionMetadata, error) {
	log.Infof("namespace controller - deleteAllContentForGroupVersionResource - namespace: %s, gvr: %v", namespace, gvr)

	// 评估下删除该gvr所需要的时间
	// estimate how long it will take for the resource to be deleted (needed for objects that support graceful delete)
	estimate, err := deleter.estimateGracefulTermination(gvr, namespace, namespaceDeletedAt)
	if err != nil {
		log.Infof("namespace controller - deleteAllContentForGroupVersionResource - unable to estimate - namespace: %s, gvr: %v, err: %v", namespace, gvr, err)
		return gvrDeletionMetadata{}, err
	}
	log.Infof("namespace controller - deleteAllContentForGroupVersionResource - estimate - namespace: %s, gvr: %v, estimate: %v", namespace, gvr, estimate)

	// first try to delete the entire collection
	deleteCollectionSupported, err := deleter.deleteCollection(gvr, namespace)
	if err != nil {
		return gvrDeletionMetadata{finalizerEstimateSeconds: estimate}, err
	}

	// delete collection was not supported, so we list and delete each item...
	if !deleteCollectionSupported {
		err = deleter.deleteEachItem(gvr, namespace)
		if err != nil {
			return gvrDeletionMetadata{finalizerEstimateSeconds: estimate}, err
		}
	}

	// 调用listCollection查看该namespace下还有没有未删除的资源对象
	// verify there are no more remaining items
	// it is not an error condition for there to be remaining items if local estimate is non-zero
	log.Infof("namespace controller - deleteAllContentForGroupVersionResource - checking for no more items in namespace: %s, gvr: %v", namespace, gvr)
	unstructuredList, listSupported, err := deleter.listCollection(gvr, namespace)
	if err != nil {
		log.Infof("namespace controller - deleteAllContentForGroupVersionResource - error verifying no items in namespace: %s, gvr: %v, err: %v", namespace, gvr, err)
		return gvrDeletionMetadata{finalizerEstimateSeconds: estimate}, err
	}
	if !listSupported {
		return gvrDeletionMetadata{finalizerEstimateSeconds: estimate}, nil
	}
	log.Infof("namespace controller - deleteAllContentForGroupVersionResource - items remaining - namespace: %s, gvr: %v, items: %v", namespace, gvr, len(unstructuredList.Items))
	if len(unstructuredList.Items) == 0 {
		// we're done
		return gvrDeletionMetadata{finalizerEstimateSeconds: 0, numRemaining: 0}, nil
	}

	// use the list to find the finalizers
	finalizersToNumRemaining := map[string]int{}
	for _, item := range unstructuredList.Items {
		for _, finalizer := range item.GetFinalizers() {
			finalizersToNumRemaining[finalizer] = finalizersToNumRemaining[finalizer] + 1
		}
	}

	if estimate != int64(0) {
		log.Infof("namespace controller - deleteAllContentForGroupVersionResource - estimate is present - namespace: %s, gvr: %v, finalizers: %v", namespace, gvr, finalizersToNumRemaining)
		return gvrDeletionMetadata{
			finalizerEstimateSeconds: estimate,
			numRemaining:             len(unstructuredList.Items),
			finalizersToNumRemaining: finalizersToNumRemaining,
		}, nil
	}

	// if any item has a finalizer, we treat that as a normal condition, and use a default estimation to allow for GC to complete.
	if len(finalizersToNumRemaining) > 0 {
		log.Infof("namespace controller - deleteAllContentForGroupVersionResource - items remaining with finalizers - namespace: %s, gvr: %v, finalizers: %v", namespace, gvr, finalizersToNumRemaining)
		return gvrDeletionMetadata{
			finalizerEstimateSeconds: finalizerEstimateSeconds,
			numRemaining:             len(unstructuredList.Items),
			finalizersToNumRemaining: finalizersToNumRemaining,
		}, nil
	}

	// nothing reported a finalizer, so something was unexpected as it should have been deleted.
	return gvrDeletionMetadata{
		finalizerEstimateSeconds: estimate,
		numRemaining:             len(unstructuredList.Items),
	}, fmt.Errorf("unexpected items still remain in namespace: %s for gvr: %v", namespace, gvr)
}

type allGVRDeletionMetadata struct {
	// gvrToNumRemaining is how many instances of the gvr remain
	gvrToNumRemaining map[schema.GroupVersionResource]int
	// finalizersToNumRemaining maps finalizers to how many resources are stuck on them
	finalizersToNumRemaining map[string]int
}

func (deleter *namespacedResourcesDeleter) deleteAllContent(namespace *v1.Namespace) (int64, error) {
	log.Infof("namespace controller - deleteAllContent - namespace: %s", namespace.Name)

	estimate := int64(0)
	conditionUpdater := namespaceConditionUpdater{}

	var errs []error
	resources, err := deleter.discoverResourcesFn()
	if err != nil {
		// discovery errors are not fatal.  We often have some set of resources we can operate against even if we don't have a complete list
		errs = append(errs, err)
		conditionUpdater.ProcessDiscoverResourcesErr(err)
	}

	// 过滤出Verb包含delete的resources
	deletableResources := discovery.FilteredBy(discovery.SupportsAllVerbs{Verbs: []string{"delete"}}, resources)
	groupVersionResources, err := discovery.GroupVersionResources(deletableResources)
	if err != nil {
		// discovery errors are not fatal.  We often have some set of resources we can operate against even if we don't have a complete list
		errs = append(errs, err)
		conditionUpdater.ProcessGroupVersionErr(err)
	}

	numRemainingTotals := allGVRDeletionMetadata{
		gvrToNumRemaining:        map[schema.GroupVersionResource]int{},
		finalizersToNumRemaining: map[string]int{},
	}
	for gvr := range groupVersionResources {
		// 根据gvr删除资源，有的资源支持批量删除，有的资源需要一个个删除
		gvrDeletionMetadata, err := deleter.deleteAllContentForGroupVersionResource(gvr, namespace.Name, *namespace.DeletionTimestamp)
		if err != nil {
			// If there is an error, hold on to it but proceed with all the remaining
			// groupVersionResources.
			errs = append(errs, err)
			conditionUpdater.ProcessDeleteContentErr(err)
		}
		if gvrDeletionMetadata.finalizerEstimateSeconds > estimate {
			estimate = gvrDeletionMetadata.finalizerEstimateSeconds
		}
		if gvrDeletionMetadata.numRemaining > 0 {
			numRemainingTotals.gvrToNumRemaining[gvr] = gvrDeletionMetadata.numRemaining
			for finalizer, numRemaining := range gvrDeletionMetadata.finalizersToNumRemaining {
				if numRemaining == 0 {
					continue
				}
				numRemainingTotals.finalizersToNumRemaining[finalizer] = numRemainingTotals.finalizersToNumRemaining[finalizer] + numRemaining
			}
		}
	}

	conditionUpdater.ProcessContentTotals(numRemainingTotals)

	// 最后要更新下status.conditions字段

	// we always want to update the conditions because if we have set a condition to "it worked" after it was previously, "it didn't work",
	// we need to reflect that information.  Recall that additional finalizers can be set on namespaces, so this finalizer may clear itself and
	// NOT remove the resource instance.
	if hasChanged := conditionUpdater.Update(namespace); hasChanged {
		if _, err = deleter.clientset.CoreV1().Namespaces().UpdateStatus(namespace); err != nil {
			utilruntime.HandleError(fmt.Errorf("couldn't update status condition for namespace %q: %v", namespace, err))
		}
	}

	log.Infof("namespace controller - deleteAllContent - namespace: %s, estimate: %v, errors: %v", namespace, estimate, utilerrors.NewAggregate(errs))

	return estimate, utilerrors.NewAggregate(errs)
}

// 删除给定namespace中的所有资源对象
// 删除前：检查namespace.DeletionTimestamp字段值；检查namespace.Status.Phase是不是"Terminating"状态
// 删除后：移除namespace.Spec.Finalizers并finalize下namespace
func (deleter *namespacedResourcesDeleter) Delete(namespaceName string) error {
	// 首先从etcd中捞出namespace对象最新状态，防止已经被删除了
	namespace, err := deleter.clientset.CoreV1().Namespaces().Get(namespaceName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}

		return err
	}
	// DeletionTimestamp字段为空，不需要删除
	if namespace.DeletionTimestamp == nil {
		return nil
	}

	log.Infof("namespace controller - syncNamespace - namespace: %s, finalizerToken: %s", namespace.Name, deleter.finalizerToken)

	// ensure that the status is up to date on the namespace
	// if we get a not found error, we assume the namespace is truly gone
	namespace, err = deleter.retryOnConflictError(namespace, deleter.updateNamespaceStatusFunc)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	// the latest view of the namespace asserts that namespace is no longer deleting..
	if namespace.DeletionTimestamp.IsZero() {
		return nil
	}

	// return if it is already finalized.
	if finalized(namespace) {
		return nil
	}

	// 开始删除该namespace下所有资源对象
	// there may still be content for us to remove
	estimate, err := deleter.deleteAllContent(namespace)
	if err != nil {
		return err
	}
	if estimate > 0 {
		return &ResourcesRemainingError{estimate}
	}

	// we have removed content, so mark it finalized by us
	_, err = deleter.retryOnConflictError(namespace, deleter.finalizeNamespace)
	if err != nil {
		// in normal practice, this should not be possible, but if a deployment is running
		// two controllers to do namespace deletion that share a common finalizer token it's
		// possible that a not found could occur since the other controller would have finished the delete.
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return nil
}

type ResourcesRemainingError struct {
	Estimate int64
}

func (e *ResourcesRemainingError) Error() string {
	return fmt.Sprintf("some content remains in the namespace, estimate %d seconds before it is removed", e.Estimate)
}
