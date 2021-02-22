package controller

import (
	"context"
	"fmt"

	"os"
	"strings"
	"time"

	"k8s-lx1036/k8s/storage/csi/csi-lib-utils/connection"
	"k8s-lx1036/k8s/storage/csi/csi-lib-utils/metrics"
	"k8s-lx1036/k8s/storage/csi/csi-lib-utils/rpc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/kubernetes"
	corev1helpers "k8s.io/component-helpers/scheduling/corev1"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/volume/util"
)

const (
	deleteVolumeRetryCount = 5

	tokenPVNameKey       = "pv.name"
	tokenPVCNameKey      = "pvc.name"
	tokenPVCNameSpaceKey = "pvc.namespace"

	prefixedDefaultSecretNameKey      = csiParameterPrefix + "secret-name"
	prefixedDefaultSecretNamespaceKey = csiParameterPrefix + "secret-namespace"

	// [Deprecated] CSI Parameters that are put into fields but
	// NOT stripped from the parameters passed to CreateVolume
	provisionerSecretNameKey      = "csiProvisionerSecretName"
	provisionerSecretNamespaceKey = "csiProvisionerSecretNamespace"

	controllerPublishSecretNameKey      = "csiControllerPublishSecretName"
	controllerPublishSecretNamespaceKey = "csiControllerPublishSecretNamespace"

	nodeStageSecretNameKey      = "csiNodeStageSecretName"
	nodeStageSecretNamespaceKey = "csiNodeStageSecretNamespace"

	nodePublishSecretNameKey      = "csiNodePublishSecretName"
	nodePublishSecretNamespaceKey = "csiNodePublishSecretNamespace"

	// PV and PVC metadata, used for sending to drivers in the  create requests, added as parameters, optional.
	pvcNameKey      = "csi.storage.k8s.io/pvc/name"
	pvcNamespaceKey = "csi.storage.k8s.io/pvc/namespace"
	pvNameKey       = "csi.storage.k8s.io/pv/name"

	prefixedProvisionerSecretNameKey      = csiParameterPrefix + "provisioner-secret-name"
	prefixedProvisionerSecretNamespaceKey = csiParameterPrefix + "provisioner-secret-namespace"

	prefixedControllerPublishSecretNameKey      = csiParameterPrefix + "controller-publish-secret-name"
	prefixedControllerPublishSecretNamespaceKey = csiParameterPrefix + "controller-publish-secret-namespace"

	prefixedNodeStageSecretNameKey      = csiParameterPrefix + "node-stage-secret-name"
	prefixedNodeStageSecretNamespaceKey = csiParameterPrefix + "node-stage-secret-namespace"

	prefixedNodePublishSecretNameKey      = csiParameterPrefix + "node-publish-secret-name"
	prefixedNodePublishSecretNamespaceKey = csiParameterPrefix + "node-publish-secret-namespace"

	prefixedControllerExpandSecretNameKey      = csiParameterPrefix + "controller-expand-secret-name"
	prefixedControllerExpandSecretNamespaceKey = csiParameterPrefix + "controller-expand-secret-namespace"
)

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

func makeVolumeName(prefix, pvcUID string, volumeNameUUIDLength int) (string, error) {
	// create persistent name based on a volumeNamePrefix and volumeNameUUIDLength
	// of PVC's UID
	if len(prefix) == 0 {
		return "", fmt.Errorf("Volume name prefix cannot be of length 0")
	}
	if len(pvcUID) == 0 {
		return "", fmt.Errorf("corrupted PVC object, it is missing UID")
	}
	if volumeNameUUIDLength == -1 {
		// Default behavior is to not truncate or remove dashes
		return fmt.Sprintf("%s-%s", prefix, pvcUID), nil
	}

	// Else we remove all dashes from UUID and truncate to volumeNameUUIDLength
	return fmt.Sprintf("%s-%s", prefix, strings.ReplaceAll(pvcUID, "-", "")[0:volumeNameUUIDLength]), nil

}

