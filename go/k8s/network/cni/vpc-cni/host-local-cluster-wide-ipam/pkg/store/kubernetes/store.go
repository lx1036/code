package kubernetes

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"net"
	"strings"

	"k8s-lx1036/k8s/network/cni/vpc-cni/host-local-cluster-wide-ipam/pkg/allocator"
	"k8s-lx1036/k8s/network/cni/vpc-cni/host-local-cluster-wide-ipam/pkg/apis/ipam.cni.io/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type Store struct {
	kubeClient *kubernetes.Clientset
	crdClient  client.Client
}

func NewClientOrDie() *Store {
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		klog.Fatal(err)
	}
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		klog.Fatal(err)
	}

	mapper, err := apiutil.NewDiscoveryRESTMapper(restConfig)
	if err != nil {
		klog.Fatal(err)
	}
	_ = v1.AddToScheme(scheme.Scheme) // INFO: 使用 scheme.Scheme，这样 crdClient 就可以 crd resource, 也可以内置的 k8s resource
	crdClient, err := client.New(restConfig, client.Options{Scheme: scheme.Scheme, Mapper: mapper})
	if err != nil {
		klog.Fatal(err)
	}

	return &Store{
		kubeClient: kubeClient,
		crdClient:  crdClient,
	}
}

func (store *Store) GetOrCreateIPPool(ctx context.Context, cidr string) (*v1.IPPool, error) {
	name := getIPPoolName(cidr)
	pool := &v1.IPPool{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	if err := store.crdClient.Get(ctx, types.NamespacedName{Name: name}, pool); err != nil {
		if errors.IsNotFound(err) {
			pool.Spec.Range = cidr
			pool.Spec.Allocations = make(map[string]v1.IPAllocation)
			if err = store.crdClient.Create(ctx, pool); err != nil && !errors.IsAlreadyExists(err) {
				return nil, err
			}

			return pool, nil
		}

		return nil, fmt.Errorf(fmt.Sprintf("get ippool %s err:%v", name, err))
	}

	return pool, nil
}

// Reserve INFO: 保存在 ippool.status.used
func (store *Store) Reserve(ctx context.Context, podKey string, ip net.IP, ipamConf allocator.IPAMConfig) (bool, error) {
	ippool, err := store.GetOrCreateIPPool(ctx, ipamConf.Range)

	if ippool.Status.Used != nil {
		if _, ok := ippool.Status.Used[podKey]; ok {
			return false, nil
		}
	}

	newIPPool := ippool.DeepCopy()
	newIPPool.Status.Used[podKey] = ip.String()

	err := store.crdClient.Status().Update(ctx, ippool)
}

func getIPPoolName(cidr string) string { // 192.168.1.0/24 -> 192.168.1.0-24
	return strings.ReplaceAll(cidr, "/", "-")
}
