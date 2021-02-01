package controller

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/kubernetes/pkg/volume/util"
	"time"

	"k8s-lx1036/k8s/storage/csi/csi-lib-utils/connection"
	"k8s-lx1036/k8s/storage/csi/csi-lib-utils/metrics"
	"k8s-lx1036/k8s/storage/csi/csi-lib-utils/rpc"
	"k8s-lx1036/k8s/storage/csi/external-provisioner/external-provisioner-lib/pkg/controller"

	"google.golang.org/grpc"

	"github.com/container-storage-interface/spec/lib/go/csi"
	snapclientset "github.com/kubernetes-csi/external-snapshotter/client/v3/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	storagelistersv1 "k8s.io/client-go/listers/storage/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
)

const (
	ResyncPeriodOfCsiNodeInformer = 1 * time.Hour
)

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

func Connect(address string, metricsManager metrics.CSIMetricsManager) (*grpc.ClientConn, error) {
	return connection.Connect(address, metricsManager, connection.OnConnectionLoss(connection.ExitOnConnectionLoss()))
}

func Probe(conn *grpc.ClientConn, singleCallTimeout time.Duration) error {
	return rpc.ProbeForever(conn, singleCallTimeout)
}

func GetDriverName(conn *grpc.ClientConn, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return rpc.GetDriverName(ctx, conn)
}

func GetNodeInfo(conn *grpc.ClientConn, timeout time.Duration) (*csi.NodeGetInfoResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	client := csi.NewNodeClient(conn)
	return client.NodeGetInfo(ctx, &csi.NodeGetInfoRequest{})
}

func GetDriverCapabilities(conn *grpc.ClientConn, timeout time.Duration) (rpc.PluginCapabilitySet, rpc.ControllerCapabilitySet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	pluginCapabilities, err := rpc.GetPluginCapabilities(ctx, conn)
	if err != nil {
		return nil, nil, err
	}

	/* Each CSI operation gets its own timeout / context */
	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()
	controllerCapabilities, err := rpc.GetControllerCapabilities(ctx, conn)
	if err != nil {
		return nil, nil, err
	}

	return pluginCapabilities, controllerCapabilities, nil
}

type csiProvisioner struct {
	client                                kubernetes.Interface
	csiClient                             csi.ControllerClient
	grpcClient                            *grpc.ClientConn
	snapshotClient                        snapclientset.Interface
	timeout                               time.Duration
	identity                              string
	volumeNamePrefix                      string
	defaultFSType                         string
	volumeNameUUIDLength                  int
	config                                *rest.Config
	driverName                            string
	pluginCapabilities                    rpc.PluginCapabilitySet
	controllerCapabilities                rpc.ControllerCapabilitySet
	supportsMigrationFromInTreePluginName string
	strictTopology                        bool
	immediateTopology                     bool
	translator                            ProvisionerCSITranslator
	scLister                              storagelistersv1.StorageClassLister
	csiNodeLister                         storagelistersv1.CSINodeLister
	nodeLister                            corelisters.NodeLister
	claimLister                           corelisters.PersistentVolumeClaimLister
	vaLister                              storagelistersv1.VolumeAttachmentLister
	extraCreateMetadata                   bool
	eventRecorder                         record.EventRecorder
	nodeDeployment                        *internalNodeDeployment
}

