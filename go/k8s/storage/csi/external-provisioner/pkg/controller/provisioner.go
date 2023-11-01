package controller

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"k8s-lx1036/k8s/storage/csi/csi-lib-utils/rpc"

	"google.golang.org/grpc"

	"github.com/container-storage-interface/spec/lib/go/csi"
	snapapi "github.com/kubernetes-csi/external-snapshotter/client/v3/apis/volumesnapshot/v1beta1"
	snapclientset "github.com/kubernetes-csi/external-snapshotter/client/v3/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	storagelistersv1 "k8s.io/client-go/listers/storage/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/volume/util"
)

const (
	ResyncPeriodOfCsiNodeInformer = 1 * time.Hour

	annMigratedTo = "pv.kubernetes.io/migrated-to"

	snapshotKind     = "VolumeSnapshot"
	snapshotAPIGroup = snapapi.GroupName       // "snapshot.storage.k8s.io"
	pvcKind          = "PersistentVolumeClaim" // Native types don't require an API group

	// CSI Parameters prefixed with csiParameterPrefix are not passed through
	// to the driver on CreateVolumeRequest calls. Instead they are intended
	// to used by the CSI external-provisioner and maybe used to populate
	// fields in subsequent CSI calls or Kubernetes API objects.
	csiParameterPrefix = "csi.storage.k8s.io/"

	prefixedFsTypeKey = csiParameterPrefix + "fstype"

	snapshotNotBound = "snapshot %s not bound"

	pvcCloneFinalizer = "provisioner.storage.kubernetes.io/cloning-protection"

	// Each provisioner have a identify string to distinguish with others. This
	// identify string will be added in PV annoations under this key.
	provisionerIDKey = "storage.kubernetes.io/csiProvisionerIdentity"
)

type internalNodeDeployment struct {
	NodeDeployment
	rateLimiter workqueue.RateLimiter
}

// becomeOwner updates the PVC with the current node as selected node.
// Returns an error if something unexpectedly failed, otherwise an updated PVC with
// the current node selected or nil if not the owner.
func (nc *internalNodeDeployment) becomeOwner(ctx context.Context, p *csiProvisioner, claim *v1.PersistentVolumeClaim) error {
	requeues := nc.rateLimiter.NumRequeues(claim.UID)
	delay := nc.rateLimiter.When(claim.UID)
	klog.V(5).Infof("will try to become owner of PVC %s/%s with resource version %s in %s (attempt #%d)", claim.Namespace, claim.Name, claim.ResourceVersion, delay, requeues)
	sleep, cancel := context.WithTimeout(ctx, delay)
	defer cancel()
	// When the base delay is high we also should check less often.
	// With multiple provisioners running in parallel, it becomes more
	// likely that one of them became the owner quickly, so we don't
	// want to check too slowly either.
	pollInterval := nc.BaseDelay / 100
	if pollInterval < 10*time.Millisecond {
		pollInterval = 10 * time.Millisecond
	}
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	check := func() (bool, *v1.PersistentVolumeClaim, error) {
		current, err := nc.ClaimInformer.Lister().PersistentVolumeClaims(claim.Namespace).Get(claim.Name)
		if err != nil {
			return false, nil, fmt.Errorf("PVC not found: %v", err)
		}
		if claim.UID != current.UID {
			return false, nil, errors.New("PVC was replaced")
		}
		if current.Annotations != nil && current.Annotations[annSelectedNode] != "" && current.Annotations[annSelectedNode] != nc.NodeName {
			return true, current, nil
		}
		return false, current, nil
	}
	var stop bool
	var current *v1.PersistentVolumeClaim
	var err error
loop:
	for {
		select {
		case <-ctx.Done():
			return errors.New("timed out waiting to become PVC owner")
		case <-sleep.Done():
			stop, current, err = check()
			break loop
		case <-ticker.C:
			// Abort the waiting early if we know that someone else is the owner.
			stop, current, err = check()
			if err != nil || stop {
				break loop
			}
		}
	}
	if err != nil {
		return err
	}
	if stop {
		// Some other instance was faster and we don't need to provision for
		// this PVC. If the PVC needs to be rescheduled, we start the delay from scratch.
		nc.rateLimiter.Forget(claim.UID)
		klog.V(5).Infof("did not become owner of PVC %s/%s with resource revision %s, now owned by %s with resource revision %s",
			claim.Namespace, claim.Name, claim.ResourceVersion,
			current.Annotations[annSelectedNode], current.ResourceVersion)
		return nil
	}

	// Check capacity as late as possible before trying to become the owner, because that is a
	// relatively expensive operation.
	//
	// The exact same parameters are computed here as if we were provisioning. If a precondition
	// is violated, like "storage class does not exist", then we have two options:
	// - silently ignore the problem, but if all instances do that, the problem is not surfaced
	//   to the user
	// - try to become the owner and let provisioning start, which then will probably
	//   fail the same way, but then has a chance to inform the user via events
	//
	// We do the latter.
	/*hasCapacity, err := p.checkCapacity(ctx, claim, p.nodeDeployment.NodeName)
	if err != nil {
		klog.V(3).Infof("proceeding with becoming owner although the capacity check failed: %v", err)
	} else if !hasCapacity {
		// Don't try to provision.
		klog.V(5).Infof("not enough capacity for PVC %s/%s with resource revision %s", claim.Namespace, claim.Name, claim.ResourceVersion)
		return nil
	}*/

	// Update PVC with our node as selected node if necessary.
	current = current.DeepCopy()
	if current.Annotations == nil {
		current.Annotations = map[string]string{}
	}
	if current.Annotations[annSelectedNode] == nc.NodeName {
		// A mere sanity check. Should not happen.
		klog.V(5).Infof("already owner of PVC %s/%s with updated resource version %s", current.Namespace, current.Name, current.ResourceVersion)
		return nil
	}
	current.Annotations[annSelectedNode] = nc.NodeName
	klog.V(5).Infof("trying to become owner of PVC %s/%s with resource version %s now", current.Namespace, current.Name, current.ResourceVersion)
	current, err = p.client.CoreV1().PersistentVolumeClaims(current.Namespace).Update(ctx, current, metav1.UpdateOptions{})
	if err != nil {
		// Next attempt will use a longer delay and most likely
		// stop quickly once we see who owns the PVC now.
		if apierrors.IsConflict(err) {
			// Lost the race or some other concurrent modification. Repeat the attempt.
			klog.V(3).Infof("conflict during PVC %s/%s update, will try again", claim.Namespace, claim.Name)
			return nc.becomeOwner(ctx, p, claim)
		}
		// Some unexpected error. Report it.
		return fmt.Errorf("selecting node %q for PVC failed: %v", nc.NodeName, err)
	}

	// Successfully became owner. Future delays will be smaller again.
	nc.rateLimiter.Forget(claim.UID)
	klog.V(5).Infof("became owner of PVC %s/%s with updated resource version %s", current.Namespace, current.Name, current.ResourceVersion)
	return nil
}

