package cgroups

import (
	"bytes"
	"fmt"
	"github.com/opencontainers/runtime-spec/specs-go"
	"io/ioutil"
	"os"
	"path/filepath"
)

type cpusetController struct {
	root string // fixtures/cpuset
}

func (c *cpusetController) Name() Name {
	return Cpuset
}

// Path cpuset/test
func (c *cpusetController) Path(path string) string {
	return filepath.Join(c.root, path)
}

func (c *cpusetController) Create(path string, resources *specs.LinuxResources) error {
	if err := c.ensureParent(c.Path(path), c.root); err != nil {
		return err
	}
	if err := os.MkdirAll(c.Path(path), defaultDirPerm); err != nil {
		return err
	}

	// 从 parent cgroup 中copy个值
	if err := c.copyIfNeeded(c.Path(path), filepath.Dir(c.Path(path))); err != nil {
		return err
	}

	// 然后根据自定义的值，更新下cpuset.cpus cpuset.mems
	// INFO: 这里的 ioutil.WriteFile 是覆盖写，不是 append 追加写，所以也就是 Update()
	if resources.CPU != nil {
		if len(resources.CPU.Cpus) != 0 {
			if err := ioutil.WriteFile(filepath.Join(c.Path(path), "cpuset.cpus"), []byte(resources.CPU.Cpus), os.FileMode(0666)); err != nil {
				return err
			}
		}

		if len(resources.CPU.Mems) != 0 {
			if err := ioutil.WriteFile(filepath.Join(c.Path(path), "cpuset.mems"), []byte(resources.CPU.Mems), os.FileMode(0666)); err != nil {
				return err
			}
		}
	}

	return nil
}

// 当创建级联cgroup /test1/test2 时，确保 test1 parent cgroup 下得有 cpus/mems files，从parent cgroup 拷贝
func (c *cpusetController) ensureParent(current, root string) error {
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
		if err := c.ensureParent(parent, root); err != nil {
			return err
		}
	}

	if err := os.MkdirAll(current, defaultDirPerm); err != nil {
		return err
	}

	return c.copyIfNeeded(current, parent)
}

// child cgroups 如果 cpuset.cpus/cpuset.mems 内容为空，从 parent cgroups 中拷贝
func (c *cpusetController) copyIfNeeded(current, parent string) error {
	var (
		err                      error
		currentCpus, currentMems []byte
		parentCpus, parentMems   []byte
	)
	if currentCpus, currentMems, err = c.getValues(current); err != nil {
		return err
	}
	if parentCpus, parentMems, err = c.getValues(parent); err != nil {
		return err
	}

	// 写 cpuset.cpus 和 cpuset.mems 文件
	if isEmpty(currentCpus) {
		// INFO: 生产环境里这里应该是 os.FileMode(0)
		if err := ioutil.WriteFile(filepath.Join(current, "cpuset.cpus"), parentCpus, os.FileMode(0666)); err != nil {
			return err
		}
	}
	if isEmpty(currentMems) {
		if err := ioutil.WriteFile(filepath.Join(current, "cpuset.mems"), parentMems, os.FileMode(0666)); err != nil {
			return err
		}
	}

	return nil
}

func isEmpty(b []byte) bool {
	return len(bytes.Trim(b, "\n")) == 0
}

func (c *cpusetController) getValues(path string) (cpus []byte, mems []byte, err error) {
	if cpus, err = ioutil.ReadFile(filepath.Join(path, "cpuset.cpus")); err != nil && !os.IsNotExist(err) {
		return
	}
	if mems, err = ioutil.ReadFile(filepath.Join(path, "cpuset.mems")); err != nil && !os.IsNotExist(err) {
		return
	}

	return cpus, mems, nil
}

func (c *cpusetController) Update(path string, resources *specs.LinuxResources) error {
	return c.Create(path, resources)
}

func NewCpuset(root string) *cpusetController {
	return &cpusetController{
		root: filepath.Join(root, string(Cpuset)),
	}
}
