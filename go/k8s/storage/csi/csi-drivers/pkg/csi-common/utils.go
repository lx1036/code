package csicommon

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"k8s.io/klog/v2"
)

func logGRPC(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	klog.Infof("GRPC call: %s", info.FullMethod)
	klog.Infof("GRPC request: %s", protosanitizer.StripSecrets(req))
	resp, err := handler(ctx, req)
	if err != nil {
		klog.Errorf("GRPC error: %v", err)
	} else {
		klog.Infof("GRPC response: %s", protosanitizer.StripSecrets(resp))
	}
	return resp, err
}