// NodeDeployment contains additional parameters for running external-provisioner alongside a
// CSI driver on one or more nodes in the cluster.
type NodeDeployment struct {
	// NodeName is the name of the node in Kubernetes on which the external-provisioner runs.
	NodeName string
	// ClaimInformer is needed to detect when some other external-provisioner
	// became the owner of a PVC while the local one is still waiting before
	// trying to become the owner itself.
	ClaimInformer coreinformers.PersistentVolumeClaimInformer
	// NodeInfo is the result of NodeGetInfo. It is need to determine which
	// PVs were created for the node.
	NodeInfo csi.NodeGetInfoResponse
	// ImmediateBinding enables support for PVCs with immediate binding.
	ImmediateBinding bool
	// BaseDelay is the initial time that the external-provisioner waits
	// before trying to become the owner of a PVC with immediate binding.
	BaseDelay time.Duration
	// MaxDelay is the maximum for the initial wait time.
	MaxDelay time.Duration
}

// requiredCapabilities provides a set of extra capabilities required for special/optional features provided by a plugin
type requiredCapabilities struct {
	snapshot bool
	clone    bool
}

type prepareProvisionResult struct {
	fsType         string
	migratedVolume bool
	req            *csi.CreateVolumeRequest
	csiPVSource    *v1.CSIPersistentVolumeSource
}

type csiProvisioner struct {
	client                 kubernetes.Interface
	csiClient              csi.ControllerClient
	grpcClient             *grpc.ClientConn
	snapshotClient         snapclientset.Interface
	timeout                time.Duration
	identity               string
	volumeNamePrefix       string
	defaultFSType          string
	volumeNameUUIDLength   int
	config                 *rest.Config
	driverName             string
	pluginCapabilities     rpc.PluginCapabilitySet
	controllerCapabilities rpc.ControllerCapabilitySet
	strictTopology         bool
	immediateTopology      bool
	scLister               storagelistersv1.StorageClassLister
	csiNodeLister          storagelistersv1.CSINodeLister
	nodeLister             corelisters.NodeLister
	claimLister            corelisters.PersistentVolumeClaimLister
	vaLister               storagelistersv1.VolumeAttachmentLister
	extraCreateMetadata    bool
	eventRecorder          record.EventRecorder
	nodeDeployment         *internalNodeDeployment
}

// checkNode optionally checks whether the PVC is assigned to the current node.
// If the PVC uses immediate binding, it will try to take the PVC for provisioning
// on the current node. Returns true if provisioning can proceed, an error
// in case of a failure that prevented checking.
func (p *csiProvisioner) checkNode(ctx context.Context, claim *v1.PersistentVolumeClaim, sc *storagev1.StorageClass, caller string) (provision bool, err error) {
	if p.nodeDeployment == nil {
		return true, nil
	}

	var selectedNode string
	if claim.Annotations != nil {
		selectedNode = claim.Annotations[annSelectedNode]
	}
	switch selectedNode {
	case "":
		logger := klog.V(5)
		if logger.Enabled() {
			logger.Infof("%s: checking node for PVC %s/%s with resource version %s", caller, claim.Namespace, claim.Name, claim.ResourceVersion)
			defer func() {
				logger.Infof("%s: done checking node for PVC %s/%s with resource version %s: provision %v, err %v", caller, claim.Namespace, claim.Name, claim.ResourceVersion, provision, err)
			}()
		}

		if sc == nil {
			var err error
			sc, err = p.scLister.Get(*claim.Spec.StorageClassName)
			if err != nil {
				return false, err
			}
		}
		if sc.VolumeBindingMode == nil ||
			*sc.VolumeBindingMode != storagev1.VolumeBindingImmediate ||
			!p.nodeDeployment.ImmediateBinding {
			return false, nil
		}

		// Try to select the current node if there is a chance of it
		// being created there, i.e. there is currently enough free space (checked in becomeOwner).
		//
		// If later volume provisioning fails on this node, the annotation will be unset and node
		// selection will happen again. If no other node picks up the volume, then the PVC remains
		// in the queue and this check will be repeated from time to time.
		//
		// A lot of different external-provisioner instances will try to do this at the same time.
		// To avoid the thundering herd problem, we sleep in becomeOwner for a short random amount of time
		// (for new PVCs) or exponentially increasing time (for PVCs were we already had a conflict).
		if err := p.nodeDeployment.becomeOwner(ctx, p, claim); err != nil {
			return false, fmt.Errorf("PVC %s/%s: %v", claim.Namespace, claim.Name, err)
		}

		// We are now either the owner or someone else is. We'll check when the updated PVC
		// enters the workqueue and gets processed by sig-storage-lib-external-provisioner.
		return false, nil
	case p.nodeDeployment.NodeName:
		// Our node is selected.
		return true, nil
	default:
		// Some other node is selected, ignore it.
		return false, nil
	}
}