func deprecationWarning(deprecatedParam, newParam, removalVersion string) string {
	if removalVersion == "" {
		removalVersion = "a future release"
	}
	newParamPhrase := ""
	if len(newParam) != 0 {
		newParamPhrase = fmt.Sprintf(`, please use "%s" instead`, newParam)
	}

	return fmt.Sprintf(`"%s" is deprecated and will be removed in %s%s`, deprecatedParam, removalVersion, newParamPhrase)
}

func getAccessTypeBlock() *csi.VolumeCapability_Block {
	return &csi.VolumeCapability_Block{
		Block: &csi.VolumeCapability_BlockVolume{},
	}
}

func getAccessTypeMount(fsType string, mountFlags []string) *csi.VolumeCapability_Mount {
	return &csi.VolumeCapability_Mount{
		Mount: &csi.VolumeCapability_MountVolume{
			FsType:     fsType,
			MountFlags: mountFlags,
		},
	}
}

func getAccessMode(pvcAccessMode v1.PersistentVolumeAccessMode) *csi.VolumeCapability_AccessMode {
	switch pvcAccessMode {
	case v1.ReadWriteOnce:
		return &csi.VolumeCapability_AccessMode{
			Mode: csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
		}
	case v1.ReadWriteMany:
		return &csi.VolumeCapability_AccessMode{
			Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
		}
	case v1.ReadOnlyMany:
		return &csi.VolumeCapability_AccessMode{
			Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY,
		}
	default:
		return nil
	}
}

func getVolumeCapability(
	claim *v1.PersistentVolumeClaim,
	sc *storagev1.StorageClass,
	pvcAccessMode v1.PersistentVolumeAccessMode,
	fsType string,
) *csi.VolumeCapability {
	if util.CheckPersistentVolumeClaimModeBlock(claim) {
		return &csi.VolumeCapability{
			AccessType: getAccessTypeBlock(),
			AccessMode: getAccessMode(pvcAccessMode),
		}
	}

	return &csi.VolumeCapability{
		AccessType: getAccessTypeMount(fsType, sc.MountOptions),
		AccessMode: getAccessMode(pvcAccessMode),
	}
}

//secretParamsMap provides a mapping of current as well as deprecated secret keys
type secretParamsMap struct {
	name                         string
	deprecatedSecretNameKey      string
	deprecatedSecretNamespaceKey string
	secretNameKey                string
	secretNamespaceKey           string
}

var (
	defaultSecretParams = secretParamsMap{
		name:               "Default",
		secretNameKey:      prefixedDefaultSecretNameKey,
		secretNamespaceKey: prefixedDefaultSecretNamespaceKey,
	}

	provisionerSecretParams = secretParamsMap{
		name:                         "Provisioner",
		deprecatedSecretNameKey:      provisionerSecretNameKey,
		deprecatedSecretNamespaceKey: provisionerSecretNamespaceKey,
		secretNameKey:                prefixedProvisionerSecretNameKey,
		secretNamespaceKey:           prefixedProvisionerSecretNamespaceKey,
	}

	nodePublishSecretParams = secretParamsMap{
		name:                         "NodePublish",
		deprecatedSecretNameKey:      nodePublishSecretNameKey,
		deprecatedSecretNamespaceKey: nodePublishSecretNamespaceKey,
		secretNameKey:                prefixedNodePublishSecretNameKey,
		secretNamespaceKey:           prefixedNodePublishSecretNamespaceKey,
	}

	controllerPublishSecretParams = secretParamsMap{
		name:                         "ControllerPublish",
		deprecatedSecretNameKey:      controllerPublishSecretNameKey,
		deprecatedSecretNamespaceKey: controllerPublishSecretNamespaceKey,
		secretNameKey:                prefixedControllerPublishSecretNameKey,
		secretNamespaceKey:           prefixedControllerPublishSecretNamespaceKey,
	}

	nodeStageSecretParams = secretParamsMap{
		name:                         "NodeStage",
		deprecatedSecretNameKey:      nodeStageSecretNameKey,
		deprecatedSecretNamespaceKey: nodeStageSecretNamespaceKey,
		secretNameKey:                prefixedNodeStageSecretNameKey,
		secretNamespaceKey:           prefixedNodeStageSecretNamespaceKey,
	}

	controllerExpandSecretParams = secretParamsMap{
		name:               "ControllerExpand",
		secretNameKey:      prefixedControllerExpandSecretNameKey,
		secretNamespaceKey: prefixedControllerExpandSecretNamespaceKey,
	}
)

