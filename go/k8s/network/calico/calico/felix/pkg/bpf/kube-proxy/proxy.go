package proxy

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/proxy/apis"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	k8sproxy "k8s.io/kubernetes/pkg/proxy"
	"k8s.io/kubernetes/pkg/proxy/config"
	"k8s.io/kubernetes/pkg/util/async"
)

type Proxy interface {
	Stop()
}

type DPSyncerState struct {
	SvcMap k8sproxy.ServiceMap
	EpsMap k8sproxy.EndpointsMap
}

type Proxier struct {
	// ensures that only one invocation runs at any time
	runnerLock sync.Mutex

	k8sClient kubernetes.Interface
	runner    *async.BoundedFrequencyRunner

	// how often to fully sync with k8s - 0 is never
	syncPeriod        time.Duration
	UseEndpointSlices bool

	svcMap     k8sproxy.ServiceMap
	svcChanges *k8sproxy.ServiceChangeTracker
	epsMap     k8sproxy.EndpointsMap
	epsChanges *k8sproxy.EndpointChangeTracker

	syncer Syncer

	stopCh chan struct{}
}

func New(k8sClient kubernetes.Interface, syncer *Syncer, hostname string, opts ...Option) (*Proxier, error) {

	proxier := &Proxier{
		k8sClient:         k8sClient,
		UseEndpointSlices: false, // TODO: 先 hard code

		svcMap: make(k8sproxy.ServiceMap),
		epsMap: make(k8sproxy.EndpointsMap),

		stopCh: make(chan struct{}),
	}

	proxier.runner = async.NewBoundedFrequencyRunner("kube-proxy-syncer",
		proxier.invokeSync, proxier.syncPeriod, time.Hour, 1) // 周期 syncPeriod 触发 invokeSync()

	proxier.epsChanges = k8sproxy.NewEndpointChangeTracker(proxier.hostname, nil, corev1.IPv4Protocol, proxier.recorder, nil)
	proxier.svcChanges = k8sproxy.NewServiceChangeTracker(nil, corev1.IPv4Protocol, proxier.recorder, nil)

	// INFO: @see https://github.com/kubernetes/kubernetes/blob/v1.24.2/cmd/kube-proxy/app/server.go#L733-L772
	noProxyName, err := labels.NewRequirement(apis.LabelServiceProxyName, selection.DoesNotExist, nil)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("noProxyName selector: %v", err))
	}
	noHeadlessEndpoints, err := labels.NewRequirement(corev1.IsHeadlessService, selection.DoesNotExist, nil)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("noHeadlessEndpoints selector: %v", err))
	}
	labelSelector := labels.NewSelector()
	labelSelector = labelSelector.Add(*noProxyName, *noHeadlessEndpoints)
	informerFactory := informers.NewSharedInformerFactoryWithOptions(k8sClient, proxier.syncPeriod,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = labelSelector.String()
		}))
	svcConfig := config.NewServiceConfig(informerFactory.Core().V1().Services(), proxier.syncPeriod)
	svcConfig.RegisterEventHandler(p)
	var epFn func(<-chan struct{})
	if !proxier.UseEndpointSlices {
		endpointsConfig := config.NewEndpointsConfig(informerFactory.Core().V1().Endpoints(), proxier.syncPeriod)
		endpointsConfig.RegisterEventHandler(p)
		epFn = endpointsConfig.Run
	} else {
		endpointSliceConfig := config.NewEndpointSliceConfig(informerFactory.Discovery().V1().EndpointSlices(), proxier.syncPeriod)
		endpointSliceConfig.RegisterEventHandler(p)
		epFn = endpointSliceConfig.Run
	}

	go informerFactory.Start(proxier.stopCh) // 包含 informer run
	go svcConfig.Run(proxier.stopCh)         // 注册 service event handler
	go epFn(proxier.stopCh)                  // 注册 endpoint/endpointSlice event handler
	go proxier.runner.Loop(proxier.stopCh)   // async runner loop 运行起来

}

// touch bpf datapath
func (proxier *Proxier) invokeSync() {
	if !proxier.isInitialized() {
		return
	}

	proxier.runnerLock.Lock()
	defer proxier.runnerLock.Unlock()

	svcUpdateResult := proxier.svcMap.Update(proxier.svcChanges)
	epsUpdateResult := proxier.epsMap.Update(proxier.epsChanges)
	// Update service healthchecks.  The endpoints list might include services that are
	// not "OnlyLocal", but the services list will not, and the serviceHealthServer
	// will just drop those endpoints.
	if err := proxier.serviceHealthServer.SyncServices(serviceUpdateResult.HCServiceNodePorts); err != nil {
		klog.ErrorS(err, "Error syncing healthcheck services")
	}
	if err := proxier.serviceHealthServer.SyncEndpoints(endpointUpdateResult.HCEndpointsLocalIPSize); err != nil {
		klog.ErrorS(err, "Error syncing healthcheck endpoints")
	}

	err := proxier.syncer.Apply(DPSyncerState{
		SvcMap: proxier.svcMap,
		EpsMap: proxier.epsMap,
	})
	if err != nil {
		log.WithError(err).Errorf("applying changes failed")
	}
}

// touch bpf datapath
func (proxier *Proxier) syncDP() {
	proxier.runner.Run()
}

func (proxier *Proxier) OnServiceAdd(service *corev1.Service) {
	proxier.OnServiceUpdate(nil, service)
}

func (proxier *Proxier) OnServiceUpdate(oldService, service *corev1.Service) {
	if proxier.svcChanges.Update(old, curr) && proxier.isInitialized() { // 如果 svc 的确有变化，则 invoke aync runner
		proxier.syncDP()
	}
}

func (proxier *Proxier) OnServiceDelete(service *corev1.Service) {
	//TODO implement me
	panic("implement me")
}

func (proxier *Proxier) OnServiceSynced() {
	//TODO implement me
	panic("implement me")
}

func (proxier *Proxier) OnEndpointSliceAdd(endpointSlice *discoveryv1.EndpointSlice) {
	//TODO implement me
	panic("implement me")
}

func (proxier *Proxier) OnEndpointSliceUpdate(oldEndpointSlice, newEndpointSlice *discoveryv1.EndpointSlice) {
	//TODO implement me
	panic("implement me")
}

func (proxier *Proxier) OnEndpointSliceDelete(endpointSlice *discoveryv1.EndpointSlice) {
	//TODO implement me
	panic("implement me")
}

func (proxier *Proxier) OnEndpointSlicesSynced() {
	//TODO implement me
	panic("implement me")
}

func (proxier *Proxier) OnEndpointsAdd(endpoints *corev1.Endpoints)                  {}
func (proxier *Proxier) OnEndpointsUpdate(oldEndpoints, endpoints *corev1.Endpoints) {}
func (proxier *Proxier) OnEndpointsDelete(endpoints *corev1.Endpoints)               {}
func (proxier *Proxier) OnEndpointsSynced()                                          {}
