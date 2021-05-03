package manager

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/cache/memory"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container/docker"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/container/raw"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/events"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/fs"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v1"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/info/v2"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/machine"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/utils/sysfs"
	"k8s-lx1036/k8s/kubelet/pkg/cadvisor/pkg/watcher"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
)

var updateMachineInfoInterval = flag.Duration("update_machine_info_interval", 5*time.Minute, "Interval between machine info updates.")
var globalHousekeepingInterval = flag.Duration("global_housekeeping_interval", 1*time.Minute, "Interval between global housekeepings")
var logCadvisorUsage = flag.Bool("log_cadvisor_usage", false, "Whether to log the usage of the cAdvisor container")
var eventStorageAgeLimit = flag.String("event_storage_age_limit", "default=24h", "Max length of time for which to store events (per type). Value is a comma separated list of key values, where the keys are event types (e.g.: creation, oom) or \"default\" and the value is a duration. Default is applied to all non-specified event types")
var eventStorageEventLimit = flag.String("event_storage_event_limit", "default=100000", "Max number of events to store (per type). Value is a comma separated list of key values, where the keys are event types (e.g.: creation, oom) or \"default\" and the value is an integer. Default is applied to all non-specified event types")

// The Manager interface defines operations for starting a manager and getting
// container and machine information.
type Manager interface {
	// Start the manager. Calling other manager methods before this returns
	// may produce undefined behavior.
	Start() error

	// Stops the manager.
	Stop() error

	//  information about a container.
	GetContainerInfo(containerName string, query *v1.ContainerInfoRequest) (*v1.ContainerInfo, error)

	// Get V2 information about a container.
	// Recursive (subcontainer) requests are best-effort, and may return a partial result alongside an
	// error in the partial failure case.
	GetContainerInfoV2(containerName string, options v2.RequestOptions) (map[string]v2.ContainerInfo, error)

	// Get information about all subcontainers of the specified container (includes self).
	SubcontainersInfo(containerName string, query *v1.ContainerInfoRequest) ([]*v1.ContainerInfo, error)

	// Gets all the Docker containers. Return is a map from full container name to ContainerInfo.
	AllDockerContainers(query *v1.ContainerInfoRequest) (map[string]v1.ContainerInfo, error)

	// Gets information about a specific Docker container. The specified name is within the Docker namespace.
	DockerContainer(dockerName string, query *v1.ContainerInfoRequest) (v1.ContainerInfo, error)

	// Gets spec for all containers based on request options.
	GetContainerSpec(containerName string, options v2.RequestOptions) (map[string]v2.ContainerSpec, error)

	// Gets summary stats for all containers based on request options.
	//GetDerivedStats(containerName string, options v2.RequestOptions) (map[string]v2.DerivedStats, error)

	// Get info for all requested containers based on the request options.
	GetRequestedContainersInfo(containerName string, options v2.RequestOptions) (map[string]*v1.ContainerInfo, error)

	// Returns true if the named container exists.
	Exists(containerName string) bool

	// Get information about the machine.
	GetMachineInfo() (*v1.MachineInfo, error)

	// Get version information about different components we depend on.
	GetVersionInfo() (*v1.VersionInfo, error)

	// GetFsInfoByFsUUID returns the information of the device having the
	// specified filesystem uuid. If no such device with the UUID exists, this
	// function will return the fs.ErrNoSuchDevice error.
	GetFsInfoByFsUUID(uuid string) (v2.FsInfo, error)

	// Get filesystem information for the filesystem that contains the given directory
	GetDirFsInfo(dir string) (v2.FsInfo, error)

	// Get filesystem information for a given label.
	// Returns information for all global filesystems if label is empty.
	GetFsInfo(label string) ([]v2.FsInfo, error)

	// Get ps output for a container.
	GetProcessList(containerName string, options v2.RequestOptions) ([]v2.ProcessInfo, error)

	// Get events streamed through passedChannel that fit the request.
	WatchForEvents(request *events.Request) (*events.EventChannel, error)

	// Get past events that have been detected and that fit the request.
	GetPastEvents(request *events.Request) ([]*v1.Event, error)

	CloseEventChannel(watchID int)

	// Get status information about docker.
	DockerInfo() (v1.DockerStatus, error)

	// Get details about interesting docker images.
	DockerImages() ([]v1.DockerImage, error)

	// Returns debugging information. Map of lines per category.
	DebugInfo() map[string][]string
}

