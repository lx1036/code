package volcano

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"

	v1 "k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/apis/sparkoperator.k9s.io/v1"
	"k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/batchscheduler/schedulerinterface"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"volcano.sh/apis/pkg/apis/scheduling/v1beta1"
	"volcano.sh/apis/pkg/client/clientset/versioned"
)

const (
	Name = "volcano"

	// INFO: 看 volcano 的 podgroups crd 定义
	PodGroupName = "podgroups.scheduling.volcano.sh"
)

func GetPluginName() string {
	return Name
}

type VolcanoBatchScheduler struct {
	crdClient     apiextensionsclient.Interface
	volcanoClient versioned.Interface
}

func (scheduler *VolcanoBatchScheduler) Name() string {
	return GetPluginName()
}

func (scheduler *VolcanoBatchScheduler) ShouldSchedule(app *v1.SparkApplication) bool {
	//NOTE: There is no additional requirement for volcano scheduler
	return true
}

// INFO: DoBatchSchedulingOnSubmission 做了两个逻辑：1. 创建或更新 podgroup 对象；2. 更新 SparkApplication driver/executor annotation 值
func (scheduler *VolcanoBatchScheduler) DoBatchSchedulingOnSubmission(app *v1.SparkApplication) error {
	if app.Spec.Executor.Annotations == nil {
		app.Spec.Executor.Annotations = make(map[string]string)
	}

	if app.Spec.Driver.Annotations == nil {
		app.Spec.Driver.Annotations = make(map[string]string)
	}

	if app.Spec.Mode == v1.ClientMode {
		return scheduler.syncPodGroupInClientMode(app)
	} else if app.Spec.Mode == v1.ClusterMode {
		return scheduler.syncPodGroupInClusterMode(app)
	}

	return nil
}

func (scheduler *VolcanoBatchScheduler) syncPodGroupInClientMode(app *v1.SparkApplication) error {
	return nil
}

func (scheduler *VolcanoBatchScheduler) syncPodGroupInClusterMode(app *v1.SparkApplication) error {
	//We need both mark Driver and Executor when submitting
	//NOTE: In cluster mode, the initial size of PodGroup is set to 1 in order to schedule driver pod first.
	if _, ok := app.Spec.Driver.Annotations[v1beta1.KubeGroupNameAnnotationKey]; !ok {
		//Both driver and executor resource will be considered.
		totalResource := sumResourceList([]corev1.ResourceList{
			getExecutorRequestResource(app),
			getDriverRequestResource(app),
		})
		if app.Spec.BatchSchedulerOptions != nil && len(app.Spec.BatchSchedulerOptions.Resources) > 0 {
			totalResource = app.Spec.BatchSchedulerOptions.Resources
		}

		if err := scheduler.syncPodGroup(app, 1, totalResource); err == nil {
			app.Spec.Executor.Annotations[v1beta1.KubeGroupNameAnnotationKey] = scheduler.getAppPodGroupName(app)
			app.Spec.Driver.Annotations[v1beta1.KubeGroupNameAnnotationKey] = scheduler.getAppPodGroupName(app)
		} else {
			return err
		}
	}

	return nil
}

func (scheduler *VolcanoBatchScheduler) CleanupOnCompletion(app *v1.SparkApplication) error {
	panic("implement me")
}

func New(config *rest.Config) (schedulerinterface.BatchScheduler, error) {
	volcanoClient, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	crdClient, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// 检查 podgroup crd 是否存在
	_, err = crdClient.ApiextensionsV1().CustomResourceDefinitions().Get(context.TODO(), PodGroupName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("podGroup CRD is required to exists in current cluster error: %s", err)
	}

	return &VolcanoBatchScheduler{
		crdClient:     crdClient,
		volcanoClient: volcanoClient,
	}, nil
}

// INFO: 这个函数计算 cpu/memory 值之和，可以复用!!!
func sumResourceList(list []corev1.ResourceList) corev1.ResourceList {
	totalResource := make(corev1.ResourceList)
	for _, l := range list {
		for name, quantity := range l {
			if value, ok := totalResource[name]; !ok {
				totalResource[name] = quantity.DeepCopy()
			} else {
				value.Add(quantity)
				totalResource[name] = value
			}
		}
	}

	return totalResource
}
