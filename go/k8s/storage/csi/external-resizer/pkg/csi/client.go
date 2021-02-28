package csi

import (
	"fmt"
	"time"

	"k8s-lx1036/k8s/storage/csi/csi-lib-utils/connection"
	"k8s-lx1036/k8s/storage/csi/csi-lib-utils/rpc"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"
)

type Client struct {
	conn       *grpc.ClientConn
	nodeClient csi.NodeClient
	ctrlClient csi.ControllerClient
}

func New(address string, timeout time.Duration) (*Client, error) {
	conn, err := connection.Connect(address, nil, connection.OnConnectionLoss(connection.ExitOnConnectionLoss()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to CSI driver: %v", err)
	}

	err = rpc.ProbeForever(conn, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed probing CSI driver: %v", err)
	}

	return &client{
		conn:       conn,
		nodeClient: csi.NewNodeClient(conn),
		ctrlClient: csi.NewControllerClient(conn),
	}, nil
}