func (c csiProvisioner) Provision(ctx context.Context, options controller.ProvisionOptions) (*v1.PersistentVolume, controller.ProvisioningState, error) {
	claim := options.PVC
	if claim.Annotations[annStorageProvisioner] != c.driverName && claim.Annotations[annMigratedTo] != c.driverName {
		// The storage provisioner annotation may not equal driver name but the
		// PVC could have annotation "migrated-to" which is the new way to
		// signal a PVC is migrated (k8s v1.17+)
		return nil, controller.ProvisioningFinished, &controller.IgnoredError{
			Reason: fmt.Sprintf("PVC annotated with external-provisioner name %s does not match provisioner driver name %s. This could mean the PVC is not migrated",
				claim.Annotations[annStorageProvisioner],
				c.driverName),
		}
	}

	// The same check already ran in ShouldProvision, but perhaps
	// it couldn't complete due to some unexpected error.
	owned, err := c.checkNode(ctx, claim, options.StorageClass, "provision")
	if err != nil {
		return nil, controller.ProvisioningNoChange,
			fmt.Errorf("node check failed: %v", err)
	}
	if !owned {
		return nil, controller.ProvisioningNoChange, &controller.IgnoredError{
			Reason: fmt.Sprintf("not responsible for provisioning of PVC %s/%s because it is not assigned to node %q", claim.Namespace, claim.Name, p.nodeDeployment.NodeName),
		}
	}

	result, state, err := c.prepareProvision(ctx, claim, options.StorageClass, options.SelectedNode)
	if result == nil {
		return nil, state, err
	}
	req := result.req
	volSizeBytes := req.CapacityRange.RequiredBytes
	pvName := req.Name
	provisionerCredentials := req.Secrets

	createCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	klog.V(5).Infof("CreateVolumeRequest %+v", req)
	rep, err := c.csiClient.CreateVolume(createCtx, req)
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
		mayReschedule := c.supportsTopology() &&
			options.SelectedNode != nil
		state := checkError(err, mayReschedule)
		klog.V(5).Infof("CreateVolume failed, supports topology = %v, node selected %v => may reschedule = %v => state = %v: %v",
			c.supportsTopology(),
			options.SelectedNode != nil,
			mayReschedule,
			state,
			err)
		return nil, state, err
	}

	if rep.Volume != nil {
		klog.V(3).Infof("create volume rep: %+v", *rep.Volume)
	}
	volumeAttributes := map[string]string{provisionerIDKey: c.identity}
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
		err = cleanupVolume(ctx, c, delReq, provisionerCredentials)
		if err != nil {
			capErr = fmt.Errorf("%v. Cleanup of volume %s failed, volume is orphaned: %v", capErr, pvName, err)
		}
		// use InBackground to retry the call, hoping the volume is deleted correctly next time.
		return nil, controller.ProvisioningInBackground, capErr
	}

	if options.PVC.Spec.DataSource != nil {
		contentSource := rep.GetVolume().ContentSource
		if contentSource == nil {
			sourceErr := fmt.Errorf("volume content source missing")
			delReq := &csi.DeleteVolumeRequest{
				VolumeId: rep.GetVolume().GetVolumeId(),
			}
			err = cleanupVolume(ctx, c, delReq, provisionerCredentials)
			if err != nil {
				sourceErr = fmt.Errorf("%v. cleanup of volume %s failed, volume is orphaned: %v", sourceErr, pvName, err)
			}
			return nil, controller.ProvisioningInBackground, sourceErr
		}
	}

	if options.PVC.Spec.DataSource != nil {
		contentSource := rep.GetVolume().ContentSource
		if contentSource == nil {
			sourceErr := fmt.Errorf("volume content source missing")
			delReq := &csi.DeleteVolumeRequest{
				VolumeId: rep.GetVolume().GetVolumeId(),
			}
			err = cleanupVolume(ctx, c, delReq, provisionerCredentials)
			if err != nil {
				sourceErr = fmt.Errorf("%v. cleanup of volume %s failed, volume is orphaned: %v", sourceErr, pvName, err)
			}
			return nil, controller.ProvisioningInBackground, sourceErr
		}
	}

	result.csiPVSource.VolumeHandle = c.volumeIdToHandle(rep.Volume.VolumeId)
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

	if c.supportsTopology() {
		pv.Spec.NodeAffinity = GenerateVolumeNodeAffinity(rep.Volume.AccessibleTopology)
	}

	// Set VolumeMode to PV if it is passed via PVC spec when Block feature is enabled
	if options.PVC.Spec.VolumeMode != nil {
		pv.Spec.VolumeMode = options.PVC.Spec.VolumeMode
	}
	// Set FSType if PV is not Block Volume
	if !util.CheckPersistentVolumeClaimModeBlock(options.PVC) {
		pv.Spec.PersistentVolumeSource.CSI.FSType = result.fsType
	}

	klog.V(2).Infof("successfully created PV %v for PVC %v and csi volume name %v", pv.Name, options.PVC.Name, pv.Spec.CSI.VolumeHandle)

	if result.migratedVolume {
		pv, err = c.translator.TranslateCSIPVToInTree(pv)
		if err != nil {
			klog.Warningf("failed to translate CSI PV to in-tree due to: %v. Deleting provisioned PV", err)
			deleteErr := c.Delete(ctx, pv)
			if deleteErr != nil {
				klog.Warningf("failed to delete partly provisioned PV: %v", deleteErr)
				// Retry the call again to clean up the orphan
				return nil, controller.ProvisioningInBackground, err
			}
			return nil, controller.ProvisioningFinished, err
		}
	}

	klog.V(5).Infof("successfully created PV %+v", pv.Spec.PersistentVolumeSource)
	return pv, controller.ProvisioningFinished, nil
}

