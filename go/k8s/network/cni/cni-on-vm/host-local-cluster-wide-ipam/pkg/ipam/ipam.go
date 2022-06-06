package ipam

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"net"
	"sync"
	"time"

	"k8s-lx1036/k8s/network/cni/cni-on-vm/host-local-cluster-wide-ipam/pkg/allocator"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

// IPManagement manages ip allocation and deallocation from a storage perspective
func IPManagement(ctx context.Context, mode int, ipamConf allocator.IPAMConfig, containerID string,
	podRef string) (net.IPNet, error) {
	var err error

	restConfig, err := clientcmd.BuildConfigFromFlags("", option.Kubeconfig)
	if err != nil {
		return nil, err
	}

	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to construct lister client: %v", err)
	}

	ipam, err := NewKubernetesIPAM(containerID, ipamConf)
	if err != nil {
		return newip, logging.Errorf("IPAM %s client initialization error: %v", ipamConf.Datastore, err)
	}
	defer ipam.Close()

	leader, leaderOK, deposed := newLeaderElector(ipam.clientSet, ipam.namespace, ipamConf.PodNamespace, ipamConf.PodName,
		ipamConf.LeaderLeaseDuration, ipamConf.LeaderRenewDeadline, ipamConf.LeaderRetryPeriod)

	var wg sync.WaitGroup
	wg.Add(2)
	stopM := make(chan struct{})

	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				err = fmt.Errorf("time limit exceeded while waiting to become leader")
				stopM <- struct{}{}
				return

			case <-leaderOK: // get leader
				newip, err = IPManagementKubernetesUpdate(ctx, mode, ipam, ipamConf, containerID, podRef)
				stopM <- struct{}{}
				return

			case <-deposed:
				stopM <- struct{}{}
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		go leader.Run(context.TODO())
		// wait for stop which tells us when IP allocation occurred or context deadline exceeded
		<-stopM
	}()

	wg.Wait()
	close(stopM)

	return newip, err
}

func IPManagementKubernetesUpdate(ctx context.Context, mode int, ipam *Client, ipamConf allocator.IPAMConfig,
	containerID string, podRef string) (net.IPNet, error) {

	var newip net.IPNet
	// Skip invalid modes
	switch mode {
	case allocator.Allocate, allocator.Deallocate:
	default:
		return newip, fmt.Errorf("got an unknown mode passed to IPManagement: %v", mode)
	}

	pool, err := ipam.GetOrCreateIPPool(requestCtx, ipamConf.Range)
	reservelist := pool.Allocations()
	switch mode {
	case allocator.Allocate:
		newip, updatedreservelist, err = allocator.AssignIP(ipamConf, reservelist, containerID, podRef)

	case allocator.Deallocate:
		updatedreservelist, ipforoverlappingrangeupdate, err = allocator.DeallocateIP(reservelist, containerID)
	}

	err = pool.Update(requestCtx, usereservelist)

}

func newLeaderElector(clientset *kubernetes.Clientset, namespace string, podNamespace string, podID string,
	leaseDuration int, renewDeadline int, retryPeriod int) (*leaderelection.LeaderElector, chan struct{}, chan struct{}) {
	leaderOK := make(chan struct{})
	deposed := make(chan struct{})

	rl := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      "host-local",
			Namespace: namespace,
		},
		Client: clientset.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: fmt.Sprintf("%s/%s", podNamespace, podID),
		},
	}
	leader, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock:            rl,
		LeaseDuration:   time.Duration(leaseDuration) * time.Millisecond,
		RenewDeadline:   time.Duration(renewDeadline) * time.Millisecond,
		RetryPeriod:     time.Duration(retryPeriod) * time.Millisecond,
		ReleaseOnCancel: true,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(_ context.Context) {
				klog.Info("get leader")
				close(leaderOK)
			},
			OnStoppedLeading: func() {
				klog.Info("lost leader")
				close(deposed)
			},
		},
	})
	if err != nil {
		klog.Errorf(fmt.Sprintf("Failed to create leader elector: %v", err))
		return nil, leaderOK, deposed
	}

	return leader, leaderOK, deposed
}