// verifyAndGetSecretNameAndNamespaceTemplate gets the values (templates) associated
// with the parameters specified in "secret" and verifies that they are specified correctly.
func verifyAndGetSecretNameAndNamespaceTemplate(secret secretParamsMap, storageClassParams map[string]string) (nameTemplate, namespaceTemplate string, err error) {
	numName := 0
	numNamespace := 0

	if t, ok := storageClassParams[secret.deprecatedSecretNameKey]; ok {
		nameTemplate = t
		numName++
		klog.Warning(deprecationWarning(secret.deprecatedSecretNameKey, secret.secretNameKey, ""))
	}
	if t, ok := storageClassParams[secret.deprecatedSecretNamespaceKey]; ok {
		namespaceTemplate = t
		numNamespace++
		klog.Warning(deprecationWarning(secret.deprecatedSecretNamespaceKey, secret.secretNamespaceKey, ""))
	}
	if t, ok := storageClassParams[secret.secretNameKey]; ok {
		nameTemplate = t
		numName++
	}
	if t, ok := storageClassParams[secret.secretNamespaceKey]; ok {
		namespaceTemplate = t
		numNamespace++
	}

	if numName > 1 || numNamespace > 1 {
		// Double specified error
		return "", "", fmt.Errorf("%s secrets specified in parameters with both \"csi\" and \"%s\" keys", secret.name, csiParameterPrefix)
	} else if numName != numNamespace {
		// Not both 0 or both 1
		return "", "", fmt.Errorf("either name and namespace for %s secrets specified, Both must be specified", secret.name)
	} else if numName == 1 {
		// Case where we've found a name and a namespace template
		if nameTemplate == "" || namespaceTemplate == "" {
			return "", "", fmt.Errorf("%s secrets specified in parameters but value of either namespace or name is empty", secret.name)
		}
		return nameTemplate, namespaceTemplate, nil
	} else if numName == 0 {
		// No secrets specified
		return "", "", nil
	} else {
		// THIS IS NOT A VALID CASE
		return "", "", fmt.Errorf("unknown error with getting secret name and namespace templates")
	}
}

