package docker

import (
	"context"
	"flag"
	"fmt"
	"sync"
	"time"

	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container/libcontainer"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/fs"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/watcher"

	"k8s.io/klog/v2"
)

const dockerClientTimeout = 10 * time.Second

type plugin struct{}

// NewPlugin returns an implementation of container.Plugin suitable for passing to container.RegisterPlugin()
func NewPlugin() container.Plugin {
	return &plugin{}
}

// Register root container before running this function!
func (p *plugin) Register(factory v1.MachineInfoFactory, fsInfo fs.FsInfo, includedMetrics container.MetricSet) (watcher.ContainerWatcher, error) {
	client, err := Client()
	if err != nil {
		return nil, fmt.Errorf("unable to communicate with docker daemon: %v", err)
	}

	dockerInfo, err := ValidateInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to validate Docker info: %v", err)
	}
	// Version already validated above, assume no error here.
	dockerVersion, _ := parseVersion(dockerInfo.ServerVersion, versionRe, 3)
	dockerAPIVersion, _ := APIVersion()
	cgroupSubsystems, err := libcontainer.GetCgroupSubsystems(includedMetrics)
	if err != nil {
		return nil, fmt.Errorf("failed to get cgroup subsystems: %v", err)
	}

	/*var (
		thinPoolWatcher *devicemapper.ThinPoolWatcher
		thinPoolName    string
	)
	if storageDriver(dockerInfo.Driver) == devicemapperStorageDriver {
		thinPoolWatcher, err = startThinPoolWatcher(dockerInfo)
		if err != nil {
			klog.Errorf("devicemapper filesystem stats will not be reported: %v", err)
		}

		// Safe to ignore error - driver status should always be populated.
		status, _ := StatusFromDockerInfo(*dockerInfo)
		thinPoolName = status.DriverStatus[dockerutil.DriverStatusPoolName]
	}*/

	klog.V(1).Infof("Registering Docker factory")
	f := &dockerFactory{
		cgroupSubsystems:   cgroupSubsystems,
		client:             client,
		dockerVersion:      dockerVersion,
		dockerAPIVersion:   dockerAPIVersion,
		fsInfo:             fsInfo,
		machineInfoFactory: factory,
		storageDriver:      storageDriver(dockerInfo.Driver),
		storageDir:         RootDir(),
		includedMetrics:    includedMetrics,
		//thinPoolName:       thinPoolName,
		//thinPoolWatcher:    thinPoolWatcher,
		//zfsWatcher:         zfsWatcher,
	}

	container.RegisterContainerHandlerFactory(f, []watcher.ContainerWatchSource{watcher.Raw})

	return nil, nil
}

func (p *plugin) InitializeFSContext(context *fs.Context) error {
	SetTimeout(dockerClientTimeout)
	// Try to connect to docker indefinitely on startup.
	dockerStatus := retryDockerStatus()
	context.Docker = fs.DockerContext{
		Root:         RootDir(),
		Driver:       dockerStatus.Driver,
		DriverStatus: dockerStatus.DriverStatus,
	}
	return nil
}

func retryDockerStatus() v1.DockerStatus {
	startupTimeout := dockerClientTimeout
	maxTimeout := 4 * startupTimeout
	for {
		ctx, _ := context.WithTimeout(context.Background(), startupTimeout)
		dockerStatus, err := StatusWithContext(ctx)
		if err == nil {
			return dockerStatus
		}

		switch err {
		case context.DeadlineExceeded:
			klog.Warningf("Timeout trying to communicate with docker during initialization, will retry")
		default:
			klog.V(5).Infof("Docker not connected: %v", err)
			return v1.DockerStatus{}
		}

		startupTimeout = 2 * startupTimeout
		if startupTimeout > maxTimeout {
			startupTimeout = maxTimeout
		}
	}
}

var (
	// Basepath to all container specific information that libcontainer stores.
	dockerRootDir string

	dockerRootDirFlag = flag.String("docker_root", "/var/lib/docker", "DEPRECATED: docker root is read from docker info (this is a fallback, default: /var/lib/docker)")

	dockerRootDirOnce sync.Once

	// flag that controls globally disabling thin_ls pending future enhancements.
	// in production, it has been found that thin_ls makes excessive use of iops.
	// in an iops restricted environment, usage of thin_ls must be controlled via blkio.
	// pending that enhancement, disable its usage.
	disableThinLs = true
)

// The retry times for getting docker root dir
const rootDirRetries = 5

//The retry period for getting docker root dir, Millisecond
const rootDirRetryPeriod time.Duration = 1000 * time.Millisecond

func RootDir() string {
	dockerRootDirOnce.Do(func() {
		for i := 0; i < rootDirRetries; i++ {
			status, err := Status()
			if err == nil && status.RootDir != "" {
				dockerRootDir = status.RootDir
				break
			} else {
				time.Sleep(rootDirRetryPeriod)
			}
		}
		if dockerRootDir == "" {
			dockerRootDir = *dockerRootDirFlag
		}
	})
	return dockerRootDir
}
