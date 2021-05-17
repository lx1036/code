package dockershim

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	kubeletconfig "k8s-lx1036/k8s/kubelet/pkg/apis/config"
	"k8s-lx1036/k8s/kubelet/pkg/checkpointmanager"
	kubecontainer "k8s-lx1036/k8s/kubelet/pkg/container"
	"k8s-lx1036/k8s/kubelet/pkg/cri/streaming"
	"k8s-lx1036/k8s/kubelet/pkg/dockershim/cm"
	"k8s-lx1036/k8s/kubelet/pkg/dockershim/libdocker"
	"k8s-lx1036/k8s/kubelet/pkg/dockershim/network"
	"k8s-lx1036/k8s/kubelet/pkg/dockershim/network/cni"
	"k8s-lx1036/k8s/kubelet/pkg/dockershim/network/hostport"
	"k8s-lx1036/k8s/kubelet/pkg/util/cache"

	runtimeapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
)

const (
	// Internal docker labels used to identify whether a container is a sandbox
	// or a regular container.
	// TODO: This is not backward compatible with older containers. We will
	// need to add filtering based on names.
	containerTypeLabelKey       = "io.kubernetes.docker.type"
	containerTypeLabelSandbox   = "podsandbox"
	containerTypeLabelContainer = "container"
	containerLogPathLabelKey    = "io.kubernetes.container.logpath"
	sandboxIDLabelKey           = "io.kubernetes.sandbox.id"
)

// CRIService includes all methods necessary for a CRI server.
type CRIService interface {
	runtimeapi.RuntimeServiceServer
	runtimeapi.ImageServiceServer
	Start() error
}

// NetworkPluginSettings is the subset of kubelet runtime args we pass
// to the container runtime shim so it can probe for network plugins.
// In the future we will feed these directly to a standalone container
// runtime process.
type NetworkPluginSettings struct {
	// HairpinMode is best described by comments surrounding the kubelet arg
	HairpinMode kubeletconfig.HairpinMode
	// NonMasqueradeCIDR is the range of ips which should *not* be included
	// in any MASQUERADE rules applied by the plugin
	NonMasqueradeCIDR string
	// PluginName is the name of the plugin, runtime shim probes for
	PluginName string
	// PluginBinDirString is a list of directiores delimited by commas, in
	// which the binaries for the plugin with PluginName may be found.
	PluginBinDirString string
	// PluginBinDirs is an array of directories in which the binaries for
	// the plugin with PluginName may be found. The admin is responsible for
	// provisioning these binaries before-hand.
	PluginBinDirs []string
	// PluginConfDir is the directory in which the admin places a CNI conf.
	// Depending on the plugin, this may be an optional field, eg: kubenet
	// generates its own plugin conf.
	PluginConfDir string
	// PluginCacheDir is the directory in which CNI should store cache files.
	PluginCacheDir string
	// MTU is the desired MTU for network devices created by the plugin.
	MTU int
}

// DockerService is an interface that embeds the new RuntimeService and
// ImageService interfaces.
type DockerService interface {
	CRIService

	// For serving streaming calls.
	http.Handler
}

type dockerService struct {
	client          libdocker.Interface
	os              kubecontainer.OSInterface
	podSandboxImage string
	//streamingRuntime *streamingRuntime
	//streamingServer  streaming.Server

	streamingServer streaming.Server

	network *network.PluginManager
	// Map of podSandboxID :: network-is-ready
	networkReady     map[string]bool
	networkReadyLock sync.Mutex

	containerManager cm.ContainerManager
	// cgroup driver used by Docker runtime.
	cgroupDriver      string
	checkpointManager checkpointmanager.CheckpointManager
	// caches the version of the runtime.
	// To be compatible with multiple docker versions, we need to perform
	// version checking for some operations. Use this cache to avoid querying
	// the docker daemon every time we need to do such checks.
	versionCache *cache.ObjectCache

	// containerCleanupInfos maps container IDs to the `containerCleanupInfo` structs
	// needed to clean up after containers have been removed.
	// (see `applyPlatformSpecificDockerConfig` and `performPlatformSpecificContainerCleanup`
	// methods for more info).
	containerCleanupInfos map[string]*containerCleanupInfo
	cleanupInfosLock      sync.RWMutex
}

func (ds *dockerService) Version(ctx context.Context, request *runtimeapi.VersionRequest) (*runtimeapi.VersionResponse, error) {
	panic("implement me")
}

func (ds *dockerService) UpdateRuntimeConfig(ctx context.Context, request *runtimeapi.UpdateRuntimeConfigRequest) (*runtimeapi.UpdateRuntimeConfigResponse, error) {
	panic("implement me")
}

