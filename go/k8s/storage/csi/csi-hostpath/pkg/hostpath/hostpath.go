package hostpath

import (
	"errors"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"os/exec"

	"github.com/golang/glog"
	"github.com/golang/protobuf/ptypes/timestamp"
	
	"k8s.io/kubernetes/pkg/volume/util/fs"
	utilexec "k8s.io/utils/exec"
	"k8s.io/kubernetes/pkg/volume/util/volumepathhandler"
)

const (
	kib    int64 = 1024
	mib    int64 = kib * 1024
	gib    int64 = mib * 1024
	gib100 int64 = gib * 100
	tib    int64 = gib * 1024
	tib100 int64 = tib * 100
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
	dataRoot = "/csi-data-dir"

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

func (hp *hostPath) Run() error {
	// Create GRPC servers
	if err := discoveryExistingVolumes(); err != nil {
		return err
	}
	
	hp.ids = NewIdentityServer(hp.name, hp.version)
	hp.ns = NewNodeServer(hp.nodeID, hp.ephemeral, hp.maxVolumesPerNode)
	hp.cs = NewControllerServer(hp.ephemeral, hp.nodeID)
	
	discoverExistingSnapshots()
	s := NewNonBlockingGRPCServer()
	s.Start(hp.endpoint, hp.ids, hp.cs, hp.ns)
	s.Wait()
	
	return nil
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

func discoverExistingSnapshots() {
	glog.V(4).Infof("discovering existing snapshots in %s", dataRoot)
	files, err := ioutil.ReadDir(dataRoot)
	if err != nil {
		glog.Errorf("failed to discover snapshots under %s: %v", dataRoot, err)
	}
	for _, file := range files {
		isSnapshot, snapshotID := getSnapshotID(file.Name())
		if isSnapshot {
			glog.V(4).Infof("adding snapshot %s from file %s", snapshotID, getSnapshotPath(snapshotID))
			hostPathVolumeSnapshots[snapshotID] = hostPathSnapshot{
				Id:         snapshotID,
				Path:       getSnapshotPath(snapshotID),
				ReadyToUse: true,
			}
		}
	}
}

// getSnapshotPath returns the full path to where the snapshot is stored
func getSnapshotPath(snapshotID string) string {
	return filepath.Join(dataRoot, fmt.Sprintf("%s%s", snapshotID, snapshotExt))
}

func getSnapshotID(file string) (bool, string) {
	glog.V(4).Infof("file: %s", file)
	// Files with .snap extension are volumesnapshot files.
	// e.g. foo.snap, foo.bar.snap
	if filepath.Ext(file) == snapshotExt {
		return true, strings.TrimSuffix(file, snapshotExt)
	}
	return false, ""
}

// loadFromSnapshot populates the given destPath with data from the snapshotID
func loadFromSnapshot(size int64, snapshotId, destPath string, mode accessType) error {
	snapshot, ok := hostPathVolumeSnapshots[snapshotId]
	if !ok {
		return status.Errorf(codes.NotFound, "cannot find snapshot %v", snapshotId)
	}
	if snapshot.ReadyToUse != true {
		return status.Errorf(codes.Internal, "snapshot %v is not yet ready to use.", snapshotId)
	}
	if snapshot.SizeBytes > size {
		return status.Errorf(codes.InvalidArgument, "snapshot %v size %v is greater than requested volume size %v", snapshotId, snapshot.SizeBytes, size)
	}
	snapshotPath := snapshot.Path
	
	var cmd []string
	switch mode {
	case mountAccess:
		cmd = []string{"tar", "zxvf", snapshotPath, "-C", destPath}
	case blockAccess:
		cmd = []string{"dd", "if=" + snapshotPath, "of=" + destPath}
	default:
		return status.Errorf(codes.InvalidArgument, "unknown accessType: %d", mode)
	}
	
	executor := utilexec.New()
	glog.V(4).Infof("Command Start: %v", cmd)
	out, err := executor.Command(cmd[0], cmd[1:]...).CombinedOutput()
	glog.V(4).Infof("Command Finish: %v", string(out))
	if err != nil {
		return status.Errorf(codes.Internal, "failed pre-populate data from snapshot %v: %v: %s", snapshotId, err, out)
	}
	return nil
}

// loadFromVolume populates the given destPath with data from the srcVolumeID
func loadFromVolume(size int64, srcVolumeId, destPath string, mode accessType) error {
	hostPathVolume, ok := hostPathVolumes[srcVolumeId]
	if !ok {
		return status.Error(codes.NotFound, "source volumeId does not exist, are source/destination in the same storage class?")
	}
	if hostPathVolume.VolSize > size {
		return status.Errorf(codes.InvalidArgument, "volume %v size %v is greater than requested volume size %v", srcVolumeId, hostPathVolume.VolSize, size)
	}
	if mode != hostPathVolume.VolAccessType {
		return status.Errorf(codes.InvalidArgument, "volume %v mode is not compatible with requested mode", srcVolumeId)
	}
	
	switch mode {
	case mountAccess:
		return loadFromFilesystemVolume(hostPathVolume, destPath)
	case blockAccess:
		return loadFromBlockVolume(hostPathVolume, destPath)
	default:
		return status.Errorf(codes.InvalidArgument, "unknown accessType: %d", mode)
	}
}

// hostPathIsEmpty is a simple check to determine if the specified hostpath directory
// is empty or not.
func hostPathIsEmpty(p string) (bool, error) {
	f, err := os.Open(p)
	if err != nil {
		return true, fmt.Errorf("unable to open hostpath volume, error: %v", err)
	}
	defer f.Close()
	
	_, err = f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	
	return false, err
}

func loadFromFilesystemVolume(hostPathVolume hostPathVolume, destPath string) error {
	srcPath := hostPathVolume.VolPath
	isEmpty, err := hostPathIsEmpty(srcPath)
	if err != nil {
		return status.Errorf(codes.Internal, "failed verification check of source hostpath volume %v: %v", hostPathVolume.VolID, err)
	}
	
	// If the source hostpath volume is empty it's a noop and we just move along, otherwise the cp call will fail with a a file stat error DNE
	if !isEmpty {
		args := []string{"-a", srcPath + "/.", destPath + "/"}
		executor := utilexec.New()
		out, err := executor.Command("cp", args...).CombinedOutput()
		if err != nil {
			return status.Errorf(codes.Internal, "failed pre-populate data from volume %v: %v: %s", hostPathVolume.VolID, err, out)
		}
	}
	return nil
}

func loadFromBlockVolume(hostPathVolume hostPathVolume, destPath string) error {
	srcPath := hostPathVolume.VolPath
	args := []string{"if=" + srcPath, "of=" + destPath}
	executor := utilexec.New()
	out, err := executor.Command("dd", args...).CombinedOutput()
	if err != nil {
		return status.Errorf(codes.Internal, "failed pre-populate data from volume %v: %v: %s", hostPathVolume.VolID, err, out)
	}
	return nil
}

func getVolumeByName(volName string) (hostPathVolume, error) {
	for _, hostPathVol := range hostPathVolumes {
		if hostPathVol.VolName == volName {
			return hostPathVol, nil
		}
	}
	return hostPathVolume{}, fmt.Errorf("volume name %s does not exist in the volumes list", volName)
}

func parseVolumeInfo(volume MountPointInfo) (*hostPathVolume, error) {
	volumeName := filterVolumeName(volume.Target)
	volumeID := filterVolumeID(volume.Source)
	sourcePath := getSourcePath(volumeID)
	
	glog.V(4).Infof("parseVolumeInfo: volumeName %s, volumeID %s, sourcePath %s", volumeName, volumeID, sourcePath)
	
	_, fscapacity, _, _, _, _, err := fs.FsInfo(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get capacity info: %+v", err)
	}
	
	hp := hostPathVolume{
		VolName:       volumeName,
		VolID:         volumeID,
		VolSize:       fscapacity,
		VolPath:       getVolumePath(volumeID),
		VolAccessType: mountAccess,
	}
	
	return &hp, nil
}

func filterVolumeName(targetPath string) string {
	pathItems := strings.Split(targetPath, "kubernetes.io~csi/")
	if len(pathItems) < 2 {
		return ""
	}
	
	return strings.TrimSuffix(pathItems[1], "/mount")
}

func filterVolumeID(sourcePath string) string {
	volumeSourcePathRegex := regexp.MustCompile(`\[(.*)\]`)
	volumeSP := string(volumeSourcePathRegex.Find([]byte(sourcePath)))
	if volumeSP == "" {
		return ""
	}
	
	return strings.TrimSuffix(strings.TrimPrefix(volumeSP, "[/var/lib/csi-hostpath-data/"), "]")
}

// deleteVolume deletes the directory for the hostpath volume.
func deleteHostpathVolume(volID string) error {
	glog.V(4).Infof("deleting hostpath volume: %s", volID)
	
	vol, err := getVolumeByID(volID)
	if err != nil {
		// Return OK if the volume is not found.
		return nil
	}
	
	if vol.VolAccessType == blockAccess {
		volPathHandler := volumepathhandler.VolumePathHandler{}
		path := getVolumePath(volID)
		glog.V(4).Infof("deleting loop device for file %s if it exists", path)
		if err := volPathHandler.DetachFileDevice(path); err != nil {
			return fmt.Errorf("failed to remove loop device for file %s: %v", path, err)
		}
	}
	
	path := getVolumePath(volID)
	if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	
	delete(hostPathVolumes, volID)
	return nil
}

func getVolumeByID(volumeID string) (hostPathVolume, error) {
	if hostPathVol, ok := hostPathVolumes[volumeID]; ok {
		return hostPathVol, nil
	}
	return hostPathVolume{}, fmt.Errorf("volume id %s does not exist in the volumes list", volumeID)
}

// getVolumePath returns the canonical path for hostpath volume
func getVolumePath(volID string) string {
	return filepath.Join(dataRoot, volID)
}

func getSortedVolumeIDs() []string {
	ids := make([]string, len(hostPathVolumes))
	index := 0
	for volId := range hostPathVolumes {
		ids[index] = volId
		index += 1
	}
	
	sort.Strings(ids)
	return ids
}

// createVolume create the directory for the hostpath volume.
// It returns the volume path or err if one occurs.
func createHostpathVolume(volID, name string, cap int64, volAccessType accessType, ephemeral bool) (*hostPathVolume, error) {
	path := getVolumePath(volID)
	
	switch volAccessType {
	case mountAccess:
		err := os.MkdirAll(path, 0777)
		if err != nil {
			return nil, err
		}
	case blockAccess:
		executor := utilexec.New()
		size := fmt.Sprintf("%dM", cap/mib)
		// Create a block file.
		_, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				out, err := executor.Command("fallocate", "-l", size, path).CombinedOutput()
				if err != nil {
					return nil, fmt.Errorf("failed to create block device: %v, %v", err, string(out))
				}
			} else {
				return nil, fmt.Errorf("failed to stat block device: %v, %v", path, err)
			}
		}
		
		// Associate block file with the loop device.
		volPathHandler := volumepathhandler.VolumePathHandler{}
		_, err = volPathHandler.AttachFileDevice(path)
		if err != nil {
			// Remove the block file because it'll no longer be used again.
			if err2 := os.Remove(path); err2 != nil {
				glog.Errorf("failed to cleanup block file %s: %v", path, err2)
			}
			return nil, fmt.Errorf("failed to attach device %v: %v", path, err)
		}
	default:
		return nil, fmt.Errorf("unsupported access type %v", volAccessType)
	}
	
	hostpathVol := hostPathVolume{
		VolID:         volID,
		VolName:       name,
		VolSize:       cap,
		VolPath:       path,
		VolAccessType: volAccessType,
		Ephemeral:     ephemeral,
	}
	hostPathVolumes[volID] = hostpathVol
	return &hostpathVol, nil
}
