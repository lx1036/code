package actions

import (
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/actions/enqueue"
	"k8s-lx1036/k8s/scheduler/volcano/volcano/pkg/scheduler/framework"
)

func init() {

	framework.RegisterAction(enqueue.New())

}
