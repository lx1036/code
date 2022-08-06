package pkg

import (
	"sort"
	"strings"
	"testing"

	configv1 "k8s-lx1036/k8s/scheduler/pkg/apis/config/v1"
	"k8s-lx1036/k8s/scheduler/pkg/framework/plugins/defaultbinder"
	frameworkruntime "k8s-lx1036/k8s/scheduler/pkg/framework/runtime"

	"github.com/google/go-cmp/cmp"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/events"
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
					configv1.KubeSchedulerProfile{
						SchedulerName: "default-scheduler",
						Plugins: &configv1.Plugins{
							QueueSort: configv1.PluginSet{
								Enabled: []configv1.Plugin{{Name: "PrioritySort"}},
							},
							Bind: configv1.PluginSet{
								Enabled: []configv1.Plugin{{Name: "DefaultBinder"}},
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
					configv1.KubeSchedulerProfile{
						SchedulerName: "default-scheduler",
						Plugins: &configv1.Plugins{
							QueueSort: configv1.PluginSet{Enabled: []configv1.Plugin{{Name: "PrioritySort"}}},
							Bind:      configv1.PluginSet{Enabled: []configv1.Plugin{{Name: "DefaultBinder"}}},
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
					configv1.KubeSchedulerProfile{
						SchedulerName: "foo",
						Plugins: &configv1.Plugins{
							QueueSort: configv1.PluginSet{Enabled: []configv1.Plugin{{Name: "PrioritySort"}}},
							Bind:      configv1.PluginSet{Enabled: []configv1.Plugin{{Name: "DefaultBinder"}}},
						},
					},
					configv1.KubeSchedulerProfile{
						SchedulerName: "bar",
						Plugins: &configv1.Plugins{
							QueueSort: configv1.PluginSet{Enabled: []configv1.Plugin{{Name: "PrioritySort"}}},
							Bind:      configv1.PluginSet{Enabled: []configv1.Plugin{{Name: "DefaultBinder"}}},
						},
					},
				)},
			wantProfiles: []string{"bar", "foo"},
		},
		{
			name: "Repeated profiles",
			opts: []Option{
				WithProfiles(
					configv1.KubeSchedulerProfile{
						SchedulerName: "foo",
						Plugins: &configv1.Plugins{
							QueueSort: configv1.PluginSet{Enabled: []configv1.Plugin{{Name: "PrioritySort"}}},
							Bind:      configv1.PluginSet{Enabled: []configv1.Plugin{{Name: "DefaultBinder"}}},
						},
					},
					configv1.KubeSchedulerProfile{
						SchedulerName: "bar",
						Plugins: &configv1.Plugins{
							QueueSort: configv1.PluginSet{Enabled: []configv1.Plugin{{Name: "PrioritySort"}}},
							Bind:      configv1.PluginSet{Enabled: []configv1.Plugin{{Name: "DefaultBinder"}}},
						},
					},
					configv1.KubeSchedulerProfile{
						SchedulerName: "foo",
						Plugins: &configv1.Plugins{
							QueueSort: configv1.PluginSet{Enabled: []configv1.Plugin{{Name: "PrioritySort"}}},
							Bind:      configv1.PluginSet{Enabled: []configv1.Plugin{{Name: "DefaultBinder"}}},
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
