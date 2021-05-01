package fs

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/utils"

	"k8s.io/klog/v2"
	"k8s.io/utils/mount"
)

const (
	LabelSystemRoot          = "root"
	LabelDockerImages        = "docker-images"
	LabelCrioImages          = "crio-images"
	DriverStatusPoolName     = "Pool Name"
	DriverStatusDataLoopFile = "Data loop file"
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
	return i.GetFsInfoForPath(nil)
}

var (
	fixturesDiskstatsPath = "fixtures/proc/diskstats"
)

func SetFixturesDiskstatsPath(path string) {
	fixturesDiskstatsPath = path
}
func GetFixturesDiskstatsPath() string {
	return fixturesDiskstatsPath
}

var partitionRegex = regexp.MustCompile(`^(?:(?:s|v|xv)d[a-z]+\d*|dm-\d+)$`)

func getDiskStatsMap(diskStatsFile string) (map[string]DiskStats, error) {
	diskStatsMap := make(map[string]DiskStats)
	file, err := os.Open(diskStatsFile)
	if err != nil {
		if os.IsNotExist(err) {
			klog.Warningf("Not collecting filesystem statistics because file %q was not found", diskStatsFile)
			return diskStatsMap, nil
		}
		return nil, err
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		words := strings.Fields(line)
		if !partitionRegex.MatchString(words[2]) {
			continue
		}
		// 8      50 sdd2 40 0 280 223 7 0 22 108 0 330 330
		deviceName := path.Join("/dev", words[2])

		devInfo := make([]uint64, 2)
		for i := 0; i < len(devInfo); i++ {
			devInfo[i], err = strconv.ParseUint(words[i], 10, 64)
			if err != nil {
				return nil, err
			}
		}

		wordLength := len(words)
		offset := 3
		var stats = make([]uint64, wordLength-offset)
		if len(stats) < 11 {
			return nil, fmt.Errorf("could not parse all 11 columns of /proc/diskstats")
		}
		for i := offset; i < wordLength; i++ {
			stats[i-offset], err = strconv.ParseUint(words[i], 10, 64)
			if err != nil {
				return nil, err
			}
		}
		diskStats := DiskStats{
			MajorNum:        devInfo[0],
			MinorNum:        devInfo[1],
			ReadsCompleted:  stats[0],
			ReadsMerged:     stats[1],
			SectorsRead:     stats[2],
			ReadTime:        stats[3],
			WritesCompleted: stats[4],
			WritesMerged:    stats[5],
			SectorsWritten:  stats[6],
			WriteTime:       stats[7],
			IoInProgress:    stats[8],
			IoTime:          stats[9],
			WeightedIoTime:  stats[10],
		}
		diskStatsMap[deviceName] = diskStats
	}
	return diskStatsMap, nil
}

func (i *RealFsInfo) GetFsInfoForPath(mountSet map[string]struct{}) ([]Fs, error) {
	filesystems := make([]Fs, 0)
	deviceSet := make(map[string]struct{})
	diskStatsMap, err := getDiskStatsMap(GetFixturesDiskstatsPath())
	if err != nil {
		return nil, err
	}
	for device, partition := range i.partitions {
		_, hasMount := mountSet[partition.mountpoint]
		_, hasDevice := deviceSet[device]
		if mountSet == nil || (hasMount && !hasDevice) {
			var (
				err error
				fs  Fs
			)
			switch partition.fsType {
			case DeviceMapper.String():
				//fs.Capacity, fs.Free, fs.Available, err = getDMStats(device, partition.blockSize) // no `dmsetup` command in mac
				klog.V(5).Infof("got devicemapper fs capacity stats: capacity: %v free: %v available: %v:", fs.Capacity, fs.Free, fs.Available)
				fs.Type = DeviceMapper
			/*case ZFS.String():
			if _, devzfs := os.Stat("/dev/zfs"); os.IsExist(devzfs) {
				fs.Capacity, fs.Free, fs.Available, err = getZfstats(device)
				fs.Type = ZFS
				break
			}
			// if /dev/zfs is not present default to VFS
			fallthrough*/
			default:
				var inodes, inodesFree uint64
				if utils.FileExists(partition.mountpoint) {
					//fs.Capacity, fs.Free, fs.Available, inodes, inodesFree, err = getVfsStats(partition.mountpoint)
					fs.Inodes = &inodes
					fs.InodesFree = &inodesFree
					fs.Type = VFS
				} else {
					klog.V(4).Infof("unable to determine file system type, partition mountpoint does not exist: %v", partition.mountpoint)
				}
			}
			if err != nil {
				klog.V(4).Infof("Stat fs failed. Error: %v", err)
			} else {
				deviceSet[device] = struct{}{}
				fs.DeviceInfo = DeviceInfo{
					Device: device,
					Major:  uint(partition.major),
					Minor:  uint(partition.minor),
				}

				if val, ok := diskStatsMap[device]; ok {
					fs.DiskStats = val
				} else {
					for k, v := range diskStatsMap {
						if v.MajorNum == uint64(partition.major) && v.MinorNum == uint64(partition.minor) {
							fs.DiskStats = diskStatsMap[k]
							break
						}
					}
				}
				filesystems = append(filesystems, fs)
			}
		}
	}

	return filesystems, nil
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

var (
	fixturesMountInfoPath = "fixtures/proc/self/mountinfo"
)

func SetFixturesMountInfoPath(path string) {
	fixturesMountInfoPath = path
}

func GetFixturesMountInfoPath() string {
	return fixturesMountInfoPath
}

func NewFsInfo(context Context) (FsInfo, error) {
	mountinfoFile, err := filepath.Abs(GetFixturesMountInfoPath())
	if err != nil {
		panic(err)
	}
	mounts, err := mount.ParseMountInfo(mountinfoFile)
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
