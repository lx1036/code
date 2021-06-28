
## Volcano Metrics

```yaml
volcano_e2e_scheduling_latency_milliseconds: (histogram) volcano scheduler 调度插件调度pod和binding pod的整个过程耗费时间

volcano_e2e_job_scheduling_duration: (gauge)
  avg(volcano_e2e_job_scheduling_duration{}) by (queue): (按queue分组)过去24小时内调度属于该queue的pods平均耗时
  avg(volcano_e2e_job_scheduling_duration{}) by (job_namespace): (按namespace分组)过去24小时内调度属于该queue的pods平均耗时
  increase(volcano_e2e_job_scheduling_duration{}[24h]) != 0: 过去24小时内volcano scheduler调度pods耗时的增量
  stddev(volcano_e2e_job_scheduling_duration)/avg(volcano_e2e_job_scheduling_duration): 标准差除以平均数，为变异系数CV(Coefficient of Variation) https://baike.baidu.com/item/%E5%8F%98%E5%BC%82%E7%B3%BB%E6%95%B0


kube_pod_volcano_container_resource_requests: (gauge)
  sum(kube_pod_volcano_container_resource_requests{resource="memory", unit="byte",job="kube-state-metrics",queue!=""}) by (queue): (按queue分组)该queue下所有pod使用memory总和
  sum(kube_pod_volcano_container_resource_requests{resource="memory", unit="byte",job="kube-state-metrics"}) by (volcano_namespace): (按namespace分组)该queue下所有pod使用memory总和



缺少的metrics:
  plugins:
    plugin_scheduling_latency_microseconds: grafana dashboard 缺少每一个 plugin 的耗时
    action_scheduling_latency_microseconds: grafana dashboard 缺少每一个 action 的耗时

```



