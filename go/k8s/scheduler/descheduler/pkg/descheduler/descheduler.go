package descheduler

import (
	"context"
	"fmt"
	"io/ioutil"

	"k8s-lx1036/k8s/scheduler/descheduler/cmd/app/options"
	"k8s-lx1036/k8s/scheduler/descheduler/pkg/api"
	"k8s-lx1036/k8s/scheduler/descheduler/pkg/api/v1alpha1"
	"k8s-lx1036/k8s/scheduler/descheduler/pkg/client"
	"k8s-lx1036/k8s/scheduler/descheduler/pkg/evictions"
	"k8s-lx1036/k8s/scheduler/descheduler/pkg/node"
	"k8s-lx1036/k8s/scheduler/descheduler/pkg/scheme"
	"k8s-lx1036/k8s/scheduler/descheduler/pkg/strategies/nodeutilization"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

func Run(rs *options.Options) error {

	ctx := context.Background()
	rsclient, err := client.CreateClient(rs.KubeconfigFile)
	if err != nil {
		return err
	}
	rs.Client = rsclient

	// --policy-config-file 必须有
	deschedulerPolicy, err := LoadPolicyConfig(rs.PolicyConfigFile)
	if err != nil {
		return err
	}
	if deschedulerPolicy == nil {
		return fmt.Errorf("deschedulerPolicy is nil")
	}

	evictionPolicyGroupVersion, err := evictions.SupportEviction(rs.Client)
	if err != nil || len(evictionPolicyGroupVersion) == 0 {
		return err
	}

	stopChannel := make(chan struct{})
	return RunDeschedulerStrategies(ctx, rs, deschedulerPolicy, evictionPolicyGroupVersion, stopChannel)
}

type strategyFunction func(ctx context.Context, client clientset.Interface, strategy api.DeschedulerStrategy,
	nodes []*v1.Node, podEvictor *evictions.PodEvictor)

func RunDeschedulerStrategies(ctx context.Context, rs *options.Options, deschedulerPolicy *api.DeschedulerPolicy,
	evictionPolicyGroupVersion string, stopChannel chan struct{}) error {
	// TODO: 这里 defaultResync=0，一直没搞明白defaultResync=0用处是啥，这里暂存
	sharedInformerFactory := informers.NewSharedInformerFactory(rs.Client, 0)
	nodeInformer := sharedInformerFactory.Core().V1().Nodes()
	sharedInformerFactory.Start(stopChannel)
	sharedInformerFactory.WaitForCacheSync(stopChannel)

	nodeSelector := rs.NodeSelector
	if deschedulerPolicy.NodeSelector != nil {
		nodeSelector = *deschedulerPolicy.NodeSelector
	}

	evictLocalStoragePods := rs.EvictLocalStoragePods
	if deschedulerPolicy.EvictLocalStoragePods != nil {
		evictLocalStoragePods = *deschedulerPolicy.EvictLocalStoragePods
	}

	maxNoOfPodsToEvictPerNode := rs.MaxNoOfPodsToEvictPerNode
	if deschedulerPolicy.MaxNoOfPodsToEvictPerNode != nil {
		maxNoOfPodsToEvictPerNode = *deschedulerPolicy.MaxNoOfPodsToEvictPerNode
	}

	strategyFuncs := map[string]strategyFunction{
		//"RemoveDuplicates":                            strategies.RemoveDuplicatePods,
		nodeutilization.Name: nodeutilization.LowNodeUtilization,
		//"RemovePodsViolatingInterPodAntiAffinity":     strategies.RemovePodsViolatingInterPodAntiAffinity,
		//"RemovePodsViolatingNodeAffinity":             strategies.RemovePodsViolatingNodeAffinity,
		//"RemovePodsViolatingNodeTaints":               strategies.RemovePodsViolatingNodeTaints,
		//"RemovePodsHavingTooManyRestarts":             strategies.RemovePodsHavingTooManyRestarts,
		//"PodLifeTime":                                 strategies.PodLifeTime,
		//"RemovePodsViolatingTopologySpreadConstraint": strategies.RemovePodsViolatingTopologySpreadConstraint,
	}

	// 周期任务并block
	wait.Until(func() {
		nodes, err := node.ReadyNodes(ctx, rs.Client, nodeInformer, nodeSelector)
		if err != nil {
			klog.V(1).InfoS("Unable to get ready nodes", "err", err)
			close(stopChannel)
			return
		}

		if len(nodes) <= 1 {
			klog.V(1).InfoS("The cluster size is 0 or 1 meaning eviction causes service disruption or degradation. So aborting..")
			close(stopChannel)
			return
		}

		podEvictor := evictions.NewPodEvictor(
			rs.Client,
			evictionPolicyGroupVersion,
			rs.DryRun,
			maxNoOfPodsToEvictPerNode, // 一个node驱逐pod最大数量
			nodes,
			evictLocalStoragePods, // 带有local storage pod是否要驱逐
		)

		for name, f := range strategyFuncs {
			if strategy := deschedulerPolicy.Strategies[api.StrategyName(name)]; strategy.Enabled {
				f(ctx, rs.Client, strategy, nodes, podEvictor)
			}
		}

		// If there was no interval specified, send a signal to the stopChannel to end the wait.Until loop after 1 iteration
		// 没有设置DeschedulingInterval，则只执行一次循环
		if rs.DeschedulingInterval.Seconds() == 0 {
			close(stopChannel)
		}
	}, rs.DeschedulingInterval, stopChannel)

	return nil
}

// TODO: 读取yaml文件，然后转换成内部版本对象，这个逻辑以后直接复用
func LoadPolicyConfig(policyConfigFile string) (*api.DeschedulerPolicy, error) {
	if policyConfigFile == "" {
		klog.V(1).InfoS("Policy config file not specified")
		return nil, nil
	}

	policy, err := ioutil.ReadFile(policyConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read policy config file %q: %+v", policyConfigFile, err)
	}

	versionedPolicy := &v1alpha1.DeschedulerPolicy{}
	decoder := scheme.Codecs.UniversalDecoder(v1alpha1.SchemeGroupVersion)
	if err := runtime.DecodeInto(decoder, policy, versionedPolicy); err != nil {
		return nil, fmt.Errorf("failed decoding descheduler's policy config %q: %v", policyConfigFile, err)
	}

	// 转换成内部版本
	internalPolicy := &api.DeschedulerPolicy{}
	if err := scheme.Scheme.Convert(versionedPolicy, internalPolicy, nil); err != nil {
		return nil, fmt.Errorf("failed converting versioned policy to internal policy version: %v", err)
	}

	return internalPolicy, nil
}