// Housekeeping configuration for the manager
type HouskeepingConfig = struct {
	Interval     *time.Duration
	AllowDynamic *bool
}

// A namespaced container name.
type namespacedContainerName struct {
	// The namespace of the container. Can be empty for the root namespace.
	Namespace string

	// The name of the container in this namespace.
	Name string
}

type manager struct {
	containers               map[namespacedContainerName]*containerData
	containersLock           sync.RWMutex
	memoryCache              *memory.InMemoryCache
	fsInfo                   fs.FsInfo
	sysFs                    sysfs.SysFs
	machineMu                sync.RWMutex // protects machineInfo
	machineInfo              v1.MachineInfo
	quitChannels             []chan error
	cadvisorContainer        string
	inHostNamespace          bool
	eventHandler             events.EventManager
	startupTime              time.Time
	maxHousekeepingInterval  time.Duration
	allowDynamicHousekeeping bool
	includedMetrics          container.MetricSet
	containerWatchers        []watcher.ContainerWatcher
	eventsChannel            chan watcher.ContainerEvent
	collectorHTTPClient      *http.Client
	//nvidiaManager            stats.Manager
	//perfManager              stats.Manager
	//resctrlManager           stats.Manager
	// List of raw container cgroup path prefix whitelist.
	rawContainerCgroupPathPrefixWhiteList []string
}

func (m *manager) updateMachineInfo(quit chan error) {
	ticker := time.NewTicker(*updateMachineInfoInterval)
	for {
		select {
		case <-ticker.C:
			info, err := machine.Info(m.sysFs, m.fsInfo, m.inHostNamespace)
			if err != nil {
				klog.Errorf("Could not get machine info: %v", err)
				break
			}
			m.machineMu.Lock()
			m.machineInfo = *info
			m.machineMu.Unlock()
			klog.V(5).Infof("Update machine info: %+v", *info)
		case <-quit:
			ticker.Stop()
			quit <- nil
			return
		}
	}
}

func (m *manager) globalHousekeeping(quit chan error) {
	// 30s
	longHousekeeping := 100 * time.Millisecond
	if *globalHousekeepingInterval/2 < longHousekeeping {
		longHousekeeping = *globalHousekeepingInterval / 2
	}

	// 60s
	ticker := time.NewTicker(*globalHousekeepingInterval)
	for {
		select {
		case t := <-ticker.C:
			start := time.Now()

			// Check for new containers.
			err := m.detectSubcontainers("/")
			if err != nil {
				klog.Errorf("Failed to detect containers: %s", err)
			}

			// Log if housekeeping took too long.
			duration := time.Since(start)
			if duration >= longHousekeeping {
				klog.V(3).Infof("Global Housekeeping(%d) took %s", t.Unix(), duration)
			}
		case <-quit:
			// Quit if asked to do so.
			quit <- nil
			klog.Infof("Exiting global housekeeping thread")
			return
		}
	}
}

