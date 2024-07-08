package api

import (
	"fmt"
	"k8s.io/apimachinery/pkg/types"
)

// JobID is the type of JobInfo's ID.
type JobID types.UID

type JobInfo struct {
	UID JobID

	Name      string
	Namespace string

	Tasks tasksMap

	Queue        QueueID
	Priority     int32
	MinAvailable int32

	PodGroup    *PodGroup
	Preemptable bool

	// RevocableZone support set volcano.sh/revocable-zone annotaion or label for pod/podgroup
	// we only support empty value or * value for this version and we will support specify revocable zone name for future release
	// empty value means workload can not use revocable node
	// * value means workload can use all the revocable node for during node active revocable time.
	RevocableZone string
	Budget        *DisruptionBudget
}

func (ji JobInfo) String() string {
	res := ""

	i := 0
	for _, task := range ji.Tasks {
		res += fmt.Sprintf("\n\t %d: %v", i, task)
		i++
	}

	return fmt.Sprintf("Job (%v): namespace %v (%v), name %v, minAvailable %d, podGroup %+v, preemptable %+v, revocableZone %+v, minAvailable %+v, maxAvailable %+v",
		ji.UID, ji.Namespace, ji.Queue, ji.Name, ji.MinAvailable, ji.PodGroup, ji.Preemptable, ji.RevocableZone, ji.Budget.MinAvailable, ji.Budget.MaxUnavilable) + res
}

type DisruptionBudget struct {
	MinAvailable  string
	MaxUnavilable string
}

type tasksMap map[TaskID]*TaskInfo

type TaskID types.UID

type TaskInfo struct {
}
