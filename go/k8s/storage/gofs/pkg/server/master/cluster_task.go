package master

import (
	"fmt"

	"k8s-lx1036/k8s/storage/gofs/pkg/util/proto"

	"k8s.io/klog/v2"
)

func (cluster *Cluster) addMetaNodeTasks(tasks []*proto.AdminTask) {
	for _, task := range tasks {
		if task == nil {
			continue
		}
		if node, err := cluster.metaNode(task.OperatorAddr); err != nil {
			klog.Warningf(fmt.Sprintf("action[putTasks],nodeAddr:%v,taskID:%v,err:%v", task.OperatorAddr, task.ID, err))
		} else {
			node.Sender.AddTask(task)
		}
	}
}
