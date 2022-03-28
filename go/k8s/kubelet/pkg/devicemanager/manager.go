package devicemanager

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"

	"k8s-lx1036/k8s/kubelet/pkg/cm/topologymanager"
	"k8s-lx1036/k8s/kubelet/pkg/devicemanager/checkpoint"
	"k8s-lx1036/k8s/kubelet/pkg/lifecycle"

	cadvisorapi "github.com/google/cadvisor/info/v1"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	errorsutil "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	podresourcesapi "k8s.io/kubelet/pkg/apis/podresources/v1alpha1"
	"k8s.io/kubernetes/pkg/kubelet/checkpointmanager"
	"k8s.io/kubernetes/pkg/kubelet/checkpointmanager/errors"
	"k8s.io/kubernetes/pkg/kubelet/config"
)

type endpointInfo struct {
	endpoint Endpoint
	opts     *pluginapi.DevicePluginOptions
}

// PodReusableDevices is a map by pod name of devices to reuse.
type PodReusableDevices map[string]map[string]sets.String

// monitorCallback is the function called when a device's health state changes,
// or new devices are reported, or old devices are deleted.
// Updated contains the most recent state of the Device.
type monitorCallback func(resourceName string, devices []pluginapi.Device)
type sourcesReadyStub struct{}

func (s *sourcesReadyStub) AddSource(source string) {}
func (s *sourcesReadyStub) AllReady() bool          { return true }

// ActivePodsFunc is a function that returns a list of pods to reconcile.
type ActivePodsFunc func() []*v1.Pod

// ManagerImpl is the structure in charge of managing Device Plugins.
type ManagerImpl struct {
	socketname string
	socketdir  string

	endpoints map[string]endpointInfo // Key is ResourceName
	mutex     sync.Mutex

	server *grpc.Server
	wg     sync.WaitGroup

	// allDevices is a map by resource name of all the devices currently registered to the device manager
	allDevices map[string]map[string]pluginapi.Device

	// healthyDevices contains all of the registered healthy resourceNames and their exported device IDs.
	healthyDevices map[string]sets.String

	// unhealthyDevices contains all of the unhealthy devices and their exported device IDs.
	unhealthyDevices map[string]sets.String

	// allocatedDevices contains allocated deviceIds, keyed by resourceName.
	allocatedDevices map[string]sets.String

	// podDevices contains pod to allocated device mapping.
	podDevices podDevices

	// List of NUMA Nodes available on the underlying machine
	numaNodes []int

	// Store of Topology Affinties that the Device Manager can query.
	topologyAffinityStore topologymanager.Store

	// devicesToReuse contains devices that can be reused as they have been allocated to
	// init containers.
	devicesToReuse PodReusableDevices

	// callback is used for updating devices' states in one time call.
	// e.g. a new device is advertised, two old devices are deleted and a running device fails.
	callback monitorCallback

	// activePods is a method for listing active pods on the node
	// so the amount of pluginResources requested by existing pods
	// could be counted when updating allocated devices
	activePods ActivePodsFunc
	// sourcesReady provides the readiness of kubelet configuration sources such as apiserver update readiness.
	// We use it to determine when we can purge inactive pods from checkpointed state.
	sourcesReady config.SourcesReady

	checkpointManager checkpointmanager.CheckpointManager
}

// topology: numa拓扑结构，统计节点的numa节点个数，为容器分配设备时提供Topology hints
func newManagerImpl(socketPath string, topology []cadvisorapi.Node,
	topologyAffinityStore topologymanager.Store) (*ManagerImpl, error) {

	klog.V(2).Infof("Creating Device Plugin manager at %s", socketPath)

	if socketPath == "" || !filepath.IsAbs(socketPath) {
		return nil, fmt.Errorf("bad socketPath, must be an absolute path: %s", socketPath)
	}

	var numaNodes []int
	for _, node := range topology {
		numaNodes = append(numaNodes, node.Id)
	}

	dir, file := filepath.Split(socketPath)
	manager := &ManagerImpl{
		endpoints:             make(map[string]endpointInfo),
		socketname:            file,
		socketdir:             dir,
		allDevices:            make(map[string]map[string]pluginapi.Device),
		healthyDevices:        make(map[string]sets.String),
		unhealthyDevices:      make(map[string]sets.String),
		allocatedDevices:      make(map[string]sets.String),
		podDevices:            make(podDevices),
		numaNodes:             numaNodes,
		topologyAffinityStore: topologyAffinityStore,
		devicesToReuse:        make(PodReusableDevices),
	}

	// INFO: ???
	manager.callback = manager.genericDeviceUpdateCallback

	// The following structures are populated with real implementations in manager.Start()
	// Before that, initializes them to perform no-op operations.
	manager.activePods = func() []*v1.Pod { return []*v1.Pod{} }
	manager.sourcesReady = &sourcesReadyStub{}
	checkpointManager, err := checkpointmanager.NewCheckpointManager(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize checkpoint manager: %v", err)
	}
	manager.checkpointManager = checkpointManager

	return manager, nil
}

