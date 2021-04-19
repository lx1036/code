package fs

import (
	"k8s-lx1036/k8s/kubelet/runc/libcontainer/cgroups"
	"k8s-lx1036/k8s/kubelet/runc/libcontainer/cgroups/fscommon"
	"k8s-lx1036/k8s/kubelet/runc/libcontainer/configs"
)

type CpusetGroup struct {
}

func (s *CpusetGroup) Name() string {
	return "cpuset"
}

func (s *CpusetGroup) Apply(d *cgroupData) error {
	dir, err := d.path("cpuset")
	if err != nil && !cgroups.IsNotFound(err) {
		return err
	}
	return s.ApplyDir(dir, d.config, d.pid)
}

func (s *CpusetGroup) ApplyDir(dir string, cgroup *configs.Cgroup, pid int) error {
	return nil
}

func (s *CpusetGroup) Set(path string, cgroup *configs.Cgroup) error {
	if cgroup.Resources.CpusetCpus != "" {
		if err := fscommon.WriteFile(path, "cpuset.cpus", cgroup.Resources.CpusetCpus); err != nil {
			return err
		}
	}
	if cgroup.Resources.CpusetMems != "" {
		if err := fscommon.WriteFile(path, "cpuset.mems", cgroup.Resources.CpusetMems); err != nil {
			return err
		}
	}
	return nil

}

func (s *CpusetGroup) GetStats(path string, stats *cgroups.Stats) error {
	return nil

}
