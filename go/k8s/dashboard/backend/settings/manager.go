package settings

import (
	"k8s.io/client-go/kubernetes"
)

// SettingsManager is a structure containing all settings manager members.
type SettingsManager struct {
	settings        map[string]api.Settings
	pinnedResources []api.PinnedResource
	rawSettings     map[string]string
}

func (s2 SettingsManager) GetGlobalSettings(client kubernetes.Interface) (s api.Settings) {
	panic("implement me")
}

func (s2 SettingsManager) SaveGlobalSettings(client kubernetes.Interface, s *api.Settings) error {
	panic("implement me")
}

func (s2 SettingsManager) GetPinnedResources(client kubernetes.Interface) (r []api.PinnedResource) {
	panic("implement me")
}

func (s2 SettingsManager) SavePinnedResource(client kubernetes.Interface, r *api.PinnedResource) error {
	panic("implement me")
}

func (s2 SettingsManager) DeletePinnedResource(client kubernetes.Interface, r *api.PinnedResource) error {
	panic("implement me")
}

// NewSettingsManager creates new settings manager.
func NewSettingsManager() api.SettingsManager {
	return &SettingsManager{
		settings:        make(map[string]api.Settings),
		pinnedResources: []api.PinnedResource{},
	}
}
