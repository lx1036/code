


# 技术分享：kubelet 是如何知道一个 pod 的资源实际使用量的？
(1) kubelet 通过 summary api 暴露 pod 的资源实际使用量, metrics-server 调用 https://{node_ip}:10250/stats/summary?only_cpu_and_memory=true api
获取包含 Node/Pod stats 数据。

(2) Kubelet 对象实际上使用 StatsProvider 来获取 stats 数据，StatsProvider 主要有：cadvisorStatsProvider 和 criStatsProvider。
cadvisorStatsProvider 是从 cadvisor 里读取 stats，criStatsProvider 是从 CRI 里读取 stats。目前我们环境，默认使用的是 cadvisorStatsProvider 对象。
pkg/kubelet/kubelet.go#L665-L683 -> pkg/kubelet/stats/stats_provider.go

