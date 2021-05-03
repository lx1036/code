package common

type Mount struct {
	HostDir      string `json:"host_dir,omitempty"`
	ContainerDir string `json:"container_dir,omitempty"`
}
