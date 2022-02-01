package client

import (
	"context"
	"syscall"

	taskpb "github.com/containerd/containerd/api/services/tasks/v1"
	"github.com/containerd/containerd/cio"
)

// UnknownExitStatus is returned when containerd is unable to
// determine the exit status of a process. This can happen if the process never starts
// or if an error was encountered when obtaining the exit status, it is set to 255.
const UnknownExitStatus = 255

type Task struct {
	pid uint32
	io  cio.IO
	id  string

	client *Client
}

func (task *Task) Wait(ctx context.Context) (<-chan ExitStatus, error) {
	c := make(chan ExitStatus, 1)
	go func() {

		r, err := task.client.TaskService().Wait(ctx, &taskpb.WaitRequest{
			ContainerID: task.id,
		})

		if err != nil {
			c <- ExitStatus{
				code: UnknownExitStatus,
				err:  err,
			}
			return
		}
		c <- ExitStatus{
			code:     r.ExitStatus,
			exitedAt: r.ExitedAt,
		}

	}()

	return c, nil
}

func (task *Task) Start(ctx context.Context) error {
	resp, err := task.client.TaskService().Start(ctx, &taskpb.StartRequest{
		ContainerID: task.id,
	})
	if err != nil {
		if task.io != nil {
			task.io.Cancel()
			task.io.Close()
		}
		return err
	}

	task.pid = resp.Pid
	return nil
}

// KillInfo contains information on how to process a Kill action
type KillInfo struct {
	// All kills all processes inside the task
	// only valid on tasks, ignored on processes
	All bool
	// ExecID is the ID of a process to kill
	ExecID string
}

// KillOpts allows options to be set for the killing of a process
type KillOpts func(context.Context, *KillInfo) error

func (task *Task) Kill(ctx context.Context, s syscall.Signal, opts ...KillOpts) error {
	var i KillInfo
	for _, o := range opts {
		if err := o(ctx, &i); err != nil {
			return err
		}
	}
	_, err := task.client.TaskService().Kill(ctx, &taskpb.KillRequest{
		Signal:      uint32(s),
		ContainerID: task.id,
		ExecID:      i.ExecID,
		All:         i.All,
	})
	if err != nil {
		return err
	}
	return nil
}
