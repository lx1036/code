package client

import (
	"context"

	containerpb "github.com/containerd/containerd/api/services/containers/v1"
	taskpb "github.com/containerd/containerd/api/services/tasks/v1"
	"github.com/containerd/containerd/cio"
)

type NewTaskOpts func(context.Context, *Client, *TaskInfo) error

// Container INFO: @see containerpb.Container
type Container struct {
	client   *Client
	id       string
	metadata containerpb.Container
}

func containerFromRecord(client *Client, c containerpb.Container) *Container {
	return &Container{
		client:   client,
		id:       c.ID,
		metadata: c,
	}
}

func (c *Container) NewTask(ctx context.Context, ioCreate cio.Creator, opts ...NewTaskOpts) (*Task, error) {
	request := &taskpb.CreateTaskRequest{
		ContainerID: c.id,
		Terminal:    cfg.Terminal,
		Stdin:       cfg.Stdin,
		Stdout:      cfg.Stdout,
		Stderr:      cfg.Stderr,
	}

	task := &Task{
		client: c.client,
		io:     i,
		id:     c.id,
		c:      c,
	}

	response, err := c.client.TaskService().Create(ctx, request)
	if err != nil {
		return nil, err
	}
	task.pid = response.Pid
	return task, nil
}
