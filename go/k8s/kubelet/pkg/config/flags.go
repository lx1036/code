package config

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// ContainerRuntimeOptions defines options for the container runtime.
type ContainerRuntimeOptions struct {
	// General Options.

	// ContainerRuntime is the container runtime to use.
	ContainerRuntime string
	// RuntimeCgroups that container runtime is expected to be isolated in.
	RuntimeCgroups string

	// Docker-specific options.

	// DockershimRootDirectory is the path to the dockershim root directory. Defaults to
	// /var/lib/dockershim if unset. Exposed for integration testing (e.g. in OpenShift).
	DockershimRootDirectory string
	// PodSandboxImage is the image whose network/ipc namespaces
	// containers in each pod will use.
	PodSandboxImage string
	// DockerEndpoint is the path to the docker endpoint to communicate with.
	DockerEndpoint string
	// If no pulling progress is made before the deadline imagePullProgressDeadline,
	// the image pulling will be cancelled. Defaults to 1m0s.
	// +optional
	ImagePullProgressDeadline metav1.Duration

	// Network plugin options.

	// networkPluginName is the name of the network plugin to be invoked for
	// various events in kubelet/pod lifecycle
	NetworkPluginName string
	// NetworkPluginMTU is the MTU to be passed to the network plugin,
	// and overrides the default MTU for cases where it cannot be automatically
	// computed (such as IPSEC).
	NetworkPluginMTU int32
	// CNIConfDir is the full path of the directory in which to search for
	// CNI config files
	CNIConfDir string
	// CNIBinDir is the full path of the directory in which to search for
	// CNI plugin binaries
	CNIBinDir string
	// CNICacheDir is the full path of the directory in which CNI should store
	// cache files
	CNICacheDir string

	// Image credential provider plugin options

	// ImageCredentialProviderConfigFile is the path to the credential provider plugin config file.
	// This config file is a specification for what credential providers are enabled and invokved
	// by the kubelet. The plugin config should contain information about what plugin binary
	// to execute and what container images the plugin should be called for.
	// +optional
	ImageCredentialProviderConfigFile string
	// ImageCredentialProviderBinDir is the path to the directory where credential provider plugin
	// binaries exist. The name of each plugin binary is expected to match the name of the plugin
	// specified in imageCredentialProviderConfigFile.
	// +optional
	ImageCredentialProviderBinDir string
}