// Detect all containers that have been added or deleted from the specified container.
func (m *manager) getContainersDiff(containerName string) (added []v1.ContainerReference, removed []v1.ContainerReference, err error) {
	// Get all subcontainers recursively.
	m.containersLock.RLock()
	cont, ok := m.containers[namespacedContainerName{
		Name: containerName,
	}]
	m.containersLock.RUnlock()
	if !ok {
		return nil, nil, fmt.Errorf("failed to find container %q while checking for new containers", containerName)
	}
	allContainers, err := cont.handler.ListContainers(container.ListRecursive)

	if err != nil {
		return nil, nil, err
	}
	allContainers = append(allContainers, v1.ContainerReference{Name: containerName})

	m.containersLock.RLock()
	defer m.containersLock.RUnlock()

	// Determine which were added and which were removed.
	allContainersSet := make(map[string]*containerData)
	for name, d := range m.containers {
		// Only add the canonical name.
		if d.info.Name == name.Name {
			allContainersSet[name.Name] = d
		}
	}

	// Added containers
	for _, c := range allContainers {
		delete(allContainersSet, c.Name)
		_, ok := m.containers[namespacedContainerName{
			Name: c.Name,
		}]
		if !ok {
			added = append(added, c)
		}
	}

	// Removed ones are no longer in the container listing.
	for _, d := range allContainersSet {
		removed = append(removed, d.info.ContainerReference)
	}

	return
}

// Detect the existing subcontainers and reflect the setup here.
func (m *manager) detectSubcontainers(containerName string) error {
	added, removed, err := m.getContainersDiff(containerName)
	if err != nil {
		return err
	}

	// Add the new containers.
	for _, cont := range added {
		err = m.createContainer(cont.Name, watcher.Raw)
		if err != nil {
			klog.Errorf("Failed to create existing container: %s: %s", cont.Name, err)
		}
	}

	// Remove the old containers.
	for _, cont := range removed {
		err = m.destroyContainer(cont.Name)
		if err != nil {
			klog.Errorf("Failed to destroy existing container: %s: %s", cont.Name, err)
		}
	}

	return nil
}

// Create a container.
func (m *manager) createContainer(containerName string, watchSource watcher.ContainerWatchSource) error {
	m.containersLock.Lock()
	defer m.containersLock.Unlock()

	return m.createContainerLocked(containerName, watchSource)
}

func (m *manager) createContainerLocked(containerName string, watchSource watcher.ContainerWatchSource) error {
	namespacedName := namespacedContainerName{
		Name: containerName,
	}

	// Check that the container didn't already exist.
	if _, ok := m.containers[namespacedName]; ok {
		return nil
	}

	handler, accept, err := container.NewContainerHandler(containerName, watchSource, m.inHostNamespace)
	if err != nil {
		return err
	}
	if !accept {
		// ignoring this container.
		klog.V(4).Infof("ignoring container %q", containerName)
		return nil
	}
	/*collectorManager, err := collector.NewCollectorManager()
	if err != nil {
		return err
	}*/

	logUsage := *logCadvisorUsage && containerName == m.cadvisorContainer
	cont, err := newContainerData(containerName, m.memoryCache, handler, logUsage,
		m.maxHousekeepingInterval, m.allowDynamicHousekeeping, clock.RealClock{})
	if err != nil {
		return err
	}

	// Add collectors
	/*labels := handler.GetContainerLabels()
	collectorConfigs := collector.GetCollectorConfigs(labels)
	err = m.registerCollectors(collectorConfigs, cont)
	if err != nil {
		klog.Warningf("Failed to register collectors for %q: %v", containerName, err)
	}*/

	// Add the container name and all its aliases. The aliases must be within the namespace of the factory.
	m.containers[namespacedName] = cont
	for _, alias := range cont.info.Aliases {
		m.containers[namespacedContainerName{
			Namespace: cont.info.Namespace,
			Name:      alias,
		}] = cont
	}

	klog.V(3).Infof("Added container: %q (aliases: %v, namespace: %q)", containerName, cont.info.Aliases, cont.info.Namespace)

	contSpec, err := cont.handler.GetSpec()
	if err != nil {
		return err
	}

	contRef, err := cont.handler.ContainerReference()
	if err != nil {
		return err
	}

	newEvent := &v1.Event{
		ContainerName: contRef.Name,
		Timestamp:     contSpec.CreationTime,
		EventType:     v1.EventContainerCreation,
	}
	err = m.eventHandler.AddEvent(newEvent)
	if err != nil {
		return err
	}

	// Start the container's housekeeping.
	return cont.Start()
}

