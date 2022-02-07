package client

import (
	"context"
	"fmt"
	"k8s-lx1036/k8s/kubelet/containerd/containerd/pkg/util"
	"sync"
	"time"

	containerpb "github.com/containerd/containerd/api/services/containers/v1"
	taskpb "github.com/containerd/containerd/api/services/tasks/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	DefaultRuntime = "io.containerd.runc.v2"
)

type NewContainerOpts func(ctx context.Context, client *Client, c *containerpb.Container) error

type services struct {
	taskService      taskpb.TasksClient
	containerService containerpb.ContainersClient
}

type Client struct {
	connMu sync.Mutex
	conn   *grpc.ClientConn

	services
	runtime string
}

func New(address string) (*Client, error) {
	var err error
	c := &Client{
		runtime: DefaultRuntime,
	}

	backoffConfig := backoff.DefaultConfig
	backoffConfig.MaxDelay = 3 * time.Second
	connParams := grpc.ConnectParams{
		Backoff: backoffConfig,
	}
	clientOps := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.FailOnNonTempDialError(true),
		grpc.WithConnectParams(connParams),
		//grpc.WithContextDialer(dialer.ContextDialer),
		grpc.WithReturnConnectionError(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(util.DefaultMaxRecvMsgSize),
			grpc.MaxCallSendMsgSize(util.DefaultMaxSendMsgSize)),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	c.conn, err = grpc.DialContext(ctx, fmt.Sprintf("unix://%s", address), clientOps...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %q: %w", address, err)
	}

	return c, nil
}

func (c *Client) NewContainer(ctx context.Context, id string, opts ...NewContainerOpts) (*Container, error) {
	ctr := containerpb.Container{
		ID: id,
		Runtime: &containerpb.Container_Runtime{
			Name: c.runtime,
		},
	}

	for _, o := range opts {
		if err := o(ctx, c, &ctr); err != nil {
			return nil, err
		}
	}

	resp, err := c.ContainerService().Create(ctx, &containerpb.CreateContainerRequest{
		Container: ctr,
	})
	if err != nil {
		return nil, err
	}

	return containerFromRecord(c, resp.Container), nil
}

func (c *Client) TaskService() taskpb.TasksClient {
	if c.taskService != nil {
		return c.taskService
	}

	c.connMu.Lock()
	defer c.connMu.Unlock()
	return taskpb.NewTasksClient(c.conn)
}

func (c *Client) ContainerService() containerpb.ContainersClient {
	if c.containerService != nil {
		return c.containerService
	}

	c.connMu.Lock()
	defer c.connMu.Unlock()
	return containerpb.NewContainersClient(c.conn)
}