// This function get called before any attempt to communicate with the driver.
// Before initiating Create/Delete API calls provisioner checks if Capabilities:
// PluginControllerService,  ControllerCreateVolume sre supported and gets the  driver name.
func (p *csiProvisioner) checkDriverCapabilities(rc *requiredCapabilities) error {
	if !p.pluginCapabilities[csi.PluginCapability_Service_CONTROLLER_SERVICE] {
		return fmt.Errorf("CSI driver does not support dynamic provisioning: plugin CONTROLLER_SERVICE capability is not reported")
	}

	if !p.controllerCapabilities[csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME] {
		return fmt.Errorf("CSI driver does not support dynamic provisioning: controller CREATE_DELETE_VOLUME capability is not reported")
	}

	if rc.snapshot {
		// Check whether plugin supports create snapshot
		// If not, create volume from snapshot cannot proceed
		if !p.controllerCapabilities[csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT] {
			return fmt.Errorf("CSI driver does not support snapshot restore: controller CREATE_DELETE_SNAPSHOT capability is not reported")
		}
	}
	if rc.clone {
		// Check whether plugin supports clone operations
		// If not, create volume from pvc cannot proceed
		if !p.controllerCapabilities[csi.ControllerServiceCapability_RPC_CLONE_VOLUME] {
			return fmt.Errorf("CSI driver does not support clone operations: controller CLONE_VOLUME capability is not reported")
		}
	}

	return nil
}

// getSnapshotSource verifies DataSource.Kind of type VolumeSnapshot, making sure that the requested Snapshot is available/ready
// returns the VolumeContentSource for the requested snapshot
func (p *csiProvisioner) getSnapshotSource(ctx context.Context, claim *v1.PersistentVolumeClaim, sc *storagev1.StorageClass) (*csi.VolumeContentSource, error) {
	snapshotObj, err := p.snapshotClient.SnapshotV1beta1().VolumeSnapshots(claim.Namespace).Get(ctx, claim.Spec.DataSource.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting snapshot %s from api server: %v", claim.Spec.DataSource.Name, err)
	}

	if snapshotObj.ObjectMeta.DeletionTimestamp != nil {
		return nil, fmt.Errorf("snapshot %s is currently being deleted", claim.Spec.DataSource.Name)
	}
	klog.V(5).Infof("VolumeSnapshot %+v", snapshotObj)

	if snapshotObj.Status == nil || snapshotObj.Status.BoundVolumeSnapshotContentName == nil {
		return nil, fmt.Errorf(snapshotNotBound, claim.Spec.DataSource.Name)
	}

	snapContentObj, err := p.snapshotClient.SnapshotV1beta1().VolumeSnapshotContents().Get(ctx, *snapshotObj.Status.BoundVolumeSnapshotContentName, metav1.GetOptions{})

	if err != nil {
		klog.Warningf("error getting snapshotcontent %s for snapshot %s/%s from api server: %s", *snapshotObj.Status.BoundVolumeSnapshotContentName, snapshotObj.Namespace, snapshotObj.Name, err)
		return nil, fmt.Errorf(snapshotNotBound, claim.Spec.DataSource.Name)
	}

	if snapContentObj.Spec.VolumeSnapshotRef.UID != snapshotObj.UID || snapContentObj.Spec.VolumeSnapshotRef.Namespace != snapshotObj.Namespace || snapContentObj.Spec.VolumeSnapshotRef.Name != snapshotObj.Name {
		klog.Warningf("snapshotcontent %s for snapshot %s/%s is bound to a different snapshot", *snapshotObj.Status.BoundVolumeSnapshotContentName, snapshotObj.Namespace, snapshotObj.Name)
		return nil, fmt.Errorf(snapshotNotBound, claim.Spec.DataSource.Name)
	}

	if snapContentObj.Spec.Driver != sc.Provisioner {
		klog.Warningf("snapshotcontent %s for snapshot %s/%s is handled by a different CSI driver than requested by StorageClass %s", *snapshotObj.Status.BoundVolumeSnapshotContentName, snapshotObj.Namespace, snapshotObj.Name, sc.Name)
		return nil, fmt.Errorf(snapshotNotBound, claim.Spec.DataSource.Name)
	}

	if snapshotObj.Status.ReadyToUse == nil || *snapshotObj.Status.ReadyToUse == false {
		return nil, fmt.Errorf("snapshot %s is not Ready", claim.Spec.DataSource.Name)
	}

	klog.V(5).Infof("VolumeSnapshotContent %+v", snapContentObj)

	if snapContentObj.Status == nil || snapContentObj.Status.SnapshotHandle == nil {
		return nil, fmt.Errorf("snapshot handle %s is not available", claim.Spec.DataSource.Name)
	}

	snapshotSource := csi.VolumeContentSource_Snapshot{
		Snapshot: &csi.VolumeContentSource_SnapshotSource{
			SnapshotId: *snapContentObj.Status.SnapshotHandle,
		},
	}
	klog.V(5).Infof("VolumeContentSource_Snapshot %+v", snapshotSource)

	if snapshotObj.Status.RestoreSize != nil {
		capacity, exists := claim.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
		if !exists {
			return nil, fmt.Errorf("error getting capacity for PVC %s when creating snapshot %s", claim.Name, snapshotObj.Name)
		}
		volSizeBytes := capacity.Value()
		klog.V(5).Infof("Requested volume size is %d and snapshot size is %d for the source snapshot %s", int64(volSizeBytes), int64(snapshotObj.Status.RestoreSize.Value()), snapshotObj.Name)
		// When restoring volume from a snapshot, the volume size should
		// be equal to or larger than its snapshot size.
		if int64(volSizeBytes) < int64(snapshotObj.Status.RestoreSize.Value()) {
			return nil, fmt.Errorf("requested volume size %d is less than the size %d for the source snapshot %s", int64(volSizeBytes), int64(snapshotObj.Status.RestoreSize.Value()), snapshotObj.Name)
		}
		if int64(volSizeBytes) > int64(snapshotObj.Status.RestoreSize.Value()) {
			klog.Warningf("requested volume size %d is greater than the size %d for the source snapshot %s. Volume plugin needs to handle volume expansion.", int64(volSizeBytes), int64(snapshotObj.Status.RestoreSize.Value()), snapshotObj.Name)
		}
	}

	volumeContentSource := &csi.VolumeContentSource{
		Type: &snapshotSource,
	}

	return volumeContentSource, nil
}