func (m *manager) destroyContainer(containerName string) error {
	m.containersLock.Lock()
	defer m.containersLock.Unlock()

	return m.destroyContainerLocked(containerName)
}

func (m *manager) destroyContainerLocked(containerName string) error {
	namespacedName := namespacedContainerName{
		Name: containerName,
	}
	cont, ok := m.containers[namespacedName]
	if !ok {
		// Already destroyed, done.
		return nil
	}

	// Tell the container to stop.
	err := cont.Stop()
	if err != nil {
		return err
	}

	// Remove the container from our records (and all its aliases).
	delete(m.containers, namespacedName)
	for _, alias := range cont.info.Aliases {
		delete(m.containers, namespacedContainerName{
			Namespace: cont.info.Namespace,
			Name:      alias,
		})
	}
	klog.V(3).Infof("Destroyed container: %q (aliases: %v, namespace: %q)", containerName, cont.info.Aliases, cont.info.Namespace)

	contRef, err := cont.handler.ContainerReference()
	if err != nil {
		return err
	}

	newEvent := &v1.Event{
		ContainerName: contRef.Name,
		Timestamp:     time.Now(),
		EventType:     v1.EventContainerDeletion,
	}
	err = m.eventHandler.AddEvent(newEvent)
	if err != nil {
		return err
	}
	return nil
}

// Watches for new containers started in the system. Runs forever unless there is a setup error.
func (m *manager) watchForNewContainers(quit chan error) error {
	for _, containerWatcher := range m.containerWatchers {
		err := containerWatcher.Start(m.eventsChannel)
		if err != nil {
			return err
		}
	}

	// There is a race between starting the watch and new container creation so we do a detection before we read new containers.
	err := m.detectSubcontainers("/")
	if err != nil {
		return err
	}

	// Listen to events from the container handler.
	go func() {
		for {
			select {
			case event := <-m.eventsChannel:
				switch {
				case event.EventType == watcher.ContainerAdd:
					switch event.WatchSource {
					default:
						err = m.createContainer(event.Name, event.WatchSource)
					}
				case event.EventType == watcher.ContainerDelete:
					err = m.destroyContainer(event.Name)
				}
				if err != nil {
					klog.Warningf("Failed to process watch event %+v: %v", event, err)
				}
			case <-quit:
				var msg []string
				// Stop processing events if asked to quit.
				for i, containerWatcher := range m.containerWatchers {
					err = containerWatcher.Stop()
					if err != nil {
						err = fmt.Errorf("watcher %d err: %v", i, err)
						msg = append(msg, fmt.Sprintf("watcher %d err: %v", i, err))
					}
				}

				if len(msg) != 0 {
					quit <- fmt.Errorf("%s", strings.Join(msg, ";"))
				} else {
					quit <- nil
					klog.Infof("Exiting thread watching subcontainers")
					return
				}
			}
		}
	}()

	return nil
}

func (m *manager) Stop() error {
	panic("implement me")
}

func (m *manager) getContainerData(containerName string) (*containerData, error) {
	m.containersLock.RLock()
	defer m.containersLock.RUnlock()

	cont, ok := m.containers[namespacedContainerName{Name: containerName}]
	if !ok {
		return nil, fmt.Errorf("unknown container %q", containerName)
	}

	return cont, nil
}

func (m *manager) GetContainerInfo(containerName string, query *v1.ContainerInfoRequest) (*v1.ContainerInfo, error) {
	cont, err := m.getContainerData(containerName)
	if err != nil {
		return nil, err
	}

	return m.containerDataToContainerInfo(cont, query)
}

