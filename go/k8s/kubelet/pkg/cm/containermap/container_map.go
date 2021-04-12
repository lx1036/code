package containermap

import "fmt"

// ContainerMap maps (containerID)->(*v1.Pod, *v1.Container)
type ContainerMap map[string]struct {
	podUID        string
	containerName string
}

func (cm ContainerMap) GetContainerRef(containerID string) (string, string, error) {
	if _, exists := cm[containerID]; !exists {
		return "", "", fmt.Errorf("containerID %s not in ContainerMap", containerID)
	}

	return cm[containerID].podUID, cm[containerID].containerName, nil
}

func (cm ContainerMap) RemoveByContainerRef(podUID, containerName string) {
	containerID, err := cm.GetContainerID(podUID, containerName)
	if err == nil {
		cm.RemoveByContainerID(containerID)
	}
}

func (cm ContainerMap) GetContainerID(podUID, containerName string) (string, error) {
	for key, val := range cm {
		if val.podUID == podUID && val.containerName == containerName {
			return key, nil
		}
	}
	return "", fmt.Errorf("container %s not in ContainerMap for pod %s", containerName, podUID)
}

func (cm ContainerMap) RemoveByContainerID(containerID string) {
	delete(cm, containerID)
}

func (cm ContainerMap) Add(podUID, containerName, containerID string) {
	cm[containerID] = struct {
		podUID        string
		containerName string
	}{podUID, containerName}
}

func NewContainerMap() ContainerMap {
	return make(ContainerMap)
}