// getPVCSource verifies DataSource.Kind of type PersistentVolumeClaim, making sure that the requested PVC is available/ready
// returns the VolumeContentSource for the requested PVC
func (p *csiProvisioner) getPVCSource(ctx context.Context, claim *v1.PersistentVolumeClaim, sc *storagev1.StorageClass) (*csi.VolumeContentSource, error) {
	sourcePVC, err := p.claimLister.PersistentVolumeClaims(claim.Namespace).Get(claim.Spec.DataSource.Name)
	if err != nil {
		return nil, fmt.Errorf("error getting PVC %s (namespace %q) from api server: %v", claim.Spec.DataSource.Name, claim.Namespace, err)
	}
	if string(sourcePVC.Status.Phase) != "Bound" {
		return nil, fmt.Errorf("the PVC DataSource %s must have a status of Bound.  Got %v", claim.Spec.DataSource.Name, sourcePVC.Status)
	}
	if sourcePVC.ObjectMeta.DeletionTimestamp != nil {
		return nil, fmt.Errorf("the PVC DataSource %s is currently being deleted", claim.Spec.DataSource.Name)
	}

	if sourcePVC.Spec.StorageClassName == nil {
		return nil, fmt.Errorf("the source PVC (%s) storageclass cannot be empty", sourcePVC.Name)
	}

	if claim.Spec.StorageClassName == nil {
		return nil, fmt.Errorf("the requested PVC (%s) storageclass cannot be empty", claim.Name)
	}

	if *sourcePVC.Spec.StorageClassName != *claim.Spec.StorageClassName {
		return nil, fmt.Errorf("the source PVC and destination PVCs must be in the same storage class for cloning.  Source is in %v, but new PVC is in %v",
			*sourcePVC.Spec.StorageClassName, *claim.Spec.StorageClassName)
	}

	capacity := claim.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	requestedSize := capacity.Value()
	srcCapacity := sourcePVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	srcPVCSize := srcCapacity.Value()
	if requestedSize < srcPVCSize {
		return nil, fmt.Errorf("error, new PVC request must be greater than or equal in size to the specified PVC data source, requested %v but source is %v", requestedSize, srcPVCSize)
	}

	if sourcePVC.Spec.VolumeName == "" {
		return nil, fmt.Errorf("volume name is empty in source PVC %s", sourcePVC.Name)
	}

	sourcePV, err := p.client.CoreV1().PersistentVolumes().Get(ctx, sourcePVC.Spec.VolumeName, metav1.GetOptions{})
	if err != nil {
		klog.Warningf("error getting volume %s for PVC %s/%s: %s", sourcePVC.Spec.VolumeName, sourcePVC.Namespace, sourcePVC.Name, err)
		return nil, fmt.Errorf("claim in dataSource not bound or invalid")
	}

	if sourcePV.Spec.CSI == nil {
		klog.Warningf("error getting volume source from %s for PVC %s/%s", sourcePVC.Spec.VolumeName, sourcePVC.Namespace, sourcePVC.Name)
		return nil, fmt.Errorf("claim in dataSource not bound or invalid")
	}

	if sourcePV.Spec.CSI.Driver != sc.Provisioner {
		klog.Warningf("the source volume %s for PVC %s/%s is handled by a different CSI driver than requested by StorageClass %s", sourcePVC.Spec.VolumeName, sourcePVC.Namespace, sourcePVC.Name, *claim.Spec.StorageClassName)
		return nil, fmt.Errorf("claim in dataSource not bound or invalid")
	}

	if sourcePV.Spec.ClaimRef == nil {
		klog.Warningf("the source volume %s for PVC %s/%s is not bound", sourcePVC.Spec.VolumeName, sourcePVC.Namespace, sourcePVC.Name)
		return nil, fmt.Errorf("claim in dataSource not bound or invalid")
	}

	if sourcePV.Spec.ClaimRef.UID != sourcePVC.UID || sourcePV.Spec.ClaimRef.Namespace != sourcePVC.Namespace || sourcePV.Spec.ClaimRef.Name != sourcePVC.Name {
		klog.Warningf("the source volume %s for PVC %s/%s is bound to a different PVC than requested", sourcePVC.Spec.VolumeName, sourcePVC.Namespace, sourcePVC.Name)
		return nil, fmt.Errorf("claim in dataSource not bound or invalid")
	}

	if sourcePV.Status.Phase != v1.VolumeBound {
		klog.Warningf("the source volume %s for PVC %s/%s status is \"%s\", should instead be \"%s\"", sourcePVC.Spec.VolumeName, sourcePVC.Namespace, sourcePVC.Name, sourcePV.Status.Phase, v1.VolumeBound)
		return nil, fmt.Errorf("claim in dataSource not bound or invalid")
	}

	if claim.Spec.VolumeMode == nil || *claim.Spec.VolumeMode == v1.PersistentVolumeFilesystem {
		if sourcePV.Spec.VolumeMode != nil && *sourcePV.Spec.VolumeMode != v1.PersistentVolumeFilesystem {
			return nil, fmt.Errorf("the source PVC and destination PVCs must have the same volume mode for cloning.  Source is Block, but new PVC requested Filesystem")
		}
	}

	if claim.Spec.VolumeMode != nil && *claim.Spec.VolumeMode == v1.PersistentVolumeBlock {
		if sourcePV.Spec.VolumeMode == nil || *sourcePV.Spec.VolumeMode != v1.PersistentVolumeBlock {
			return nil, fmt.Errorf("the source PVC and destination PVCs must have the same volume mode for cloning.  Source is Filesystem, but new PVC requested Block")
		}
	}

	volumeSource := csi.VolumeContentSource_Volume{
		Volume: &csi.VolumeContentSource_VolumeSource{
			VolumeId: sourcePV.Spec.CSI.VolumeHandle,
		},
	}
	klog.V(5).Infof("VolumeContentSource_Volume %+v", volumeSource)

	volumeContentSource := &csi.VolumeContentSource{
		Type: &volumeSource,
	}
	return volumeContentSource, nil
}

