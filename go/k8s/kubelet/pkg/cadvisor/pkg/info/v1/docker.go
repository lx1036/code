package v1

type DockerStatus struct {
	Version       string            `json:"version"`
	APIVersion    string            `json:"api_version"`
	KernelVersion string            `json:"kernel_version"`
	OS            string            `json:"os"`
	Hostname      string            `json:"hostname"`
	RootDir       string            `json:"root_dir"`
	Driver        string            `json:"driver"`
	DriverStatus  map[string]string `json:"driver_status"`
	ExecDriver    string            `json:"exec_driver"`
	NumImages     int               `json:"num_images"`
	NumContainers int               `json:"num_containers"`
}

type DockerImage struct {
	ID          string   `json:"id"`
	RepoTags    []string `json:"repo_tags"` // repository name and tags.
	Created     int64    `json:"created"`   // unix time since creation.
	VirtualSize int64    `json:"virtual_size"`
	Size        int64    `json:"size"`
}
