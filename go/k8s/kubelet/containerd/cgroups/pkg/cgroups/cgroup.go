package cgroups

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	v1 "github.com/containerd/cgroups/stats/v1"
	"github.com/opencontainers/runtime-spec/specs-go"
	"k8s.io/klog/v2"
)

const (
	cgroupProcs    = "cgroup.procs"
	cgroupTasks    = "tasks"
	defaultDirPerm = 0755
)

type Process struct {
	// Subsystem is the name of the subsystem that the process is in
	Subsystem Name
	// Pid is the process id of the process
	Pid int
	// Path is the full path of the subsystem and location that the process is in
	Path string
}

type Task struct {
	// Subsystem is the name of the subsystem that the task is in
	Subsystem Name
	// Pid is the process id of the task
	Pid int
	// Path is the full path of the subsystem and location that the task is in
	Path string
}

// Cgroup handles interactions with the individual groups to perform
// actions on them as them main interface to this cgroup package
type Cgroup interface {
	// New creates a new cgroup under the calling cgroup
	New(string, *specs.LinuxResources) (Cgroup, error)
	// Add adds a process to the cgroup (cgroup.procs)
	Add(Process) error
	// AddTask adds a process to the cgroup (tasks)
	AddTask(Process) error
	// Delete removes the cgroup as a whole
	Delete() error
	// MoveTo moves all the processes under the calling cgroup to the provided one
	// subsystems are moved one at a time
	MoveTo(Cgroup) error
	// Stat returns the stats for all subsystems in the cgroup
	Stat(...ErrorHandler) (*v1.Metrics, error)
	// Update updates all the subsystems with the provided resource changes
	Update(resources *specs.LinuxResources) error
	// Processes returns all the processes in a select subsystem for the cgroup
	Processes(Name, bool) ([]Process, error)
	// Tasks returns all the tasks in a select subsystem for the cgroup
	Tasks(Name, bool) ([]Task, error)
	// Freeze freezes or pauses all processes inside the cgroup
	Freeze() error
	// Thaw thaw or resumes all processes inside the cgroup
	Thaw() error
	// OOMEventFD returns the memory subsystem's event fd for OOM events
	OOMEventFD() (uintptr, error)
	// RegisterMemoryEvent returns the memory subsystems event fd for whatever memory event was
	// registered for. Can alternatively register for the oom event with this method.
	//RegisterMemoryEvent(MemoryEvent) (uintptr, error)
	// State returns the cgroups current state
	State() State
	// Subsystems returns all the subsystems in the cgroup
	Subsystems() []Subsystem
}

type cgroup struct {
	path Path

	subsystems []Subsystem // 注册的所有 subsystem
	mu         sync.Mutex
	err        error
}

func (c *cgroup) New(s string, resources *specs.LinuxResources) (Cgroup, error) {
	panic("implement me")
}

