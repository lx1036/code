package cgroups

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"k8s-lx1036/k8s/kubelet/runc/libcontainer/configs"
)

const (
	cgroupCpusetCpus = "cpuset.cpus"
	cgroupCpusetMems = "cpuset.mems"
	cgroupProcs      = "cgroup.procs"
)

type CpusetGroup struct {
}

func (cpusetGroup *CpusetGroup) Name() string {
	return Cpuset
}

func (cpusetGroup *CpusetGroup) Apply(data *cgroupData) error {
	dir, err := data.path(Cpuset)
	if err != nil && !IsNotFound(err) {
		return err
	}

	return cpusetGroup.ApplyDir(dir, data, data.config, data.pid)
}

func (cpusetGroup *CpusetGroup) ApplyDir(dir string, data *cgroupData, cgroup *configs.Cgroup, pid int) error {
	if dir == "" {
		return nil
	}

	if err := cpusetGroup.cpusetEnsureParent(filepath.Dir(dir), data.root); err != nil {
		return err
	}
	if err := os.Mkdir(dir, 0755); err != nil && !os.IsExist(err) {
		return err
	}

	// 从 parent cgroup 先 copy 一份
	if err := cpusetGroup.copyIfNeeded(dir, filepath.Dir(dir)); err != nil {
		return err
	}

	if data.config != nil {
		if len(data.config.CpusetCpus) != 0 {
			if err := ioutil.WriteFile(filepath.Join(dir, cgroupCpusetCpus), []byte(data.config.CpusetCpus), os.FileMode(0666)); err != nil {
				return err
			}
		}

		if len(data.config.CpusetMems) != 0 {
			if err := ioutil.WriteFile(filepath.Join(dir, cgroupCpusetMems), []byte(data.config.CpusetMems), os.FileMode(0666)); err != nil {
				return err
			}
		}
	}

	if err := ioutil.WriteFile(filepath.Join(dir, cgroupProcs), []byte(strconv.Itoa(data.pid)), os.FileMode(0666)); err != nil {
		return err
	}

	return nil
}

// 当创建级联cgroup /test1/test2 时，确保 test1 parent cgroup 下得有 cpus/mems files，从parent cgroup 拷贝
// INFO: 参考 go/k8s/kubelet/containerd/cgroups/pkg/cgroups/cpuset.go::ensureParent() 函数
func (cpusetGroup *CpusetGroup) cpusetEnsureParent(current, root string) error {
	// current="fixtures/cpuset/test1/test2" root="fixtures/cpuset" parent="fixtures/cpuset/test1"
	parent := filepath.Dir(current)
	// fixtures/cpuset, dir(fixtures/cpuset/test) 两个目录必须能有相对目录，防止 current 瞎写
	if _, err := filepath.Rel(root, parent); err != nil {
		return err
	}

	// Avoid infinite recursion.
	if parent == current {
		return fmt.Errorf("cpuset: cgroup parent path outside cgroup root")
	}

	if cleanPath(parent) != root {
		if err := cpusetGroup.cpusetEnsureParent(parent, root); err != nil {
			return err
		}
	}

	if err := os.MkdirAll(current, defaultDirPerm); err != nil {
		return err
	}

	return cpusetGroup.copyIfNeeded(current, parent)
}

// child cgroups 如果 cpuset.cpus/cpuset.mems 内容为空，从 parent cgroups 中拷贝
func (cpusetGroup *CpusetGroup) copyIfNeeded(current, parent string) error {
	var (
		err                      error
		currentCpus, currentMems []byte
		parentCpus, parentMems   []byte
	)
	if currentCpus, currentMems, err = cpusetGroup.getValues(current); err != nil {
		return err
	}
	if parentCpus, parentMems, err = cpusetGroup.getValues(parent); err != nil {
		return err
	}

	// 写 cpuset.cpus 和 cpuset.mems 文件
	if isEmpty(currentCpus) {
		// INFO: 生产环境里这里应该是 os.FileMode(0)
		if err := ioutil.WriteFile(filepath.Join(current, cgroupCpusetCpus), parentCpus, os.FileMode(0666)); err != nil {
			return err
		}
	}
	if isEmpty(currentMems) {
		if err := ioutil.WriteFile(filepath.Join(current, cgroupCpusetMems), parentMems, os.FileMode(0666)); err != nil {
			return err
		}
	}

	return nil
}

func (cpusetGroup *CpusetGroup) getValues(path string) (cpus []byte, mems []byte, err error) {
	if cpus, err = ioutil.ReadFile(filepath.Join(path, cgroupCpusetCpus)); err != nil && !os.IsNotExist(err) {
		return
	}
	if mems, err = ioutil.ReadFile(filepath.Join(path, cgroupCpusetMems)); err != nil && !os.IsNotExist(err) {
		return
	}

	return cpus, mems, nil
}

func isEmpty(b []byte) bool {
	return len(bytes.Trim(b, "\n")) == 0
}

func (cpusetGroup *CpusetGroup) Set(path string, cgroup *configs.Cgroup) error {
	if cgroup.Resources.CpusetCpus != "" {
		if err := WriteFile(path, cgroupCpusetCpus, cgroup.Resources.CpusetCpus); err != nil {
			return err
		}
	}
	if cgroup.Resources.CpusetMems != "" {
		if err := WriteFile(path, cgroupCpusetMems, cgroup.Resources.CpusetMems); err != nil {
			return err
		}
	}
	return nil
}

func (cpusetGroup *CpusetGroup) GetStats(path string, stats *Stats) error {
	return nil
}
