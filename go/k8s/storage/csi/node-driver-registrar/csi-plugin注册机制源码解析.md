

# Kubernetes学习笔记之CSI Plugin注册机制源码解析

## Overview
最近在维护组内K8s CSI plugin代码时，一直对其内部原理好奇，故趁机深入学习熟悉K8s CSI相关原理。
部署K8s持久化存储插件时，需要按照CSI官网说明，部署一个daemonset pod实现插件注册，该pod内容器包含 **[node-driver-registrar](https://kubernetes-csi.github.io/docs/node-driver-registrar.html)** ，部署yaml类似如下：

```yaml

apiVersion: apps/v1
kind: DaemonSet
metadata:
  annotations:
    deprecated.daemonset.template.generation: "7"
  name: sunnyfs-csi-share-node
  namespace: sunnyfs
spec:
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app: sunnyfs-csi-share-node
  template:
    metadata:
      labels:
        app: sunnyfs-csi-share-node
    spec:
      containers:
        - args:
            - --csi-address=/csi/sunnyfs-csi-share.sock
            - --kubelet-registration-path=/csi/sunnyfs-csi-share.sock
          env:
            - name: KUBE_NODE_NAME
              valueFrom:
                fieldRef:
                  apiVersion: v1
                  fieldPath: spec.nodeName
          image: quay.io/k8scsi/csi-node-driver-registrar:v2.1.0
          imagePullPolicy: IfNotPresent
          name: node-driver-registrar
          resources:
            limits:
              cpu: "2"
              memory: 4000Mi
            requests:
              cpu: "1"
              memory: 4000Mi
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
            - mountPath: /registration
              name: registration-dir
            - mountPath: /csi
              name: socket-dir
        - args:
            - --v=5
            - --endpoint=unix:///csi/sunnyfs-csi-share/sunnyfs-csi-share.sock
            - --nodeid=$(NODE_ID)
            - --drivername=csi.sunnyfs.share.com
            - --version=v1.0.0
          env:
            - name: NODE_ID
              valueFrom:
                fieldRef:
                  apiVersion: v1
                  fieldPath: spec.nodeName
          image: sunnyfs-csi-driver:v1.0.4
          imagePullPolicy: IfNotPresent
          lifecycle:
            preStop:
              exec:
                command:
                  - /bin/sh
                  - -c
                  - rm -rf /csi/sunnyfs-csi-share.sock /registration/csi.sunnyfs.share.com-reg.sock
          name: sunnyfs-csi-driver
          resources:
            limits:
              cpu: "2"
              memory: 4000Mi
            requests:
              cpu: "1"
              memory: 4000Mi
          securityContext:
            privileged: true
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
            - mountPath: /registration
              name: registration-dir
            - mountPath: /csi
              name: socket-dir
            - mountPath: /var/lib/kubelet/pods
              mountPropagation: Bidirectional
              name: mountpoint-dir
      dnsPolicy: ClusterFirstWithHostNet
      hostNetwork: true
      imagePullSecrets:
        - name: regcred
      restartPolicy: Always
      terminationGracePeriodSeconds: 30
      tolerations:
        - operator: Exists
      volumes:
        - hostPath:
            path: /var/lib/kubelet/plugins/csi.sunnyfs.share.com
            type: DirectoryOrCreate
          name: socket-dir
        - hostPath:
            path: /var/lib/kubelet/plugins_registry
            type: Directory
          name: registration-dir
        - hostPath:
            path: /var/lib/kubelet/pods
            type: Directory
          name: mountpoint-dir
  updateStrategy:
    rollingUpdate:
      maxUnavailable: 1
    type: RollingUpdate

```

pod内部署了自定义的csi-plugin如sunnyfs-csi-driver，该csi-plugin后端实际存储引擎是一个自研的分部署文件类型存储系统，和一个sidecar container **[node-driver-registrar](https://kubernetes-csi.github.io/docs/node-driver-registrar.html)** ，该pod主要实现了自定义的csi-plugin的注册。

重要问题是，是如何做到csi-plugin注册的？

答案很简单：daemonset中的 **[node-driver-registrar](https://github.com/kubernetes-csi/node-driver-registrar)** 作为一个sidecar container，会被kubelet plugin-mamanger模块调用，
而 node-driver-registrar sidecar container又会去调用我们自研的csi-plugin即sunnyfs-csi-driver container。而kubelet在启动时就会往plugin-mamanger模块
中注册一个csi plugin handler，该handler获取sunnyfs-csi-driver container基本信息后，会做一些操作，如更新node的annotation以及创建/更新CSINode对象。

## 源码解析

### node-driver-registrar 源码解析
node-driver-registrar sidecar container代码逻辑很简单，主要做了两件事：rpc调用自研的csi-plugin插件，调用了GetPluginInfo方法，获取response.GetName即csiDriverName；
启动一个grpc server，并监听在宿主机上/var/lib/kubelet/plugins_registry/${csiDriverName}-reg.sock，供csi plugin handler来调用。

大概看下代码做的这两件事。

首先rpc调用自研的csi-plugin插件获取csiDriverName，**[L137-L152](https://github.com/kubernetes-csi/node-driver-registrar/blob/master/cmd/csi-node-driver-registrar/main.go#L137-L152)** :

```go

func main() {
	// ...
	
	// 1. rpc调用自研的csi-plugin插件，调用了GetPluginInfo方法，获取response.GetName即csiDriverName
	csiConn, err := connection.Connect(*csiAddress, cmm)
	csiDriverName, err := csirpc.GetDriverName(ctx, csiConn)
	
	// Run forever
	nodeRegister(csiDriverName, addr)
}

```

**[GetDriverName](https://github.com/kubernetes-csi/csi-lib-utils/blob/master/rpc/common.go#L38-L52)** 代码如下，主要rpc调用自研csi-plugin中identity server中的GetPluginInfo方法：

```go
import (
    "github.com/container-storage-interface/spec/lib/go/csi"
)

// GetDriverName returns name of CSI driver.
func GetDriverName(ctx context.Context, conn *grpc.ClientConn) (string, error) {
	client := csi.NewIdentityClient(conn)
	req := csi.GetPluginInfoRequest{}
	rsp, err := client.GetPluginInfo(ctx, &req)
	// ...
	name := rsp.GetName()
	//...
	return name, nil
}

```

node-driver-registrar会先调用我们自研csi-plugin中identity server中的GetPluginInfo方法，而 **[CSI(Container Storage Interface)](https://github.com/container-storage-interface/spec/blob/master/spec.md#rpc-interface)** 设计文档
详细说明了，我们的csi-plugin中主要需要实现三个service: identity service, controller service和node service。其中，node service需要实现GetPluginInfo方法，返回我们自研plugin相关的基本信息，
比如我这里的identity service GetPluginInfo实现逻辑，主要返回我们自研csi plugin name：

```go

func (ids *DefaultIdentityServer) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	klog.Infof("Using default GetPluginInfo")

	if ids.Driver.name == "" {
		return nil, status.Error(codes.Unavailable, "Driver name not configured")
	}

	if ids.Driver.version == "" {
		return nil, status.Error(codes.Unavailable, "Driver is missing version")
	}

	return &csi.GetPluginInfoResponse{
		Name:          ids.Driver.name,
		VendorVersion: ids.Driver.version,
	}, nil
}

```


然后，node-driver-registrar sidecar container就会启动一个grpc server，并监听在宿主机上/var/lib/kubelet/plugins_registry/${csiDriverName}-reg.sock 。
该rpc server遵循 **[kubelet plugin registration标准](https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/kubelet/pkg/apis/pluginregistration/v1/api.proto)** ，*registrationServer service提供GetInfo和NotifyRegistrationStatus方法供客户端调用，
其实也就是被kubelet plugin manager模块调用，代码逻辑如下：

```go
// 启动一个grpc server并监听在socket /var/lib/kubelet/plugins_registry/${csiDriverName}-reg.sock
func nodeRegister(csiDriverName, httpEndpoint string) {
	registrar := newRegistrationServer(csiDriverName, *kubeletRegistrationPath, supportedVersions)
	socketPath := buildSocketPath(csiDriverName)
	// ...
	lis, err := net.Listen("unix", socketPath)
	
	grpcServer := grpc.NewServer()
	registerapi.RegisterRegistrationServer(grpcServer, registrar)
	grpcServer.Serve(lis)
	// ...
}

// socket path为：/var/lib/kubelet/plugins_registry/${csiDriverName}-reg.sock
func buildSocketPath(csiDriverName string) string {
    return fmt.Sprintf("%s/%s-reg.sock", *pluginRegistrationPath, csiDriverName)
}

func newRegistrationServer(driverName string, endpoint string, versions []string) registerapi.RegistrationServer {
    return &registrationServer{
        driverName: driverName,
        endpoint:   endpoint,
        version:    versions,
    }
}
// GetInfo is the RPC invoked by plugin watcher
func (e registrationServer) GetInfo(ctx context.Context, req *registerapi.InfoRequest) (*registerapi.PluginInfo, error) {
    return &registerapi.PluginInfo{
        Type:              registerapi.CSIPlugin,
        Name:              e.driverName,
        Endpoint:          e.endpoint,
        SupportedVersions: e.version,
    }, nil
}
func (e registrationServer) NotifyRegistrationStatus(ctx context.Context, status *registerapi.RegistrationStatus) (*registerapi.RegistrationStatusResponse, error) {
    if !status.PluginRegistered {
        os.Exit(1)
    }
    
    return &registerapi.RegistrationStatusResponse{}, nil
}
```


总之，node-driver-registrar sidecar container 主要代码逻辑很简单，先调用我们自研的csi-plugin获取csiDriverName，然后在/var/lib/kubelet/plugins_registry/${csiDriverName}-reg.sock 启动一个grpc server，并按照kubelet plugin registration标准
提供了registrationServer供kubelet plugin manager实现rpc调用。


接下来关键就是kubelet plugin manager是如何rpc调用node-driver-registrar sidecar container的？


## kubelet plugin manager 源码解析
kubelet组件在启动时，会实例化 **[pluginManager](https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/kubelet.go#L733-L736)** 对象，这里的socket dir就是 `/var/lib/kubelet/plugins_registry/` 目录：

```go
const (
    DefaultKubeletPluginsRegistrationDirName = "plugins_registry"
)

klet.pluginManager = pluginmanager.NewPluginManager(
		klet.getPluginsRegistrationDir(), /* sockDir */
		kubeDeps.Recorder,
	)

func (kl *Kubelet) getPluginsRegistrationDir() string {
    return filepath.Join(kl.getRootDir(), config.DefaultKubeletPluginsRegistrationDirName)
}
```

同时还会注册一个CSIPlugin type的csi.RegistrationHandler{}对象，并启动pluginManager对象，代码见 **[L1385-L1391](https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/kubelet.go#L1385-L1391)** ：

```go

// Adding Registration Callback function for CSI Driver
kl.pluginManager.AddHandler(pluginwatcherapi.CSIPlugin, plugincache.PluginHandler(csi.PluginHandler))

// Start the plugin manager
klog.V(4).Infof("starting plugin manager")
go kl.pluginManager.Run(kl.sourcesReady, wait.NeverStop)

```


**[pluginmanager package](https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/pluginmanager/plugin_manager.go)** 模块代码尽管比较多，但实际上主要就实现了两个逻辑。

### plugin watcher
pluginmanager模块plugin watcher对象来 recursively watch /var/lib/kubelet/plugins_registry socket dir，而该对象实际上使用 `github.com/fsnotify/fsnotify` 包来实现该功能。
如果该socket dir增加或删除一个socket file，都会写入desiredStateOfWorld缓存对象的 `socketFileToInfo map[string]PluginInfo` 中，看下主要的watch socket dir代码，代码见 **[L50-L98](https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/pluginmanager/pluginwatcher/plugin_watcher.go#L50-L98)** ：

```go

func (w *Watcher) Start(stopCh <-chan struct{}) error {
	// ...
	fsWatcher, err := fsnotify.NewWatcher()
	w.fsWatcher = fsWatcher
	// 去watch socket dir
	if err := w.traversePluginDir(w.path); err != nil {
		klog.Errorf("failed to traverse plugin socket path %q, err: %v", w.path, err)
	}

	// 启动一个goroutine去watch socket dir中，socket文件的增加和删除
	go func(fsWatcher *fsnotify.Watcher) {
		defer close(w.stopped)
		for {
			select {
			case event := <-fsWatcher.Events:
				if event.Op&fsnotify.Create == fsnotify.Create {
					err := w.handleCreateEvent(event)
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					w.handleDeleteEvent(event)
				}
				continue
			case err := <-fsWatcher.Errors: 
				// ...
                continue
			case <-stopCh:
				// ...
				return
			}
		}
	}(fsWatcher)

	return nil
}

```

当我们daemonset部署node-driver-registrar sidecar container时，/var/lib/kubelet/plugins_registry socket dir中会多一个socket file ${csiDriverName}-reg.sock，
这时plugin watcher对象会把数据写入desiredStateOfWorld缓存中，供第二个逻辑reconcile使用

### reconciler
该reconciler就是一个定时任务，每 `rc.loopSleepDuration` 运行一次，**[L84-L90](https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/pluginmanager/reconciler/reconciler.go#L84-L90)** ：

```go

func (rc *reconciler) Run(stopCh <-chan struct{}) {
	wait.Until(func() {
		rc.reconcile()
	},
		rc.loopSleepDuration,
		stopCh)
}

```

每一次调谐，会去diff下两个缓存map对象：desiredStateOfWorld和actualStateOfWorld。desiredStateOfWorld是期望状态，actualStateOfWorld是实际状态。
如果一个plugin在actualStateOfWorld缓存中但不在desiredStateOfWorld中(表示plugin已经被删除了)，或者尽管在desiredStateOfWorld中但是plugin.Timestamp不匹配(表示plugin重新注册更新了)，
则需要从desiredStateOfWorld缓存中删除并注销插件DeRegisterPlugin；如果一个plugin在desiredStateOfWorld中但不在actualStateOfWorld缓存中，说明是新建的plugin，需要添加到desiredStateOfWorld缓存中并注册插件RegisterPlugin。
看下调谐主要逻辑 **[L110-L164](https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/pluginmanager/reconciler/reconciler.go#L110-L164)** ：

```go

func (rc *reconciler) reconcile() {

	// diff下actualStateOfWorld和desiredStateOfWorld，判断是否需要从desiredStateOfWorld缓存中删除并注销插件DeRegisterPlugin
	for _, registeredPlugin := range rc.actualStateOfWorld.GetRegisteredPlugins() {
		if !rc.desiredStateOfWorld.PluginExists(registeredPlugin.SocketPath) {
			unregisterPlugin = true
		} else {
			for _, dswPlugin := range rc.desiredStateOfWorld.GetPluginsToRegister() {
				if dswPlugin.SocketPath == registeredPlugin.SocketPath && dswPlugin.Timestamp != registeredPlugin.Timestamp {
					unregisterPlugin = true
					break
				}
			}
		}
		if unregisterPlugin {
			err := rc.UnregisterPlugin(registeredPlugin, rc.actualStateOfWorld)
		}
	}

	// diff下desiredStateOfWorld和actualStateOfWorld，查是否需要添加到desiredStateOfWorld缓存中并注册插件RegisterPlugin
	for _, pluginToRegister := range rc.desiredStateOfWorld.GetPluginsToRegister() {
		if !rc.actualStateOfWorld.PluginExistsWithCorrectTimestamp(pluginToRegister) {
			err := rc.RegisterPlugin(pluginToRegister.SocketPath, pluginToRegister.Timestamp, rc.getHandlers(), rc.actualStateOfWorld)
		}
	}
}

```

这里主要查看一个新建的plugin的注册逻辑，reconciler对象会rpc调用node-driver-registrar sidecar container中rpc server提供的的GetInfo。
然后根据返回字段的type，从最开始注册的pluginHandlers中查找对应的handler，这里就是上文说的CSIPlugin type的csi.RegistrationHandler{}对象，并调用该对象的
ValidatePlugin和RegisterPlugin来注册插件，这里的注册插件其实就是设置node annotation和创建/更新CSINode对象。最后rpc调用NotifyRegistrationStatus告知注册结果。

看下注册插件相关代码，**[L74-L134](https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/pluginmanager/operationexecutor/operation_generator.go#L74-L134)** ：

```go

// 与/var/lib/kubelet/plugins_registry/${csiDriverName}-reg.sock建立grpc通信
func dial(unixSocketPath string, timeout time.Duration) (registerapi.RegistrationClient, *grpc.ClientConn, error) {
	// ...
    c, err := grpc.DialContext(ctx, unixSocketPath, grpc.WithInsecure(), grpc.WithBlock())
    return registerapi.NewRegistrationClient(c), c, nil
}

func (og *operationGenerator) GenerateRegisterPluginFunc(/*...*/) func() error {
	registerPluginFunc := func() error {
		client, conn, err := dial(socketPath, dialTimeoutDuration)
        // 调用node-driver-registrar sidecar container中rpc server提供的的GetInfo
		infoResp, err := client.GetInfo(ctx, &registerapi.InfoRequest{})
        // 这里handler就是上文说的CSIPlugin type的csi.RegistrationHandler{}对象
		handler, ok := pluginHandlers[infoResp.Type]
        // 调用handler.ValidatePlugin
		if err := handler.ValidatePlugin(infoResp.Name, infoResp.Endpoint, infoResp.SupportedVersions); err != nil {
		}
		// 加入actualStateOfWorldUpdater缓存
		err = actualStateOfWorldUpdater.AddPlugin(cache.PluginInfo{
			SocketPath: socketPath,
			Timestamp:  timestamp,
			Handler:    handler,
			Name:       infoResp.Name,
		})
		// 这是最关键逻辑，调用handler.RegisterPlugin注册插件
		// 这里的infoResp.Endpoint是我们自研的csi-plugin监听的socket path
		if err := handler.RegisterPlugin(infoResp.Name, infoResp.Endpoint, infoResp.SupportedVersions); err != nil {
			return og.notifyPlugin(client, false, fmt.Sprintf("RegisterPlugin error -- plugin registration failed with err: %v", err))
		}
        // ...
	}
	return registerPluginFunc
}

```


总之，kubelet plugin manager模块代码逻辑比较清晰简单，主要两个逻辑：通过plugin watcher对象去watch socket dir即/var/lib/kubelet/plugins_registry，把plugin数据放入
desiredStateOfWorld缓存中；reconcile调谐desiredStateOfWorld和actualStateOfWorld缓存，调用node-driver-registrar获取plugin info，根据该plugin info查找plugin handler，
然后调用plugin handler来注册插件RegisterPlugin，plugin handler会根据传入的csi-plugin监听的socket path，直接和我们自研的csi-plugin通信(其实node-driver-registrar起到中介作用，传递
我们自研csi-plugin grpc server监听的socket path这个关键信息)。


接下来关键逻辑就是csi.RegistrationHandler{}对象是如何注册插件的？


## csi.RegistrationHandler 源码解析
csi.RegistrationHandler{}对象注册插件逻辑，主要就是更新node annotation和创建/更新CSINode对象，这里可以看下代码逻辑 **[L112-L154](https://github.com/kubernetes/kubernetes/blob/master/pkg/volume/csi/csi_plugin.go#L112-L154)** ：

```go
import (
    csipbv1 "github.com/container-storage-interface/spec/lib/go/csi"
)

func (h *RegistrationHandler) RegisterPlugin(pluginName string, endpoint string, versions []string) error {
    // ...
	// 与我们自研的csi-plugin建立grpc通信，并调用csi-plugin中node service中的NodeGetInfo()获得相关数据，供更新node annotation和创建CSINode对象使用
	csi, err := newCsiDriverClient(csiDriverName(pluginName))
	driverNodeID, maxVolumePerNode, accessibleTopology, err := csi.NodeGetInfo(ctx)
    // ...
    // 这里是主要逻辑：更新node annotation和创建/更新CSINode对象
    err = nim.InstallCSIDriver(pluginName, driverNodeID, maxVolumePerNode, accessibleTopology)
    // ...
	return nil
}

// 与我们自研的csi-plugin建立grpc通信，创建一个grpc client
func newCsiDriverClient(driverName csiDriverName) (*csiDriverClient, error) {
    // ...
    nodeV1ClientCreator := newV1NodeClient
    return &csiDriverClient{
        driverName:          driverName,
        addr:                csiAddr(existingDriver.endpoint),
        nodeV1ClientCreator: nodeV1ClientCreator,
    }, nil
}
// 这里调用csipbv1.NewNodeClient(conn)创建一个grpc client
// CSI标准文档可以参见该仓库的 https://github.com/container-storage-interface/spec/blob/master/spec.md#rpc-interface
func newV1NodeClient(addr csiAddr) (nodeClient csipbv1.NodeClient, closer io.Closer, err error) {
    var conn *grpc.ClientConn
    conn, err = newGrpcConn(addr)
    nodeClient = csipbv1.NewNodeClient(conn)
    return nodeClient, conn, nil
}
func newGrpcConn(addr csiAddr) (*grpc.ClientConn, error) {
    network := "unix"
    return grpc.Dial(string(addr), /*...*/)
}

```

以上代码中，主要包含两个逻辑：更新node annotation；创建更新CSINode对象。
更新node annotation逻辑很简单，主要是往当前Node中增加一个annotation `csi.volume.kubernetes.io/nodeid:{"$csiDriverName":"$driverNodeID"}` ，$csiDriverName是之前rpc调用node-driver-registrar sidecar container获得的，
$driverNodeID是直接rpc调用我们自定义csi-plugin的node service NodeGetInfo获得的，代码可见 **[L237-L273](https://github.com/kubernetes/kubernetes/blob/master/pkg/volume/csi/nodeinfomanager/nodeinfomanager.go#L237-L273)** 。

然后是往apiserver中创建/更新CSINode对象，创建CSINode对象逻辑可见 **[CreateCSINode](https://github.com/kubernetes/kubernetes/blob/master/pkg/volume/csi/nodeinfomanager/nodeinfomanager.go#L433-L470)** ，更新CSINode对象逻辑可见 **[installDriverToCSINode](https://github.com/kubernetes/kubernetes/blob/master/pkg/volume/csi/nodeinfomanager/nodeinfomanager.go#L513-L574)** ，
就可以通过kubectl查看CSINode对象：

![csinodes](./imgs/csinodes.png)

总之，csi.RegistrationHandler{}对象注册插件其实主要就是更新了node annotation和创建/更新该plugin相应的CSINode对象。



## 总结
本文主要学习了CSI Plugin注册机制相关原理逻辑，涉及的主要组件包括：由node-driver-registrar sidecar container和我们自研的csi-plugin组成的daemonset pod，以及
kubelet plugin manager模块框架包，和csi plugin handler模块。其中，kubelet plugin manager模块框架包是一个桥梁，会rpc调用node-driver-registrar sidecar container获取
我们自研csi-plugin相关信息如监听的rpc socket地址，然后调用csi plugin handler模块并传入csi-plugin rpc socket地址， 与csi-plugin直接rpc通信，
实现更新node annotation和创建/更新CSINode对象等相关业务逻辑。

这样，通过以上几个组件模块共同作用，我们自研的一个csi-plugin就注册进来了。
但是，我们自研的csi-plugin提供了create/delete volume等核心功能，又是如何工作的呢？后续有空再更新。


## 参考文献
**[一文读懂 K8s 持久化存储流程](https://mp.weixin.qq.com/s/jpopq16BOA_vrnLmejwEdQ)**

**[从零开始入门 K8s | Kubernetes 存储架构及插件使用](https://mp.weixin.qq.com/s/QWLGkpqpMdsY1w6npZj-yQ)**

**[Kubernetes Container Storage Interface (CSI) Documentation](https://kubernetes-csi.github.io/docs/introduction.html)**

**[node-driver-registrar](https://github.com/kubernetes-csi/node-driver-registrar)**
























```text

第 1 步，在启动的时候有一个约定，比如说在 `/var/lib/kuberlet/plugins_registry` 这个目录每新加一个文件，就相当于每新加了一个 Plugin；
启动 Node-Driver-Registrar，它首先会向 CSI-Plugin 发起一个接口调用 GetPluginInfo，这个接口会返回 CSI 所监听的地址以及 CSI-Plugin 的一个 Driver name；

第 2 步，Node-Driver-Registrar 会监听 GetInfo 和 NotifyRegistrationStatus 两个接口；

第 3 步，会在 `/var/lib/kuberlet/plugins_registry` 这个目录下启动一个 Socket，生成一个 Socket 文件 ，例如 "diskplugin.csi.alibabacloud.com-reg.sock"，此时 Kubelet 通过 Watcher 发现这个 Socket 后，它会通过该 Socket 向 Node-Driver-Registrar 的 GetInfo 接口进行调用。GetInfo 会把刚才我们所获得的的 CSI-Plugin 的信息返回给 Kubelet，该信息包含了 CSI-Plugin 的监听地址以及它的 Driver name；

第 4 步，Kubelet 通过得到的监听地址对 CSI-Plugin 的 NodeGetInfo 接口进行调用；

第 5 步，调用成功之后，Kubelet 会去更新一些状态信息，比如节点的 Annotations、Labels、status.allocatable 等信息，同时会创建一个 CSINode 对象；

第 6 步，通过对 Node-Driver-Registrar 的 NotifyRegistrationStatus 接口的调用告诉它我们已经把 CSI-Plugin 注册成功了。

通过以上 6 步就实现了 CSI Plugin 注册机制。

```
