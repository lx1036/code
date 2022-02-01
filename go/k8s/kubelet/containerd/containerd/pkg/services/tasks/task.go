package tasks

import (
	"context"

	"k8s-lx1036/k8s/kubelet/containerd/containerd/pkg/plugin"
	"k8s-lx1036/k8s/kubelet/containerd/containerd/pkg/runtime"
	"k8s-lx1036/k8s/kubelet/containerd/containerd/pkg/services"

	containerpb "github.com/containerd/containerd/api/services/containers/v1"
	taskpb "github.com/containerd/containerd/api/services/tasks/v1"
	"github.com/gogo/protobuf/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	plugin.Register(&plugin.Registration{
		Type:     plugin.ServicePlugin,
		ID:       services.TasksService,
		Requires: tasksServiceRequires,
		InitFn:   initFunc,
	})
}

var tasksServiceRequires = []plugin.Type{
	plugin.EventPlugin,
	plugin.RuntimePluginV2,
	plugin.MetadataPlugin,
	plugin.TaskMonitorPlugin,
}

func initFunc(ic *plugin.InitContext) (interface{}, error) {
	r, err := ic.GetByID(plugin.RuntimePluginV2, "task")
	if err != nil {
		return nil, err
	}

	taskService := &TaskService{
		runtime: r.(runtime.PlatformRuntime),
	}

	return taskService, nil
}

type TaskService struct {
	runtime runtime.PlatformRuntime
}

func (taskService *TaskService) Register(server *grpc.Server) error {
	taskpb.RegisterTasksServer(server, taskService)
	return nil
}

func (taskService *TaskService) Create(ctx context.Context, request *taskpb.CreateTaskRequest) (*taskpb.CreateTaskResponse, error) {
	panic("implement me")
}

func (taskService *TaskService) Start(ctx context.Context, request *taskpb.StartRequest) (*taskpb.StartResponse, error) {
	t, err := taskService.getTask(ctx, request.ContainerID)
	if err != nil {
		return nil, err
	}
	p := runtime.Process(t)
	if request.ExecID != "" {
		if p, err = t.Process(ctx, request.ExecID); err != nil {
			return nil, err
		}
	}

	if err = p.Start(ctx); err != nil {
		return nil, err
	}
	state, err := p.State(ctx)
	if err != nil {
		return nil, err
	}
	return &taskpb.StartResponse{
		Pid: state.Pid,
	}, nil
}

func (taskService *TaskService) Delete(ctx context.Context, request *taskpb.DeleteTaskRequest) (*taskpb.DeleteResponse, error) {
	panic("implement me")
}

func (taskService *TaskService) DeleteProcess(ctx context.Context, request *taskpb.DeleteProcessRequest) (*taskpb.DeleteResponse, error) {
	panic("implement me")
}

func (taskService *TaskService) Get(ctx context.Context, request *taskpb.GetRequest) (*taskpb.GetResponse, error) {
	panic("implement me")
}

func (taskService *TaskService) List(ctx context.Context, request *taskpb.ListTasksRequest) (*taskpb.ListTasksResponse, error) {
	panic("implement me")
}

func (taskService *TaskService) Kill(ctx context.Context, request *taskpb.KillRequest) (*types.Empty, error) {
	panic("implement me")
}

func (taskService *TaskService) Exec(ctx context.Context, request *taskpb.ExecProcessRequest) (*types.Empty, error) {
	panic("implement me")
}

func (taskService *TaskService) ResizePty(ctx context.Context, request *taskpb.ResizePtyRequest) (*types.Empty, error) {
	panic("implement me")
}

func (taskService *TaskService) CloseIO(ctx context.Context, request *taskpb.CloseIORequest) (*types.Empty, error) {
	panic("implement me")
}

func (taskService *TaskService) Pause(ctx context.Context, request *taskpb.PauseTaskRequest) (*types.Empty, error) {
	panic("implement me")
}

func (taskService *TaskService) Resume(ctx context.Context, request *taskpb.ResumeTaskRequest) (*types.Empty, error) {
	panic("implement me")
}

func (taskService *TaskService) ListPids(ctx context.Context, request *taskpb.ListPidsRequest) (*taskpb.ListPidsResponse, error) {
	panic("implement me")
}

func (taskService *TaskService) Checkpoint(ctx context.Context, request *taskpb.CheckpointTaskRequest) (*taskpb.CheckpointTaskResponse, error) {
	panic("implement me")
}

func (taskService *TaskService) Update(ctx context.Context, request *taskpb.UpdateTaskRequest) (*types.Empty, error) {
	panic("implement me")
}

func (taskService *TaskService) Metrics(ctx context.Context, request *taskpb.MetricsRequest) (*taskpb.MetricsResponse, error) {
	panic("implement me")
}

func (taskService *TaskService) Wait(ctx context.Context, request *taskpb.WaitRequest) (*taskpb.WaitResponse, error) {
	t, err := taskService.getTask(ctx, request.ContainerID)
	if err != nil {
		return nil, err
	}
	p := runtime.Process(t)
	if request.ExecID != "" {
		if p, err = t.Process(ctx, request.ExecID); err != nil {
			return nil, err
		}
	}
	exit, err := p.Wait(ctx)
	if err != nil {
		return nil, err
	}
	return &taskpb.WaitResponse{
		ExitStatus: exit.Status,
		ExitedAt:   exit.Timestamp,
	}, nil
}

func (taskService *TaskService) getTask(ctx context.Context, id string) (runtime.Task, error) {
	container, err := taskService.getContainer(ctx, id)
	if err != nil {
		return nil, err
	}
	return taskService.getTaskFromContainer(ctx, container)
}

func (taskService *TaskService) getContainer(ctx context.Context, id string) (*containerpb.Container, error) {
	var container containerpb.Container
	container, err := taskService.containers.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &container, nil
}

func (taskService *TaskService) getTaskFromContainer(ctx context.Context, container *containerpb.Container) (runtime.Task, error) {
	t, err := taskService.runtime.Get(ctx, container.ID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "task %v not found", container.ID)
	}
	return t, nil
}