// INFO: 这个函数很重要，"/" 可以获取所有容器的 stats 数据
func (m *manager) GetContainerInfoV2(containerName string, options v2.RequestOptions) (map[string]v2.ContainerInfo, error) {
	// 先根据 containerName="/" 获取所有 containers
	containers, err := m.getRequestedContainers(containerName, options)
	if err != nil {
		return nil, err
	}

	var errs []error
	var nilTime time.Time // Ignored.

	infos := make(map[string]v2.ContainerInfo, len(containers))
	for name, cData := range containers {
		result := v2.ContainerInfo{}
		cinfo, err := cData.GetInfo(false)
		if err != nil {
			errs = append(errs, fmt.Errorf("[GetInfo for %s with err: %v]", name, err))
			infos[name] = result
			continue
		}
		result.Spec = m.getV2Spec(cinfo)

		// INFO: container stats 是cadvisor周期定时读取 cgroup，然后存入 memoryCache 中
		stats, err := m.memoryCache.RecentStats(name, nilTime, nilTime, options.Count)
		if err != nil {
			errs = append(errs, fmt.Errorf("[RecentStats for %s with err: %v]", name, err))
			infos[name] = result
			continue
		}

		result.Stats = v2.ContainerStatsFromV1(containerName, &cinfo.Spec, stats)
		infos[name] = result
	}

	return infos, utilerrors.NewAggregate(errs)
}

// Get V2 container spec from v1 container info.
func (m *manager) getV2Spec(cinfo *containerInfo) v2.ContainerSpec {
	spec := m.getAdjustedSpec(cinfo)
	return v2.ContainerSpecFromV1(&spec, cinfo.Aliases, cinfo.Namespace)
}

func (m *manager) getRequestedContainers(containerName string, options v2.RequestOptions) (map[string]*containerData, error) {
	containersMap := make(map[string]*containerData)
	switch options.IdType {
	case v2.TypeName:
		if !options.Recursive {
			cont, err := m.getContainer(containerName)
			if err != nil {
				return containersMap, err
			}
			containersMap[cont.info.Name] = cont
		} else {
			containersMap = m.getSubcontainers(containerName)
			if len(containersMap) == 0 {
				return containersMap, fmt.Errorf("unknown container: %q", containerName)
			}
		}
	case v2.TypeDocker:
		if !options.Recursive {
			containerName = strings.TrimPrefix(containerName, "/")
			cont, err := m.getDockerContainer(containerName)
			if err != nil {
				return containersMap, err
			}
			containersMap[cont.info.Name] = cont
		} else {
			if containerName != "/" {
				return containersMap, fmt.Errorf("invalid request for docker container %q with subcontainers", containerName)
			}
			containersMap = m.getAllDockerContainers()
		}
	default:
		return containersMap, fmt.Errorf("invalid request type %q", options.IdType)
	}

	if options.MaxAge != nil {
		// update stats for all containers in containersMap
		var waitGroup sync.WaitGroup
		waitGroup.Add(len(containersMap))
		for _, cont := range containersMap {
			go func(cont *containerData) {
				cont.OnDemandHousekeeping(*options.MaxAge)
				waitGroup.Done()
			}(cont)
		}
		waitGroup.Wait()
	}

	return containersMap, nil
}
func (m *manager) getDockerContainer(containerName string) (*containerData, error) {
	m.containersLock.RLock()
	defer m.containersLock.RUnlock()

	// Check for the container in the Docker container namespace.
	cont, ok := m.containers[namespacedContainerName{
		Namespace: docker.DockerNamespace,
		Name:      containerName,
	}]

	// Look for container by short prefix name if no exact match found.
	if !ok {
		for contName, c := range m.containers {
			if contName.Namespace == docker.DockerNamespace && strings.HasPrefix(contName.Name, containerName) {
				if cont == nil {
					cont = c
				} else {
					return nil, fmt.Errorf("unable to find container. Container %q is not unique", containerName)
				}
			}
		}

		if cont == nil {
			return nil, fmt.Errorf("unable to find Docker container %q", containerName)
		}
	}

	return cont, nil
}
func (m *manager) getContainer(containerName string) (*containerData, error) {
	m.containersLock.RLock()
	defer m.containersLock.RUnlock()
	cont, ok := m.containers[namespacedContainerName{Name: containerName}]
	if !ok {
		return nil, fmt.Errorf("unknown container %q", containerName)
	}
	return cont, nil
}
func (m *manager) getAllDockerContainers() map[string]*containerData {
	m.containersLock.RLock()
	defer m.containersLock.RUnlock()
	containers := make(map[string]*containerData, len(m.containers))

	// Get containers in the Docker namespace.
	for name, cont := range m.containers {
		if name.Namespace == docker.DockerNamespace {
			containers[cont.info.Name] = cont
		}
	}
	return containers
}
func (m *manager) getSubcontainers(containerName string) map[string]*containerData {
	m.containersLock.RLock()
	defer m.containersLock.RUnlock()
	containersMap := make(map[string]*containerData, len(m.containers))

	// Get all the unique subcontainers of the specified container
	matchedName := path.Join(containerName, "/")
	for i := range m.containers {
		if m.containers[i] == nil {
			continue
		}
		name := m.containers[i].info.Name
		if name == containerName || strings.HasPrefix(name, matchedName) {
			containersMap[m.containers[i].info.Name] = m.containers[i]
		}
	}
	return containersMap
}

