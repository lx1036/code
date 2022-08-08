package podgroups

import (
	"fmt"

	"k8s-lx1036/k8s/scheduler/pkg/apis/podgroup/config"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
)

// New initializes and returns a new Coscheduling plugin.
func New(obj runtime.Object, handle framework.FrameworkHandle) (framework.Plugin, error) {

	args, ok := obj.(*config.CoschedulingArgs)
	if !ok {
		return nil, fmt.Errorf("want args to be of type CoschedulingArgs, got %T", obj)
	}

	conf, err := clientcmd.BuildConfigFromFlags(args.KubeMaster, args.KubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to init rest.Config: %v", err)
	}

}
