package runtime

import (
	"context"
	"fmt"
	"github.com/containerd/containerd/plugin"
	"github.com/containerd/containerd/runtime"
)

// PlatformRuntime is responsible for the creation and management of
// tasks and processes for a platform.
type PlatformRuntime interface {
	// ID of the runtime
	ID() string
	// Create creates a task with the provided id and options.
	Create(ctx context.Context, taskID string, opts CreateOpts) (Task, error)
	// Get returns a task.
	Get(ctx context.Context, taskID string) (Task, error)
	// Tasks returns all the current tasks for the runtime.
	// Any container runs at most one task at a time.
	Tasks(ctx context.Context, all bool) ([]Task, error)
	// Delete remove a task.
	Delete(ctx context.Context, taskID string) (*Exit, error)
}

// ShimManager manages currently running shim processes.
// It is mainly responsible for launching new shims and for proper shutdown and cleanup of existing instances.
// The manager is unaware of the underlying services shim provides and lets higher level services consume them,
// but don't care about lifecycle management.
type ShimManager struct {
	shims *TaskList
}

func (m *ShimManager) ID() string {
	return fmt.Sprintf("%s.%s", plugin.RuntimePluginV2, "shim")
}

// Create launches new shim instance and creates new task
func (m *ShimManager) Create(ctx context.Context, taskID string, opts runtime.CreateOpts) (runtime.Task, error) {

}

func (m *ShimManager) Get(ctx context.Context, id string) (ShimProcess, error) {
	proc, err := m.shims.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return proc, nil
}

func (m *ShimManager) Tasks(ctx context.Context, all bool) ([]Task, error) {

}

func (m *ShimManager) Delete(ctx context.Context, taskID string) (*Exit, error) {

}
