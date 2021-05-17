package dockershim

import (
	"context"

	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
)

func (ds *dockerService) ExecSync(ctx context.Context, request *runtimeapi.ExecSyncRequest) (*runtimeapi.ExecSyncResponse, error) {
	panic("implement me")
}

func (ds *dockerService) Exec(ctx context.Context, request *runtimeapi.ExecRequest) (*runtimeapi.ExecResponse, error) {
	panic("implement me")
}

func (ds *dockerService) Attach(ctx context.Context, request *runtimeapi.AttachRequest) (*runtimeapi.AttachResponse, error) {
	panic("implement me")
}

func (ds *dockerService) PortForward(ctx context.Context, request *runtimeapi.PortForwardRequest) (*runtimeapi.PortForwardResponse, error) {
	panic("implement me")
}