// getVolumeContentSource is a helper function to process provisioning requests that include a DataSource
// currently we provide Snapshot and PVC, the default case allows the provisioner to still create a volume
// so that an external controller can act upon it.   Additional DataSource types can be added here with
// an appropriate implementation function
func (p *csiProvisioner) getVolumeContentSource(ctx context.Context, claim *v1.PersistentVolumeClaim, sc *storagev1.StorageClass) (*csi.VolumeContentSource, error) {
	switch claim.Spec.DataSource.Kind {
	case snapshotKind:
		return p.getSnapshotSource(ctx, claim, sc)
	case pvcKind:
		return p.getPVCSource(ctx, claim, sc)
	default:
		// For now we shouldn't pass other things to this function, but treat it as a noop and extend as needed
		return nil, nil
	}
}

func (p *csiProvisioner) setCloneFinalizer(ctx context.Context, pvc *v1.PersistentVolumeClaim) error {
	claim, err := p.claimLister.PersistentVolumeClaims(pvc.Namespace).Get(pvc.Spec.DataSource.Name)
	if err != nil {
		return err
	}

	if !checkFinalizer(claim, pvcCloneFinalizer) {
		claim.Finalizers = append(claim.Finalizers, pvcCloneFinalizer)
		_, err := p.client.CoreV1().PersistentVolumeClaims(claim.Namespace).Update(ctx, claim, metav1.UpdateOptions{})
		return err
	}

	return nil
}

