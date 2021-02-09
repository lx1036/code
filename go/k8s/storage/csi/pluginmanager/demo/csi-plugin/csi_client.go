package csi_plugin

import (
	"context"
	"fmt"
	"io"
	"net"

	csipbv1 "github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/volume"
)

type csiClient interface {
	NodeGetInfo(ctx context.Context) (
		nodeID string,
		maxVolumePerNode int64,
		accessibleTopology map[string]string,
		err error)
	NodePublishVolume(
		ctx context.Context,
		volumeid string,
		readOnly bool,
		stagingTargetPath string,
		targetPath string,
		accessMode v1.PersistentVolumeAccessMode,
		publishContext map[string]string,
		volumeContext map[string]string,
		secrets map[string]string,
		fsType string,
		mountOptions []string,
	) error
	NodeExpandVolume(ctx context.Context, volumeid, volumePath string, newSize resource.Quantity) (resource.Quantity, error)
	NodeUnpublishVolume(
		ctx context.Context,
		volID string,
		targetPath string,
	) error
	NodeStageVolume(ctx context.Context,
		volID string,
		publishVolumeInfo map[string]string,
		stagingTargetPath string,
		fsType string,
		accessMode v1.PersistentVolumeAccessMode,
		secrets map[string]string,
		volumeContext map[string]string,
		mountOptions []string,
	) error

	NodeGetVolumeStats(
		ctx context.Context,
		volID string,
		targetPath string,
	) (*volume.Metrics, error)
	NodeUnstageVolume(ctx context.Context, volID, stagingTargetPath string) error
	NodeSupportsStageUnstage(ctx context.Context) (bool, error)
	NodeSupportsNodeExpand(ctx context.Context) (bool, error)
	NodeSupportsVolumeStats(ctx context.Context) (bool, error)
}

// Strongly typed address
type csiAddr string

// Strongly typed driver name
type csiDriverName string

type nodeV1ClientCreator func(addr csiAddr) (
	nodeClient csipbv1.NodeClient,
	closer io.Closer,
	err error,
)

// csiClient encapsulates all csi-plugin methods
type csiDriverClient struct {
	driverName          csiDriverName
	addr                csiAddr
	nodeV1ClientCreator nodeV1ClientCreator
}

func (c *csiDriverClient) NodeGetInfo(ctx context.Context) (
	nodeID string,
	maxVolumePerNode int64,
	accessibleTopology map[string]string,
	err error) {
	klog.Info("calling NodeGetInfo rpc")

	var getNodeInfoError error
	nodeID, maxVolumePerNode, accessibleTopology, getNodeInfoError = c.nodeGetInfoV1(ctx)
	if getNodeInfoError != nil {
		klog.Warningf("Error calling CSI NodeGetInfo(): %v", getNodeInfoError.Error())
	}
	return nodeID, maxVolumePerNode, accessibleTopology, getNodeInfoError
}

func (c *csiDriverClient) nodeGetInfoV1(ctx context.Context) (
	nodeID string,
	maxVolumePerNode int64,
	accessibleTopology map[string]string,
	err error) {
	nodeClient, closer, err := c.nodeV1ClientCreator(c.addr)
	if err != nil {
		return "", 0, nil, err
	}
	defer closer.Close()

	res, err := nodeClient.NodeGetInfo(ctx, &csipbv1.NodeGetInfoRequest{})
	if err != nil {
		return "", 0, nil, err
	}

	topology := res.GetAccessibleTopology()
	if topology != nil {
		accessibleTopology = topology.Segments
	}

	return res.GetNodeId(), res.GetMaxVolumesPerNode(), accessibleTopology, nil
}

func newGrpcConn(addr csiAddr) (*grpc.ClientConn, error) {
	network := "unix"
	klog.Infof("creating new gRPC connection for [%s://%s]", network, addr)

	return grpc.Dial(
		string(addr),
		grpc.WithInsecure(),
		grpc.WithContextDialer(func(ctx context.Context, target string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, network, target)
		}),
	)
}

// newV1NodeClient creates a new NodeClient with the internally used gRPC
// connection set up. It also returns a closer which must to be called to close
// the gRPC connection when the NodeClient is not used anymore.
// This is the default implementation for the nodeV1ClientCreator, used in
// newCsiDriverClient.
func newV1NodeClient(addr csiAddr) (nodeClient csipbv1.NodeClient, closer io.Closer, err error) {
	var conn *grpc.ClientConn
	conn, err = newGrpcConn(addr)
	if err != nil {
		return nil, nil, err
	}

	nodeClient = csipbv1.NewNodeClient(conn)
	return nodeClient, conn, nil
}

func newCsiDriverClient(driverName csiDriverName) (*csiDriverClient, error) {
	if driverName == "" {
		return nil, fmt.Errorf("driver name is empty")
	}

	existingDriver, driverExists := csiDrivers.Get(string(driverName))
	if !driverExists {
		return nil, fmt.Errorf("driver name %s not found in the list of registered CSI drivers", driverName)
	}

	nodeV1ClientCreator := newV1NodeClient
	return &csiDriverClient{
		driverName:          driverName,
		addr:                csiAddr(existingDriver.endpoint),
		nodeV1ClientCreator: nodeV1ClientCreator,
	}, nil
}
