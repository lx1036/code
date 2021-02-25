package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"k8s-lx1036/k8s/storage/csi/csi-lib-utils/leaderelection"
	"k8s-lx1036/k8s/storage/csi/external-resizer/pkg/controller"
	"k8s-lx1036/k8s/storage/csi/external-resizer/pkg/csi"
	"k8s-lx1036/k8s/storage/csi/external-resizer/pkg/resizer"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

var (
	master       = flag.String("master", "", "Master URL to build a client config from. Either this or kubeconfig needs to be set if the provisioner is being run out of cluster.")
	kubeConfig   = flag.String("kubeconfig", "", "Absolute path to the kubeconfig")
	resyncPeriod = flag.Duration("resync-period", time.Minute*10, "Resync period for cache")
	workers      = flag.Int("workers", 10, "Concurrency to process multiple resize requests")

	csiAddress = flag.String("csi-address", "/run/csi/socket", "Address of the CSI driver socket.")
	timeout    = flag.Duration("timeout", 10*time.Second, "Timeout for waiting for CSI driver socket.")
)

func main() {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")
	flag.Parse()

	var config *rest.Config
	var err error
	if *master != "" || *kubeConfig != "" {
		config, err = clientcmd.BuildConfigFromFlags(*master, *kubeConfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		klog.Fatal(err.Error())
	}
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatal(err.Error())
	}

	informerFactory := informers.NewSharedInformerFactory(kubeClient, *resyncPeriod)

	csiClient, err := csi.New(*csiAddress, *timeout)
	if err != nil {
		klog.Fatal(err.Error())
	}

	driverName, err := getDriverName(csiClient, *timeout)
	if err != nil {
		klog.Fatal(fmt.Errorf("get driver name failed: %v", err))
	}
	klog.Infof("CSI driver name: %q", driverName)

	csiResizer, err := resizer.NewResizerFromClient(
		csiClient,
		*timeout,
		kubeClient,
		informerFactory,
		driverName)
	if err != nil {
		klog.Fatal(err.Error())
	}

	resizerName := csiResizer.Name()
	rc := controller.NewResizeController(resizerName, csiResizer, kubeClient, *resyncPeriod, informerFactory,
		workqueue.NewItemExponentialFailureRateLimiter(*retryIntervalStart, *retryIntervalMax),
		*handleVolumeInUseError)
	run := func(ctx context.Context) {
		informerFactory.Start(wait.NeverStop)
		rc.Run(*workers, ctx)

	}

	if !*enableLeaderElection {
		run(context.TODO())
	} else {
		lockName := "external-resizer-" + util.SanitizeName(resizerName)
		leKubeClient, err := kubernetes.NewForConfig(config)
		if err != nil {
			klog.Fatal(err.Error())
		}
		le := leaderelection.NewLeaderElection(leKubeClient, lockName, run)
		if *leaderElectionNamespace != "" {
			le.WithNamespace(*leaderElectionNamespace)
		}

		if err := le.Run(); err != nil {
			klog.Fatalf("error initializing leader election: %v", err)
		}
	}

}

func getDriverName(client csi.Client, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return client.GetDriverName(ctx)
}