func (m *ManagerImpl) genericDeviceUpdateCallback(resourceName string, devices []pluginapi.Device) {

}

// Reads device to container allocation information from disk, and populates
// m.allocatedDevices accordingly.
func (m *ManagerImpl) readCheckpoint() error {
	registeredDevs := make(map[string][]string)
	devEntries := make([]checkpoint.PodDevicesEntry, 0)
	deviceManagerCheckpoint := checkpoint.New(devEntries, registeredDevs)
	err := m.checkpointManager.GetCheckpoint(kubeletDeviceManagerCheckpoint, deviceManagerCheckpoint)
	if err != nil {
		if err == errors.ErrCheckpointNotFound {
			klog.Warningf("Failed to retrieve checkpoint for %q: %v", kubeletDeviceManagerCheckpoint, err)
			return nil
		}
		return err
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()
	podDeviceEntry, registeredDevs := deviceManagerCheckpoint.GetData()
	m.podDevices.fromCheckpointData(podDeviceEntry)
	m.allocatedDevices = m.podDevices.devices()

	for resource := range registeredDevs {
		// During start up, creates empty healthyDevices list so that the resource capacity
		// will stay zero till the corresponding device plugin re-registers.
		m.healthyDevices[resource] = sets.NewString()
		m.unhealthyDevices[resource] = sets.NewString()
		m.endpoints[resource] = endpointInfo{endpoint: newStoppedEndpointImpl(resource), opts: nil}
	}

	return nil
}

// Start starts the Device Plugin Manager and start initialization of
// podDevices and allocatedDevices information from checkpointed state and
// starts device plugin registration service.
// Device Manager启动的时候，首先会加载checkpoint文件中内容，然后启动一个RPC Server
func (m *ManagerImpl) Start(activePods ActivePodsFunc, sourcesReady config.SourcesReady) error {
	klog.V(2).Infof("Starting Device Plugin manager")

	m.activePods = activePods
	m.sourcesReady = sourcesReady

	// Loads in allocatedDevices information from disk.
	err := m.readCheckpoint()
	if err != nil {
		klog.Warningf("Continue after failing to read checkpoint file. Device allocation info may NOT be up-to-date. Err: %v", err)
	}

	socketPath := filepath.Join(m.socketdir, m.socketname)
	if err = os.MkdirAll(m.socketdir, 0750); err != nil {
		return err
	}

	// Removes all stale sockets in m.socketdir. Device plugins can monitor
	// this and use it as a signal to re-register with the new Kubelet.
	if err := m.removeContents(m.socketdir); err != nil {
		klog.Errorf("Fail to clean up stale contents under %s: %v", m.socketdir, err)
	}

	s, err := net.Listen("unix", socketPath)
	if err != nil {
		klog.Errorf("failed to listen to socket while starting device plugin registry, with error %v", err)
		return err
	}

	m.wg.Add(1)
	m.server = grpc.NewServer([]grpc.ServerOption{}...)

	pluginapi.RegisterRegistrationServer(m.server, m)
	go func() {
		defer m.wg.Done()
		m.server.Serve(s)
	}()

	klog.V(2).Infof("Serving device plugin registration server on %q", socketPath)

	return nil
}

func (m *ManagerImpl) removeContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	var errs []error
	for _, name := range names {
		filePath := filepath.Join(dir, name)
		if filePath == m.checkpointFile() {
			continue
		}
		stat, err := os.Stat(filePath)
		if err != nil {
			klog.Errorf("Failed to stat file %s: %v", filePath, err)
			continue
		}
		if stat.IsDir() {
			continue
		}
		err = os.RemoveAll(filePath)
		if err != nil {
			errs = append(errs, err)
			klog.Errorf("Failed to remove file %s: %v", filePath, err)
			continue
		}
	}
	return errorsutil.NewAggregate(errs)
}

