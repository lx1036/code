package hostpath

import (
	"fmt"
	"os"

	csicommon "k8s-lx1036/k8s/storage/csi/csi-drivers/pkg/csi-common"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"github.com/google/uuid"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	utilexec "k8s.io/utils/exec"
)

const (
	deviceID           = "deviceID"
	provisionRoot      = "/tmp/"
	snapshotRoot       = "/tmp/"
	maxStorageCapacity = tib
)

type controllerServer struct {
	*csicommon.DefaultControllerServer
}

func (cs controllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		glog.V(3).Infof("invalid create volume req: %v", req)
		return nil, err
	}

	// Check arguments
	if len(req.GetName()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Name missing in request")
	}
	caps := req.GetVolumeCapabilities()
	if caps == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume Capabilities missing in request")
	}
	for _, capability := range caps {
		if capability.GetBlock() != nil {
			return nil, status.Error(codes.Unimplemented, "Block Volume not supported")
		}
	}

	// A real driver would also need to check that the other
	// fields in VolumeCapabilities are sane. The check above is
	// just enough to pass the "[Testpattern: Dynamic PV (block
	// volmode)] volumeMode should fail in binding dynamic
	// provisioned PV to PVC" storage E2E test.

	// Need to check for already existing volume name, and if found
	// check for the requested capacity and already allocated capacity
	if exVol, err := getVolumeByName(req.GetName()); err == nil {
		// Since err is nil, it means the volume with the same name already exists
		// need to check if the size of exisiting volume is the same as in new
		// request
		if exVol.VolSize >= int64(req.GetCapacityRange().GetRequiredBytes()) {
			// exisiting volume is compatible with new request and should be reused.
			// TODO (sbezverk) Do I need to make sure that RBD volume still exists?
			return &csi.CreateVolumeResponse{
				Volume: &csi.Volume{
					VolumeId:      exVol.VolID,
					CapacityBytes: int64(exVol.VolSize),
					VolumeContext: req.GetParameters(),
				},
			}, nil
		}
		return nil, status.Error(codes.AlreadyExists, fmt.Sprintf("Volume with the same name: %s but with different size already exist", req.GetName()))
	}
	// Check for maximum available capacity
	capacity := int64(req.GetCapacityRange().GetRequiredBytes())
	if capacity >= maxStorageCapacity {
		return nil, status.Errorf(codes.OutOfRange, "Requested capacity %d exceeds maximum allowed %d", capacity, maxStorageCapacity)
	}

	volumeID := uuid.New().String()
	path := provisionRoot + volumeID
	err := os.MkdirAll(path, 0777)
	if err != nil {
		glog.V(3).Infof("failed to create volume: %v", err)
		return nil, err
	}

	if req.GetVolumeContentSource() != nil {
		contentSource := req.GetVolumeContentSource()
		if contentSource.GetSnapshot() != nil {
			snapshotId := contentSource.GetSnapshot().GetSnapshotId()
			snapshot, ok := hostPathVolumeSnapshots[snapshotId]
			if !ok {
				return nil, status.Errorf(codes.NotFound, "cannot find snapshot %v", snapshotId)
			}
			if snapshot.ReadyToUse != true {
				return nil, status.Errorf(codes.Internal, "Snapshot %v is not yet ready to use.", snapshotId)
			}
			snapshotPath := snapshot.Path
			args := []string{"zxvf", snapshotPath, "-C", path}
			out, err := utilexec.New().Command("tar", args...).CombinedOutput()
			if err != nil {
				return nil, status.Error(codes.Internal, fmt.Sprintf("failed pre-populate data for volume: %v: %s", err, out))
			}
		}
	}

	glog.V(4).Infof("create volume %s", path)
	hostPathVol := hostPathVolume{}
	hostPathVol.VolName = req.GetName()
	hostPathVol.VolID = volumeID
	hostPathVol.VolSize = capacity
	hostPathVol.VolPath = path
	hostPathVolumes[volumeID] = hostPathVol

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volumeID,
			CapacityBytes: req.GetCapacityRange().GetRequiredBytes(),
			VolumeContext: req.GetParameters(),
		},
	}, nil
}

func (cs controllerServer) DeleteVolume(ctx context.Context, request *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	panic("implement me")
}

func (cs controllerServer) ValidateVolumeCapabilities(ctx context.Context, request *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	panic("implement me")
}

func (cs controllerServer) ListVolumes(ctx context.Context, request *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	panic("implement me")
}

func (cs controllerServer) CreateSnapshot(ctx context.Context, request *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	panic("implement me")
}

func (cs controllerServer) DeleteSnapshot(ctx context.Context, request *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	panic("implement me")
}

func (cs controllerServer) ListSnapshots(ctx context.Context, request *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	panic("implement me")
}

func (cs controllerServer) ControllerExpandVolume(ctx context.Context, request *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	panic("implement me")
}

func (cs controllerServer) ControllerGetVolume(ctx context.Context, request *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	panic("implement me")
}

func NewControllerServer(d *csicommon.CSIDriver) *controllerServer {
	return &controllerServer{
		DefaultControllerServer: csicommon.NewDefaultControllerServer(d),
	}
}
