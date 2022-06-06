package leader_election

import (
	"context"
	"flag"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
)

// INFO: @see k8s-lx1036/k8s/storage/csi/csi-lib-utils/leaderelection/leader_election.go
//  对于 CNI 需要使用分布式锁来中心化分配 IP，可以使用这段逻辑

var (
	kubeconfig = flag.String("kubeconfig", "", "absolute path to kubeconfig file")
)

func TestLeaderElectionForHostLocalClusterWideIPAM(test *testing.T) {
	flag.Parse()

	if len(*kubeconfig) == 0 {
		klog.Errorf("--kubeconfig should be required")
		return
	}

	leader, leaderOk, deposed := newLeaderElector(*kubeconfig)
	var wg sync.WaitGroup
	var newip net.IP
	wg.Add(2)
	stopM := make(chan struct{})

	go func() {
		defer wg.Done()
		for {
			select {
			case <-leaderOk:
				klog.Info("Elected as leader, do processing")
				newip = net.ParseIP("192.168.0.1")
				stopM <- struct{}{}
				return
			case <-deposed:
				klog.Info("Deposed as leader, shutting down")
				return
			}
		}
	}()

	go func() {
		defer wg.Done()

		go func() {
			leader.Run(context.TODO())
		}()

		// wait for stop which tells us when IP allocation occurred or context deadline exceeded
		<-stopM
	}()

	wg.Wait()

	klog.Info(fmt.Sprintf("ip is %s", newip.String()))
}

func newLeaderElector(kubeconfig string) (*leaderelection.LeaderElector, chan struct{}, chan struct{}) {
	leaderOK := make(chan struct{})
	deposed := make(chan struct{})

	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		klog.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		klog.Fatal(err)
	}

	name := "lease-test"
	id := name + "_" + string(uuid.NewUUID())
	eventBroadcaster := record.NewBroadcaster()
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{
		Component: "leader-election-controller",
		Host:      "nodeName-1",
	})
	var rl = &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Client: clientset.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity:      id,
			EventRecorder: recorder,
		},
	}
	leader, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock:            rl,
		LeaseDuration:   time.Duration(1500) * time.Millisecond, // INFO: 对于 CNI 需要以 ms 计算
		RenewDeadline:   time.Duration(1000) * time.Millisecond,
		RetryPeriod:     time.Duration(500) * time.Millisecond,
		ReleaseOnCancel: true,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(_ context.Context) {
				klog.Info("get leader")
				close(leaderOK)
			},
			OnStoppedLeading: func() {
				klog.Info("lost leader")
				// The context being canceled will trigger a handler that will
				// deal with being deposed.
				close(deposed)
			},
			OnNewLeader: func(identity string) {
				klog.Infof("%s is the new leader", identity)
			},
		},
	})
	if err != nil {
		klog.Errorf(fmt.Sprintf("Failed to create leader elector: %v", err))
		return nil, leaderOK, deposed
	}

	return leader, leaderOK, deposed
}
