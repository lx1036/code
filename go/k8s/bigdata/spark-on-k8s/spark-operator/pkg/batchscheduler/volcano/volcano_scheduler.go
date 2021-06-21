package volcano

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"

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

		/*
			INFO: ./bin/spark-submit \
			    --master k8s://https://<k8s-apiserver-host>:<k8s-apiserver-port> \
			    --deploy-mode cluster \
			    --name spark-pi \
			    --class org.apache.spark.examples.SparkPi \
			    --conf spark.executor.instances=5 \
			    --conf spark.kubernetes.container.image=<spark-image> \
			    local:///path/to/examples.jar
		*/

		// INFO: https://spark.apache.org/docs/latest/running-on-kubernetes.html#cluster-mode
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
		// INFO: 统计driver和executor(cores*instances)资源使用总和
		totalResource := sumResourceList([]corev1.ResourceList{
			getExecutorRequestResource(app),
			getDriverRequestResource(app),
		})
		if app.Spec.BatchSchedulerOptions != nil && len(app.Spec.BatchSchedulerOptions.Resources) > 0 {
			totalResource = app.Spec.BatchSchedulerOptions.Resources
		}

		// INFO: podgroup MinMember 设置为 1，为了先调度 driver pod
		if err := scheduler.syncPodGroup(app, 1, totalResource); err == nil {
			app.Spec.Executor.Annotations[v1beta1.KubeGroupNameAnnotationKey] = scheduler.getAppPodGroupName(app)
			app.Spec.Driver.Annotations[v1beta1.KubeGroupNameAnnotationKey] = scheduler.getAppPodGroupName(app)
		} else {
			return err
		}
	}

	return nil
}

func (scheduler *VolcanoBatchScheduler) getAppPodGroupName(app *v1.SparkApplication) string {
	return fmt.Sprintf("spark-%s-pg", app.Name)
}

// INFO: 创建一个新 podgroup，或者更新设置 MinMember 为 1(为了先创建 driver pod)
func (scheduler *VolcanoBatchScheduler) syncPodGroup(app *v1.SparkApplication, size int32, minResource corev1.ResourceList) error {
	var err error
	podGroupName := scheduler.getAppPodGroupName(app)
	if pg, err := scheduler.volcanoClient.SchedulingV1beta1().PodGroups(app.Namespace).Get(context.TODO(), podGroupName, metav1.GetOptions{}); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		// INFO: 创建一个新的 podgroup，最后会在 CleanupOnCompletion 中删除
		podGroup := v1beta1.PodGroup{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: app.Namespace,
				Name:      podGroupName,
				OwnerReferences: []metav1.OwnerReference{
					*metav1.NewControllerRef(app, v1.SchemeGroupVersion.WithKind("SparkApplication")),
				},
			},
			Spec: v1beta1.PodGroupSpec{
				MinMember:    size,
				MinResources: &minResource,
			},
			Status: v1beta1.PodGroupStatus{
				Phase: v1beta1.PodGroupPending,
			},
		}
		if app.Spec.BatchSchedulerOptions != nil {
			//Update pod group queue if it's specified in Spark Application
			if app.Spec.BatchSchedulerOptions.Queue != nil {
				podGroup.Spec.Queue = *app.Spec.BatchSchedulerOptions.Queue
			}
			//Update pod group priorityClassName if it's specified in Spark Application
			if app.Spec.BatchSchedulerOptions.PriorityClassName != nil {
				podGroup.Spec.PriorityClassName = *app.Spec.BatchSchedulerOptions.PriorityClassName
			}
		}
		_, err = scheduler.volcanoClient.SchedulingV1beta1().PodGroups(app.Namespace).Create(context.TODO(), &podGroup, metav1.CreateOptions{})
	} else {
		// https://volcano.sh/en/docs/podgroup/#minmember
		if pg.Spec.MinMember != size {
			pg.Spec.MinMember = size
			_, err = scheduler.volcanoClient.SchedulingV1beta1().PodGroups(app.Namespace).Update(context.TODO(), pg, metav1.UpdateOptions{})
		}
	}

	if err != nil {
		return fmt.Errorf("failed to sync PodGroup with error: %s. Abandon schedule pods via volcano", err)
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

func getExecutorRequestResource(app *v1.SparkApplication) corev1.ResourceList {
	minResource := corev1.ResourceList{}

	//CoreRequest correspond to executor's core request
	if app.Spec.Executor.CoreRequest != nil {
		if value, err := resource.ParseQuantity(*app.Spec.Executor.CoreRequest); err == nil {
			minResource[corev1.ResourceCPU] = value
		}
	}

	//Use Core attribute if CoreRequest is empty
	if app.Spec.Executor.Cores != nil {
		if _, ok := minResource[corev1.ResourceCPU]; !ok {
			if value, err := resource.ParseQuantity(fmt.Sprintf("%d", *app.Spec.Executor.Cores)); err == nil {
				minResource[corev1.ResourceCPU] = value
			}
		}
	}

	//CoreLimit correspond to executor's core limit, this attribute will be used only when core request is empty.
	if app.Spec.Executor.CoreLimit != nil {
		if _, ok := minResource[corev1.ResourceCPU]; !ok {
			if value, err := resource.ParseQuantity(*app.Spec.Executor.CoreLimit); err == nil {
				minResource[corev1.ResourceCPU] = value
			}
		}
	}

	//Memory + MemoryOverhead correspond to executor's memory request
	if app.Spec.Executor.Memory != nil {
		if value, err := resource.ParseQuantity(*app.Spec.Executor.Memory); err == nil {
			minResource[corev1.ResourceMemory] = value
		}
	}
	if app.Spec.Executor.MemoryOverhead != nil {
		if value, err := resource.ParseQuantity(*app.Spec.Executor.MemoryOverhead); err == nil {
			if existing, ok := minResource[corev1.ResourceMemory]; ok {
				existing.Add(value)
				minResource[corev1.ResourceMemory] = existing
			}
		}
	}

	resourceList := []corev1.ResourceList{{}}
	for i := int32(0); i < *app.Spec.Executor.Instances; i++ {
		resourceList = append(resourceList, minResource)
	}

	return sumResourceList(resourceList)
}

func getDriverRequestResource(app *v1.SparkApplication) corev1.ResourceList {
	minResource := corev1.ResourceList{}

	//Cores correspond to driver's core request
	if app.Spec.Driver.Cores != nil {
		if value, err := resource.ParseQuantity(fmt.Sprintf("%d", *app.Spec.Driver.Cores)); err == nil {
			minResource[corev1.ResourceCPU] = value
		}
	}

	//CoreLimit correspond to driver's core limit, this attribute will be used only when core request is empty.
	if app.Spec.Driver.CoreLimit != nil {
		if _, ok := minResource[corev1.ResourceCPU]; !ok {
			if value, err := resource.ParseQuantity(*app.Spec.Driver.CoreLimit); err == nil {
				minResource[corev1.ResourceCPU] = value
			}
		}
	}

	//Memory + MemoryOverhead correspond to driver's memory request
	if app.Spec.Driver.Memory != nil {
		if value, err := resource.ParseQuantity(*app.Spec.Driver.Memory); err == nil {
			minResource[corev1.ResourceMemory] = value
		}
	}
	if app.Spec.Driver.MemoryOverhead != nil {
		if value, err := resource.ParseQuantity(*app.Spec.Driver.MemoryOverhead); err == nil {
			if existing, ok := minResource[corev1.ResourceMemory]; ok {
				existing.Add(value)
				minResource[corev1.ResourceMemory] = existing
			}
		}
	}

	return minResource
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
