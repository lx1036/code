package sparkapplication

import (
	"fmt"
	v1 "k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/apis/sparkoperator.k9s.io/v1"
	"k8s-lx1036/k8s/bigdata/spark-on-k8s/spark-operator/pkg/config"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/apis/policy"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	sparkHomeEnvVar             = "SPARK_HOME"
	kubernetesServiceHostEnvVar = "KUBERNETES_SERVICE_HOST"
	kubernetesServicePortEnvVar = "KUBERNETES_SERVICE_PORT"
)

const (
	podAlreadyExistsErrorCode = "code=409"
)

// submission includes information of a Spark application to be submitted.
type submission struct {
	namespace string
	name      string
	args      []string
}

func newSubmission(args []string, app *v1.SparkApplication) *submission {
	return &submission{
		namespace: app.Namespace,
		name:      app.Name,
		args:      args,
	}
}

func buildSubmissionCommandArgs(app *v1.SparkApplication, driverPodName string, submissionID string) ([]string, error) {
	var args []string
	if app.Spec.MainClass != nil {
		args = append(args, "--class", *app.Spec.MainClass)
	}
	masterURL, err := getMasterURL()
	if err != nil {
		return nil, err
	}

	args = append(args, "--master", masterURL)
	args = append(args, "--deploy-mode", string(app.Spec.Mode))

	// Add proxy user
	if app.Spec.ProxyUser != nil {
		args = append(args, "--proxy-user", *app.Spec.ProxyUser)
	}

	args = append(args, "--conf", fmt.Sprintf("%s=%s", config.SparkAppNamespaceKey, app.Namespace))
	args = append(args, "--conf", fmt.Sprintf("%s=%s", config.SparkAppNameKey, app.Name))
	args = append(args, "--conf", fmt.Sprintf("%s=%s", config.SparkDriverPodNameKey, driverPodName))

	// TODO: ignore application dependencies conf

	if app.Spec.Image != nil {
		args = append(args, "--conf",
			fmt.Sprintf("%s=%s", config.SparkContainerImageKey, *app.Spec.Image))
	}
	if app.Spec.ImagePullPolicy != nil {
		args = append(args, "--conf",
			fmt.Sprintf("%s=%s", config.SparkContainerImagePullPolicyKey, *app.Spec.ImagePullPolicy))
	}
	if len(app.Spec.ImagePullSecrets) > 0 {
		secretNames := strings.Join(app.Spec.ImagePullSecrets, ",")
		args = append(args, "--conf", fmt.Sprintf("%s=%s", config.SparkImagePullSecretKey, secretNames))
	}
	if app.Spec.PythonVersion != nil {
		args = append(args, "--conf",
			fmt.Sprintf("%s=%s", config.SparkPythonVersion, *app.Spec.PythonVersion))
	}
	if app.Spec.MemoryOverheadFactor != nil {
		args = append(args, "--conf",
			fmt.Sprintf("%s=%s", config.SparkMemoryOverheadFactor, *app.Spec.MemoryOverheadFactor))
	}

	// Operator triggered spark-submit should never wait for App completion
	args = append(args, "--conf", fmt.Sprintf("%s=false", config.SparkWaitAppCompletion))

	// Add Spark configuration properties.
	for key, value := range app.Spec.SparkConf {
		// Configuration property for the driver pod name has already been set.
		if key != config.SparkDriverPodNameKey {
			args = append(args, "--conf", fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Add Hadoop configuration properties.
	for key, value := range app.Spec.HadoopConf {
		args = append(args, "--conf", fmt.Sprintf("spark.hadoop.%s=%s", key, value))
	}

	// TODO: ignore dynamic allocation conf

	// Add the driver and executor configuration options.
	// Note that when the controller submits the application, it expects that all dependencies are local
	// so init-container is not needed and therefore no init-container image needs to be specified.
	options, err := addDriverConfOptions(app, submissionID)
	if err != nil {
		return nil, err
	}
	for _, option := range options {
		args = append(args, "--conf", option)
	}

	options, err = addExecutorConfOptions(app, submissionID)
	if err != nil {
		return nil, err
	}
	for _, option := range options {
		args = append(args, "--conf", option)
	}

	for key, value := range app.Spec.NodeSelector {
		conf := fmt.Sprintf("%s%s=%s", config.SparkNodeSelectorKeyPrefix, key, value)
		args = append(args, "--conf", conf)
	}

	if app.Spec.Volumes != nil {
		options, err = addLocalDirConfOptions(app)
		if err != nil {
			return nil, err
		}

		for _, option := range options {
			args = append(args, "--conf", option)
		}
	}

	if app.Spec.MainApplicationFile != nil {
		// Add the main application file if it is present.
		args = append(args, *app.Spec.MainApplicationFile)
	}

	// Add application arguments.
	for _, argument := range app.Spec.Arguments {
		args = append(args, argument)
	}

	return args, nil
}

func getMasterURL() (string, error) {
	kubernetesServiceHost := os.Getenv(kubernetesServiceHostEnvVar)
	if kubernetesServiceHost == "" {
		return "", fmt.Errorf("environment variable %s is not found", kubernetesServiceHostEnvVar)
	}
	kubernetesServicePort := os.Getenv(kubernetesServicePortEnvVar)
	if kubernetesServicePort == "" {
		return "", fmt.Errorf("environment variable %s is not found", kubernetesServicePortEnvVar)
	}

	return fmt.Sprintf("k8s://https://%s:%s", kubernetesServiceHost, kubernetesServicePort), nil
}

// addLocalDirConfOptions excludes local dir volumes, update SparkApplication and returns local dir config options
func addLocalDirConfOptions(app *v1.SparkApplication) ([]string, error) {
	var localDirConfOptions []string

	sparkLocalVolumes := map[string]corev1.Volume{}
	var mutateVolumes []corev1.Volume
	// Filter local dir volumes
	for _, volume := range app.Spec.Volumes {
		if strings.HasPrefix(volume.Name, config.SparkLocalDirVolumePrefix) {
			sparkLocalVolumes[volume.Name] = volume // 带有 "spark-local-dir-" 前缀的 volume，为 spark-local volume
		} else {
			mutateVolumes = append(mutateVolumes, volume)
		}
	}

	app.Spec.Volumes = mutateVolumes

	// Filter local dir volumeMounts and set mutate volume mounts to driver and executor
	if app.Spec.Driver.VolumeMounts != nil {
		driverMutateVolumeMounts, driverLocalDirConfConfOptions := filterMutateMountVolumes(app.Spec.Driver.VolumeMounts,
			config.SparkDriverVolumesPrefix, sparkLocalVolumes)
		app.Spec.Driver.VolumeMounts = driverMutateVolumeMounts                             // driver 自己的 volume
		localDirConfOptions = append(localDirConfOptions, driverLocalDirConfConfOptions...) // spark-local 的 volume
	}

	if app.Spec.Executor.VolumeMounts != nil {
		executorMutateVolumeMounts, executorLocalDirConfConfOptions := filterMutateMountVolumes(app.Spec.Executor.VolumeMounts, config.SparkExecutorVolumesPrefix, sparkLocalVolumes)
		app.Spec.Executor.VolumeMounts = executorMutateVolumeMounts
		localDirConfOptions = append(localDirConfOptions, executorLocalDirConfConfOptions...)
	}

	return localDirConfOptions, nil
}

// INFO: 看看 driver/executor volume mount 里哪些是 spark-local volume，哪些是自己的 volume。愚蠢的设计！！！
func filterMutateMountVolumes(volumeMounts []corev1.VolumeMount, prefix string,
	sparkLocalVolumes map[string]corev1.Volume) ([]corev1.VolumeMount, []string) {

	var mutateMountVolumes []corev1.VolumeMount
	var localDirConfOptions []string
	for _, volumeMount := range volumeMounts {
		if volume, ok := sparkLocalVolumes[volumeMount.Name]; ok { // 是 spark-local volume
			options := buildLocalVolumeOptions(prefix, volume, volumeMount)
			for _, option := range options {
				localDirConfOptions = append(localDirConfOptions, option)
			}
		} else {
			mutateMountVolumes = append(mutateMountVolumes, volumeMount)
		}
	}

	return mutateMountVolumes, localDirConfOptions
}

func buildLocalVolumeOptions(prefix string, volume corev1.Volume, volumeMount corev1.VolumeMount) []string {
	VolumeMountPathTemplate := prefix + "%s.%s.mount.path=%s"
	VolumeMountOptionTemplate := prefix + "%s.%s.options.%s=%s"

	var options []string
	switch {
	case volume.HostPath != nil:
		options = append(options, fmt.Sprintf(VolumeMountPathTemplate, string(policy.HostPath), volume.Name, volumeMount.MountPath))
		options = append(options, fmt.Sprintf(VolumeMountOptionTemplate, string(policy.HostPath), volume.Name, "path", volume.HostPath.Path))
		if volume.HostPath.Type != nil {
			options = append(options, fmt.Sprintf(VolumeMountOptionTemplate, string(policy.HostPath), volume.Name, "type", *volume.HostPath.Type))
		}
	case volume.EmptyDir != nil:
		options = append(options, fmt.Sprintf(VolumeMountPathTemplate, string(policy.EmptyDir), volume.Name, volumeMount.MountPath))
	case volume.PersistentVolumeClaim != nil:
		options = append(options, fmt.Sprintf(VolumeMountPathTemplate, string(policy.PersistentVolumeClaim), volume.Name, volumeMount.MountPath))
		options = append(options, fmt.Sprintf(VolumeMountOptionTemplate, string(policy.PersistentVolumeClaim), volume.Name, "claimName", volume.PersistentVolumeClaim.ClaimName))
	}

	return options
}

func addDriverConfOptions(app *v1.SparkApplication, submissionID string) ([]string, error) {
	var driverConfOptions []string

	driverConfOptions = append(driverConfOptions,
		fmt.Sprintf("%s%s=%s", config.SparkDriverLabelKeyPrefix, config.SparkAppNameLabel, app.Name))
	driverConfOptions = append(driverConfOptions,
		fmt.Sprintf("%s%s=%s", config.SparkDriverLabelKeyPrefix, config.LaunchedBySparkOperatorLabel, "true"))
	driverConfOptions = append(driverConfOptions,
		fmt.Sprintf("%s%s=%s", config.SparkDriverLabelKeyPrefix, config.SubmissionIDLabel, submissionID))

	if app.Spec.Driver.Image != nil {
		driverConfOptions = append(driverConfOptions,
			fmt.Sprintf("%s=%s", config.SparkDriverContainerImageKey, *app.Spec.Driver.Image))
	}

	if app.Spec.Driver.Cores != nil {
		driverConfOptions = append(driverConfOptions,
			fmt.Sprintf("spark.driver.cores=%d", *app.Spec.Driver.Cores))
	}
	if app.Spec.Driver.CoreRequest != nil {
		driverConfOptions = append(driverConfOptions,
			fmt.Sprintf("%s=%s", config.SparkDriverCoreRequestKey, *app.Spec.Driver.CoreRequest))
	}
	if app.Spec.Driver.CoreLimit != nil {
		driverConfOptions = append(driverConfOptions,
			fmt.Sprintf("%s=%s", config.SparkDriverCoreLimitKey, *app.Spec.Driver.CoreLimit))
	}
	if app.Spec.Driver.Memory != nil {
		driverConfOptions = append(driverConfOptions,
			fmt.Sprintf("spark.driver.memory=%s", *app.Spec.Driver.Memory))
	}
	if app.Spec.Driver.MemoryOverhead != nil {
		driverConfOptions = append(driverConfOptions,
			fmt.Sprintf("spark.driver.memoryOverhead=%s", *app.Spec.Driver.MemoryOverhead))
	}

	if app.Spec.Driver.ServiceAccount != nil {
		driverConfOptions = append(driverConfOptions,
			fmt.Sprintf("%s=%s", config.SparkDriverServiceAccountName, *app.Spec.Driver.ServiceAccount))
	}

	if app.Spec.Driver.JavaOptions != nil {
		driverConfOptions = append(driverConfOptions,
			fmt.Sprintf("%s=%s", config.SparkDriverJavaOptions, *app.Spec.Driver.JavaOptions))
	}

	if app.Spec.Driver.KubernetesMaster != nil {
		driverConfOptions = append(driverConfOptions,
			fmt.Sprintf("%s=%s", config.SparkDriverKubernetesMaster, *app.Spec.Driver.KubernetesMaster))
	}

	//Populate SparkApplication Labels to Driver
	driverLabels := make(map[string]string)
	for key, value := range app.Labels {
		driverLabels[key] = value
	}
	for key, value := range app.Spec.Driver.Labels {
		driverLabels[key] = value
	}

	for key, value := range driverLabels {
		driverConfOptions = append(driverConfOptions,
			fmt.Sprintf("%s%s=%s", config.SparkDriverLabelKeyPrefix, key, value))
	}

	for key, value := range app.Spec.Driver.Annotations {
		driverConfOptions = append(driverConfOptions,
			fmt.Sprintf("%s%s=%s", config.SparkDriverAnnotationKeyPrefix, key, value))
	}

	/*for key, value := range app.Spec.Driver.EnvSecretKeyRefs {
		driverConfOptions = append(driverConfOptions,
			fmt.Sprintf("%s%s=%s:%s", config.SparkDriverSecretKeyRefKeyPrefix, key, value.Name, value.Key))
	}*/

	for key, value := range app.Spec.Driver.ServiceAnnotations {
		driverConfOptions = append(driverConfOptions,
			fmt.Sprintf("%s%s=%s", config.SparkDriverServiceAnnotationKeyPrefix, key, value))
	}

	//driverConfOptions = append(driverConfOptions, config.GetDriverSecretConfOptions(app)...)
	//driverConfOptions = append(driverConfOptions, config.GetDriverEnvVarConfOptions(app)...)

	return driverConfOptions, nil
}

func addExecutorConfOptions(app *v1.SparkApplication, submissionID string) ([]string, error) {
	var executorConfOptions []string

	executorConfOptions = append(executorConfOptions,
		fmt.Sprintf("%s%s=%s", config.SparkExecutorLabelKeyPrefix, config.SparkAppNameLabel, app.Name))
	executorConfOptions = append(executorConfOptions,
		fmt.Sprintf("%s%s=%s", config.SparkExecutorLabelKeyPrefix, config.LaunchedBySparkOperatorLabel, "true"))
	executorConfOptions = append(executorConfOptions,
		fmt.Sprintf("%s%s=%s", config.SparkExecutorLabelKeyPrefix, config.SubmissionIDLabel, submissionID))

	if app.Spec.Executor.Instances != nil {
		conf := fmt.Sprintf("spark.executor.instances=%d", *app.Spec.Executor.Instances)
		executorConfOptions = append(executorConfOptions, conf)
	}

	if app.Spec.Executor.Image != nil {
		executorConfOptions = append(executorConfOptions,
			fmt.Sprintf("%s=%s", config.SparkExecutorContainerImageKey, *app.Spec.Executor.Image))
	}

	if app.Spec.Executor.Cores != nil {
		// Property "spark.executor.cores" does not allow float values.
		executorConfOptions = append(executorConfOptions,
			fmt.Sprintf("spark.executor.cores=%d", int32(*app.Spec.Executor.Cores)))
	}
	if app.Spec.Executor.CoreRequest != nil {
		executorConfOptions = append(executorConfOptions,
			fmt.Sprintf("%s=%s", config.SparkExecutorCoreRequestKey, *app.Spec.Executor.CoreRequest))
	}
	if app.Spec.Executor.CoreLimit != nil {
		executorConfOptions = append(executorConfOptions,
			fmt.Sprintf("%s=%s", config.SparkExecutorCoreLimitKey, *app.Spec.Executor.CoreLimit))
	}
	if app.Spec.Executor.Memory != nil {
		executorConfOptions = append(executorConfOptions,
			fmt.Sprintf("spark.executor.memory=%s", *app.Spec.Executor.Memory))
	}
	if app.Spec.Executor.MemoryOverhead != nil {
		executorConfOptions = append(executorConfOptions,
			fmt.Sprintf("spark.executor.memoryOverhead=%s", *app.Spec.Executor.MemoryOverhead))
	}

	if app.Spec.Executor.DeleteOnTermination != nil {
		executorConfOptions = append(executorConfOptions,
			fmt.Sprintf("%s=%t", config.SparkExecutorDeleteOnTermination, *app.Spec.Executor.DeleteOnTermination))
	}

	//Populate SparkApplication Labels to Executors
	executorLabels := make(map[string]string)
	for key, value := range app.Labels {
		executorLabels[key] = value
	}
	for key, value := range app.Spec.Executor.Labels {
		executorLabels[key] = value
	}
	for key, value := range executorLabels {
		executorConfOptions = append(executorConfOptions,
			fmt.Sprintf("%s%s=%s", config.SparkExecutorLabelKeyPrefix, key, value))
	}

	for key, value := range app.Spec.Executor.Annotations {
		executorConfOptions = append(executorConfOptions,
			fmt.Sprintf("%s%s=%s", config.SparkExecutorAnnotationKeyPrefix, key, value))
	}

	/*for key, value := range app.Spec.Executor.EnvSecretKeyRefs {
		executorConfOptions = append(executorConfOptions,
			fmt.Sprintf("%s%s=%s:%s", config.SparkExecutorSecretKeyRefKeyPrefix, key, value.Name, value.Key))
	}*/

	if app.Spec.Executor.JavaOptions != nil {
		executorConfOptions = append(executorConfOptions,
			fmt.Sprintf("%s=%s", config.SparkExecutorJavaOptions, *app.Spec.Executor.JavaOptions))
	}

	//executorConfOptions = append(executorConfOptions, config.GetExecutorSecretConfOptions(app)...)
	//executorConfOptions = append(executorConfOptions, config.GetExecutorEnvVarConfOptions(app)...)

	return executorConfOptions, nil
}

/*
INFO: ` /opt/spark/bin/spark-submit --class xxx --master xxx --deploy-mode cluster
	--conf spark.kubernetes.namespace=xxx --conf spark.app.name=xxx --conf spark.kubernetes.driver.pod.name=xxx
	--conf ...
`
*/
func runSparkSubmit(submission *submission) (bool, error) {
	// INFO: 在 spark-operator pod 里 path: /opt/spark/bin/spark-submit
	sparkHome, present := os.LookupEnv(sparkHomeEnvVar)
	if !present {
		klog.Error("SPARK_HOME is not specified")
	}
	var command = filepath.Join(sparkHome, "/bin/spark-submit")

	klog.V(2).Info(fmt.Sprintf("[runSparkSubmit]%s %s", command, strings.Join(submission.args, " ")))
	output, err := exec.Command(command, submission.args...).Output()
	if err != nil {
		var errorMsg string
		if exitErr, ok := err.(*exec.ExitError); ok {
			errorMsg = string(exitErr.Stderr)
		}
		// INFO: 这种情况是由于state发生了变化，重新submit了一次
		// The driver pod of the application already exists.
		if strings.Contains(errorMsg, podAlreadyExistsErrorCode) {
			klog.Warningf("trying to resubmit an already submitted SparkApplication %s/%s", submission.namespace, submission.name)
			return false, nil
		}

		if errorMsg != "" {
			return false, fmt.Errorf("failed to run spark-submit for SparkApplication %s/%s: %s", submission.namespace, submission.name, errorMsg)
		}
		return false, fmt.Errorf("failed to run spark-submit for SparkApplication %s/%s: %v", submission.namespace, submission.name, err)
	}

	klog.V(2).Info(fmt.Sprintf("[runSparkSubmit]output: %s", string(output)))
	return true, nil
}
