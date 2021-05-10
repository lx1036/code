package options

import (
	"time"

	"k8s-lx1036/k8s/kubelet/isolation-agent/pkg/server"

	"github.com/spf13/cobra"
)

type Options struct {
	MetricResolution time.Duration

	Kubeconfig string
	Nodename   string

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
	flags.DurationVar(&o.MetricResolution, "metric-resolution", o.MetricResolution, "The resolution at which metrics-server will retain metrics.")
	flags.StringVar(&o.Kubeconfig, "kubeconfig", o.Kubeconfig, "The path to the kubeconfig used to connect to the Kubernetes API server and the Kubelets (defaults to in-cluster config)")
	flags.StringVar(&o.Nodename, "nodename", o.Nodename, "current node name")

	// @see https://github.com/kubernetes/kubernetes/blob/release-1.19/cmd/kubelet/app/options/options.go#L186
	flags.StringVar(&o.RootDirectory, "root-dir", "/var/lib/kubelet", "Directory path for managing kubelet files (volume mounts,etc).")
	// @see https://github.com/kubernetes/kubernetes/blob/release-1.19/pkg/kubelet/apis/config/fuzzer/fuzzer.go#L94-L95
	flags.BoolVar(&o.CgroupsPerQOS, "cgroups-per-qos", true, "Enable creation of QoS cgroup hierarchy, if true top level QoS and pod cgroups are created.")
	flags.StringVar(&o.CgroupDriver, "cgroup-driver", "cgroupfs", "Driver that the kubelet uses to manipulate cgroups on the host.  Possible values: 'cgroupfs', 'systemd'")
	flags.StringVar(&o.CgroupRoot, "cgroup-root", "/", "Optional root cgroup to use for pods. This is handled by the container runtime on a best effort basis. Default: '', which means use the container runtime default.")
}

func (o *Options) ServerConfig() (*server.Config, error) {

}

func NewOptions() *Options {
	o := &Options{
		//SecureServing:  genericoptions.NewSecureServingOptions().WithLoopback(),
		//Authentication: genericoptions.NewDelegatingAuthenticationOptions(),
		//Authorization:  genericoptions.NewDelegatingAuthorizationOptions(),
		//Features:       genericoptions.NewFeatureOptions(),
		MetricResolution: 60 * time.Second,
		//KubeletPort:                  10250,
	}

	return o
}