// prepareProvision does non-destructive parameter checking and preparations for provisioning a volume.
func (p *csiProvisioner) prepareProvision(ctx context.Context, claim *v1.PersistentVolumeClaim, sc *storagev1.StorageClass, selectedNode *v1.Node) (*prepareProvisionResult, ProvisioningState, error) {
	if sc == nil {
		return nil, ProvisioningFinished, errors.New("storage class was nil")
	}

	migratedVolume := false

	// Make sure the plugin is capable of fulfilling the requested options
	rc := &requiredCapabilities{}
	if claim.Spec.DataSource != nil {
		// PVC.Spec.DataSource.Name is the name of the VolumeSnapshot API object
		if claim.Spec.DataSource.Name == "" {
			return nil, ProvisioningFinished, fmt.Errorf("the PVC source not found for PVC %s", claim.Name)
		}

		switch claim.Spec.DataSource.Kind {
		case snapshotKind:
			if *(claim.Spec.DataSource.APIGroup) != snapshotAPIGroup {
				return nil, ProvisioningFinished, fmt.Errorf("the PVC source does not belong to the right APIGroup. Expected %s, Got %s", snapshotAPIGroup, *(claim.Spec.DataSource.APIGroup))
			}
			rc.snapshot = true
		case pvcKind:
			rc.clone = true
		default:
			// DataSource is not VolumeSnapshot and PVC
			// Assume external data populator to create the volume, and there is no more work for us to do
			p.eventRecorder.Event(claim, v1.EventTypeNormal, "Provisioning", fmt.Sprintf("Assuming an external populator will provision the volume"))
			return nil, ProvisioningFinished, &IgnoredError{
				Reason: fmt.Sprintf("data source (%s) is not handled by the provisioner, assuming an external populator will provision it",
					claim.Spec.DataSource.Kind),
			}
		}
	}
	if err := p.checkDriverCapabilities(rc); err != nil {
		return nil, ProvisioningFinished, err
	}

	if claim.Spec.Selector != nil {
		return nil, ProvisioningFinished, fmt.Errorf("claim Selector is not supported")
	}

	pvName, err := makeVolumeName(p.volumeNamePrefix, fmt.Sprintf("%s", claim.ObjectMeta.UID), p.volumeNameUUIDLength)
	if err != nil {
		return nil, ProvisioningFinished, err
	}

	fsTypesFound := 0
	fsType := ""
	for k, v := range sc.Parameters {
		if strings.ToLower(k) == "fstype" || k == prefixedFsTypeKey {
			fsType = v
			fsTypesFound++
		}
		if strings.ToLower(k) == "fstype" {
			klog.Warningf(deprecationWarning("fstype", prefixedFsTypeKey, ""))
		}
	}
	if fsTypesFound > 1 {
		return nil, ProvisioningFinished, fmt.Errorf("fstype specified in parameters with both \"fstype\" and \"%s\" keys", prefixedFsTypeKey)
	}
	if fsType == "" && p.defaultFSType != "" {
		fsType = p.defaultFSType
	}

	capacity := claim.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)]
	volSizeBytes := capacity.Value()

	// Get access mode
	volumeCaps := make([]*csi.VolumeCapability, 0)
	for _, pvcAccessMode := range claim.Spec.AccessModes {
		volumeCaps = append(volumeCaps, getVolumeCapability(claim, sc, pvcAccessMode, fsType))
	}

	// Create a CSI CreateVolumeRequest and Response
	req := csi.CreateVolumeRequest{
		Name:               pvName,
		Parameters:         sc.Parameters,
		VolumeCapabilities: volumeCaps,
		CapacityRange: &csi.CapacityRange{
			RequiredBytes: int64(volSizeBytes),
		},
	}

	if claim.Spec.DataSource != nil && (rc.clone || rc.snapshot) {
		volumeContentSource, err := p.getVolumeContentSource(ctx, claim, sc)
		if err != nil {
			return nil, ProvisioningNoChange, fmt.Errorf("error getting handle for DataSource Type %s by Name %s: %v", claim.Spec.DataSource.Kind, claim.Spec.DataSource.Name, err)
		}
		req.VolumeContentSource = volumeContentSource
	}

	if claim.Spec.DataSource != nil && rc.clone {
		err = p.setCloneFinalizer(ctx, claim)
		if err != nil {
			return nil, ProvisioningNoChange, err
		}
	}

	/*if p.supportsTopology() {
		requirements, err := GenerateAccessibilityRequirements(
			p.client,
			p.driverName,
			claim.Name,
			sc.AllowedTopologies,
			selectedNode,
			p.strictTopology,
			p.immediateTopology,
			p.csiNodeLister,
			p.nodeLister)
		if err != nil {
			return nil, ProvisioningNoChange, fmt.Errorf("error generating accessibility requirements: %v", err)
		}
		req.AccessibilityRequirements = requirements
	}*/

	// Resolve provision secret credentials.
	provisionerSecretRef, err := getSecretReference(provisionerSecretParams, sc.Parameters, pvName, &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      claim.Name,
			Namespace: claim.Namespace,
		},
	})
	if err != nil {
		return nil, ProvisioningNoChange, err
	}
	provisionerCredentials, err := getCredentials(ctx, p.client, provisionerSecretRef)
	if err != nil {
		return nil, ProvisioningNoChange, err
	}
	req.Secrets = provisionerCredentials

	// Resolve controller publish, node stage, node publish secret references
	controllerPublishSecretRef, err := getSecretReference(controllerPublishSecretParams, sc.Parameters, pvName, claim)
	if err != nil {
		return nil, ProvisioningNoChange, err
	}
	nodeStageSecretRef, err := getSecretReference(nodeStageSecretParams, sc.Parameters, pvName, claim)
	if err != nil {
		return nil, ProvisioningNoChange, err
	}
	nodePublishSecretRef, err := getSecretReference(nodePublishSecretParams, sc.Parameters, pvName, claim)
	if err != nil {
		return nil, ProvisioningNoChange, err
	}
	controllerExpandSecretRef, err := getSecretReference(controllerExpandSecretParams, sc.Parameters, pvName, claim)
	if err != nil {
		return nil, ProvisioningNoChange, err
	}
	csiPVSource := &v1.CSIPersistentVolumeSource{
		Driver: p.driverName,
		// VolumeHandle and VolumeAttributes will be added after provisioning.
		ControllerPublishSecretRef: controllerPublishSecretRef,
		NodeStageSecretRef:         nodeStageSecretRef,
		NodePublishSecretRef:       nodePublishSecretRef,
		ControllerExpandSecretRef:  controllerExpandSecretRef,
	}

	req.Parameters, err = removePrefixedParameters(sc.Parameters)
	if err != nil {
		return nil, ProvisioningFinished, fmt.Errorf("failed to strip CSI Parameters of prefixed keys: %v", err)
	}

	if p.extraCreateMetadata {
		// add pvc and pv metadata to request for use by the plugin
		req.Parameters[pvcNameKey] = claim.GetName()
		req.Parameters[pvcNamespaceKey] = claim.GetNamespace()
		req.Parameters[pvNameKey] = pvName
	}

	return &prepareProvisionResult{
		fsType:         fsType,
		migratedVolume: migratedVolume,
		req:            &req,
		csiPVSource:    csiPVSource,
	}, ProvisioningNoChange, nil
}

// TODO use a unique volume handle from and to Id
func (p *csiProvisioner) volumeIdToHandle(id string) string {
	return id
}

func (p *csiProvisioner) volumeHandleToId(handle string) string {
	return handle
}

