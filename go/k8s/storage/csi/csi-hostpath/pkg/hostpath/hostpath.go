package hostpath

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/golang/protobuf/ptypes/timestamp"
)

type hostPath struct {
	name              string
	nodeID            string
	version           string
	endpoint          string
	ephemeral         bool
	maxVolumesPerNode int64

	ids *identityServer
	ns  *nodeServer
	cs  *controllerServer
}

type hostPathVolume struct {
	VolName       string     `json:"volName"`
	VolID         string     `json:"volID"`
	VolSize       int64      `json:"volSize"`
	VolPath       string     `json:"volPath"`
	VolAccessType accessType `json:"volAccessType"`
	ParentVolID   string     `json:"parentVolID,omitempty"`
	ParentSnapID  string     `json:"parentSnapID,omitempty"`
	Ephemeral     bool       `json:"ephemeral"`
	NodeID        string     `json:"nodeID"`
}

type hostPathSnapshot struct {
	Name         string               `json:"name"`
	Id           string               `json:"id"`
	VolID        string               `json:"volID"`
	Path         string               `json:"path"`
	CreationTime *timestamp.Timestamp `json:"creationTime"`
	SizeBytes    int64                `json:"sizeBytes"`
	ReadyToUse   bool                 `json:"readyToUse"`
}

var (
	vendorVersion = "dev"

	hostPathVolumes         = map[string]hostPathVolume{}
	hostPathVolumeSnapshots = map[string]hostPathSnapshot{}
)

const (
	// Directory where data for volumes and snapshots are persisted.
	// This can be ephemeral within the container or persisted if
	// backed by a Pod volume.
	dataRoot = "/tmp/csi-data-dir"

	// Extension with which snapshot files will be saved.
	snapshotExt = ".snap"
)

func NewHostPathDriver(driverName, nodeID, endpoint string, ephemeral bool, maxVolumesPerNode int64, version string) (*hostPath, error) {
	if driverName == "" {
		return nil, errors.New("no driver name provided")
	}

	if nodeID == "" {
		return nil, errors.New("no node id provided")
	}

	if endpoint == "" {
		return nil, errors.New("no driver endpoint provided")
	}
	if version != "" {
		vendorVersion = version
	}

	if err := os.MkdirAll(dataRoot, 0750); err != nil {
		return nil, fmt.Errorf("failed to create dataRoot: %v", err)
	}

	glog.Infof("Driver: %v ", driverName)
	glog.Infof("Version: %s", vendorVersion)

	return &hostPath{
		name:              driverName,
		version:           vendorVersion,
		nodeID:            nodeID,
		endpoint:          endpoint,
		ephemeral:         ephemeral,
		maxVolumesPerNode: maxVolumesPerNode,
	}, nil
}

/*
findmnt --json
{
"filesystems": [
	{"target": "/", "source": "/dev/vda1", "fstype": "ext4", "options": "rw,relatime,errors=remount-ro,data=ordered",
		"children": [
		{"target": "/sys", "source": "sysfs", "fstype": "sysfs", "options": "rw,nosuid,nodev,noexec,relatime",
		...
*/
func discoveryExistingVolumes() error {
	cmdPath := locateCommandPath("findmnt")
	out, err := exec.Command(cmdPath, "--json").CombinedOutput()
	if err != nil {
		glog.V(3).Infof("failed to execute command: %+v", cmdPath)
		return err
	}

	if len(out) < 1 {
		return fmt.Errorf("mount point info is nil")
	}

	mountInfos, err := parseMountInfo(out)
	if err != nil {
		return fmt.Errorf("failed to parse the mount infos: %+v", err)
	}

	mountInfosOfPod := MountPointInfo{}
	for _, mountInfo := range mountInfos {
		if mountInfo.Target == podVolumeTargetPath {
			mountInfosOfPod = mountInfo
			break
		}
	}

	// getting existing volumes based on the mount point infos.
	// It's a temporary solution to recall volumes.
	for _, pv := range mountInfosOfPod.ContainerFileSystem {
		if !strings.Contains(pv.Target, csiSignOfVolumeTargetPath) {
			continue
		}

		hp, err := parseVolumeInfo(pv)
		if err != nil {
			return err
		}

		hostPathVolumes[hp.VolID] = *hp
	}

	glog.V(4).Infof("Existing Volumes: %+v", hostPathVolumes)
	return nil

}

func (hp *hostPath) Run() error {
	// Create GRPC servers
	if err := discoveryExistingVolumes(); err != nil {
		return err
	}

}
