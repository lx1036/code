

# Core Concepts 19%

(1) Node

**[nodes](https://kubernetes.io/zh/docs/concepts/architecture/nodes/)**:
一个 Node Status 主要包含：
* addresses:
```json
[
    {
        "address": "192.168.64.36",
        "type": "InternalIP"
    },
    {
        "address": "minikube",
        "type": "Hostname"
    }
]
```
* 容量和可分配资源 capacity and allocatable:
```json
{
    "allocatable": {
        "cpu": "2",
        "ephemeral-storage": "16954224Ki",
        "hugepages-2Mi": "0",
        "memory": "5954220Ki",
        "pods": "110"
    },
    "capacity": {
        "cpu": "2",
        "ephemeral-storage": "16954224Ki",
        "hugepages-2Mi": "0",
        "memory": "5954220Ki",
        "pods": "110"
    }
}
``` 
* 条件 conditions
```json
[
    {
        "lastHeartbeatTime": "2020-05-30T14:45:24Z",
        "lastTransitionTime": "2020-05-30T14:45:24Z",
        "message": "Calico is running on this node",
        "reason": "CalicoIsUp",
        "status": "False",
        "type": "NetworkUnavailable"
    },
    {
        "lastHeartbeatTime": "2020-05-30T15:05:22Z",
        "lastTransitionTime": "2020-05-28T14:30:13Z",
        "message": "kubelet has sufficient memory available",
        "reason": "KubeletHasSufficientMemory",
        "status": "False",
        "type": "MemoryPressure"
    },
    {
        "lastHeartbeatTime": "2020-05-30T15:05:22Z",
        "lastTransitionTime": "2020-05-28T14:30:13Z",
        "message": "kubelet has no disk pressure",
        "reason": "KubeletHasNoDiskPressure",
        "status": "False",
        "type": "DiskPressure"
    },
    {
        "lastHeartbeatTime": "2020-05-30T15:05:22Z",
        "lastTransitionTime": "2020-05-28T14:30:13Z",
        "message": "kubelet has sufficient PID available",
        "reason": "KubeletHasSufficientPID",
        "status": "False",
        "type": "PIDPressure"
    },
    {
        "lastHeartbeatTime": "2020-05-30T15:05:22Z",
        "lastTransitionTime": "2020-05-28T14:30:21Z",
        "message": "kubelet is posting ready status",
        "reason": "KubeletReady",
        "status": "True",
        "type": "Ready"
    }
]
```
* 节点信息 nodeInfo:
```json
{
    "nodeInfo": {
        "architecture": "amd64",
        "bootID": "93ad0808-a0ff-460b-8b42-106c1534f641",
        "containerRuntimeVersion": "docker://19.3.8",
        "kernelVersion": "4.19.107",
        "kubeProxyVersion": "v1.18.2",
        "kubeletVersion": "v1.18.2",
        "machineID": "167395c2cd4f46deb567c973af427a11",
        "operatingSystem": "linux",
        "osImage": "Buildroot 2019.02.10",
        "systemUUID": "ab1b11ea-0000-0000-8570-787b8aaa73af"
    }
}
```


1. 列出 minikube node 上所有正在运行的 pod?
列出 minikube node 上可分配资源，包括可用 CPU/Memory？
```shell script
kubectl get pods -A \
  --field-selector="spec.nodeName=minikube,status.phase!=Succeeded,status.phase!=Failed"
```






