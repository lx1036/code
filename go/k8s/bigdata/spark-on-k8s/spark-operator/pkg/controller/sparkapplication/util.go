package sparkapplication

import (
	"encoding/json"
	"fmt"

	v1 "k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/apis/sparkoperator.k9s.io/v1"
	"k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/config"

	corev1 "k8s.io/api/core/v1"
)

func getDriverPodName(app *v1.SparkApplication) string {
	name := app.Spec.Driver.PodName
	if name != nil && len(*name) > 0 {
		return *name
	}

	sparkConf := app.Spec.SparkConf
	if sparkConf[config.SparkDriverPodNameKey] != "" {
		return sparkConf[config.SparkDriverPodNameKey]
	}

	return fmt.Sprintf("%s-driver", app.Name)
}

// INFO: 这个函数可以复用
func printStatus(status *v1.SparkApplicationStatus) (string, error) {
	marshalled, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return "", err
	}
	return string(marshalled), nil
}

func podPhaseToDriverState(podStatus corev1.PodStatus) v1.DriverState {
	switch podStatus.Phase {
	case corev1.PodPending:
		return v1.DriverPendingState
	case corev1.PodRunning:
		state := getDriverContainerTerminatedState(podStatus)
		if state != nil {
			if state.ExitCode == 0 {
				return v1.DriverCompletedState
			}
			return v1.DriverFailedState
		}
		return v1.DriverRunningState
	case corev1.PodSucceeded:
		return v1.DriverCompletedState
	case corev1.PodFailed:
		state := getDriverContainerTerminatedState(podStatus)
		if state != nil && state.ExitCode == 0 {
			return v1.DriverCompletedState
		}
		return v1.DriverFailedState
	default:
		return v1.DriverUnknownState
	}
}

// INFO: 如果 pod template name没设置，driver name 默认是 spark-kubernetes-driver, executor 默认是 spark-kubernetes-executor
//  @see https://spark.apache.org/docs/latest/running-on-kubernetes.html#container-spec
func getDriverContainerTerminatedState(podStatus corev1.PodStatus) *corev1.ContainerStateTerminated {
	for _, c := range podStatus.ContainerStatuses {
		if c.Name == config.SparkDriverContainerName {
			if c.State.Terminated != nil {
				return c.State.Terminated
			}
			return nil
		}
	}
	return nil
}

func getSparkApplicationID(pod *corev1.Pod) string {
	return pod.Labels[config.SparkApplicationSelectorLabel]
}

func hasDriverTerminated(driverState v1.DriverState) bool {
	return driverState == v1.DriverCompletedState || driverState == v1.DriverFailedState
}

func driverStateToApplicationState(driverState v1.DriverState) v1.ApplicationStateType {
	switch driverState {
	case v1.DriverPendingState:
		return v1.SubmittedState
	case v1.DriverCompletedState:
		return v1.SucceedingState
	case v1.DriverFailedState:
		return v1.FailingState
	case v1.DriverRunningState:
		return v1.RunningState
	default:
		return v1.UnknownState
	}
}

func isDriverRunning(app *v1.SparkApplication) bool {
	return app.Status.AppState.State == v1.RunningState
}

// IsExecutorPod returns whether the given pod is a Spark executor Pod.
func IsExecutorPod(pod *corev1.Pod) bool {
	return pod.Labels[config.SparkRoleLabel] == config.SparkExecutorRole
}

func getResourceLabels(app *v1.SparkApplication) map[string]string {
	labels := map[string]string{config.SparkAppNameLabel: app.Name}
	if app.Status.SubmissionID != "" {
		labels[config.SubmissionIDLabel] = app.Status.SubmissionID
	}
	return labels
}

func podPhaseToExecutorState(podPhase corev1.PodPhase) v1.ExecutorState {
	switch podPhase {
	case corev1.PodPending:
		return v1.ExecutorPendingState
	case corev1.PodRunning:
		return v1.ExecutorRunningState
	case corev1.PodSucceeded:
		return v1.ExecutorCompletedState
	case corev1.PodFailed:
		return v1.ExecutorFailedState
	default:
		return v1.ExecutorUnknownState
	}
}

func isExecutorTerminated(executorState v1.ExecutorState) bool {
	return executorState == v1.ExecutorCompletedState || executorState == v1.ExecutorFailedState
}

// ShouldRetry determines if SparkApplication in a given state should be retried.
// INFO: https://github.com/GoogleCloudPlatform/spark-on-k8s-operator/blob/master/docs/user-guide.md#configuring-automatic-application-restart-and-failure-handling
func shouldRetry(app *v1.SparkApplication) bool {
	switch app.Status.AppState.State {
	case v1.SucceedingState:
		return app.Spec.RestartPolicy.Type == v1.Always
	case v1.FailingState:
		if app.Spec.RestartPolicy.Type == v1.Always {
			return true
		} else if app.Spec.RestartPolicy.Type == v1.OnFailure {
			// We retry if we haven't hit the retry limit.
			if app.Spec.RestartPolicy.OnFailureRetries != nil && app.Status.ExecutionAttempts <= *app.Spec.RestartPolicy.OnFailureRetries {
				return true
			}
		}
	case v1.FailedSubmissionState:
		if app.Spec.RestartPolicy.Type == v1.Always {
			return true
		} else if app.Spec.RestartPolicy.Type == v1.OnFailure {
			// We retry if we haven't hit the retry limit.
			if app.Spec.RestartPolicy.OnSubmissionFailureRetries != nil && app.Status.SubmissionAttempts <= *app.Spec.RestartPolicy.OnSubmissionFailureRetries {
				return true
			}
		}
	}
	return false
}

func int64ptr(n int64) *int64 {
	return &n
}
