package main

import (
	"flag"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

	"github.com/spf13/pflag"

	cliflag "k8s-lx1036/k8s/kubelet/cmd/flag"
	"k8s-lx1036/k8s/kubelet/pkg"

	"k8s.io/klog/v2"
)

var c *pkg.KubeletConfiguration

func init() {
	c = &pkg.KubeletConfiguration{}

	klog.InitFlags(nil)
	fs := pflag.NewFlagSet("", pflag.ExitOnError)

	fs.StringSliceVar(&c.EnforceNodeAllocatable, "enforce-node-allocatable", []string{"pods"},
		"A comma separated list of levels of node allocatable enforcement to be enforced by kubelet. Acceptable options are 'none', 'pods', 'system-reserved', and 'kube-reserved'. If the latter two options are specified, '--system-reserved-cgroup' and '--kube-reserved-cgroup' must also be set, respectively. If 'none' is specified, no additional options should be set. See https://kubernetes.io/docs/tasks/administer-cluster/reserve-compute-resources/ for more details.")

	fs.Var(cliflag.NewLangleSeparatedMapStringString(&c.EvictionHard), "eviction-hard",
		"A set of eviction thresholds (e.g. memory.available<1Gi) that if met would trigger a pod eviction.")
	// DefaultEvictionHard includes default options for hard eviction.
	var DefaultEvictionHard = map[string]string{
		"memory.available":  "100Mi",
		"nodefs.available":  "10%",
		"imagefs.available": "15%",
	}
	c.EvictionHard = DefaultEvictionHard

	fs.Var(cliflag.NewLangleSeparatedMapStringString(&c.EvictionSoft), "eviction-soft",
		"A set of eviction thresholds (e.g. memory.available<1.5Gi) that if met over a corresponding grace period would trigger a pod eviction.")
	fs.Var(cliflag.NewMapStringString(&c.EvictionSoftGracePeriod), "eviction-soft-grace-period",
		"A set of eviction grace periods (e.g. memory.available=1m30s) that correspond to how long a soft eviction threshold must hold before triggering a pod eviction.")
	fs.Int32Var(&c.EvictionMaxPodGracePeriod, "eviction-max-pod-grace-period", c.EvictionMaxPodGracePeriod,
		"Maximum allowed grace period (in seconds) to use when terminating pods in response to a soft eviction threshold being met.  If negative, defer to pod specified value.")
	fs.Var(cliflag.NewMapStringString(&c.EvictionMinimumReclaim), "eviction-minimum-reclaim",
		"A set of minimum reclaims (e.g. imagefs.available=2Gi) that describes the minimum amount of resource the kubelet will reclaim when performing a pod eviction if that resource is under pressure.")

	fs.DurationVar(&c.EvictionPressureTransitionPeriod.Duration, "eviction-pressure-transition-period", c.EvictionPressureTransitionPeriod.Duration, "Duration for which the kubelet has to wait before transitioning out of an eviction pressure condition.")
	c.EvictionPressureTransitionPeriod = metav1.Duration{Duration: 5 * time.Minute}

	fs.DurationVar(&c.VolumeStatsAggPeriod.Duration, "volume-stats-agg-period", c.VolumeStatsAggPeriod.Duration,
		"Specifies interval for kubelet to calculate and cache the volume disk usage for all pods and volumes.  To disable volume calculations, set to 0.")
	c.VolumeStatsAggPeriod = metav1.Duration{Duration: 1 * time.Minute}

	fs.DurationVar(&c.ImageMinimumGCAge.Duration, "minimum-image-ttl-duration", c.ImageMinimumGCAge.Duration, "Minimum age for an unused image before it is garbage collected.  Examples: '300ms', '10s' or '2h45m'.")
	fs.Int32Var(&c.ImageGCHighThresholdPercent, "image-gc-high-threshold", c.ImageGCHighThresholdPercent, "The percent of disk usage after which image garbage collection is always run. Values must be within the range [0, 100], To disable image garbage collection, set to 100. ")
	fs.Int32Var(&c.ImageGCLowThresholdPercent, "image-gc-low-threshold", c.ImageGCLowThresholdPercent, "The percent of disk usage before which image garbage collection is never run. Lowest disk usage to garbage collect to. Values must be within the range [0, 100] and should not be larger than that of --image-gc-high-threshold.")
	c.ImageMinimumGCAge = metav1.Duration{Duration: 2 * time.Minute}
	c.ImageGCHighThresholdPercent = 85
	c.ImageGCLowThresholdPercent = 80

	fs.StringVar(&c.PodSandboxImage, "pod-infra-container-image", c.PodSandboxImage, fmt.Sprintf("The image whose network/ipc namespaces containers in each pod will use. %s", "This docker-specific flag only works when container-runtime is set to docker."))
	c.PodSandboxImage = "k8s.gcr.io/pause:3.2"

	flag.Set("logtostderr", "true")
	flag.Parse()
}

func main() {
	kubelet, err := pkg.NewMainKubelet(c)
	if err != nil {
		klog.Fatal(err)
	}

	klog.Info("Started kubelet as runonce")
	err = kubelet.RunOnce()
	if err != nil {
		klog.Fatal(err)
	}

}
