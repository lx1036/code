package plugin

import (
	"k8s-lx1036/k8s-ui/backend/dashboard/api"
	"k8s-lx1036/k8s-ui/backend/dashboard/controllers/plugin/apis/v1alpha1"
	pluginclientset "k8s-lx1036/k8s-ui/backend/dashboard/controllers/plugin/client/clientset/versioned"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PluginList holds only necessary information and is used to
// map v1alpha1.PluginList to plugin.PluginList
type PluginList struct {
	ListMeta api.ListMeta `json:"listMeta"`
	Items    []Plugin     `json:"items"`
	Errors   []error      `json:"errors"`
}

// PluginList holds only necessary information and is used to
// map v1alpha1.Plugin to plugin.Plugin
type Plugin struct {
	ObjectMeta   api.ObjectMeta `json:"objectMeta"`
	TypeMeta     api.TypeMeta   `json:"typeMeta"`
	Name         string         `json:"name"`
	Path         string         `json:"path"`
	Dependencies []string       `json:"dependencies"`
}

// GetPluginList returns all the registered plugins
func GetPluginList(client pluginclientset.Interface, namespace string) (*PluginList, error) {
	plugins, err := client.DashboardV1alpha1().Plugins(namespace).List(v1.ListOptions{})

	toPluginList(plugins.Items)

}

func toPluginList(plugins []v1alpha1.Plugin, nonCriticalErrors []error, dsQuery *dataselect.DataSelectQuery) *PluginList {
	result := &PluginList{
		Items:    make([]Plugin, 0),
		ListMeta: api.ListMeta{TotalItems: len(plugins)},
		Errors:   nonCriticalErrors,
	}

	for _, item := range plugins {
		result.Items = append(result.Items, toPlugin(item))
	}

	return result
}
