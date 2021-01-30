package hostpath

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	"github.com/google/uuid"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
)

const (
	deviceID           = "deviceID"
	maxStorageCapacity = tib
)

type accessType int

const (
	mountAccess accessType = iota
	blockAccess
)

type controllerServer struct {
	caps   []*csi.ControllerServiceCapability
	nodeID string
}

func (cs controllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	if err := cs.validateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
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

	// Keep a record of the requested access types.
	var accessTypeMount, accessTypeBlock bool
	for _, ca := range caps {
		if ca.GetBlock() != nil {
			accessTypeBlock = true
		}
		if ca.GetMount() != nil {
			accessTypeMount = true
		}
	}

	// A real driver would also need to check that the other
	// fields in VolumeCapabilities are sane. The check above is
	// just enough to pass the "[Testpattern: Dynamic PV (block
	// volmode)] volumeMode should fail in binding dynamic
	// provisioned PV to PVC" storage E2E test.

	if accessTypeBlock && accessTypeMount {
		return nil, status.Error(codes.InvalidArgument, "cannot have both block and mount access type")
	}

	var requestedAccessType accessType

	if accessTypeBlock {
		requestedAccessType = blockAccess
	} else {
		// Default to mount.
		requestedAccessType = mountAccess
	}

	// Check for maximum available capacity
	capacity := int64(req.GetCapacityRange().GetRequiredBytes())
	if capacity >= maxStorageCapacity {
		return nil, status.Errorf(codes.OutOfRange, "Requested capacity %d exceeds maximum allowed %d", capacity, maxStorageCapacity)
	}

	topologies := []*csi.Topology{
		&csi.Topology{
			Segments: map[string]string{TopologyKeyNode: cs.nodeID},
		},
	}

	// Need to check for already existing volume name, and if found
	// check for the requested capacity and already allocated capacity
	if exVol, err := getVolumeByName(req.GetName()); err == nil {
		// Since err is nil, it means the volume with the same name already exists
		// need to check if the size of existing volume is the same as in new
		// request
		if exVol.VolSize < capacity {
			return nil, status.Errorf(codes.AlreadyExists, "Volume with the same name: %s but with different size already exist", req.GetName())
		}
		if req.GetVolumeContentSource() != nil {
			volumeSource := req.VolumeContentSource
			switch volumeSource.Type.(type) {
			case *csi.VolumeContentSource_Snapshot:
				if volumeSource.GetSnapshot() != nil && exVol.ParentSnapID != "" && exVol.ParentSnapID != volumeSource.GetSnapshot().GetSnapshotId() {
					return nil, status.Error(codes.AlreadyExists, "existing volume source snapshot id not matching")
				}
			case *csi.VolumeContentSource_Volume:
				if volumeSource.GetVolume() != nil && exVol.ParentVolID != volumeSource.GetVolume().GetVolumeId() {
					return nil, status.Error(codes.AlreadyExists, "existing volume source volume id not matching")
				}
			default:
				return nil, status.Errorf(codes.InvalidArgument, "%v not a proper volume source", volumeSource)
			}
		}
		// TODO (sbezverk) Do I need to make sure that volume still exists?
		return &csi.CreateVolumeResponse{
			Volume: &csi.Volume{
				VolumeId:           exVol.VolID,
				CapacityBytes:      int64(exVol.VolSize),
				VolumeContext:      req.GetParameters(),
				ContentSource:      req.GetVolumeContentSource(),
				AccessibleTopology: topologies,
			},
		}, nil
	}

	volumeID := uuid.New().String()
	vol, err := createHostpathVolume(volumeID, req.GetName(), capacity, requestedAccessType, false /* ephemeral */)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create volume %v: %v", volumeID, err)
	}
	glog.V(4).Infof("created volume %s at path %s", vol.VolID, vol.VolPath)

	if req.GetVolumeContentSource() != nil {
		path := getVolumePath(volumeID)
		volumeSource := req.VolumeContentSource
		switch volumeSource.Type.(type) {
		case *csi.VolumeContentSource_Snapshot:
			if snapshot := volumeSource.GetSnapshot(); snapshot != nil {
				err = loadFromSnapshot(capacity, snapshot.GetSnapshotId(), path, requestedAccessType)
				vol.ParentSnapID = snapshot.GetSnapshotId()
			}
		case *csi.VolumeContentSource_Volume:
			if srcVolume := volumeSource.GetVolume(); srcVolume != nil {
				err = loadFromVolume(capacity, srcVolume.GetVolumeId(), path, requestedAccessType)
				vol.ParentVolID = srcVolume.GetVolumeId()
			}
		default:
			err = status.Errorf(codes.InvalidArgument, "%v not a proper volume source", volumeSource)
		}
		if err != nil {
			glog.V(4).Infof("VolumeSource error: %v", err)
			if delErr := deleteHostpathVolume(volumeID); delErr != nil {
				glog.V(2).Infof("deleting hostpath volume %v failed: %v", volumeID, delErr)
			}
			return nil, err
		}
		glog.V(4).Infof("successfully populated volume %s", vol.VolID)
	}

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:           volumeID,
			CapacityBytes:      req.GetCapacityRange().GetRequiredBytes(),
			VolumeContext:      req.GetParameters(),
			ContentSource:      req.GetVolumeContentSource(),
			AccessibleTopology: topologies,
		},
	}, nil

}

