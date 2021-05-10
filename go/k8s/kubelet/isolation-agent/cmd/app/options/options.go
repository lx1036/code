package options

import (
	"time"

	"github.com/spf13/cobra"

	kubetypes "k8s-lx1036/k8s/kubelet/pkg/types"
)

type Options struct {
	// path to the kubeconfig used to connect to the Kubernetes API server
	Kubeconfig string
	// current node name
	Nodename string

	Debug bool

	// reconciliation period to calculate cpuset and update container cpuset
	ReconcilePeriod time.Duration
	// ContainerRuntime is the container runtime to use.
	ContainerRuntime string
	// remoteRuntimeEndpoint is the endpoint of remote runtime service
	RemoteRuntimeEndpoint string
	// runtimeRequestTimeout is the timeout for all runtime requests except long running
	// requests - pull, logs, exec and attach.
	RuntimeRequestTimeout time.Duration
	// rootDirectory is the directory path to place kubelet files (volume
	// mounts,etc).
	RootDirectory string
	// CgroupRoot is the root cgroup to use for pods.
	// If CgroupsPerQOS is enabled, this is the root of the QoS cgroup hierarchy.
	CgroupRoot string
	// Enable QoS based Cgroup hierarchy: top level cgroups for QoS Classes
	// And all Burstable and BestEffort pods are brought up under their
	// specific top level QoS cgroup.
	CgroupsPerQOS bool
	// driver that the kubelet uses to manipulate cgroups on the host (cgroupfs or systemd)
	CgroupDriver string
}

func (o *Options) Flags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.BoolVar(&o.Debug, "debug", false, "debug for skip updating container cpuset")
	flags.StringVar(&o.Kubeconfig, "kubeconfig", o.Kubeconfig, "The path to the kubeconfig used to connect to the Kubernetes API server and the Kubelets (defaults to in-cluster config)")
	cobra.MarkFlagRequired(flags, "kubeconfig")
	flags.StringVar(&o.Nodename, "nodename", o.Nodename, "current node name")
	cobra.MarkFlagRequired(flags, "nodename")

	// @see https://github.com/kubernetes/kubernetes/blob/release-1.19/pkg/kubelet/apis/config/fuzzer/fuzzer.go#L66-L70
	flags.DurationVar(&o.ReconcilePeriod, "reconcile-period", 10*time.Second, "The period at which agent will calculate cpuset and update container cpuset.")
	// @see https://github.com/kubernetes/kubernetes/blob/release-1.19/cmd/kubelet/app/options/options.go#L186
	flags.StringVar(&o.RootDirectory, "root-dir", "/var/lib/kubelet", "Directory path for managing kubelet files (volume mounts,etc).")
	// @see https://github.com/kubernetes/kubernetes/blob/release-1.19/cmd/kubelet/app/options/container_runtime.go#L48
	flags.StringVar(&o.ContainerRuntime, "container-runtime", kubetypes.DockerContainerRuntime, "The container runtime to use. Possible values: 'docker', 'remote'.")
	// @see https://github.com/kubernetes/kubernetes/blob/release-1.19/cmd/kubelet/app/options/options.go#L174-L200
	flags.StringVar(&o.RemoteRuntimeEndpoint, "container-runtime-endpoint", "unix:///var/run/dockershim.sock", "[Experimental] The endpoint of remote runtime service. Currently unix socket endpoint is supported on Linux, while npipe and tcp endpoints are supported on windows.  Examples:'unix:///var/run/dockershim.sock', 'npipe:////./pipe/dockershim'")
	// @see https://github.com/kubernetes/kubernetes/blob/release-1.19/pkg/kubelet/apis/config/fuzzer/fuzzer.go#L49
	flags.DurationVar(&o.RuntimeRequestTimeout, "runtime-request-timeout", 2*time.Minute, "Timeout of all runtime requests except long running request - pull, logs, exec and attach. When timeout exceeded, kubelet will cancel the request, throw out an error and retry later.")
	// @see https://github.com/kubernetes/kubernetes/blob/release-1.19/pkg/kubelet/apis/config/fuzzer/fuzzer.go#L94-L95
	flags.BoolVar(&o.CgroupsPerQOS, "cgroups-per-qos", true, "Enable creation of QoS cgroup hierarchy, if true top level QoS and pod cgroups are created.")
	flags.StringVar(&o.CgroupDriver, "cgroup-driver", "cgroupfs", "Driver that the kubelet uses to manipulate cgroups on the host.  Possible values: 'cgroupfs', 'systemd'")
	flags.StringVar(&o.CgroupRoot, "cgroup-root", "/", "Optional root cgroup to use for pods. This is handled by the container runtime on a best effort basis. Default: '', which means use the container runtime default.")
}

func NewOptions() *Options {
	return &Options{}
}
