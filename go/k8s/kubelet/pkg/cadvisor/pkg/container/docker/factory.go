package docker

import (
	"context"
	"flag"
	"fmt"
	"path"
	"regexp"
	"strings"

	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container/libcontainer"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/fs"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"

	docker "github.com/docker/docker/client"
)

// The namespace under which Docker aliases are unique.
const DockerNamespace = "docker"

var ArgDockerEndpoint = flag.String("docker", "unix:///var/run/docker.sock", "docker endpoint")
var ArgDockerTLS = flag.Bool("docker-tls", false, "use TLS to connect to docker")
var ArgDockerCert = flag.String("docker-tls-cert", "cert.pem", "path to client certificate")
var ArgDockerKey = flag.String("docker-tls-key", "key.pem", "path to private key")
var ArgDockerCA = flag.String("docker-tls-ca", "ca.pem", "path to trusted CA")

type storageDriver string

const (
	devicemapperStorageDriver storageDriver = "devicemapper"
	aufsStorageDriver         storageDriver = "aufs"
	overlayStorageDriver      storageDriver = "overlay"
	overlay2StorageDriver     storageDriver = "overlay2"
	zfsStorageDriver          storageDriver = "zfs"
	vfsStorageDriver          storageDriver = "vfs"
)

type dockerFactory struct {
	machineInfoFactory v1.MachineInfoFactory

	storageDriver storageDriver
	storageDir    string

	client *docker.Client

	// Information about the mounted cgroup subsystems.
	cgroupSubsystems libcontainer.CgroupSubsystems

	// Information about mounted filesystems.
	fsInfo fs.FsInfo

	dockerVersion []int

	dockerAPIVersion []int

	includedMetrics container.MetricSet

	thinPoolName string
	//thinPoolWatcher *devicemapper.ThinPoolWatcher

	//zfsWatcher *zfs.ZfsWatcher
}

var dockerEnvWhitelist = flag.String("docker_env_metadata_whitelist", "",
	"a comma-separated list of environment variable keys matched with specified prefix that needs to be collected for docker containers")

func (f *dockerFactory) NewContainerHandler(name string, inHostNamespace bool) (handler container.ContainerHandler, err error) {
	client, err := Client()
	if err != nil {
		return
	}

	metadataEnvs := strings.Split(*dockerEnvWhitelist, ",")

	handler, err = newDockerContainerHandler(
		client,
		name,
		f.machineInfoFactory,
		f.fsInfo,
		f.storageDriver,
		f.storageDir,
		&f.cgroupSubsystems,
		inHostNamespace,
		metadataEnvs,
		f.dockerVersion,
		f.includedMetrics,
		f.thinPoolName,
		//f.thinPoolWatcher,
		//f.zfsWatcher,
	)

	return
}

// Regexp that identifies docker cgroups, containers started with
// --cgroup-parent have another prefix than 'docker'
// 包含64位字符
var dockerCgroupRegexp = regexp.MustCompile(`([a-z0-9]{64})`)

// isContainerName returns true if the cgroup with associated name
// corresponds to a docker container.
func isContainerName(name string) bool {
	// always ignore .mount cgroup even if associated with docker and delegate to systemd
	if strings.HasSuffix(name, ".mount") {
		return false
	}

	base := path.Base(name)
	return dockerCgroupRegexp.MatchString(base)
}

// Returns the Docker ID from the full container name.
func ContainerNameToDockerId(name string) string {
	id := path.Base(name)

	if matches := dockerCgroupRegexp.FindStringSubmatch(id); matches != nil {
		return matches[1]
	}

	return id
}

func (f *dockerFactory) CanHandleAndAccept(name string) (bool, bool, error) {
	// if the container is not associated with docker, we can't handle it or accept it.
	if !isContainerName(name) {
		return false, false, nil
	}

	// Check if the container is known to docker and it is active.
	id := ContainerNameToDockerId(name)

	// We assume that if Inspect fails then the container is not known to docker.
	ctnr, err := f.client.ContainerInspect(context.Background(), id)
	if err != nil || !ctnr.State.Running {
		return false, true, fmt.Errorf("error inspecting container: %v", err)
	}

	return true, true, nil

}

func (f *dockerFactory) String() string {
	return DockerNamespace
}

func (f *dockerFactory) DebugInfo() map[string][]string {
	return map[string][]string{}
}