func (m *manager) containerDataSliceToContainerInfoSlice(containers []*containerData,
	query *v1.ContainerInfoRequest) ([]*v1.ContainerInfo, error) {
	if len(containers) == 0 {
		return nil, fmt.Errorf("no containers found")
	}

	// Get the info for each container.
	output := make([]*v1.ContainerInfo, 0, len(containers))
	for i := range containers {
		cinfo, err := m.containerDataToContainerInfo(containers[i], query)
		if err != nil {
			// Skip containers with errors, we try to degrade gracefully.
			klog.V(4).Infof("convert container data to container info failed with error %s", err.Error())
			continue
		}
		output = append(output, cinfo)
	}

	return output, nil
}

func (m *manager) getAdjustedSpec(cinfo *containerInfo) v1.ContainerSpec {
	spec := cinfo.Spec

	// Set default value to an actual value
	if spec.HasMemory {
		// Memory.Limit is 0 means there's no limit
		if spec.Memory.Limit == 0 {
			m.machineMu.RLock()
			spec.Memory.Limit = uint64(m.machineInfo.MemoryCapacity)
			m.machineMu.RUnlock()
		}
	}

	return spec
}

func (m *manager) containerDataToContainerInfo(cont *containerData, query *v1.ContainerInfoRequest) (*v1.ContainerInfo, error) {
	// Get the info from the container.
	cinfo, err := cont.GetInfo(true)
	if err != nil {
		return nil, err
	}

	stats, err := m.memoryCache.RecentStats(cinfo.Name, query.Start, query.End, query.NumStats)
	if err != nil {
		return nil, err
	}

	// Make a copy of the info for the user.
	ret := &v1.ContainerInfo{
		ContainerReference: cinfo.ContainerReference,
		Subcontainers:      cinfo.Subcontainers,
		Spec:               m.getAdjustedSpec(cinfo),
		Stats:              stats,
	}

	return ret, nil
}

func (m *manager) SubcontainersInfo(containerName string, query *v1.ContainerInfoRequest) ([]*v1.ContainerInfo, error) {
	containersMap := m.getSubcontainers(containerName)

	containers := make([]*containerData, 0, len(containersMap))
	for _, cont := range containersMap {
		containers = append(containers, cont)
	}

	return m.containerDataSliceToContainerInfoSlice(containers, query)
}

func (m *manager) AllDockerContainers(query *v1.ContainerInfoRequest) (map[string]v1.ContainerInfo, error) {
	panic("implement me")
}