// Add 把 pid 写入 cgroup.procs 文件
func (c *cgroup) Add(process Process) error {
	if process.Pid <= 0 {
		return ErrInvalidPid
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.err != nil {
		return c.err
	}

	for _, s := range pathers(c.subsystems) {
		p, err := c.path(s.Name())
		if err != nil {
			return err
		}

		if err := ioutil.WriteFile(filepath.Join(s.Path(p), cgroupProcs), []byte(strconv.Itoa(process.Pid)), os.FileMode(0666)); err != nil {
			return err
		}
	}

	return nil
}

func (c *cgroup) AddTask(process Process) error {
	panic("implement me")
}

// Delete will remove the control group from each of the subsystems registered
func (c *cgroup) Delete() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.err != nil {
		return c.err
	}

	var errors []string
	for _, s := range c.subsystems {
		if d, ok := s.(deleter); ok { // 主要是为了扩展
			sp, err := c.path(s.Name())
			if err != nil {
				return err
			}
			if err := d.Delete(sp); err != nil {
				errors = append(errors, string(s.Name()))
			}
		} else if p, ok := s.(pather); ok { // 否则使用默认的 os.RemoveAll(path)
			sp, err := c.path(s.Name())
			if err != nil {
				return err
			}
			path := p.Path(sp)
			if err := remove(path); err != nil {
				errors = append(errors, path)
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("cgroups: unable to remove paths %s", strings.Join(errors, ", "))
	}
	c.err = ErrCgroupDeleted

	return nil
}

func (c *cgroup) MoveTo(c2 Cgroup) error {
	panic("implement me")
}

func (c *cgroup) Stat(handler ...ErrorHandler) (*v1.Metrics, error) {
	panic("implement me")
}

// Update updates the cgroup with the new resource values provided
// Be prepared to handle EBUSY when trying to update a cgroup with
// live processes and other operations like Stats being performed at the
// same time
func (c *cgroup) Update(resources *specs.LinuxResources) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.err != nil {
		return c.err
	}

	// INFO: 设计这些接口 pather/updater/creater/deleter 主要是为了面向接口编程，也方便扩展。cpuset controller 得实现 Update() 函数
	for _, s := range c.subsystems {
		if u, ok := s.(updater); ok {
			sp, err := c.path(s.Name())
			if err != nil {
				return err
			}
			if err := u.Update(sp, resources); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *cgroup) Processes(subsystem Name, recursive bool) ([]Process, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.err != nil {
		return nil, c.err
	}

	return c.processes(subsystem, recursive)
}

func (c *cgroup) processes(subsystem Name, recursive bool) ([]Process, error) {
	s := c.getSubsystem(subsystem)
	relativePath, err := c.path(subsystem)
	if err != nil {
		return nil, err
	}
	absPath := s.(pather).Path(relativePath)

	// TODO: 这里递归读取 cgroup.procs 文件并返回所有 process 逻辑暂时不写了
	klog.Infof("cgroup path: %s", absPath)

	return nil, err
}

func (c *cgroup) getSubsystem(n Name) Subsystem {
	for _, s := range c.subsystems {
		if s.Name() == n {
			return s
		}
	}
	return nil
}

func (c *cgroup) Tasks(name Name, b bool) ([]Task, error) {
	panic("implement me")
}

func (c *cgroup) Freeze() error {
	panic("implement me")
}

func (c *cgroup) Thaw() error {
	panic("implement me")
}

func (c *cgroup) OOMEventFD() (uintptr, error) {
	panic("implement me")
}

func (c *cgroup) State() State {
	return Deleted
}

func (c *cgroup) Subsystems() []Subsystem {
	return c.subsystems
}

func initializeSubsystem(s Subsystem, path Path, resources *specs.LinuxResources) error {
	if c, ok := s.(creator); ok {
		p, err := path(s.Name())
		if err != nil {
			return err
		}
		if err := c.Create(p, resources); err != nil {
			return err
		}
	} else if c, ok := s.(pather); ok { // 否则直接 mkdir 创建
		p, err := path(s.Name())
		if err != nil {
			return err
		}
		// do the default create if the group does not have a custom one
		if err := os.MkdirAll(c.Path(p), defaultDirPerm); err != nil {
			return err
		}
	}

	return nil
}

// New returns a new control via the cgroup cgroups interface
func New(hierarchy Hierarchy, path Path, resources *specs.LinuxResources) (Cgroup, error) {
	subsystems, err := hierarchy()
	if err != nil {
		return nil, err
	}

	var active []Subsystem
	for _, s := range subsystems {
		// check if subsystem exists
		if err := initializeSubsystem(s, path, resources); err != nil {
			return nil, err
		}
		active = append(active, s)
	}

	return &cgroup{
		path:       path,
		subsystems: active,
	}, nil
}

// Load will load an existing cgroup and allow it to be controlled
func Load(hierarchy Hierarchy, path Path) (Cgroup, error) {
	var active []Subsystem

	subsystems, err := hierarchy()
	if err != nil {
		return nil, err
	}

	// INFO: 这块 check 逻辑与 V1() 里的逻辑有重复
	for _, subsystem := range pathers(subsystems) {
		p, err := path(subsystem.Name())
		if err != nil {
			return nil, err
		}
		if _, err := os.Lstat(subsystem.Path(p)); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		active = append(active, subsystem)
	}

	return &cgroup{
		path:       path,
		subsystems: active,
	}, nil
}
