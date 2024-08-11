

# custom container runtime
docker/containerd 默认使用 runc runtime，但是可以使用自定义 runtime，来修改 OCI spec config.json 文件内容，然后执行 runc 命令。

比如，使用 containerd 运行 ascend-container-runtime 的配置：

```toml
# cgroup v2

[plugins."io.containerd.grpc.v1.cri".containerd.runtimes]
    [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
        base_runtime_spec = ""
        cni_conf_dir = ""
        cni_max_conf_num = 0
        container_annotations = []
        pod_annotations = []
        privileged_without_host_devices = false
        runtime_engine = ""
        runtime_path = ""
        runtime_root = ""
        runtime_type = "io.containerd.runc.v2"
        [plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options]
        BinaryName = "/usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime"
        CriuImagePath = ""
        CriuPath = ""
        CriuWorkPath = ""
        IoGid = 0
        IoUid = 0
        NoNewKeyring = false
        NoPivotRoot = false
        Root = ""
        ShimCgroup = ""
        SystemdCgroup = true

```

containerd 调用 ascend-docker-runtime 的代码在：

```shell
# /usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime create --bundle bundle xxx
https://github.com/containerd/containerd/blob/v2.0.0-rc.3/cmd/containerd-shim-runc-v2/runc/container.go#L122-L131 ->
https://github.com/containerd/containerd/blob/v2.0.0-rc.3/cmd/containerd-shim-runc-v2/runc/container.go#L218 ->
https://github.com/containerd/containerd/blob/v2.0.0-rc.3/cmd/containerd-shim-runc-v2/process/init.go#L144 ->
https://github.com/containerd/go-runc/blob/v1.1.0/runc.go#L180-L191
```

代码仓库：
```md
https://gitee.com/ascend/ascend-docker-runtime
https://github.com/NVIDIA/nvidia-container-toolkit
https://github.com/containerd/containerd
https://github.com/containerd/go-runc
```
