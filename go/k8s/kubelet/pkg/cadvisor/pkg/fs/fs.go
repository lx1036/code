package fs

import (
	"fmt"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/utils/mount"
)

const (
	LabelSystemRoot = "root"
)

type partition struct {
	mountpoint string
	major      uint
	minor      uint
	fsType     string
	blockSize  uint
}

type RealFsInfo struct {
	// Map from block device path to partition information.
	partitions map[string]partition
	// Map from label to block device path.
	// Labels are intent-specific tags that are auto-detected.
	labels map[string]string
	// Map from mountpoint to mount information.
	mounts map[string]mount.MountInfo
	// devicemapper client
	//dmsetup devicemapper.DmsetupClient
	// fsUUIDToDeviceName is a map from the filesystem UUID to its device name.
	fsUUIDToDeviceName map[string]string
}

func processMounts(mounts []mount.MountInfo, excludedMountpointPrefixes []string) map[string]partition {
	partitions := make(map[string]partition)

	supportedFsType := map[string]bool{
		// all ext systems are checked through prefix.
		"btrfs":   true,
		"overlay": true,
		"tmpfs":   true,
		"xfs":     true,
		"zfs":     true,
	}

	for _, mount := range mounts {
		if !strings.HasPrefix(mount.FsType, "ext") && !supportedFsType[mount.FsType] {
			continue
		}
		// Avoid bind mounts, exclude tmpfs.
		if _, ok := partitions[mount.Source]; ok {
			if mount.FsType != "tmpfs" {
				continue
			}
		}

		hasPrefix := false
		for _, prefix := range excludedMountpointPrefixes {
			if strings.HasPrefix(mount.MountPoint, prefix) {
				hasPrefix = true
				break
			}
		}
		if hasPrefix {
			continue
		}

		// using mountpoint to replace device once fstype it tmpfs
		if mount.FsType == "tmpfs" {
			mount.Source = mount.MountPoint
		}

		// overlay fix: Making mount source unique for all overlay mounts, using the mount's major and minor ids.
		if mount.FsType == "overlay" {
			mount.Source = fmt.Sprintf("%s_%d-%d", mount.Source, mount.Major, mount.Minor)
		}

		partitions[mount.Source] = partition{
			fsType:     mount.FsType,
			mountpoint: mount.MountPoint,
			major:      uint(mount.Major),
			minor:      uint(mount.Minor),
		}
	}

	return partitions
}

func (i *RealFsInfo) GetGlobalFsInfo() ([]Fs, error) {
	panic("implement me")
}

func (i *RealFsInfo) GetFsInfoForPath(mountSet map[string]struct{}) ([]Fs, error) {
	panic("implement me")
}

func (i *RealFsInfo) GetDirUsage(dir string) (UsageInfo, error) {
	panic("implement me")
}

func (i *RealFsInfo) GetDeviceInfoByFsUUID(uuid string) (*DeviceInfo, error) {
	panic("implement me")
}

func (i *RealFsInfo) GetDirFsDevice(dir string) (*DeviceInfo, error) {
	panic("implement me")
}

func (i *RealFsInfo) GetDeviceForLabel(label string) (string, error) {
	panic("implement me")
}

func (i *RealFsInfo) GetLabelsForDevice(device string) ([]string, error) {
	panic("implement me")
}

func (i *RealFsInfo) GetMountpointForDevice(device string) (string, error) {
	panic("implement me")
}

// addSystemRootLabel attempts to determine which device contains the mount for /.
func (i *RealFsInfo) addSystemRootLabel(mounts []mount.MountInfo) {
	for _, m := range mounts {
		if m.MountPoint == "/" {
			i.partitions[m.Source] = partition{
				fsType:     m.FsType,
				mountpoint: m.MountPoint,
				major:      uint(m.Major),
				minor:      uint(m.Minor),
			}
			i.labels[LabelSystemRoot] = m.Source
			return
		}
	}
}

func NewFsInfo(context Context) (FsInfo, error) {
	mounts, err := mount.ParseMountInfo("fixtures/proc/self/mountinfo")
	if err != nil {
		return nil, err
	}

	// Avoid devicemapper container mounts - these are tracked by the ThinPoolWatcher
	excluded := []string{fmt.Sprintf("%s/devicemapper/mnt", context.Docker.Root)}
	fsInfo := &RealFsInfo{
		partitions: processMounts(mounts, excluded),
		labels:     make(map[string]string),
		mounts:     make(map[string]mount.MountInfo),
	}

	for _, mount := range mounts {
		fsInfo.mounts[mount.MountPoint] = mount
	}

	klog.V(1).Infof("Filesystem UUIDs: %+v", fsInfo.fsUUIDToDeviceName)
	klog.V(1).Infof("Filesystem partitions: %+v", fsInfo.partitions)
	fsInfo.addSystemRootLabel(mounts)

	return fsInfo, nil
}
