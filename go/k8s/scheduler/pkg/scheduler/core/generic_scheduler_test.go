package core

import (
	"fmt"
	"testing"
	"context"
	"reflect"
	"time"
	
	st "k8s-lx1036/k8s/scheduler/pkg/scheduler/testing"
	"k8s-lx1036/k8s/scheduler/pkg/scheduler/apis/config"
	internalcache "k8s-lx1036/k8s/scheduler/pkg/scheduler/internal/cache"
	
	
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/defaultbinder"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/podtopologyspread"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/queuesort"
	"k8s.io/apimachinery/pkg/util/wait"

	
)

func TestGenericScheduler(test *testing.T) {
	
	fixtures := []struct {
		name            string
		registerPlugins []st.RegisterPluginFunc
		nodes           []string
		pvcs            []v1.PersistentVolumeClaim
		pod             *v1.Pod
		pods            []*v1.Pod
		expectedHosts   sets.String
		wErr            error
	}{
		{
			registerPlugins: []st.RegisterPluginFunc{
				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
				st.RegisterFilterPlugin("FalseFilter", st.NewFalseFilterPlugin),
				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
			},
			nodes: []string{"machine1", "machine2"},
			pod:   &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "2", UID: types.UID("2")}},
			name:  "test 1",
			wErr: &FitError{
				Pod:         &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "2", UID: types.UID("2")}},
				NumAllNodes: 2,
				FilteredNodesStatuses: framework.NodeToStatusMap{
					"machine1": framework.NewStatus(framework.Unschedulable, st.ErrReasonFake),
					"machine2": framework.NewStatus(framework.Unschedulable, st.ErrReasonFake),
				},
			},
		},
		{
			registerPlugins: []st.RegisterPluginFunc{
				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
				st.RegisterFilterPlugin("TrueFilter", st.NewTrueFilterPlugin),
				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
			},
			nodes:         []string{"machine1", "machine2"},
			pod:           &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "ignore", UID: types.UID("ignore")}},
			expectedHosts: sets.NewString("machine1", "machine2"),
			name:          "test 2",
			wErr:          nil,
		},
		{
			// Fits on a machine where the pod ID matches the machine name
			registerPlugins: []st.RegisterPluginFunc{
				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
				st.RegisterFilterPlugin("MatchFilter", st.NewMatchFilterPlugin),
				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
			},
			nodes:         []string{"machine1", "machine2"},
			pod:           &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "machine2", UID: types.UID("machine2")}},
			expectedHosts: sets.NewString("machine2"),
			name:          "test 3",
			wErr:          nil,
		},
		{
			registerPlugins: []st.RegisterPluginFunc{
				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
				st.RegisterFilterPlugin("TrueFilter", st.NewTrueFilterPlugin),
				st.RegisterScorePlugin("NumericMap", newNumericMapPlugin(), 1),
				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
			},
			nodes:         []string{"3", "2", "1"},
			pod:           &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "ignore", UID: types.UID("ignore")}},
			expectedHosts: sets.NewString("3"),
			name:          "test 4",
			wErr:          nil,
		},
		{
			registerPlugins: []st.RegisterPluginFunc{
				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
				st.RegisterFilterPlugin("MatchFilter", st.NewMatchFilterPlugin),
				st.RegisterScorePlugin("NumericMap", newNumericMapPlugin(), 1),
				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
			},
			nodes:         []string{"3", "2", "1"},
			pod:           &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "2", UID: types.UID("2")}},
			expectedHosts: sets.NewString("2"),
			name:          "test 5",
			wErr:          nil,
		},
		{
			registerPlugins: []st.RegisterPluginFunc{
				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
				st.RegisterFilterPlugin("TrueFilter", st.NewTrueFilterPlugin),
				st.RegisterScorePlugin("NumericMap", newNumericMapPlugin(), 1),
				st.RegisterScorePlugin("ReverseNumericMap", newReverseNumericMapPlugin(), 2),
				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
			},
			nodes:         []string{"3", "2", "1"},
			pod:           &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "2", UID: types.UID("2")}},
			expectedHosts: sets.NewString("1"),
			name:          "test 6",
			wErr:          nil,
		},
		{
			registerPlugins: []st.RegisterPluginFunc{
				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
				st.RegisterFilterPlugin("TrueFilter", st.NewTrueFilterPlugin),
				st.RegisterFilterPlugin("FalseFilter", st.NewFalseFilterPlugin),
				st.RegisterScorePlugin("NumericMap", newNumericMapPlugin(), 1),
				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
			},
			nodes: []string{"3", "2", "1"},
			pod:   &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "2", UID: types.UID("2")}},
			name:  "test 7",
			wErr: &FitError{
				Pod:         &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "2", UID: types.UID("2")}},
				NumAllNodes: 3,
				FilteredNodesStatuses: framework.NodeToStatusMap{
					"3": framework.NewStatus(framework.Unschedulable, st.ErrReasonFake),
					"2": framework.NewStatus(framework.Unschedulable, st.ErrReasonFake),
					"1": framework.NewStatus(framework.Unschedulable, st.ErrReasonFake),
				},
			},
		},
		{
			registerPlugins: []st.RegisterPluginFunc{
				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
				st.RegisterFilterPlugin("NoPodsFilter", NewNoPodsFilterPlugin),
				st.RegisterFilterPlugin("MatchFilter", st.NewMatchFilterPlugin),
				st.RegisterScorePlugin("NumericMap", newNumericMapPlugin(), 1),
				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
			},
			pods: []*v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "2", UID: types.UID("2")},
					Spec: v1.PodSpec{
						NodeName: "2",
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				},
			},
			pod:   &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "2", UID: types.UID("2")}},
			nodes: []string{"1", "2"},
			name:  "test 8",
			wErr: &FitError{
				Pod:         &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "2", UID: types.UID("2")}},
				NumAllNodes: 2,
				FilteredNodesStatuses: framework.NodeToStatusMap{
					"1": framework.NewStatus(framework.Unschedulable, st.ErrReasonFake),
					"2": framework.NewStatus(framework.Unschedulable, st.ErrReasonFake),
				},
			},
		},
		{
			// Pod with existing PVC
			registerPlugins: []st.RegisterPluginFunc{
				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
				st.RegisterFilterPlugin("TrueFilter", st.NewTrueFilterPlugin),
				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
			},
			nodes: []string{"machine1", "machine2"},
			pvcs:  []v1.PersistentVolumeClaim{{ObjectMeta: metav1.ObjectMeta{Name: "existingPVC"}}},
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "ignore", UID: types.UID("ignore")},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						{
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: "existingPVC",
								},
							},
						},
					},
				},
			},
			expectedHosts: sets.NewString("machine1", "machine2"),
			name:          "existing PVC",
			wErr:          nil,
		},
		{
			// Pod with non existing PVC
			registerPlugins: []st.RegisterPluginFunc{
				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
				st.RegisterFilterPlugin("TrueFilter", st.NewTrueFilterPlugin),
				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
			},
			nodes: []string{"machine1", "machine2"},
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "ignore", UID: types.UID("ignore")},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						{
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: "unknownPVC",
								},
							},
						},
					},
				},
			},
			name: "unknown PVC",
			wErr: fmt.Errorf("persistentvolumeclaim \"unknownPVC\" not found"),
		},
		{
			// Pod with deleting PVC
			registerPlugins: []st.RegisterPluginFunc{
				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
				st.RegisterFilterPlugin("TrueFilter", st.NewTrueFilterPlugin),
				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
			},
			nodes: []string{"machine1", "machine2"},
			pvcs:  []v1.PersistentVolumeClaim{{ObjectMeta: metav1.ObjectMeta{Name: "existingPVC", DeletionTimestamp: &metav1.Time{}}}},
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "ignore", UID: types.UID("ignore")},
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						{
							VolumeSource: v1.VolumeSource{
								PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
									ClaimName: "existingPVC",
								},
							},
						},
					},
				},
			},
			name: "deleted PVC",
			wErr: fmt.Errorf("persistentvolumeclaim \"existingPVC\" is being deleted"),
		},
		{
			registerPlugins: []st.RegisterPluginFunc{
				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
				st.RegisterFilterPlugin("TrueFilter", st.NewTrueFilterPlugin),
				st.RegisterScorePlugin("FalseMap", newFalseMapPlugin(), 1),
				st.RegisterScorePlugin("TrueMap", newTrueMapPlugin(), 2),
				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
			},
			nodes: []string{"2", "1"},
			pod:   &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "2"}},
			name:  "test error with priority map",
			wErr:  fmt.Errorf("error while running score plugin for pod \"2\": %+v", errPrioritize),
		},
		{
			name: "test podtopologyspread plugin - 2 nodes with maxskew=1",
			registerPlugins: []st.RegisterPluginFunc{
				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
				st.RegisterPluginAsExtensions(
					podtopologyspread.Name,
					podtopologyspread.New,
					"PreFilter",
					"Filter",
				),
				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
			},
			nodes: []string{"machine1", "machine2"},
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "p", UID: types.UID("p"), Labels: map[string]string{"foo": ""}},
				Spec: v1.PodSpec{
					TopologySpreadConstraints: []v1.TopologySpreadConstraint{
						{
							MaxSkew:           1,
							TopologyKey:       "hostname",
							WhenUnsatisfiable: v1.DoNotSchedule,
							LabelSelector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      "foo",
										Operator: metav1.LabelSelectorOpExists,
									},
								},
							},
						},
					},
				},
			},
			pods: []*v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod1", UID: types.UID("pod1"), Labels: map[string]string{"foo": ""}},
					Spec: v1.PodSpec{
						NodeName: "machine1",
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				},
			},
			expectedHosts: sets.NewString("machine2"),
			wErr:          nil,
		},
		{
			name: "test podtopologyspread plugin - 3 nodes with maxskew=2",
			registerPlugins: []st.RegisterPluginFunc{
				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
				st.RegisterPluginAsExtensions(
					podtopologyspread.Name,
					podtopologyspread.New,
					"PreFilter",
					"Filter",
				),
				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
			},
			nodes: []string{"machine1", "machine2", "machine3"},
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "p", UID: types.UID("p"), Labels: map[string]string{"foo": ""}},
				Spec: v1.PodSpec{
					TopologySpreadConstraints: []v1.TopologySpreadConstraint{
						{
							MaxSkew:           2,
							TopologyKey:       "hostname",
							WhenUnsatisfiable: v1.DoNotSchedule,
							LabelSelector: &metav1.LabelSelector{
								MatchExpressions: []metav1.LabelSelectorRequirement{
									{
										Key:      "foo",
										Operator: metav1.LabelSelectorOpExists,
									},
								},
							},
						},
					},
				},
			},
			pods: []*v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod1a", UID: types.UID("pod1a"), Labels: map[string]string{"foo": ""}},
					Spec: v1.PodSpec{
						NodeName: "machine1",
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod1b", UID: types.UID("pod1b"), Labels: map[string]string{"foo": ""}},
					Spec: v1.PodSpec{
						NodeName: "machine1",
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "pod2", UID: types.UID("pod2"), Labels: map[string]string{"foo": ""}},
					Spec: v1.PodSpec{
						NodeName: "machine2",
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				},
			},
			expectedHosts: sets.NewString("machine2", "machine3"),
			wErr:          nil,
		},
		{
			name: "test with filter plugin returning Unschedulable status",
			registerPlugins: []st.RegisterPluginFunc{
				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
				st.RegisterFilterPlugin(
					"FakeFilter",
					st.NewFakeFilterPlugin(map[string]framework.Code{"3": framework.Unschedulable}),
				),
				st.RegisterScorePlugin("NumericMap", newNumericMapPlugin(), 1),
				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
			},
			nodes:         []string{"3"},
			pod:           &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test-filter", UID: types.UID("test-filter")}},
			expectedHosts: nil,
			wErr: &FitError{
				Pod:         &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test-filter", UID: types.UID("test-filter")}},
				NumAllNodes: 1,
				FilteredNodesStatuses: framework.NodeToStatusMap{
					"3": framework.NewStatus(framework.Unschedulable, "injecting failure for pod test-filter"),
				},
			},
		},
		{
			name: "test with filter plugin returning UnschedulableAndUnresolvable status",
			registerPlugins: []st.RegisterPluginFunc{
				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
				st.RegisterFilterPlugin(
					"FakeFilter",
					st.NewFakeFilterPlugin(map[string]framework.Code{"3": framework.UnschedulableAndUnresolvable}),
				),
				st.RegisterScorePlugin("NumericMap", newNumericMapPlugin(), 1),
				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
			},
			nodes:         []string{"3"},
			pod:           &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test-filter", UID: types.UID("test-filter")}},
			expectedHosts: nil,
			wErr: &FitError{
				Pod:         &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test-filter", UID: types.UID("test-filter")}},
				NumAllNodes: 1,
				FilteredNodesStatuses: framework.NodeToStatusMap{
					"3": framework.NewStatus(framework.UnschedulableAndUnresolvable, "injecting failure for pod test-filter"),
				},
			},
		},
		{
			name: "test with partial failed filter plugin",
			registerPlugins: []st.RegisterPluginFunc{
				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
				st.RegisterFilterPlugin(
					"FakeFilter",
					st.NewFakeFilterPlugin(map[string]framework.Code{"1": framework.Unschedulable}),
				),
				st.RegisterScorePlugin("NumericMap", newNumericMapPlugin(), 1),
				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
			},
			nodes:         []string{"1", "2"},
			pod:           &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test-filter", UID: types.UID("test-filter")}},
			expectedHosts: nil,
			wErr:          nil,
		},
		{
			name: "test prefilter plugin returning Unschedulable status",
			registerPlugins: []st.RegisterPluginFunc{
				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
				st.RegisterPreFilterPlugin(
					"FakePreFilter",
					st.NewFakePreFilterPlugin(framework.NewStatus(framework.UnschedulableAndUnresolvable, "injected unschedulable status")),
				),
				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
			},
			nodes:         []string{"1", "2"},
			pod:           &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test-prefilter", UID: types.UID("test-prefilter")}},
			expectedHosts: nil,
			wErr: &FitError{
				Pod:         &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test-prefilter", UID: types.UID("test-prefilter")}},
				NumAllNodes: 2,
				FilteredNodesStatuses: framework.NodeToStatusMap{
					"1": framework.NewStatus(framework.UnschedulableAndUnresolvable, "injected unschedulable status"),
					"2": framework.NewStatus(framework.UnschedulableAndUnresolvable, "injected unschedulable status"),
				},
			},
		},
		{
			name: "test prefilter plugin returning error status",
			registerPlugins: []st.RegisterPluginFunc{
				st.RegisterQueueSortPlugin(queuesort.Name, queuesort.New),
				st.RegisterPreFilterPlugin(
					"FakePreFilter",
					st.NewFakePreFilterPlugin(framework.NewStatus(framework.Error, "injected error status")),
				),
				st.RegisterBindPlugin(defaultbinder.Name, defaultbinder.New),
			},
			nodes:         []string{"1", "2"},
			pod:           &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test-prefilter", UID: types.UID("test-prefilter")}},
			expectedHosts: nil,
			wErr:          fmt.Errorf(`prefilter plugin "FakePreFilter" failed for pod "test-prefilter": injected error status`),
		},
	}
	
	for _, fixture := range fixtures {
		
		test.Run(fixture.name, func(t *testing.T) {
			cache := internalcache.New(time.Duration(0), wait.NeverStop)
			
			
			snapshot := internalcache.NewSnapshot(test.pods, nodes)
			
			scheduler := NewGenericScheduler(
				cache,
				snapshot,
				[]framework.Extender{},
				pvcLister,
				false,
				config.DefaultPercentageOfNodesToScore)
			result, err := scheduler.Schedule(context.Background(), prof, framework.NewCycleState(), test.pod)
			if !reflect.DeepEqual(err, fixture.wErr) {
				t.Errorf("want: %v, got: %v", fixture.wErr, err)
			}
			
			
			
		})
		
	}
	
	
	
}