func (cs controllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	// Check arguments
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if err := cs.validateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		glog.V(3).Infof("invalid delete volume req: %v", req)
		return nil, err
	}

	volId := req.GetVolumeId()
	if err := deleteHostpathVolume(volId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete volume %v: %v", volId, err)
	}

	glog.V(4).Infof("volume %v successfully deleted", volId)

	return &csi.DeleteVolumeResponse{}, nil
}

func (cs controllerServer) ControllerPublishVolume(ctx context.Context, request *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	panic("implement me")
}

func (cs controllerServer) ControllerUnpublishVolume(ctx context.Context, request *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	panic("implement me")
}

func (cs controllerServer) ValidateVolumeCapabilities(ctx context.Context, request *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	panic("implement me")
}

func (cs controllerServer) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	volumeRes := &csi.ListVolumesResponse{
		Entries: []*csi.ListVolumesResponse_Entry{},
	}

	var (
		startIdx, volumesLength, maxLength int64
		hpVolume                           hostPathVolume
	)
	volumeIds := getSortedVolumeIDs()
	if req.StartingToken == "" {
		req.StartingToken = "1"
	}

	startIdx, err := strconv.ParseInt(req.StartingToken, 10, 32)
	if err != nil {
		return nil, status.Error(codes.Aborted, "The type of startingToken should be integer")
	}

	volumesLength = int64(len(volumeIds))
	maxLength = int64(req.MaxEntries)

	if maxLength > volumesLength || maxLength <= 0 {
		maxLength = volumesLength
	}

	for index := startIdx - 1; index < volumesLength && index < maxLength; index++ {
		hpVolume = hostPathVolumes[volumeIds[index]]
		healthy, msg := doHealthCheckInControllerSide(volumeIds[index])
		glog.V(3).Infof("Healthy state: %s Volume: %t", hpVolume.VolName, healthy)
		volumeRes.Entries = append(volumeRes.Entries, &csi.ListVolumesResponse_Entry{
			Volume: &csi.Volume{
				VolumeId:      hpVolume.VolID,
				CapacityBytes: hpVolume.VolSize,
			},
			Status: &csi.ListVolumesResponse_VolumeStatus{
				PublishedNodeIds: []string{hpVolume.NodeID},
				VolumeCondition: &csi.VolumeCondition{
					Abnormal: !healthy,
					Message:  msg,
				},
			},
		})
	}

	glog.V(5).Infof("Volumes are: %+v", *volumeRes)
	return volumeRes, nil
}

func (cs controllerServer) GetCapacity(ctx context.Context, request *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	panic("implement me")
}

func (cs controllerServer) ControllerGetCapabilities(ctx context.Context, request *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
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

func NewControllerServer(ephemeral bool, nodeID string) *controllerServer {
	if ephemeral {
		return &controllerServer{caps: getControllerServiceCapabilities(nil), nodeID: nodeID}
	}
	return &controllerServer{
		caps: getControllerServiceCapabilities(
			[]csi.ControllerServiceCapability_RPC_Type{
				csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
				csi.ControllerServiceCapability_RPC_GET_VOLUME,
				csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
				csi.ControllerServiceCapability_RPC_LIST_SNAPSHOTS,
				csi.ControllerServiceCapability_RPC_LIST_VOLUMES,
				csi.ControllerServiceCapability_RPC_CLONE_VOLUME,
				csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
				csi.ControllerServiceCapability_RPC_VOLUME_CONDITION,
			}),
		nodeID: nodeID,
	}
}

func getControllerServiceCapabilities(cl []csi.ControllerServiceCapability_RPC_Type) []*csi.ControllerServiceCapability {
	var csc []*csi.ControllerServiceCapability

	for _, ca := range cl {
		glog.Infof("Enabling controller service capability: %v", ca.String())
		csc = append(csc, &csi.ControllerServiceCapability{
			Type: &csi.ControllerServiceCapability_Rpc{
				Rpc: &csi.ControllerServiceCapability_RPC{
					Type: ca,
				},
			},
		})
	}

	return csc
}

func (cs *controllerServer) validateControllerServiceRequest(c csi.ControllerServiceCapability_RPC_Type) error {
	if c == csi.ControllerServiceCapability_RPC_UNKNOWN {
		return nil
	}

	for _, ca := range cs.caps {
		if c == ca.GetRpc().GetType() {
			return nil
		}
	}
	return status.Errorf(codes.InvalidArgument, "unsupported capability %s", c)
}
