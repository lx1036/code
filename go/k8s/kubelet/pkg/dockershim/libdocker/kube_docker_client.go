package libdocker

import (
	"context"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	dockercontainer "github.com/docker/docker/api/types/container"
	dockerapi "github.com/docker/docker/client"
)

// There are 2 kinds of docker operations categorized by running time:
// * Long running operation: The long running operation could run for arbitrary long time, and the running time
// usually depends on some uncontrollable factors. These operations include: PullImage, Logs, StartExec, AttachToContainer.
// * Non-long running operation: Given the maximum load of the system, the non-long running operation should finish
// in expected and usually short time. These include all other operations.
// kubeDockerClient only applies timeout on non-long running operations.
const (
	// defaultTimeout is the default timeout of short running docker operations.
	// Value is slightly offset from 2 minutes to make timeouts due to this
	// constant recognizable.
	defaultTimeout = 2*time.Minute - 1*time.Second

	// defaultShmSize is the default ShmSize to use (in bytes) if not specified.
	defaultShmSize = int64(1024 * 1024 * 64)

	// defaultImagePullingProgressReportInterval is the default interval of image pulling progress reporting.
	defaultImagePullingProgressReportInterval = 10 * time.Second
)

// kubeDockerClient is a wrapped layer of docker client for kubelet internal use. This layer is added to:
//	1) Redirect stream for exec and attach operations.
//	2) Wrap the context in this layer to make the Interface cleaner.
type kubeDockerClient struct {
	// timeout is the timeout of short running docker operations.
	timeout time.Duration
	// If no pulling progress is made before imagePullProgressDeadline, the image pulling will be cancelled.
	// Docker reports image progress for every 512kB block, so normally there shouldn't be too long interval
	// between progress updates.
	imagePullProgressDeadline time.Duration
	client                    *dockerapi.Client
}

func (d *kubeDockerClient) ListContainers(options dockertypes.ContainerListOptions) ([]dockertypes.Container, error) {
	panic("implement me")
}

func (d *kubeDockerClient) InspectContainer(id string) (*dockertypes.ContainerJSON, error) {
	panic("implement me")
}

func (d *kubeDockerClient) InspectContainerWithSize(id string) (*dockertypes.ContainerJSON, error) {
	panic("implement me")
}

func (d *kubeDockerClient) CreateContainer(config dockertypes.ContainerCreateConfig) (*dockercontainer.ContainerCreateCreatedBody, error) {
	panic("implement me")
}

func (d *kubeDockerClient) StartContainer(id string) error {
	panic("implement me")
}

func (d *kubeDockerClient) StopContainer(id string, timeout time.Duration) error {
	panic("implement me")
}

func (d *kubeDockerClient) UpdateContainerResources(id string, updateConfig dockercontainer.UpdateConfig) error {
	panic("implement me")
}

func (d *kubeDockerClient) RemoveContainer(id string, opts dockertypes.ContainerRemoveOptions) error {
	panic("implement me")
}

func (d *kubeDockerClient) InspectImageByRef(imageRef string) (*dockertypes.ImageInspect, error) {
	panic("implement me")
}

func (d *kubeDockerClient) InspectImageByID(imageID string) (*dockertypes.ImageInspect, error) {
	panic("implement me")
}

func (d *kubeDockerClient) ListImages(opts dockertypes.ImageListOptions) ([]dockertypes.ImageSummary, error) {
	panic("implement me")
}

func (d *kubeDockerClient) PullImage(image string, auth dockertypes.AuthConfig, opts dockertypes.ImagePullOptions) error {
	panic("implement me")
}

func (d *kubeDockerClient) RemoveImage(image string, opts dockertypes.ImageRemoveOptions) ([]dockertypes.ImageDeleteResponseItem, error) {
	panic("implement me")
}

func (d *kubeDockerClient) Version() (*dockertypes.Version, error) {
	panic("implement me")
}

func (d *kubeDockerClient) Info() (*dockertypes.Info, error) {
	panic("implement me")
}

func (d *kubeDockerClient) CreateExec(s string, config dockertypes.ExecConfig) (*dockertypes.IDResponse, error) {
	panic("implement me")
}

func (d *kubeDockerClient) InspectExec(id string) (*dockertypes.ContainerExecInspect, error) {
	panic("implement me")
}

func (d *kubeDockerClient) ResizeContainerTTY(id string, height, width uint) error {
	panic("implement me")
}

func (d *kubeDockerClient) ResizeExecTTY(id string, height, width uint) error {
	panic("implement me")
}

func (d *kubeDockerClient) GetContainerStats(id string) (*dockertypes.StatsJSON, error) {
	panic("implement me")
}

// getTimeoutContext returns a new context with default request timeout
func (d *kubeDockerClient) getTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d.timeout)
}

// newKubeDockerClient creates an kubeDockerClient from an existing docker client. If requestTimeout is 0,
// defaultTimeout will be applied.
func newKubeDockerClient(dockerClient *dockerapi.Client, requestTimeout, imagePullProgressDeadline time.Duration) Interface {
	if requestTimeout == 0 {
		requestTimeout = defaultTimeout
	}

	k := &kubeDockerClient{
		client:                    dockerClient,
		timeout:                   requestTimeout,
		imagePullProgressDeadline: imagePullProgressDeadline,
	}

	// Notice that this assumes that docker is running before kubelet is started.
	ctx, cancel := k.getTimeoutContext()
	defer cancel()
	dockerClient.NegotiateAPIVersion(ctx)

	return k
}