func (p *csiProvisioner) Provision(ctx context.Context, options ProvisionOptions) (*v1.PersistentVolume, ProvisioningState, error) {
	claim := options.PVC
	if claim.Annotations[annStorageProvisioner] != p.driverName && claim.Annotations[annMigratedTo] != p.driverName {
		// The storage provisioner annotation may not equal driver name but the
		// PVC could have annotation "migrated-to" which is the new way to
		// signal a PVC is migrated (k8s v1.17+)
		return nil, ProvisioningFinished, &IgnoredError{
			Reason: fmt.Sprintf("PVC annotated with external-provisioner name %s does not match provisioner driver name %s. This could mean the PVC is not migrated",
				claim.Annotations[annStorageProvisioner],
				p.driverName),
		}
	}

	// The same check already ran in ShouldProvision, but perhaps
	// it couldn't complete due to some unexpected error.
	owned, err := p.checkNode(ctx, claim, options.StorageClass, "provision")
	if err != nil {
		return nil, ProvisioningNoChange,
			fmt.Errorf("node check failed: %v", err)
	}
	if !owned {
		return nil, ProvisioningNoChange, &IgnoredError{
			Reason: fmt.Sprintf("not responsible for provisioning of PVC %s/%s because it is not assigned to node %q",
				claim.Namespace, claim.Name, p.nodeDeployment.NodeName),
		}
	}

	result, state, err := p.prepareProvision(ctx, claim, options.StorageClass, options.SelectedNode)
	if result == nil {
		return nil, state, err
	}
	req := result.req
	volSizeBytes := req.CapacityRange.RequiredBytes
	pvName := req.Name
	provisionerCredentials := req.Secrets

	// rpc调用csi-driver，创建volume
	createCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()
	klog.V(5).Infof("CreateVolumeRequest %+v", req)
	rep, err := p.csiClient.CreateVolume(createCtx, req)
	if err != nil {
		// Giving up after an error and telling the pod scheduler to retry with a different node
		// only makes sense if:
		// - The CSI driver supports topology: without that, the next CreateVolume call after
		//   rescheduling will be exactly the same.
		// - We are working on a volume with late binding: only in that case will
		//   provisioning be retried if we give up for now.
		// - The error is one where rescheduling is
		//   a) allowed (i.e. we don't have to keep calling CreateVolume because the operation might be running) and
		//   b) it makes sense (typically local resource exhausted).
		//   isFinalError is going to check this.
		//
		// We do this regardless whether the driver has asked for strict topology because
		// even drivers which did not ask for it explicitly might still only look at the first
		// topology entry and thus succeed after rescheduling.
		//mayReschedule := p.supportsTopology() && options.SelectedNode != nil
		mayReschedule := options.SelectedNode != nil
		state := checkError(err, mayReschedule)
		klog.V(5).Infof("CreateVolume failed, supports topology = , node selected %v => may reschedule = %v => state = %v: %v",
			//p.supportsTopology(),
			options.SelectedNode != nil,
			mayReschedule,
			state,
			err)
		return nil, state, err
	}

	if rep.Volume != nil {
		klog.V(3).Infof("create volume rep: %+v", *rep.Volume)
	}
	volumeAttributes := map[string]string{provisionerIDKey: p.identity}
	for k, v := range rep.Volume.VolumeContext {
		volumeAttributes[k] = v
	}
	respCap := rep.GetVolume().GetCapacityBytes()
	//According to CSI spec CreateVolume should be able to return capacity = 0, which means it is unknown. for example NFS/FTP
	if respCap == 0 {
		respCap = volSizeBytes
		klog.V(3).Infof("csiClient response volume with size 0, which is not supported by apiServer, will use claim size:%d", respCap)
	} else if respCap < volSizeBytes {
		capErr := fmt.Errorf("created volume capacity %v less than requested capacity %v", respCap, volSizeBytes)
		delReq := &csi.DeleteVolumeRequest{
			VolumeId: rep.GetVolume().GetVolumeId(),
		}
		err = cleanupVolume(ctx, p, delReq, provisionerCredentials)
		if err != nil {
			capErr = fmt.Errorf("%v. Cleanup of volume %s failed, volume is orphaned: %v", capErr, pvName, err)
		}
		// use InBackground to retry the call, hoping the volume is deleted correctly next time.
		return nil, ProvisioningInBackground, capErr
	}

	if options.PVC.Spec.DataSource != nil {
		contentSource := rep.GetVolume().ContentSource
		if contentSource == nil {
			sourceErr := fmt.Errorf("volume content source missing")
			delReq := &csi.DeleteVolumeRequest{
				VolumeId: rep.GetVolume().GetVolumeId(),
			}
			err = cleanupVolume(ctx, p, delReq, provisionerCredentials)
			if err != nil {
				sourceErr = fmt.Errorf("%v. cleanup of volume %s failed, volume is orphaned: %v", sourceErr, pvName, err)
			}
			return nil, ProvisioningInBackground, sourceErr
		}
	}

	result.csiPVSource.VolumeHandle = p.volumeIdToHandle(rep.Volume.VolumeId)
	result.csiPVSource.VolumeAttributes = volumeAttributes
	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: pvName,
		},
		Spec: v1.PersistentVolumeSpec{
			AccessModes:  options.PVC.Spec.AccessModes,
			MountOptions: options.StorageClass.MountOptions,
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): bytesToQuantity(respCap),
			},
			// TODO wait for CSI VolumeSource API
			PersistentVolumeSource: v1.PersistentVolumeSource{
				CSI: result.csiPVSource,
			},
		},
	}

	if options.StorageClass.ReclaimPolicy != nil {
		pv.Spec.PersistentVolumeReclaimPolicy = *options.StorageClass.ReclaimPolicy
	}

	/*if p.supportsTopology() {
		pv.Spec.NodeAffinity = GenerateVolumeNodeAffinity(rep.Volume.AccessibleTopology)
	}*/

	// Set VolumeMode to PV if it is passed via PVC spec when Block feature is enabled
	if options.PVC.Spec.VolumeMode != nil {
		pv.Spec.VolumeMode = options.PVC.Spec.VolumeMode
	}
	// Set FSType if PV is not Block Volume
	if !util.CheckPersistentVolumeClaimModeBlock(options.PVC) {
		pv.Spec.PersistentVolumeSource.CSI.FSType = result.fsType
	}

	klog.V(2).Infof("successfully created PV %v for PVC %v and csi volume name %v", pv.Name, options.PVC.Name, pv.Spec.CSI.VolumeHandle)

	klog.V(5).Infof("successfully created PV %+v", pv.Spec.PersistentVolumeSource)
	return pv, ProvisioningFinished, nil
}

