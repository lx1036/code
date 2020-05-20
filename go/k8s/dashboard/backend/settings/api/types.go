package api

import "k8s.io/client-go/kubernetes"

// SettingsManager is used for user settings management.
type SettingsManager interface {
	// GetGlobalSettings gets current global settings from config map.
	GetGlobalSettings(client kubernetes.Interface) (s Settings)
	// SaveGlobalSettings saves provided global settings in config map.
	SaveGlobalSettings(client kubernetes.Interface, s *Settings) error
	// GetPinnedResources gets the pinned resources from config map.
	GetPinnedResources(client kubernetes.Interface) (r []PinnedResource)
	// SavePinnedResource adds a new pinned resource to config map.
	SavePinnedResource(client kubernetes.Interface, r *PinnedResource) error
	// DeletePinnedResource removes a pinned resource from config map.
	DeletePinnedResource(client kubernetes.Interface, r *PinnedResource) error
}

// PinnedResource represents a pinned resource.
type PinnedResource struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

// Settings is a single instance of settings without context.
type Settings struct {
	ClusterName                     string `json:"clusterName"`
	ItemsPerPage                    int    `json:"itemsPerPage"`
	LogsAutoRefreshTimeInterval     int    `json:"logsAutoRefreshTimeInterval"`
	ResourceAutoRefreshTimeInterval int    `json:"resourceAutoRefreshTimeInterval"`
}
