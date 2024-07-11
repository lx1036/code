package priority

import (
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/api"
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/framework"
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/plugins/util"
	"k8s.io/klog/v2"
)

const PluginName = "priority"

type priorityPlugin struct {
	// Arguments given for the plugin
	pluginArguments framework.Arguments
}

func New(arguments framework.Arguments) framework.Plugin {
	return &priorityPlugin{pluginArguments: arguments}
}

func (pp *priorityPlugin) Name() string {
	return PluginName
}

func (pp *priorityPlugin) OnSessionOpen(ssn *framework.Session) {
	// Add Task Order function
	taskOrderFn := func(l interface{}, r interface{}) int {
		lv := l.(*api.TaskInfo)
		rv := r.(*api.TaskInfo)

		klog.V(4).Infof("Priority TaskOrder: <%v/%v> priority is %v, <%v/%v> priority is %v",
			lv.Namespace, lv.Name, lv.Priority, rv.Namespace, rv.Name, rv.Priority)

		if lv.Priority == rv.Priority {
			return 0
		}

		if lv.Priority > rv.Priority {
			return -1
		}

		return 1
	}
	ssn.AddTaskOrderFn(pp.Name(), taskOrderFn)

	// Add Job Order function
	jobOrderFn := func(l, r interface{}) int {
		lv := l.(*api.JobInfo)
		rv := r.(*api.JobInfo)

		klog.V(4).Infof("Priority JobOrderFn: <%v/%v> priority: %d, <%v/%v> priority: %d",
			lv.Namespace, lv.Name, lv.Priority, rv.Namespace, rv.Name, rv.Priority)

		if lv.Priority > rv.Priority {
			return -1
		}

		if lv.Priority < rv.Priority {
			return 1
		}

		return 0
	}
	ssn.AddJobOrderFn(pp.Name(), jobOrderFn)

	// Add Job Preempt function
	preemptableFn := func(preemptor *api.TaskInfo, preemptees []*api.TaskInfo) ([]*api.TaskInfo, int) {
		preemptorJob := ssn.Jobs[preemptor.Job]
		var victims []*api.TaskInfo
		for _, preemptee := range preemptees {
			preempteeJob := ssn.Jobs[preemptee.Job]
			if preempteeJob.UID != preemptorJob.UID {
				if preempteeJob.Priority >= preemptorJob.Priority { // Preemption between Jobs within Queue
					klog.V(4).Infof("Can not preempt task <%v/%v>"+
						"because preemptee job has greater or equal job priority (%d) than preemptor (%d)",
						preemptee.Namespace, preemptee.Name, preempteeJob.Priority, preemptorJob.Priority)
				} else {
					victims = append(victims, preemptee)
				}
			} else { // same job's different tasks should compare task's priority
				if preemptee.Priority >= preemptor.Priority {
					klog.V(4).Infof("Can not preempt task <%v/%v>"+
						"because preemptee task has greater or equal task priority (%d) than preemptor (%d)",
						preemptee.Namespace, preemptee.Name, preemptee.Priority, preemptor.Priority)
				} else {
					victims = append(victims, preemptee)
				}
			}
		}

		klog.V(4).Infof("Victims from %s plugins are %+v", pp.Name(), victims)
		return victims, util.Permit
	}
	ssn.AddPreemptableFn(pp.Name(), preemptableFn)

	jobStarvingFn := func(obj interface{}) bool {
		jobInfo := obj.(*api.JobInfo)
		return jobInfo.ReadyTaskNum()+jobInfo.WaitingTaskNum() < int32(len(jobInfo.Tasks))
	}
	ssn.AddJobStarvingFns(pp.Name(), jobStarvingFn)
}

func (pp *priorityPlugin) OnSessionClose(ssn *framework.Session) {}
