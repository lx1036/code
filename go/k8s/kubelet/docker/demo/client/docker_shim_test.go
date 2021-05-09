package main

import (
	"flag"
	"testing"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	dockercontainer "github.com/docker/docker/api/types/container"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/kubelet/dockershim"
)

var (
	containerID = flag.String("container_id", "", "")
)

// GOOS=linux GOARCH=amd64 go build .
func TestDockerClient(test *testing.T) {
	flag.Parse()

	klog.Infof("container ID: %s", *containerID)

	dockerEndpoint := "unix:///var/run/docker.sock" // linux/mac
	dockerClient := dockershim.NewDockerClientFromConfig(&dockershim.ClientConfig{
		DockerEndpoint:            dockerEndpoint,
		RuntimeRequestTimeout:     time.Minute * 2,
		ImagePullProgressDeadline: time.Minute * 1,
	})

	containers, err := dockerClient.ListContainers(dockertypes.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	klog.Infof("%d containers", len(containers))

	imagesSummary, err := dockerClient.ListImages(dockertypes.ImageListOptions{All: true})
	if err != nil {
		panic(err)
	}

	for _, image := range imagesSummary {
		klog.Infof("image: %v", image)
	}

	if len(*containerID) == 0 {
		return
	}

	for _, container := range containers {
		if container.ID != *containerID {
			continue
		}

		klog.Infof("container: %v", container)

		// docker-on-mac 上修改 cpuset.cpus
		err = dockerClient.UpdateContainerResources(container.ID, dockercontainer.UpdateConfig{
			Resources: dockercontainer.Resources{
				CpusetCpus: "1,17",
			},
		})
		if err != nil {
			klog.Error(err)
		}
	}
}
