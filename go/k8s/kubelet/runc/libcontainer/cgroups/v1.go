package cgroups

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Gets the cgroupRoot.
func getCgroupRoot() (string, error) {
	cgroupRootLock.Lock()
	defer cgroupRootLock.Unlock()

	// root="/sys/fs/cgroup", 这里用的测试数据root="fixtures"
	root, err := v1MountPoint()
	if err != nil {
		return "", err
	}

	return root, nil
}

const MountInfo = "fixtures/proc/self/mountinfo"

// v1MountPoint returns the mount point where the cgroup
// mountpoints are mounted in a single hiearchy
func v1MountPoint() (string, error) {
	absFilepath, err := filepath.Abs(MountInfo)
	if err != nil {
		return "", err
	}
	f, err := os.Open(absFilepath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var (
			text      = scanner.Text()
			fields    = strings.Split(text, " ")
			numFields = len(fields)
		)
		if numFields < 10 {
			return "", fmt.Errorf("mountinfo: bad entry %q", text)
		}
		if fields[numFields-3] == "cgroup" {
			return filepath.Dir(fields[4]), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("cgroups: cgroup mountpoint does not exist")
}