func cleanupVolume(ctx context.Context, p *csiProvisioner, delReq *csi.DeleteVolumeRequest, provisionerCredentials map[string]string) error {
	var err error
	delReq.Secrets = provisionerCredentials
	deleteCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()
	for i := 0; i < deleteVolumeRetryCount; i++ {
		_, err = p.csiClient.DeleteVolume(deleteCtx, delReq)
		if err == nil {
			break
		}
	}
	return err
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

func (c csiProvisioner) Delete(ctx context.Context, volume *v1.PersistentVolume) error {
	if volume == nil {
		return fmt.Errorf("invalid CSI PV")
	}

	var err error
	if c.translator.IsPVMigratable(volume) {
		// we end up here only if CSI migration is enabled in-tree (both overall
		// and for the specific plugin that is migratable) causing in-tree PV
		// controller to yield deletion of PVs with in-tree source to external provisioner
		// based on AnnDynamicallyProvisioned annotation.
		volume, err = c.translator.TranslateInTreePVToCSI(volume)
		if err != nil {
			return err
		}
	}

	if volume.Spec.CSI == nil {
		return fmt.Errorf("invalid CSI PV")
	}

	// If we run on a single node, then we shouldn't delete volumes
	// that we didn't create. In practice, that means that the volume
	// is accessible (only!) on this node.
	if c.nodeDeployment != nil {
		accessible, err := VolumeIsAccessible(volume.Spec.NodeAffinity, c.nodeDeployment.NodeInfo.AccessibleTopology)
		if err != nil {
			return fmt.Errorf("checking volume affinity failed: %v", err)
		}
		if !accessible {
			return &controller.IgnoredError{
				Reason: "PV was not provisioned on this node",
			}
		}
	}

	volumeId := c.volumeHandleToId(volume.Spec.CSI.VolumeHandle)

	rc := &requiredCapabilities{}
	if err := c.checkDriverCapabilities(rc); err != nil {
		return err
	}

	req := csi.DeleteVolumeRequest{
		VolumeId: volumeId,
	}
	// get secrets if StorageClass specifies it
	storageClassName := util.GetPersistentVolumeClass(volume)
	if len(storageClassName) != 0 {
		if storageClass, err := c.scLister.Get(storageClassName); err == nil {
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

			credentials, err := getCredentials(ctx, c.client, provisionerSecretRef)
			if err != nil {
				// Continue with deletion, as the secret may have already been deleted.
				klog.Errorf("Failed to get credentials for volume %s: %s", volume.Name, err.Error())
			}
			req.Secrets = credentials
		} else {
			klog.Warningf("failed to get storageclass: %s, proceeding to delete without secrets. %v", storageClassName, err)
		}
	}
	deleteCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	if err := c.canDeleteVolume(volume); err != nil {
		return err
	}

	_, err = c.csiClient.DeleteVolume(deleteCtx, &req)

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
	supportsMigrationFromInTreePluginName string,
	strictTopology bool,
	immediateTopology bool,
	translator ProvisionerCSITranslator,
	scLister storagelistersv1.StorageClassLister,
	csiNodeLister storagelistersv1.CSINodeLister,
	nodeLister corelisters.NodeLister,
	claimLister corelisters.PersistentVolumeClaimLister,
	vaLister storagelistersv1.VolumeAttachmentLister,
	extraCreateMetadata bool,
	defaultFSType string,
	nodeDeployment *NodeDeployment,
) controller.Provisioner {
	broadcaster := record.NewBroadcaster()
	broadcaster.StartLogging(klog.Infof)
	broadcaster.StartRecordingToSink(&corev1.EventSinkImpl{Interface: client.CoreV1().Events(v1.NamespaceAll)})
	eventRecorder := broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: fmt.Sprintf("external-provisioner")})

	csiClient := csi.NewControllerClient(grpcClient)

	provisioner := &csiProvisioner{
		client:                                client,
		grpcClient:                            grpcClient,
		csiClient:                             csiClient,
		snapshotClient:                        snapshotClient,
		timeout:                               connectionTimeout,
		identity:                              identity,
		volumeNamePrefix:                      volumeNamePrefix,
		defaultFSType:                         defaultFSType,
		volumeNameUUIDLength:                  volumeNameUUIDLength,
		driverName:                            driverName,
		pluginCapabilities:                    pluginCapabilities,
		controllerCapabilities:                controllerCapabilities,
		supportsMigrationFromInTreePluginName: supportsMigrationFromInTreePluginName,
		strictTopology:                        strictTopology,
		immediateTopology:                     immediateTopology,
		translator:                            translator,
		scLister:                              scLister,
		csiNodeLister:                         csiNodeLister,
		nodeLister:                            nodeLister,
		claimLister:                           claimLister,
		vaLister:                              vaLister,
		extraCreateMetadata:                   extraCreateMetadata,
		eventRecorder:                         eventRecorder,
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