// getSecretReference returns a reference to the secret specified in the given nameTemplate
//  and namespaceTemplate, or an error if the templates are not specified correctly.
// no lookup of the referenced secret is performed, and the secret may or may not exist.
//
// supported tokens for name resolution:
// - ${pv.name}
// - ${pvc.namespace}
// - ${pvc.name}
// - ${pvc.annotations['ANNOTATION_KEY']} (e.g. ${pvc.annotations['example.com/node-publish-secret-name']})
//
// supported tokens for namespace resolution:
// - ${pv.name}
// - ${pvc.namespace}
//
// an error is returned in the following situations:
// - the nameTemplate or namespaceTemplate contains a token that cannot be resolved
// - the resolved name is not a valid secret name
// - the resolved namespace is not a valid namespace name
func getSecretReference(secretParams secretParamsMap, storageClassParams map[string]string, pvName string, pvc *v1.PersistentVolumeClaim) (*v1.SecretReference, error) {
	nameTemplate, namespaceTemplate, err := verifyAndGetSecretNameAndNamespaceTemplate(secretParams, storageClassParams)
	if err != nil {
		return nil, fmt.Errorf("failed to get name and namespace template from params: %v", err)
	}

	// if didn't find secrets for specific call, try to check default values
	if nameTemplate == "" && namespaceTemplate == "" {
		nameTemplate, namespaceTemplate, err = verifyAndGetSecretNameAndNamespaceTemplate(defaultSecretParams, storageClassParams)
		if err != nil {
			return nil, fmt.Errorf("failed to get default name and namespace template from params: %v", err)
		}
	}

	if nameTemplate == "" && namespaceTemplate == "" {
		return nil, nil
	}

	ref := &v1.SecretReference{}
	{
		// Secret namespace template can make use of the PV name or the PVC namespace.
		// Note that neither of those things are under the control of the PVC user.
		namespaceParams := map[string]string{tokenPVNameKey: pvName}
		if pvc != nil {
			namespaceParams[tokenPVCNameSpaceKey] = pvc.Namespace
		}

		resolvedNamespace, err := resolveTemplate(namespaceTemplate, namespaceParams)
		if err != nil {
			return nil, fmt.Errorf("error resolving value %q: %v", namespaceTemplate, err)
		}
		if len(validation.IsDNS1123Label(resolvedNamespace)) > 0 {
			if namespaceTemplate != resolvedNamespace {
				return nil, fmt.Errorf("%q resolved to %q which is not a valid namespace name", namespaceTemplate, resolvedNamespace)
			}
			return nil, fmt.Errorf("%q is not a valid namespace name", namespaceTemplate)
		}
		ref.Namespace = resolvedNamespace
	}

	{
		// Secret name template can make use of the PV name, PVC name or namespace, or a PVC annotation.
		// Note that PVC name and annotations are under the PVC user's control.
		nameParams := map[string]string{tokenPVNameKey: pvName}
		if pvc != nil {
			nameParams[tokenPVCNameKey] = pvc.Name
			nameParams[tokenPVCNameSpaceKey] = pvc.Namespace
			for k, v := range pvc.Annotations {
				nameParams["pvc.annotations['"+k+"']"] = v
			}
		}
		resolvedName, err := resolveTemplate(nameTemplate, nameParams)
		if err != nil {
			return nil, fmt.Errorf("error resolving value %q: %v", nameTemplate, err)
		}
		if len(validation.IsDNS1123Subdomain(resolvedName)) > 0 {
			if nameTemplate != resolvedName {
				return nil, fmt.Errorf("%q resolved to %q which is not a valid secret name", nameTemplate, resolvedName)
			}
			return nil, fmt.Errorf("%q is not a valid secret name", nameTemplate)
		}
		ref.Name = resolvedName
	}

	return ref, nil
}

func resolveTemplate(template string, params map[string]string) (string, error) {
	missingParams := sets.NewString()
	resolved := os.Expand(template, func(k string) string {
		v, ok := params[k]
		if !ok {
			missingParams.Insert(k)
		}
		return v
	})
	if missingParams.Len() > 0 {
		return "", fmt.Errorf("invalid tokens: %q", missingParams.List())
	}
	return resolved, nil
}

func getCredentials(ctx context.Context, k8s kubernetes.Interface, ref *v1.SecretReference) (map[string]string, error) {
	if ref == nil {
		return nil, nil
	}

	secret, err := k8s.CoreV1().Secrets(ref.Namespace).Get(ctx, ref.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting secret %s in namespace %s: %v", ref.Name, ref.Namespace, err)
	}

	credentials := map[string]string{}
	for key, value := range secret.Data {
		credentials[key] = string(value)
	}
	return credentials, nil
}

