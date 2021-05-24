package dockershim

import (
	"context"

	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
)

func (ds *dockerService) ListImages(ctx context.Context, request *runtimeapi.ListImagesRequest) (*runtimeapi.ListImagesResponse, error) {
	panic("implement me")
}

func (ds *dockerService) ImageStatus(ctx context.Context, request *runtimeapi.ImageStatusRequest) (*runtimeapi.ImageStatusResponse, error) {
	panic("implement me")
}

func (ds *dockerService) PullImage(ctx context.Context, request *runtimeapi.PullImageRequest) (*runtimeapi.PullImageResponse, error) {
	panic("implement me")
}

func (ds *dockerService) RemoveImage(ctx context.Context, request *runtimeapi.RemoveImageRequest) (*runtimeapi.RemoveImageResponse, error) {
	panic("implement me")
}

func (ds *dockerService) ImageFsInfo(ctx context.Context, request *runtimeapi.ImageFsInfoRequest) (*runtimeapi.ImageFsInfoResponse, error) {
	panic("implement me")
}