func (p *csiProvisioner) canDeleteVolume(volume *v1.PersistentVolume) error {
	if p.vaLister == nil {
		// Nothing to check.
		return nil
	}

	// Verify if volume is attached to a node before proceeding with deletion
	vaList, err := p.vaLister.List(labels.Everything())
	if err != nil {
		return fmt.Errorf("failed to list volumeattachments: %v", err)
	}

	for _, va := range vaList {
		if va.Spec.Source.PersistentVolumeName != nil && *va.Spec.Source.PersistentVolumeName == volume.Name {
			return fmt.Errorf("persistentvolume %s is still attached to node %s", volume.Name, va.Spec.NodeName)
		}
	}

	return nil
}

func (p *csiProvisioner) Delete(ctx context.Context, volume *v1.PersistentVolume) error {
	if volume == nil {
		return fmt.Errorf("invalid CSI PV")
	}

	var err error

	if volume.Spec.CSI == nil {
		return fmt.Errorf("invalid CSI PV")
	}

	// If we run on a single node, then we shouldn't delete volumes
	// that we didn't create. In practice, that means that the volume
	// is accessible (only!) on this node.
	if p.nodeDeployment != nil {
		accessible, err := VolumeIsAccessible(volume.Spec.NodeAffinity, p.nodeDeployment.NodeInfo.AccessibleTopology)
		if err != nil {
			return fmt.Errorf("checking volume affinity failed: %v", err)
		}
		if !accessible {
			return &IgnoredError{
				Reason: "PV was not provisioned on this node",
			}
		}
	}

	volumeId := p.volumeHandleToId(volume.Spec.CSI.VolumeHandle)

	rc := &requiredCapabilities{}
	if err := p.checkDriverCapabilities(rc); err != nil {
		return err
	}

	req := csi.DeleteVolumeRequest{
		VolumeId: volumeId,
	}
	// get secrets if StorageClass specifies it
	storageClassName := GetPersistentVolumeClass(volume)
	if len(storageClassName) != 0 {
		if storageClass, err := p.scLister.Get(storageClassName); err == nil {
			// Resolve provision secret credentials.
			provisionerSecretRef, err := getSecretReference(provisionerSecretParams, storageClass.Parameters, volume.Name, &v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:      volume.Spec.ClaimRef.Name,
					Namespace: volume.Spec.ClaimRef.Namespace,
				},
			})
			if err != nil {
				return fmt.Errorf("failed to get secretreference for volume %s: %v", volume.Name, err)
			}

			credentials, err := getCredentials(ctx, p.client, provisionerSecretRef)
			if err != nil {
				// Continue with deletion, as the secret may have already been deleted.
				klog.Errorf("Failed to get credentials for volume %s: %s", volume.Name, err.Error())
			}
			req.Secrets = credentials
		} else {
			klog.Warningf("failed to get storageclass: %s, proceeding to delete without secrets. %v", storageClassName, err)
		}
	}
	deleteCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	if err := p.canDeleteVolume(volume); err != nil {
		return err
	}

	_, err = p.csiClient.DeleteVolume(deleteCtx, &req)

	return err
}

// NewCSIProvisioner creates new CSI provisioner.
// vaLister is optional and only needed when VolumeAttachments are
// meant to be checked before deleting a volume.
func NewCSIProvisioner(client kubernetes.Interface,
	connectionTimeout time.Duration,
	identity string,
	volumeNamePrefix string,
	volumeNameUUIDLength int,
	grpcClient *grpc.ClientConn,
	snapshotClient snapclientset.Interface,
	driverName string,
	pluginCapabilities rpc.PluginCapabilitySet,
	controllerCapabilities rpc.ControllerCapabilitySet,
	strictTopology bool,
	immediateTopology bool,
	scLister storagelistersv1.StorageClassLister,
	csiNodeLister storagelistersv1.CSINodeLister,
	nodeLister corelisters.NodeLister,
	claimLister corelisters.PersistentVolumeClaimLister,
	vaLister storagelistersv1.VolumeAttachmentLister,
	extraCreateMetadata bool,
	defaultFSType string,
	nodeDeployment *NodeDeployment,
) Provisioner {
	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(klog.Infof)
	broadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: client.CoreV1().Events(v1.NamespaceAll)})
	eventRecorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: fmt.Sprintf("external-provisioner")})

	csiClient := csi.NewControllerClient(grpcClient)

	provisioner := &csiProvisioner{
		client:                 client,
		grpcClient:             grpcClient,
		csiClient:              csiClient,
		snapshotClient:         snapshotClient,
		timeout:                connectionTimeout,
		identity:               identity,
		volumeNamePrefix:       volumeNamePrefix,
		defaultFSType:          defaultFSType,
		volumeNameUUIDLength:   volumeNameUUIDLength,
		driverName:             driverName,
		pluginCapabilities:     pluginCapabilities,
		controllerCapabilities: controllerCapabilities,
		strictTopology:         strictTopology,
		immediateTopology:      immediateTopology,
		scLister:               scLister,
		csiNodeLister:          csiNodeLister,
		nodeLister:             nodeLister,
		claimLister:            claimLister,
		vaLister:               vaLister,
		extraCreateMetadata:    extraCreateMetadata,
		eventRecorder:          eventRecorder,
	}

	if nodeDeployment != nil {
		provisioner.nodeDeployment = &internalNodeDeployment{
			NodeDeployment: *nodeDeployment,
			rateLimiter:    newItemExponentialFailureRateLimiterWithJitter(nodeDeployment.BaseDelay, nodeDeployment.MaxDelay),
		}
		// Remove deleted PVCs from rate limiter.
		claimHandler := cache.ResourceEventHandlerFuncs{
			DeleteFunc: func(obj interface{}) {
				if claim, ok := obj.(*v1.PersistentVolumeClaim); ok {
					provisioner.nodeDeployment.rateLimiter.Forget(claim.UID)
				}
			},
		}
		provisioner.nodeDeployment.ClaimInformer.Informer().AddEventHandler(claimHandler)
	}

	return provisioner
}
