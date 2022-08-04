package runtime

import (
	"context"
	"testing"

	configv1 "k8s-lx1036/k8s/scheduler/pkg/apis/config/v1"
	"k8s-lx1036/k8s/scheduler/pkg/framework"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	preFilterPluginName               = "test-prefilter-plugin"
	preFilterWithExtensionsPluginName = "test-prefilter-with-extensions-plugin"
)

type TestPreFilterPlugin struct {
	PreFilterCalled int
}

func (pl *TestPreFilterPlugin) Name() string {
	return preFilterPluginName
}
func (pl *TestPreFilterPlugin) PreFilter(ctx context.Context, state *framework.CycleState,
	p *corev1.Pod) (*framework.PreFilterResult, *framework.Status) {
	pl.PreFilterCalled++
	return nil, nil
}
func (pl *TestPreFilterPlugin) PreFilterExtensions() framework.PreFilterExtensions {
	return nil
}

type TestPreFilterWithExtensionsPlugin struct {
	PreFilterCalled int
	AddCalled       int
	RemoveCalled    int
}

func (pl *TestPreFilterWithExtensionsPlugin) Name() string {
	return preFilterWithExtensionsPluginName
}
func (pl *TestPreFilterWithExtensionsPlugin) PreFilter(ctx context.Context, state *framework.CycleState,
	p *corev1.Pod) (*framework.PreFilterResult, *framework.Status) {
	pl.PreFilterCalled++
	return nil, nil
}
func (pl *TestPreFilterWithExtensionsPlugin) PreFilterExtensions() framework.PreFilterExtensions {
	return pl
}
func (pl *TestPreFilterWithExtensionsPlugin) AddPod(ctx context.Context, state *framework.CycleState, podToSchedule *corev1.Pod,
	podInfoToAdd *framework.PodInfo, nodeInfo *framework.NodeInfo) *framework.Status {
	pl.AddCalled++
	return nil
}
func (pl *TestPreFilterWithExtensionsPlugin) RemovePod(ctx context.Context, state *framework.CycleState, podToSchedule *corev1.Pod,
	podInfoToRemove *framework.PodInfo, nodeInfo *framework.NodeInfo) *framework.Status {
	pl.RemoveCalled++
	return nil
}

const (
	queueSortPlugin = "queue-sort-plugin"
	bindPlugin      = "bind-plugin"
)

type TestQueueSortPlugin struct{}

func (pl *TestQueueSortPlugin) Name() string {
	return queueSortPlugin
}
func newQueueSortPlugin(_ runtime.Object, _ *Framework) (framework.Plugin, error) {
	return &TestQueueSortPlugin{}, nil
}

type TestBindPlugin struct{}

func (t TestBindPlugin) Name() string {
	return bindPlugin
}
func newBindPlugin(_ runtime.Object, _ *Framework) (framework.Plugin, error) {
	return &TestBindPlugin{}, nil
}
func newFrameworkWithQueueSortAndBind(r Registry, profile configv1.KubeSchedulerProfile, opts ...Option) (*Framework, error) {
	if _, ok := r[queueSortPlugin]; !ok {
		r[queueSortPlugin] = newQueueSortPlugin
	}
	if _, ok := r[bindPlugin]; !ok {
		r[bindPlugin] = newBindPlugin
	}
	if len(profile.Plugins.QueueSort.Enabled) == 0 {
		profile.Plugins.QueueSort.Enabled = append(profile.Plugins.QueueSort.Enabled, configv1.Plugin{Name: queueSortPlugin})
	}
	if len(profile.Plugins.Bind.Enabled) == 0 {
		profile.Plugins.Bind.Enabled = append(profile.Plugins.Bind.Enabled, configv1.Plugin{Name: bindPlugin})
	}

	return NewFramework(r, &profile, opts...)
}

func TestPreFilterPlugins(test *testing.T) {
	preFilter1 := &TestPreFilterPlugin{}
	preFilter2 := &TestPreFilterWithExtensionsPlugin{}
	r := make(Registry)
	r.Register(preFilterPluginName, func(configuration runtime.Object, f *Framework) (framework.Plugin, error) {
		return preFilter1, nil
	})
	r.Register(preFilterWithExtensionsPluginName, func(configuration runtime.Object, f *Framework) (framework.Plugin, error) {
		return preFilter2, nil
	})
	profile := configv1.KubeSchedulerProfile{
		Plugins: &configv1.Plugins{
			PreFilter: configv1.PluginSet{
				Enabled: []configv1.Plugin{
					{Name: preFilterPluginName},
					{Name: preFilterWithExtensionsPluginName},
				},
			},
		},
	}
	test.Run("TestPreFilterPlugins", func(t *testing.T) {
		f, err := newFrameworkWithQueueSortAndBind(r, profile)
		if err != nil {
			t.Fatalf("Failed to create framework for testing: %v", err)
		}
		f.RunPreFilterPlugins(context.Background(), nil, nil)
		f.RunPreFilterExtensionAddPod(context.Background(), nil, nil, nil, nil)
		f.RunPreFilterExtensionRemovePod(context.Background(), nil, nil, nil, nil)
		if preFilter1.PreFilterCalled != 1 {
			t.Errorf("preFilter1 called %v, expected: 1", preFilter1.PreFilterCalled)
		}
		if preFilter2.PreFilterCalled != 1 {
			t.Errorf("preFilter2 called %v, expected: 1", preFilter2.PreFilterCalled)
		}
		if preFilter2.AddCalled != 1 {
			t.Errorf("AddPod called %v, expected: 1", preFilter2.AddCalled)
		}
		if preFilter2.RemoveCalled != 1 {
			t.Errorf("AddPod called %v, expected: 1", preFilter2.RemoveCalled)
		}
	})
}
