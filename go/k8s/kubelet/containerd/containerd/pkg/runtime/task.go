package runtime

import (
	"context"
	"github.com/gogo/protobuf/types"
	"sync"
)

// Process is a runtime object for an executing process inside a container
type Process interface {
	// ID of the process
	ID() string
	// State returns the process state
	State(ctx context.Context) (State, error)
	// Kill signals a container
	Kill(ctx context.Context, signal uint32, all bool) error
	// ResizePty resizes the processes pty/console
	ResizePty(ctx context.Context, size ConsoleSize) error
	// CloseIO closes the processes IO
	CloseIO(ctx context.Context) error
	// Start the container's user defined process
	Start(ctx context.Context) error
	// Wait for the process to exit
	Wait(ctx context.Context) (*Exit, error)
}

// Task is the runtime object for an executing container
type Task interface {
	Process

	// PID of the process
	PID(ctx context.Context) (uint32, error)
	// Namespace that the task exists in
	Namespace() string
	// Pause pauses the container process
	Pause(ctx context.Context) error
	// Resume unpauses the container process
	Resume(ctx context.Context) error
	// Exec adds a process into the container
	Exec(ctx context.Context, id string, opts ExecOpts) (ExecProcess, error)
	// Pids returns all pids
	Pids(ctx context.Context) ([]ProcessInfo, error)
	// Checkpoint checkpoints a container to an image with live system data
	Checkpoint(ctx context.Context, path string, opts *types.Any) error
	// Update sets the provided resources to a running task
	Update(ctx context.Context, resources *types.Any, annotations map[string]string) error
	// Process returns a process within the task for the provided id
	Process(ctx context.Context, id string) (ExecProcess, error)
	// Stats returns runtime specific metrics for a task
	Stats(ctx context.Context) (*types.Any, error)
}

// TaskList holds and provides locking around tasks
type TaskList struct {
	mu    sync.Mutex
	tasks map[string]map[string]Task
}

func NewTaskList() *TaskList {
	return &TaskList{
		tasks: make(map[string]map[string]Task),
	}
}
