



参考代码：pkg/kubelet/pluginmanager/plugin_manager.go

**[kubelet plugin registration mechanism](https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/pluginmanager/pluginwatcher/README.md)** :
It discovers plugins by monitoring inotify events under the directory returned by kubelet.getPluginsDir().
Plugins are expected to implement the gRPC registration service specified in staging/src/k8s.io/kubelet/pkg/apis/pluginregistration/v1/api.proto
对于plugin实现，可以参考 pkg/kubelet/pluginmanager/pluginwatcher/example_plugin.go


pluginmanager使用示例：
```go
// (1) 实例化：https://github.com/kubernetes/kubernetes/blob/release-1.19/pkg/kubelet/kubelet.go#L716-L719
klet.pluginManager = pluginmanager.NewPluginManager(
    klet.getPluginsRegistrationDir(), /* sockDir */
    kubeDeps.Recorder,
)

// (2) 添加handler和run pluginmanager: https://github.com/kubernetes/kubernetes/blob/release-1.19/pkg/kubelet/kubelet.go#L1307-L1313
// Adding Registration Callback function for CSI Driver
kl.pluginManager.AddHandler(pluginwatcherapi.CSIPlugin, plugincache.PluginHandler(csi.PluginHandler))
// Adding Registration Callback function for Device Manager
kl.pluginManager.AddHandler(pluginwatcherapi.DevicePlugin, kl.containerManager.GetPluginRegistrationHandler())
// Start the plugin manager
klog.V(4).Infof("starting plugin manager")
go kl.pluginManager.Run(kl.sourcesReady, wait.NeverStop)

```
