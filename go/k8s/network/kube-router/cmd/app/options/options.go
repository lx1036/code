package options

import (
	"github.com/spf13/cobra"
	apiv1 "k8s.io/api/core/v1"
)

type Options struct {
	// path to the kubeconfig used to connect to the Kubernetes API server
	Kubeconfig string

	Namespace string

	ControllerThreads int

	EnableOverlays          bool
	EnablePodEgress         bool
	AutoMTU                 bool
	AdvertiseClusterIP      bool
	AdvertiseExternalIP     bool
	AdvertiseLoadBalancerIP bool
}

func (o *Options) Flags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.StringVar(&o.Kubeconfig, "kubeconfig", o.Kubeconfig, "The path to the kubeconfig used to connect to the Kubernetes API server and the Kubelets (defaults to in-cluster config)")
	flags.StringVar(&o.Namespace, "namespace", apiv1.NamespaceAll, "The Kubernetes namespace to manage. Will manage custom resource objects of the managed CRD types for the whole cluster if unset.")
	flags.IntVar(&o.ControllerThreads, "controller-threads", 1, "Number of worker threads used by the SparkApplication controller.")

	flags.BoolVar(&o.EnableOverlays, "enable-overlay", true, `When enable-overlay is set to true,
		IP-in-IP tunneling is used for pod-to-pod networking across
		nodes in different subnets. When set to false no tunneling is used and routing infrastructure is
		expected to route traffic for pod-to-pod networking across nodes in different subnets`)
	flags.BoolVar(&o.EnablePodEgress, "enable-pod-egress", true, "SNAT traffic from Pods to destinations outside the cluster.")
	flags.BoolVar(&o.AutoMTU, "auto-mtu", true, "Auto detect and set the largest possible MTU for pod interfaces.")

	flags.BoolVar(&o.AdvertiseClusterIP, "advertise-cluster-ip", true,
		"Add Cluster IP of the service to the RIB so that it gets advertises to the BGP peers.")
	flags.BoolVar(&o.AdvertiseExternalIP, "advertise-external-ip", true,
		"Add External IP of service to the RIB so that it gets advertised to the BGP peers.")
	flags.BoolVar(&o.AdvertiseLoadBalancerIP, "advertise-loadbalancer-ip", true,
		"Add LoadbBalancer IP of service status as set by the LB provider to the RIB so that it gets advertised to the BGP peers.")
}

func NewOptions() *Options {
	return &Options{}
}
