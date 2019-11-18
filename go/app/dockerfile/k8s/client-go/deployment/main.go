package main

import (
	"flag"
	"fmt"
	"k8s-lx1036/app/dockerfile/k8s/client-go/util"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/retry"
	"path/filepath"
)

func main() {
	crudDeployment()
}

func crudDeployment() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "absolute path to kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to kubeconfig file")
	}

	fmt.Println("kube config path: " + *kubeconfig)

	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	const deploymentName = "2019-10-26-deployment"
	deploymentsClient := clientSet.AppsV1().Deployments(apiv1.NamespaceDefault)

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels:      map[string]string{"app": "2019-10-26-deployment"},
				MatchExpressions: nil,
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:                       "",
					GenerateName:               "",
					Namespace:                  "",
					SelfLink:                   "",
					UID:                        "",
					ResourceVersion:            "",
					Generation:                 0,
					CreationTimestamp:          metav1.Time{},
					DeletionTimestamp:          nil,
					DeletionGracePeriodSeconds: nil,
					Labels:                     map[string]string{"app": "2019-10-26-deployment"},
					Annotations:                nil,
					OwnerReferences:            nil,
					Finalizers:                 nil,
					ClusterName:                "",
					ManagedFields:              nil,
				},
				Spec: apiv1.PodSpec{
					Volumes:        nil,
					InitContainers: nil,
					Containers: []apiv1.Container{
						{
							Name:  "web",
							Image: "nginx:1.12",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
					EphemeralContainers:           nil,
					RestartPolicy:                 "",
					TerminationGracePeriodSeconds: nil,
					ActiveDeadlineSeconds:         nil,
					DNSPolicy:                     "",
					NodeSelector:                  nil,
					ServiceAccountName:            "",
					DeprecatedServiceAccount:      "",
					AutomountServiceAccountToken:  nil,
					NodeName:                      "",
					HostNetwork:                   false,
					HostPID:                       false,
					HostIPC:                       false,
					ShareProcessNamespace:         nil,
					SecurityContext:               nil,
					ImagePullSecrets:              nil,
					Hostname:                      "",
					Subdomain:                     "",
					Affinity:                      nil,
					SchedulerName:                 "",
					Tolerations:                   nil,
					HostAliases:                   nil,
					PriorityClassName:             "",
					Priority:                      nil,
					DNSConfig:                     nil,
					ReadinessGates:                nil,
					RuntimeClassName:              nil,
					EnableServiceLinks:            nil,
					PreemptionPolicy:              nil,
					Overhead:                      nil,
					TopologySpreadConstraints:     nil,
				},
			},
			Strategy:                appsv1.DeploymentStrategy{},
			MinReadySeconds:         0,
			RevisionHistoryLimit:    nil,
			Paused:                  false,
			ProgressDeadlineSeconds: nil,
		},
		Status: appsv1.DeploymentStatus{},
	}

	// Create
	fmt.Println("Creating deployment: " + deployment.Name + " ...")
	newDeployment, getErr := deploymentsClient.Get(deploymentName, metav1.GetOptions{
		TypeMeta:        metav1.TypeMeta{},
		ResourceVersion: "",
	})
	if getErr != nil {
		newDeployment, err = deploymentsClient.Create(deployment)
		if err != nil {
			panic(err)
		}
		fmt.Println("Created deployment: " + deployment.Name + " ...")
	}

	util.Prompt()

	// Update
	fmt.Println("Updating deployment: " + newDeployment.Name + " ...")
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		newDeployment.Spec.Replicas = int32Ptr(3)
		newDeployment.Spec.Template.Spec.Containers[0].Image = "nginx:1.13"
		_, updateErr := deploymentsClient.Update(newDeployment)
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("updated failed: %v", retryErr))
	}
	fmt.Println("Updated deployment: " + newDeployment.Name + " ...")

	util.Prompt()

	// List
	fmt.Println("Listing deployment in namespace[" + apiv1.NamespaceDefault + "]:")
	list, err := deploymentsClient.List(metav1.ListOptions{
		TypeMeta:            metav1.TypeMeta{},
		LabelSelector:       "",
		FieldSelector:       "",
		Watch:               false,
		AllowWatchBookmarks: false,
		ResourceVersion:     "",
		TimeoutSeconds:      nil,
		Limit:               0,
		Continue:            "",
	})
	if err != nil {
		panic(fmt.Errorf("list deployment failed: %v", err))
	}
	for _, item := range list.Items {
		fmt.Printf("%s(%d replicas)\n", item.Name, *item.Spec.Replicas)
	}

	util.Prompt()

	// Delete
	fmt.Println("Deleting deployment: " + newDeployment.Name + " ...")
	deletePolicy := metav1.DeletePropagationForeground
	deleteErr := deploymentsClient.Delete(deploymentName, &metav1.DeleteOptions{
		TypeMeta:           metav1.TypeMeta{},
		GracePeriodSeconds: nil,
		Preconditions:      nil,
		OrphanDependents:   nil,
		PropagationPolicy:  &deletePolicy,
		DryRun:             nil,
	})
	if deleteErr != nil {
		panic(fmt.Errorf("deleted failed: %v", deleteErr))
	}
	fmt.Println("Deleted deployment: " + newDeployment.Name + " ...")
}

func int32Ptr(i int32) *int32 {
	return &i
}