func (ds *dockerService) Status(ctx context.Context, request *runtimeapi.StatusRequest) (*runtimeapi.StatusResponse, error) {
	panic("implement me")
}

func (ds *dockerService) Start() error {
	return ds.containerManager.Start()
}

func (ds *dockerService) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if ds.streamingServer != nil {
		ds.streamingServer.ServeHTTP(writer, request)
	} else {
		http.NotFound(writer, request)
	}
}

// GetNetNS returns the network namespace of the given containerID. The ID
// supplied is typically the ID of a pod sandbox. This getter doesn't try
// to map non-sandbox IDs to their respective sandboxes.
func (ds *dockerService) GetNetNS(podSandboxID string) (string, error) {
	// `docker inspect ${container_id}` 获取 ".State.Pid"
	r, err := ds.client.InspectContainer(podSandboxID)
	if err != nil {
		return "", err
	}

	return getNetworkNamespace(r)
}

// GetPodPortMappings returns the port mappings of the given podSandbox ID.
func (ds *dockerService) GetPodPortMappings(podSandboxID string) ([]*hostport.PortMapping, error) {
	panic("implement me")
}

// NewDockerService creates a new `DockerService` struct.
// NOTE: Anything passed to DockerService should be eventually handled in another way when we switch to running the shim as a different process.
func NewDockerService(config *ClientConfig, podSandboxImage string, pluginSettings *NetworkPluginSettings,
	cgroupsName string, kubeCgroupDriver string, dockershimRootDir string) (DockerService, error) {
	// Create docker client.
	dockerClient := libdocker.ConnectToDockerOrDie(
		config.DockerEndpoint,
		config.RuntimeRequestTimeout,
		config.ImagePullProgressDeadline,
	)

	checkpointManager, err := checkpointmanager.NewCheckpointManager(filepath.Join(dockershimRootDir, sandboxCheckpointDir))
	if err != nil {
		return nil, err
	}
	ds := &dockerService{
		client:                dockerClient,
		os:                    kubecontainer.RealOS{},
		podSandboxImage:       podSandboxImage,
		containerManager:      cm.NewContainerManager(cgroupsName, dockerClient),
		checkpointManager:     checkpointManager,
		networkReady:          make(map[string]bool),
		containerCleanupInfos: make(map[string]*containerCleanupInfo),
	}

	cniPlugins := cni.ProbeNetworkPlugins(pluginSettings.PluginConfDir, pluginSettings.PluginCacheDir, pluginSettings.PluginBinDirs)
	netHost := &dockerNetworkHost{
		&namespaceGetter{ds},
		&portMappingGetter{ds},
	}
	plugin, err := network.InitNetworkPlugin(cniPlugins, pluginSettings.PluginName, netHost, pluginSettings.HairpinMode, pluginSettings.NonMasqueradeCIDR, pluginSettings.MTU)
	if err != nil {
		return nil, fmt.Errorf("didn't find compatible CNI plugin with given settings %+v: %v", pluginSettings, err)
	}
	// INFO: dockerService network 模块，很重要，在创建/查询 sandbox container network status 即 pod network 时需要
	ds.network = network.NewPluginManager(plugin)

	return ds, nil
}

// ClientConfig is parameters used to initialize docker client
type ClientConfig struct {
	DockerEndpoint            string
	RuntimeRequestTimeout     time.Duration
	ImagePullProgressDeadline time.Duration

	// Configuration for fake docker client
	EnableSleep       bool
	WithTraceDisabled bool
}

// INFO: 实现接口 go/k8s/kubelet/pkg/dockershim/network/plugins.go::Host
// dockerNetworkHost implements network.Host by wrapping the legacy host passed in by the kubelet
// and dockerServices which implements the rest of the network host interfaces.
// The legacy host methods are slated for deletion.
type dockerNetworkHost struct {
	*namespaceGetter
	*portMappingGetter
}

// namespaceGetter is a wrapper around the dockerService that implements
// the network.NamespaceGetter interface.
type namespaceGetter struct {
	ds *dockerService
}

func (n *namespaceGetter) GetNetNS(containerID string) (string, error) {
	return n.ds.GetNetNS(containerID)
}

// portMappingGetter is a wrapper around the dockerService that implements
// the network.PortMappingGetter interface.
type portMappingGetter struct {
	ds *dockerService
}

func (p *portMappingGetter) GetPodPortMappings(containerID string) ([]*hostport.PortMapping, error) {
	return p.ds.GetPodPortMappings(containerID)
}
