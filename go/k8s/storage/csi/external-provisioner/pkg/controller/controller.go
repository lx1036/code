package controller

import (
	"context"
	"time"

	"k8s-lx1036/k8s/storage/csi/csi-lib-utils/connection"
	"k8s-lx1036/k8s/storage/csi/csi-lib-utils/metrics"
	"k8s-lx1036/k8s/storage/csi/csi-lib-utils/rpc"

	"google.golang.org/grpc"

	"github.com/container-storage-interface/spec/lib/go/csi"
	coreinformers "k8s.io/client-go/informers/core/v1"
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
