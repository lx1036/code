package api

import (
	"fmt"
	"strconv"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"volcano.sh/apis/pkg/apis/scheduling"
)

const (
	TaskPriorityAnnotation = "volcano.sh/task-priority"
)

// TaskStatus defines the status of a task/pod.
type TaskStatus int

const (
	// Pending means the task is pending in the apiserver.
	Pending TaskStatus = 1 << iota

	// Allocated means the scheduler assigns a host to it.
	Allocated

	// Pipelined means the scheduler assigns a host to wait for releasing resource.
	Pipelined

	// Binding means the scheduler send Bind request to apiserver.
	Binding

	// Bound means the task/Pod bounds to a host.
	Bound

	// Running means a task is running on the host.
	Running

	// Releasing means a task/pod is deleted.
	Releasing

	// Succeeded means that all containers in the pod have voluntarily terminated
	// with a container exit code of 0, and the system is not going to restart any of these containers.
	Succeeded

	// Failed means that all containers in the pod have terminated, and at least one container has
	// terminated in a failure (exited with a non-zero exit code or was stopped by the system).
	Failed

	// Unknown means the status of task/pod is unknown to the scheduler.
	Unknown
)

func (ts TaskStatus) String() string {
	switch ts {
	case Pending:
		return "Pending"
	case Allocated:
		return "Allocated"
	case Pipelined:
		return "Pipelined"
	case Binding:
		return "Binding"
	case Bound:
		return "Bound"
	case Running:
		return "Running"
	case Releasing:
		return "Releasing"
	case Succeeded:
		return "Succeeded"
	case Failed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// JobID is the type of JobInfo's ID.
type JobID types.UID

type JobInfo struct {
	UID JobID

	Name      string
	Namespace string

	Tasks tasksMap
	// All tasks of the Job.
	TaskStatusIndex map[TaskStatus]tasksMap
	Allocated       *Resource
	TotalRequest    *Resource

	Queue        QueueID
	Priority     int32
	MinAvailable int32
	PodGroup     *PodGroup
	Preemptable  bool

	// RevocableZone support set volcano.sh/revocable-zone annotaion or label for pod/podgroup
	// we only support empty value or * value for this version and we will support specify revocable zone name for future release
	// empty value means workload can not use revocable node
	// * value means workload can use all the revocable node for during node active revocable time.
	RevocableZone string
	Budget        *DisruptionBudget

	ScheduleStartTimestamp metav1.Time
	CreationTimestamp      metav1.Time
}

// ReadyTaskNum returns the number of tasks that are ready or that is best-effort.
func (ji *JobInfo) ReadyTaskNum() int32 {
	occupied := 0
	occupied += len(ji.TaskStatusIndex[Bound])
	occupied += len(ji.TaskStatusIndex[Binding])
	occupied += len(ji.TaskStatusIndex[Running])
	occupied += len(ji.TaskStatusIndex[Allocated])
	occupied += len(ji.TaskStatusIndex[Succeeded])

	return int32(occupied)
}

// WaitingTaskNum returns the number of tasks that are pipelined.
func (ji *JobInfo) WaitingTaskNum() int32 {
	return int32(len(ji.TaskStatusIndex[Pipelined]))
}

func (ji *JobInfo) IsPending() bool {
	return ji.PodGroup == nil ||
		ji.PodGroup.Status.Phase == scheduling.PodGroupPending ||
		ji.PodGroup.Status.Phase == ""
}

// AddTaskInfo is used to add a task to a job
func (ji *JobInfo) AddTaskInfo(ti *TaskInfo) {
	ji.Tasks[ti.UID] = ti
	ji.addTaskIndex(ti)
	ji.TotalRequest.Add(ti.Resreq)
	if AllocatedStatus(ti.Status) {
		ji.Allocated.Add(ti.Resreq)
	}
}

func (ji *JobInfo) addTaskIndex(ti *TaskInfo) {
	if _, found := ji.TaskStatusIndex[ti.Status]; !found {
		ji.TaskStatusIndex[ti.Status] = tasksMap{}
	}
	ji.TaskStatusIndex[ti.Status][ti.UID] = ti
}

func (ji *JobInfo) String() string {
	res := ""

	i := 0
	for _, task := range ji.Tasks {
		res += fmt.Sprintf("\n\t %d: %v", i, task)
		i++
	}

	return fmt.Sprintf("Job (%v): namespace %v (%v), name %v, minAvailable %d, podGroup %+v, preemptable %+v, revocableZone %+v, minAvailable %+v, maxAvailable %+v",
		ji.UID, ji.Namespace, ji.Queue, ji.Name, ji.MinAvailable, ji.PodGroup, ji.Preemptable, ji.RevocableZone, ji.Budget.MinAvailable, ji.Budget.MaxUnavilable) + res
}

// NewJobInfo creates a new jobInfo for set of tasks
func NewJobInfo(uid JobID, tasks ...*TaskInfo) *JobInfo {
	job := &JobInfo{
		UID:          uid,
		MinAvailable: 0,
		//NodesFitErrors:   make(map[TaskID]*FitErrors),
		Allocated:       EmptyResource(),
		TotalRequest:    EmptyResource(),
		TaskStatusIndex: map[TaskStatus]tasksMap{},
		Tasks:           tasksMap{},
		//TaskMinAvailable: map[TaskID]int32{},
	}

	for _, task := range tasks {
		job.AddTaskInfo(task)
	}

	return job
}

type DisruptionBudget struct {
	MinAvailable  string
	MaxUnavilable string
}

type tasksMap map[TaskID]*TaskInfo

type TaskID types.UID

// TransactionContext holds all the fields that needed by scheduling transaction
type TransactionContext struct {
	NodeName string
	Status   TaskStatus
}

type TaskInfo struct {
	UID TaskID
	Job JobID

	Name      string
	Namespace string

	Priority int32
	// Resreq is the resource that used when task running.
	Resreq *Resource
	// InitResreq is the resource that used to launch a task.
	InitResreq *Resource

	Preemptable bool
	BestEffort  bool

	TransactionContext
	// LastTransaction holds the context of last scheduling transaction
	LastTransaction *TransactionContext

	Pod *v1.Pod
}

// NewTaskInfo creates new taskInfo object for a Pod
func NewTaskInfo(pod *v1.Pod) *TaskInfo {
	initResReq := GetPodResourceRequest(pod)
	resReq := initResReq
	bestEffort := initResReq.IsEmpty()
	preemptable := GetPodPreemptable(pod)
	jobID := getJobID(pod)

	ti := &TaskInfo{
		UID:         TaskID(pod.UID),
		Job:         jobID,
		Name:        pod.Name,
		Namespace:   pod.Namespace,
		Priority:    1,
		Pod:         pod,
		Resreq:      resReq,
		InitResreq:  initResReq,
		Preemptable: preemptable,
		BestEffort:  bestEffort,
		//RevocableZone: revocableZone,
		//NumaInfo:      topologyInfo,
		//TransactionContext: TransactionContext{
		//	NodeName: pod.Spec.NodeName,
		//	Status:   getTaskStatus(pod),
		//},
	}

	if pod.Spec.Priority != nil {
		ti.Priority = *pod.Spec.Priority
	}
	if taskPriority, ok := pod.Annotations[TaskPriorityAnnotation]; ok {
		if priority, err := strconv.ParseInt(taskPriority, 10, 32); err == nil {
			ti.Priority = int32(priority)
		}
	}

	return ti
}

func getJobID(pod *v1.Pod) JobID {
	if gn, found := pod.Annotations[v1beta1.KubeGroupNameAnnotationKey]; found && len(gn) != 0 {
		// Make sure Pod and PodGroup belong to the same namespace.
		jobID := fmt.Sprintf("%s/%s", pod.Namespace, gn)
		return JobID(jobID)
	}

	return ""
}
