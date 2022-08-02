package pkg

import (
	"github.com/google/go-cmp/cmp"
	schedulerapi "k8s-lx1036/k8s/scheduler/pkg/apis/config"
	"k8s.io/client-go/tools/events"
	"k8s.io/kubernetes/pkg/scheduler/framework/plugins/defaultbinder"
	"sort"
	"strings"
	"testing"

	frameworkruntime "k8s-lx1036/k8s/scheduler/pkg/framework/runtime"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

func TestSchedulerNew(test *testing.T) {
	invalidRegistry := map[string]frameworkruntime.PluginFactory{
		defaultbinder.Name: defaultbinder.New,
	}
	validRegistry := map[string]frameworkruntime.PluginFactory{
		"Foo": defaultbinder.New,
	}
	fixtures := []struct {
		name         string
		opts         []Option
		wantErr      string
		wantProfiles []string
	}{
		{
			name: "valid out-of-tree registry",
			opts: []Option{
				WithFrameworkOutOfTreeRegistry(validRegistry),
				WithProfiles(
					schedulerapi.KubeSchedulerProfile{
						SchedulerName: "default-scheduler",
						Plugins: &schedulerapi.Plugins{
							QueueSort: schedulerapi.PluginSet{
								Enabled: []schedulerapi.Plugin{{Name: "PrioritySort"}},
							},
							Bind: schedulerapi.PluginSet{
								Enabled: []schedulerapi.Plugin{{Name: "DefaultBinder"}},
							},
						},
					},
				)},
			wantProfiles: []string{"default-scheduler"},
		},
		{
			name: "repeated plugin name in out-of-tree plugin",
			opts: []Option{
				WithFrameworkOutOfTreeRegistry(invalidRegistry),
				WithProfiles(
					schedulerapi.KubeSchedulerProfile{
						SchedulerName: "default-scheduler",
						Plugins: &schedulerapi.Plugins{
							QueueSort: schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "PrioritySort"}}},
							Bind:      schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "DefaultBinder"}}},
						},
					},
				)},
			wantProfiles: []string{"default-scheduler"},
			wantErr:      "a plugin named DefaultBinder already exists",
		},
		{
			name: "multiple profiles",
			opts: []Option{
				WithProfiles(
					schedulerapi.KubeSchedulerProfile{
						SchedulerName: "foo",
						Plugins: &schedulerapi.Plugins{
							QueueSort: schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "PrioritySort"}}},
							Bind:      schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "DefaultBinder"}}},
						},
					},
					schedulerapi.KubeSchedulerProfile{
						SchedulerName: "bar",
						Plugins: &schedulerapi.Plugins{
							QueueSort: schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "PrioritySort"}}},
							Bind:      schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "DefaultBinder"}}},
						},
					},
				)},
			wantProfiles: []string{"bar", "foo"},
		},
		{
			name: "Repeated profiles",
			opts: []Option{
				WithProfiles(
					schedulerapi.KubeSchedulerProfile{
						SchedulerName: "foo",
						Plugins: &schedulerapi.Plugins{
							QueueSort: schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "PrioritySort"}}},
							Bind:      schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "DefaultBinder"}}},
						},
					},
					schedulerapi.KubeSchedulerProfile{
						SchedulerName: "bar",
						Plugins: &schedulerapi.Plugins{
							QueueSort: schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "PrioritySort"}}},
							Bind:      schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "DefaultBinder"}}},
						},
					},
					schedulerapi.KubeSchedulerProfile{
						SchedulerName: "foo",
						Plugins: &schedulerapi.Plugins{
							QueueSort: schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "PrioritySort"}}},
							Bind:      schedulerapi.PluginSet{Enabled: []schedulerapi.Plugin{{Name: "DefaultBinder"}}},
						},
					},
				)},
			wantErr: `duplicate profile with scheduler name "foo"`,
		},
	}

	for _, fixture := range fixtures {
		test.Run(fixture.name, func(t *testing.T) {
			client := fake.NewSimpleClientset()
			informerFactory := informers.NewSharedInformerFactory(client, 0)
			eventBroadcaster := events.NewBroadcaster(&events.EventSinkImpl{Interface: client.EventsV1()})
			stopCh := make(chan struct{})
			defer close(stopCh)

			scheduler, err := New(
				client,
				informerFactory,
				nil,
				frameworkruntime.NewRecorderFactory(eventBroadcaster),
				stopCh,
				fixture.opts...,
			)
			if len(fixture.wantErr) != 0 {
				if err == nil || !strings.Contains(err.Error(), fixture.wantErr) {
					t.Errorf("got error %q, want %q", err, fixture.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Failed to create scheduler: %v", err)
			}

			// Profiles
			profiles := make([]string, 0, len(scheduler.Frameworks))
			for name := range scheduler.Frameworks {
				profiles = append(profiles, name)
			}
			sort.Strings(profiles)
			if diff := cmp.Diff(fixture.wantProfiles, profiles); diff != "" { // INFO: diff string slice，以后可以使用
				t.Errorf("unexpected profiles (-want, +got):\n%s", diff)
			}
		})
	}
}