func (m *manager) DockerContainer(dockerName string, query *v1.ContainerInfoRequest) (v1.ContainerInfo, error) {
	panic("implement me")
}

func (m *manager) GetContainerSpec(containerName string, options v2.RequestOptions) (map[string]v2.ContainerSpec, error) {
	panic("implement me")
}

func (m *manager) GetRequestedContainersInfo(containerName string, options v2.RequestOptions) (map[string]*v1.ContainerInfo, error) {
	panic("implement me")
}

func (m *manager) Exists(containerName string) bool {
	panic("implement me")
}

func (m *manager) GetMachineInfo() (*v1.MachineInfo, error) {
	panic("implement me")
}

func (m *manager) GetVersionInfo() (*v1.VersionInfo, error) {
	panic("implement me")
}

func (m *manager) GetFsInfoByFsUUID(uuid string) (v2.FsInfo, error) {
	panic("implement me")
}

func (m *manager) GetDirFsInfo(dir string) (v2.FsInfo, error) {
	panic("implement me")
}

func (m *manager) GetFsInfo(label string) ([]v2.FsInfo, error) {
	panic("implement me")
}

func (m *manager) GetProcessList(containerName string, options v2.RequestOptions) ([]v2.ProcessInfo, error) {
	panic("implement me")
}

func (m *manager) WatchForEvents(request *events.Request) (*events.EventChannel, error) {
	panic("implement me")
}

func (m *manager) GetPastEvents(request *events.Request) ([]*v1.Event, error) {
	return m.eventHandler.GetEvents(request)
}

func (m *manager) CloseEventChannel(watchID int) {
	panic("implement me")
}

func (m *manager) DockerInfo() (v1.DockerStatus, error) {
	panic("implement me")
}

func (m *manager) DockerImages() ([]v1.DockerImage, error) {
	panic("implement me")
}

func (m *manager) DebugInfo() map[string][]string {
	panic("implement me")
}

// Parses the events StoragePolicy from the flags.
func parseEventsStoragePolicy() events.StoragePolicy {
	policy := events.DefaultStoragePolicy()

	// Parse max age.
	parts := strings.Split(*eventStorageAgeLimit, ",")
	for _, part := range parts {
		items := strings.Split(part, "=")
		if len(items) != 2 {
			klog.Warningf("Unknown event storage policy %q when parsing max age", part)
			continue
		}
		dur, err := time.ParseDuration(items[1])
		if err != nil {
			klog.Warningf("Unable to parse event max age duration %q: %v", items[1], err)
			continue
		}
		if items[0] == "default" {
			policy.DefaultMaxAge = dur
			continue
		}
		policy.PerTypeMaxAge[v1.EventType(items[0])] = dur
	}

	// Parse max number.
	parts = strings.Split(*eventStorageEventLimit, ",")
	for _, part := range parts {
		items := strings.Split(part, "=")
		if len(items) != 2 {
			klog.Warningf("Unknown event storage policy %q when parsing max event limit", part)
			continue
		}
		val, err := strconv.Atoi(items[1])
		if err != nil {
			klog.Warningf("Unable to parse integer from %q: %v", items[1], err)
			continue
		}
		if items[0] == "default" {
			policy.DefaultMaxNumEvents = val
			continue
		}
		policy.PerTypeMaxNumEvents[v1.EventType(items[0])] = val
	}

	return policy
}

