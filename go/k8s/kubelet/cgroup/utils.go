package cgroup

import (
	"fmt"
	"github.com/opencontainers/runc/libcontainer/system"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"os"
	"sync"
)

const (
	CgroupProcesses   = "cgroup.procs"
	unifiedMountpoint = "/sys/fs/cgroup"
)

var (
	isUnifiedOnce sync.Once
	isUnified     bool
)

// @see https://github.com/opencontainers/runc/blob/master/libcontainer/cgroups/utils.go

// IsCgroup2UnifiedMode returns whether we are running in cgroup v2 unified mode.
func IsCgroup2UnifiedMode() bool {
	isUnifiedOnce.Do(func() {
		var st unix.Statfs_t
		err := unix.Statfs(unifiedMountpoint, &st)
		if err != nil {
			if os.IsNotExist(err) && system.RunningInUserNS() {
				// ignore the "not found" error if running in userns
				fmt.Println(fmt.Sprintf("%s missing, assuming cgroup v1", unifiedMountpoint))
				logrus.WithError(err).Debugf("%s missing, assuming cgroup v1", unifiedMountpoint)
				isUnified = false
				return
			}
			panic(fmt.Sprintf("cannot statfs cgroup root: %s", err))
		}
		fmt.Println(st.Type)
		isUnified = st.Type == unix.CGROUP2_SUPER_MAGIC
	})

	return isUnified
}
