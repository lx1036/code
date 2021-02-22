package controller

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corelistersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// LeaderElection determines whether to enable leader election or not. Defaults
// to true.
func LeaderElection(leaderElection bool) func(*ProvisionController) error {
	return func(c *ProvisionController) error {
		if c.HasRun() {
			return errRuntime
		}
		c.leaderElection = leaderElection
		return nil
	}
}

// FailedProvisionThreshold is the threshold for max number of retries on
// failures of Provision. Set to 0 to retry indefinitely. Defaults to 15.
func FailedProvisionThreshold(failedProvisionThreshold int) func(*ProvisionController) error {
	return func(c *ProvisionController) error {
		if c.HasRun() {
			return errRuntime
		}
		c.failedProvisionThreshold = failedProvisionThreshold
		return nil
	}
}

// FailedDeleteThreshold is the threshold for max number of retries on failures
// of Delete. Set to 0 to retry indefinitely. Defaults to 15.
func FailedDeleteThreshold(failedDeleteThreshold int) func(*ProvisionController) error {
	return func(c *ProvisionController) error {
		if c.HasRun() {
			return errRuntime
		}
		c.failedDeleteThreshold = failedDeleteThreshold
		return nil
	}
}

// RateLimiter is the workqueue.RateLimiter to use for the provisioning and
// deleting work queues. If set, ExponentialBackOffOnError is ignored.
func RateLimiter(rateLimiter workqueue.RateLimiter) func(*ProvisionController) error {
	return func(c *ProvisionController) error {
		if c.HasRun() {
			return errRuntime
		}
		c.rateLimiter = rateLimiter
		return nil
	}
}

// Threadiness is the number of claim and volume workers each to launch.
// Defaults to 4.
func Threadiness(threadiness int) func(*ProvisionController) error {
	return func(c *ProvisionController) error {
		if c.HasRun() {
			return errRuntime
		}
		c.threadiness = threadiness
		return nil
	}
}

// CreateProvisionedPVRetryCount is the number of retries when we create a PV
// object for a provisioned volume. Defaults to 5.
// If PV is not saved after given number of retries, corresponding storage asset (volume) is deleted!
// Only one of CreateProvisionedPVInterval+CreateProvisionedPVRetryCount or CreateProvisionedPVBackoff or
// CreateProvisionedPVLimiter can be used.
// Deprecated: Use CreateProvisionedPVLimiter instead, it tries indefinitely.
func CreateProvisionedPVRetryCount(createProvisionedPVRetryCount int) func(*ProvisionController) error {
	return func(c *ProvisionController) error {
		if c.HasRun() {
			return errRuntime
		}
		if c.createProvisionedPVBackoff != nil {
			return fmt.Errorf("CreateProvisionedPVBackoff cannot be used together with CreateProvisionedPVRetryCount")
		}
		if c.createProvisionerPVLimiter != nil {
			return fmt.Errorf("CreateProvisionedPVBackoff cannot be used together with CreateProvisionedPVLimiter")
		}
		c.createProvisionedPVRetryCount = createProvisionedPVRetryCount
		return nil
	}
}

// CreateProvisionedPVLimiter is the configuration of rate limiter for queue of unsaved PersistentVolumes.
// If set, PVs that fail to be saved to Kubernetes API server will be re-enqueued to a separate workqueue
// with this limiter and re-tried until they are saved to API server. There is no limit of retries.
// The main difference to other CreateProvisionedPV* option is that the storage asset is never deleted
// and the controller continues saving PV to API server indefinitely.
// This option cannot be used with CreateProvisionedPVBackoff or CreateProvisionedPVInterval
// or CreateProvisionedPVRetryCount.
func CreateProvisionedPVLimiter(limiter workqueue.RateLimiter) func(*ProvisionController) error {
	return func(c *ProvisionController) error {
		if c.HasRun() {
			return errRuntime
		}
		if c.createProvisionedPVRetryCount != 0 {
			return fmt.Errorf("CreateProvisionedPVLimiter cannot be used together with CreateProvisionedPVRetryCount")
		}
		if c.createProvisionedPVInterval != 0 {
			return fmt.Errorf("CreateProvisionedPVLimiter cannot be used together with CreateProvisionedPVInterval")
		}
		if c.createProvisionedPVBackoff != nil {
			return fmt.Errorf("CreateProvisionedPVLimiter cannot be used together with CreateProvisionedPVBackoff")
		}
		c.createProvisionerPVLimiter = limiter
		return nil
	}
}

// ClaimsInformer sets the informer to use for accessing PersistentVolumeClaims.
// Defaults to using a internal informer.
func ClaimsInformer(informer cache.SharedIndexInformer) func(*ProvisionController) error {
	return func(c *ProvisionController) error {
		if c.HasRun() {
			return errRuntime
		}
		c.claimInformer = informer
		c.customClaimInformer = true
		return nil
	}
}

// NodesLister sets the informer to use for accessing Nodes.
// This is needed only for PVCs which have a selected node.
// Defaults to using a GET instead of an informer.
//
// Which approach is better depends on factors like cluster size and
// ratio of PVCs with a selected node.
func NodesLister(nodeLister corelistersv1.NodeLister) func(*ProvisionController) error {
	return func(c *ProvisionController) error {
		if c.HasRun() {
			return errRuntime
		}
		c.nodeLister = nodeLister
		return nil
	}
}

// AdditionalProvisionerNames sets additional names for the provisioner
func AdditionalProvisionerNames(additionalProvisionerNames []string) func(*ProvisionController) error {
	return func(c *ProvisionController) error {
		if c.HasRun() {
			return errRuntime
		}
		c.additionalProvisionerNames = additionalProvisionerNames
		return nil
	}
}

// getInClusterNamespace returns the namespace in which the controller runs.
func getInClusterNamespace() string {
	if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
		return ns
	}

	// Fall back to the namespace associated with the service account token, if available
	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns
		}
	}

	return "default"
}

func getObjectUID(obj interface{}) (string, error) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			return "", fmt.Errorf("error decoding object, invalid type")
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			return "", fmt.Errorf("error decoding object tombstone, invalid type")
		}
	}

	return string(object.GetUID()), nil
}

// GetPersistentVolumeClaimClass returns StorageClassName. If no storage class was
// requested, it returns "".
func GetPersistentVolumeClaimClass(claim *v1.PersistentVolumeClaim) string {
	// Use beta annotation first
	if class, found := claim.Annotations[v1.BetaStorageClassAnnotation]; found {
		return class
	}

	if claim.Spec.StorageClassName != nil {
		return *claim.Spec.StorageClassName
	}

	return ""
}

func logOperation(operation, format string, a ...interface{}) string {
	return fmt.Sprintf(fmt.Sprintf("%s: %s", operation, format), a...)
}

func claimToClaimKey(claim *v1.PersistentVolumeClaim) string {
	return fmt.Sprintf("%s/%s", claim.Namespace, claim.Name)
}

func getString(m map[string]string, key string, alts ...string) (string, bool) {
	if m == nil {
		return "", false
	}
	keys := append([]string{key}, alts...)
	for _, k := range keys {
		if v, ok := m[k]; ok {
			return v, true
		}
	}
	return "", false
}
