package main

import (
	"context"
	"fmt"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc"
)

// GetDriverName returns name of CSI driver.
func GetDriverName(ctx context.Context, conn *grpc.ClientConn) (string, error) {
	client := csi.NewIdentityClient(conn)

	req := csi.GetPluginInfoRequest{}
	rsp, err := client.GetPluginInfo(ctx, &req)
	if err != nil {
		return "", err
	}
	name := rsp.GetName()
	if name == "" {
		return "", fmt.Errorf("driver name is empty")
	}
	return name, nil
}
