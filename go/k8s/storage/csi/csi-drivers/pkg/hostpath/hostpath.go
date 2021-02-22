package hostpath

import (
	"fmt"
	csicommon "k8s-lx1036/k8s/storage/csi/csi-drivers/pkg/csi-common"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/protobuf/ptypes/timestamp"
	"k8s.io/klog/v2"
)

const (
	kib    int64 = 1024
	mib    int64 = kib * 1024
	gib    int64 = mib * 1024
	gib100 int64 = gib * 100
	tib    int64 = gib * 1024
	tib100 int64 = tib * 100
)

type hostPath struct {
	driver *csicommon.CSIDriver

	ids *identityServer
	ns  *nodeServer
	cs  *controllerServer

	cap   []*csi.VolumeCapability_AccessMode
	cscap []*csi.ControllerServiceCapability
}

type hostPathVolume struct {
	VolName string `json:"volName"`
	VolID   string `json:"volID"`
	VolSize int64  `json:"volSize"`
	VolPath string `json:"volPath"`
}

type hostPathSnapshot struct {
	Name         string               `json:"name"`
	Id           string               `json:"id"`
	VolID        string               `json:"volID"`
	Path         string               `json:"path"`
	CreationTime *timestamp.Timestamp `json:"creationTime"`
	SizeBytes    int64                `json:"sizeBytes"`
	ReadyToUse   bool                 `json:"readyToUse"`
}

var hostPathVolumes map[string]hostPathVolume
var hostPathVolumeSnapshots map[string]hostPathSnapshot

var (
	hostPathDriver *hostPath
	vendorVersion  = "dev"
)

func init() {
	hostPathVolumes = map[string]hostPathVolume{}
	hostPathVolumeSnapshots = map[string]hostPathSnapshot{}
}

func GetHostPathDriver() *hostPath {
	return &hostPath{}
}

func (hp *hostPath) Run(driverName, nodeID, endpoint string) {
	klog.Infof("Driver: %v ", driverName)
	klog.Infof("Version: %s", vendorVersion)

	// Initialize default library driver
	hp.driver = csicommon.NewCSIDriver(driverName, vendorVersion, nodeID)
	if hp.driver == nil {
		klog.Fatalln("Failed to initialize CSI Driver.")
	}
	hp.driver.AddControllerServiceCapabilities(
		[]csi.ControllerServiceCapability_RPC_Type{
			csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
			csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
			csi.ControllerServiceCapability_RPC_LIST_SNAPSHOTS,
		})
	hp.driver.AddVolumeCapabilityAccessModes(
		[]csi.VolumeCapability_AccessMode_Mode{
			csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER,
		})

	// Create GRPC servers
	hp.ids = NewIdentityServer(hp.driver)
	hp.ns = NewNodeServer(hp.driver)
	hp.cs = NewControllerServer(hp.driver)

	s := csicommon.NewNonBlockingGRPCServer()
	s.Start(endpoint, hp.ids, hp.cs, hp.ns)
}

func getVolumeByID(volumeID string) (hostPathVolume, error) {
	if hostPathVol, ok := hostPathVolumes[volumeID]; ok {
		return hostPathVol, nil
	}
	return hostPathVolume{}, fmt.Errorf("volume id %s does not exit in the volumes list", volumeID)
}

func getVolumeByName(volName string) (hostPathVolume, error) {
	for _, hostPathVol := range hostPathVolumes {
		if hostPathVol.VolName == volName {
			return hostPathVol, nil
		}
	}
	return hostPathVolume{}, fmt.Errorf("volume name %s does not exit in the volumes list", volName)
}

func getSnapshotByName(name string) (hostPathSnapshot, error) {
	for _, snapshot := range hostPathVolumeSnapshots {
		if snapshot.Name == name {
			return snapshot, nil
		}
	}
	return hostPathSnapshot{}, fmt.Errorf("snapshot name %s does not exit in the snapshots list", name)
}