func removePrefixedParameters(param map[string]string) (map[string]string, error) {
	newParam := map[string]string{}
	for k, v := range param {
		if strings.HasPrefix(k, csiParameterPrefix) {
			// Check if its well known
			switch k {
			case prefixedFsTypeKey:
			case prefixedProvisionerSecretNameKey:
			case prefixedProvisionerSecretNamespaceKey:
			case prefixedControllerPublishSecretNameKey:
			case prefixedControllerPublishSecretNamespaceKey:
			case prefixedNodeStageSecretNameKey:
			case prefixedNodeStageSecretNamespaceKey:
			case prefixedNodePublishSecretNameKey:
			case prefixedNodePublishSecretNamespaceKey:
			case prefixedControllerExpandSecretNameKey:
			case prefixedControllerExpandSecretNamespaceKey:
			case prefixedDefaultSecretNameKey:
			case prefixedDefaultSecretNamespaceKey:
			default:
				return map[string]string{}, fmt.Errorf("found unknown parameter key \"%s\" with reserved namespace %s", k, csiParameterPrefix)
			}
		} else {
			// Don't strip, add this key-value to new map
			// Deprecated parameters prefixed with "csi" are not stripped to preserve backwards compatibility
			newParam[k] = v
		}
	}
	return newParam, nil
}

func checkFinalizer(obj metav1.Object, finalizer string) bool {
	for _, f := range obj.GetFinalizers() {
		if f == finalizer {
			return true
		}
	}
	return false
}

func checkError(err error, mayReschedule bool) ProvisioningState {
	// Sources:
	// https://github.com/grpc/grpc/blob/master/doc/statuscodes.md
	// https://github.com/container-storage-interface/spec/blob/master/spec.md
	st, ok := status.FromError(err)
	if !ok {
		// This is not gRPC error. The operation must have failed before gRPC
		// method was called, otherwise we would get gRPC error.
		// We don't know if any previous CreateVolume is in progress, be on the safe side.
		return ProvisioningInBackground
	}
	switch st.Code() {
	case codes.ResourceExhausted:
		// CSI: operation not pending, "Unable to provision in `accessible_topology`"
		// However, it also could be from the transport layer for "message size exceeded".
		// Cannot be decided properly here and needs to be resolved in the spec
		// https://github.com/container-storage-interface/spec/issues/419.
		// What we assume here for now is that message size limits are large enough that
		// the error really comes from the CSI driver.
		if mayReschedule {
			// may succeed elsewhere -> give up for now
			return ProvisioningReschedule
		}
		// may still succeed at a later time -> continue
		return ProvisioningInBackground
	case codes.Canceled, // gRPC: Client Application cancelled the request
		codes.DeadlineExceeded, // gRPC: Timeout
		codes.Unavailable,      // gRPC: Server shutting down, TCP connection broken - previous CreateVolume() may be still in progress.
		codes.Aborted:          // CSI: Operation pending for volume
		return ProvisioningInBackground
	}
	// All other errors mean that provisioning either did not
	// even start or failed. It is for sure not in progress.
	return ProvisioningFinished
}

func bytesToQuantity(bytes int64) resource.Quantity {
	quantity := resource.NewQuantity(bytes, resource.BinarySI)
	return *quantity
}

// GetPersistentVolumeClass returns StorageClassName.
func GetPersistentVolumeClass(volume *v1.PersistentVolume) string {
	// Use beta annotation first
	if class, found := volume.Annotations[v1.BetaStorageClassAnnotation]; found {
		return class
	}

	return volume.Spec.StorageClassName
}

// VolumeIsAccessible checks whether the generated volume affinity is satisfied by
// a the node topology that a CSI driver reported in GetNodeInfoResponse.
func VolumeIsAccessible(affinity *v1.VolumeNodeAffinity, nodeTopology *csi.Topology) (bool, error) {
	if nodeTopology == nil || affinity == nil || affinity.Required == nil {
		// No topology information -> all volumes accessible.
		return true, nil
	}

	nodeLabels := labels.Set{}
	for k, v := range nodeTopology.Segments {
		nodeLabels[k] = v
	}
	node := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Labels: nodeLabels,
		},
	}
	return corev1helpers.MatchNodeSelectorTerms(&node, affinity.Required)
}