// Start the container manager.
func (m *manager) Start() error {
	// INFO: 这里初始化所有 plugins，这里是初始化 docker/plugin.go::Register()
	m.containerWatchers = container.InitializePlugins(m, m.fsInfo, m.includedMetrics)

	err := raw.Register(m, m.fsInfo, m.includedMetrics, m.rawContainerCgroupPathPrefixWhiteList)
	if err != nil {
		klog.Errorf("Registration of the raw container factory failed: %v", err)
	}
	rawWatcher, err := raw.NewRawContainerWatcher()
	if err != nil {
		return err
	}
	m.containerWatchers = append(m.containerWatchers, rawWatcher)

	// Watch for OOMs.
	/*err := m.watchForNewOoms()
	if err != nil {
		klog.Warningf("Could not configure a source for OOM detection, disabling OOM events: %v", err)
	}*/

	// If there are no factories, don't start any housekeeping and serve the information we do have.
	if !container.HasFactories() {
		return nil
	}

	// Create root and then recover all containers.
	err = m.createContainer("/", watcher.Raw)
	if err != nil {
		return err
	}
	klog.V(2).Infof("Starting recovery of all containers")
	err = m.detectSubcontainers("/")
	if err != nil {
		return err
	}
	klog.V(2).Infof("Recovery completed")

	// Watch for new container.
	quitWatcher := make(chan error)
	err = m.watchForNewContainers(quitWatcher)
	if err != nil {
		return err
	}
	m.quitChannels = append(m.quitChannels, quitWatcher)

	// Look for new containers in the main housekeeping thread.
	// INFO: 定时sync containers, 即 add/destroy containers
	quitGlobalHousekeeping := make(chan error)
	m.quitChannels = append(m.quitChannels, quitGlobalHousekeeping)
	go m.globalHousekeeping(quitGlobalHousekeeping)

	// INFO: 定时获取 machineInfo，默认 5 mins
	quitUpdateMachineInfo := make(chan error)
	m.quitChannels = append(m.quitChannels, quitUpdateMachineInfo)
	go m.updateMachineInfo(quitUpdateMachineInfo)

	return nil
}

// New takes a memory storage and returns a new manager.
func New(memoryCache *memory.InMemoryCache, sysfs sysfs.SysFs, houskeepingConfig HouskeepingConfig,
	includedMetricsSet container.MetricSet, collectorHTTPClient *http.Client,
	rawContainerCgroupPathPrefixWhiteList []string, perfEventsFile string) (Manager, error) {
	// Detect the container we are running on.
	selfContainer := "/"

	context := fs.Context{}
	fsInfo, err := fs.NewFsInfo(context)
	if err != nil {
		return nil, err
	}

	// If cAdvisor was started with host's rootfs mounted, assume that its running
	// in its own namespaces.
	inHostNamespace := false
	if _, err := os.Stat("/rootfs/proc"); os.IsNotExist(err) {
		inHostNamespace = true
	}

	// Register for new subcontainers.
	eventsChannel := make(chan watcher.ContainerEvent, 16)

	newManager := &manager{
		containers:               make(map[namespacedContainerName]*containerData),
		quitChannels:             make([]chan error, 0, 2),
		memoryCache:              memoryCache,
		fsInfo:                   fsInfo,
		sysFs:                    sysfs,
		cadvisorContainer:        selfContainer,
		inHostNamespace:          inHostNamespace,
		startupTime:              time.Now(),
		maxHousekeepingInterval:  *houskeepingConfig.Interval,
		allowDynamicHousekeeping: *houskeepingConfig.AllowDynamic,
		includedMetrics:          includedMetricsSet,
		containerWatchers:        []watcher.ContainerWatcher{},
		eventsChannel:            eventsChannel,
		collectorHTTPClient:      collectorHTTPClient,
		//nvidiaManager:                         accelerators.NewNvidiaManager(includedMetricsSet),
		rawContainerCgroupPathPrefixWhiteList: rawContainerCgroupPathPrefixWhiteList,
	}

	machineInfo, err := machine.Info(sysfs, fsInfo, inHostNamespace)
	if err != nil {
		return nil, err
	}
	newManager.machineInfo = *machineInfo
	klog.Infof("Machine: %+v", newManager.machineInfo)

	newManager.eventHandler = events.NewEventManager(parseEventsStoragePolicy())

	return newManager, nil
}
