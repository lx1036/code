package cgroups

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// V1 returns all the groups in the default cgroups mountpoint in a single hierarchy
func V1() ([]Subsystem, error) {
	// root="/sys/fs/cgroup", 这里用的测试数据root="fixtures"
	root, err := v1MountPoint()
	if err != nil {
		return nil, err
	}

	subsystems, err := defaults(root)
	if err != nil {
		return nil, err
	}
	var enabled []Subsystem
	for _, subsystem := range pathers(subsystems) {
		// check and remove the default groups that do not exist
		if _, err := os.Lstat(subsystem.Path("/")); err == nil {
			enabled = append(enabled, subsystem)
		}
	}

	return enabled, nil
}

var (
	fixturesMemInfoPath = "fixtures/proc/meminfo"
)

func SetFixturesMemInfoPath(path string) {
	fixturesMemInfoPath = path
}
func GetFixturesMemInfoPath() string {
	return fixturesMemInfoPath
}

// v1MountPoint returns the mount point where the cgroup
// mountpoints are mounted in a single hiearchy
func v1MountPoint() (string, error) {
	absFilepath, err := filepath.Abs(GetFixturesMemInfoPath())
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
	return "", ErrMountPointNotExist
}
