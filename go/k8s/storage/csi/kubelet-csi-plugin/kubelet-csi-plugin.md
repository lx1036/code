




**[kubelet plugin registration mechanism](https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/pluginmanager/pluginwatcher/README.md)** :
It discovers plugins by monitoring inotify events under the directory returned by kubelet.getPluginsDir().
Plugins are expected to implement the gRPC registration service specified in staging/src/k8s.io/kubelet/pkg/apis/pluginregistration/v1/api.proto
对于plugin实现，可以参考 pkg/kubelet/pluginmanager/pluginwatcher/example_plugin.go


