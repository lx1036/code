package utils

import (
	"context"
	"fmt"

	"k8s-lx1036/k8s/scheduler/descheduler/pkg/api"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/apis/scheduling"
)

// 从传参中获取 pod priority
func GetPriorityFromStrategyParams(ctx context.Context, client clientset.Interface,
	params *api.StrategyParameters) (priority int32, err error) {

	if params == nil {
		return scheduling.SystemCriticalPriority, nil
	}

	if params.ThresholdPriority != nil {
		priority = *params.ThresholdPriority
	} else {
		priority, err = GetPriorityFromPriorityClass(ctx, client, params.ThresholdPriorityClassName)
		if err != nil {
			return
		}
	}
	if priority > scheduling.SystemCriticalPriority {
		return 0, fmt.Errorf("priority threshold can't be greater than %d", scheduling.SystemCriticalPriority)
	}

	return
}

func GetPriorityFromPriorityClass(ctx context.Context, client clientset.Interface, name string) (int32, error) {
	if name != "" {
		// 从 apiserver 取值
		priorityClass, err := client.SchedulingV1().PriorityClasses().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return 0, err
		}
		return priorityClass.Value, nil
	}

	return scheduling.SystemCriticalPriority, nil
}
