



# container stats
由于历史原因，目前kubelet默认使用 cadvisor->cgroup 来读取 container stats。同时也提供了从 cri 读取 container stats，以后可能会建议
使用从 cri 读取 container stats 数据，可以见设计文档： 
**[Container Runtime Interface: Container Metrics](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-node/cri-container-stats.md)**