// checkpointFile returns device plugin checkpoint file path.
func (m *ManagerImpl) checkpointFile() string {
	return filepath.Join(m.socketdir, kubeletDeviceManagerCheckpoint)
}

func (m *ManagerImpl) Allocate(pod *v1.Pod, container *v1.Container) error {
	panic("implement me")
}

func (m *ManagerImpl) UpdatePluginResources(node *schedulerframework.NodeInfo, attrs *lifecycle.PodAdmitAttributes) error {
	panic("implement me")
}

func (m *ManagerImpl) Stop() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for _, eInfo := range m.endpoints {
		eInfo.endpoint.stop()
	}

	if m.server == nil {
		return nil
	}
	m.server.Stop()
	m.wg.Wait()
	m.server = nil

	return nil
}

func (m *ManagerImpl) GetDeviceRunContainerOptions(pod *v1.Pod, container *v1.Container) (*DeviceRunContainerOptions, error) {
	panic("implement me")
}

func (m *ManagerImpl) GetCapacity() (v1.ResourceList, v1.ResourceList, []string) {
	panic("implement me")
}

func (m *ManagerImpl) GetWatcherHandler() cache.PluginHandler {
	panic("implement me")
}

func (m *ManagerImpl) GetDevices(podUID, containerName string) []*podresourcesapi.ContainerDevices {
	panic("implement me")
}

func (m *ManagerImpl) ShouldResetExtendedResourceCapacity() bool {
	panic("implement me")
}

func (m *ManagerImpl) GetTopologyHints(pod *v1.Pod, container *v1.Container) map[string][]topologymanager.TopologyHint {
	panic("implement me")
}

func (m *ManagerImpl) UpdateAllocatedDevices() {
	panic("implement me")
}

// INFO: 实现接口 k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1/api.pb.go::RegistrationServer
// 客户端代码在 device_plugin.go::Register(kubeletEndpoint, resourceName string, pluginSockDir string)
func (m *ManagerImpl) Register(ctx context.Context, request *pluginapi.RegisterRequest) (*pluginapi.Empty, error) {
	klog.Infof("Got registration request from device plugin with resource name %q", request.ResourceName)

	// TODO: for now, always accepts newest device plugin. Later may consider to
	// add some policies here, e.g., verify whether an old device plugin with the
	// same resource name is still alive to determine whether we want to accept
	// the new registration.
	go m.addEndpoint(request)

	return &pluginapi.Empty{}, nil
}

func (m *ManagerImpl) addEndpoint(r *pluginapi.RegisterRequest) {
	e, err := newEndpointImpl(filepath.Join(m.socketdir, r.Endpoint), r.ResourceName, m.callback)
	if err != nil {
		klog.Errorf("Failed to dial device plugin with request %v: %v", r, err)
		return
	}

	m.registerEndpoint(r.ResourceName, r.Options, e)
	go func() {
		m.runEndpoint(r.ResourceName, e)
	}()
}

func (m *ManagerImpl) runEndpoint(resourceName string, e Endpoint) {
	e.run()
	e.stop()

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if old, ok := m.endpoints[resourceName]; ok && old.endpoint == e {
		m.markResourceUnhealthy(resourceName)
	}

	klog.V(2).Infof("Endpoint (%s, %v) became unhealthy", resourceName, e)
}

func (m *ManagerImpl) markResourceUnhealthy(resourceName string) {
	klog.V(2).Infof("Mark all resources Unhealthy for resource %s", resourceName)
	healthyDevices := sets.NewString()
	if _, ok := m.healthyDevices[resourceName]; ok {
		healthyDevices = m.healthyDevices[resourceName]
		m.healthyDevices[resourceName] = sets.NewString()
	}
	if _, ok := m.unhealthyDevices[resourceName]; !ok {
		m.unhealthyDevices[resourceName] = sets.NewString()
	}
	m.unhealthyDevices[resourceName] = m.unhealthyDevices[resourceName].Union(healthyDevices)
}

func (m *ManagerImpl) registerEndpoint(resourceName string, options *pluginapi.DevicePluginOptions, e Endpoint) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.endpoints[resourceName] = endpointInfo{endpoint: e, opts: options}
	klog.V(2).Infof("Registered endpoint %v", e)
}
