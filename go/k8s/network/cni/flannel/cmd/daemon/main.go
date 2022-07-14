package main

import (
	"flag"
	"fmt"
	"k8s-lx1036/k8s/network/cni/flannel/pkg/backend"
	"k8s-lx1036/k8s/network/cni/flannel/pkg/subnet"
	"k8s.io/klog/v2"

	"k8s.io/component-base/logs"
)

var (
	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file (only needed when running outside of k8s)")
	subnetFile = flag.String("subnet-file", "/run/flannel/subnet.env", "filename where env variables (subnet, MTU, ... ) will be written to")
	ipMasq     = flag.Bool("ip-masq", false, "setup IP masquerade rule for traffic destined outside of overlay network")
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	flag.Parse()

	subnetMgr, err := subnet.NewSubnetManager(ctx, opts.kubeApiUrl, opts.kubeConfigFile,
		opts.kubeAnnotationPrefix, opts.netConfPath, opts.setNodeNetworkUnavailable)

	config, err := subnetMgr.GetNetworkConfig(ctx)

	// Create a backend manager then use it to create the backend and register the network with it.
	backendMgr := backend.NewManager(ctx, sm, extIface)
	be, err := backendMgr.GetBackend(config.BackendType)
	if err != nil {

	}
	backendNetwork, err := be.RegisterNetwork(ctx, &wg, config)
	if err != nil {

	}
	go backendNetwork.Run(ctx)

	if err := WriteSubnetFile(*subnetFile, config, *ipMasq, backendNetwork); err != nil {
		// Continue, even though it failed.
		klog.Warningf(fmt.Sprintf("Failed to write subnet file: %v", err))
	} else {
		klog.Infof(fmt.Sprintf("Wrote subnet file to %s", *subnetFile))
	}

}
