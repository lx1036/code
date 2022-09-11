

# CPU 超线程干扰测试(阿里内核)
不同于 linux 社区内核的 CFS 公平调度器，每个进程的调度优先级是一样的，阿里内核实现了可以通过 cgroup 设置进程调度优先级，这个很强!!!
就是阿里内核 cpu cgroup GroupIdentity 功能。

## 背景
cpu numa 架构下，如果一个物理核两个逻辑核下，一个逻辑核跑在线 pod，对端逻辑核跑离线 pod，离线 pod 会干扰在线 pod，这就是超线程干扰问题。
其根本原因是因为 linux CFS 进程调度器是公平调度器，没有根据进程优先级来调度，使得在线 pod 的进程和离线 pod 的进程公平调度，
会极大影响在线 pod P95/P99，这在生产中不可接受，最好能使得干扰率在 5% 以内。

解决方案：(1)直接物理核隔离，通过设置 cpuset 使得当前物理核即两个逻辑核，要么都跑在线 pod，要么都跑离线 pod，通过这样的绑核操作直接物理隔离，
然后 agent 根据机器的在线 pod 实际资源使用率来动态设置大框界限，来实现在线 pod 和离线 pod 的 cpu 隔离；
(2)使用阿里内核，该内核在内核层通过实现进程调度优先级 cpu.bvt_warp_ns cgroup，来实现更高优先级进程优先被 linux 调度器调度，从而解决超线程干扰问题。


## 验证

设置 online 和 offline cgroup 的进程优先级 cpu.bvt_warp_ns:

```shell
echo 2 > /sys/fs/cgroup/cpu/online/cpu.bvt_warp_ns
echo -1 > /sys/fs/cgroup/cpu/offline/cpu.bvt_warp_ns

# 运行离线 pod
echo 1920000 > /sys/fs/cgroup/cpu/offline/cpu.cfs_quota_us # 离线 pod 只打满 80% cpu, 100000 * 24 * 0.8 = 1920000
docker run -it --cgroup-parent=/offline lx1036/ubuntu:stress-ng bash
stress-ng -c 24 # 打满 24 个 cpu

# 运行在线 pod server 端
docker run -it -p 8001:8000 --cpus 4 --cpuset-cpus 0-3  --cgroup-parent=/online lx1036/vivo_test:v0.0.4
cd
./serverTest

```

然后运行在有离线 pod 打满 80% cpu 条件下，给在线 pod server 打流量，查看 P95 情况。client 脚本为：

```shell
#!/bin/sh
IP=`hostname -i`
for ((i=0;i<1000;i++)); # 循环 1000次
do
#sleep 0.0001;
curl -s "http://$IP:8001/cpu?cpu=1&count=4000000" > /dev/null ; # 每次计算 4000000
done
curl -g http://$IP:8001/metrics | grep response
```

得到结果：

```shell
# HELP http_response_time Duration of HTTP requests.
# TYPE http_response_time summary
http_response_time{path="/cpu",quantile="0.5"} 0.001557837
http_response_time{path="/cpu",quantile="0.9"} 0.001566839
http_response_time{path="/cpu",quantile="0.95"} 0.001574852
http_response_time{path="/cpu",quantile="0.99"} 0.00167129
http_response_time_sum{path="/cpu"} 7.139152625000009
http_response_time_count{path="/cpu"} 4560
```

然后关闭离线 pod，测试没有离线 pod 条件下在线 pod 的 P95，得到结果:

```shell
# HELP http_response_time Duration of HTTP requests.
# TYPE http_response_time summary
http_response_time{path="/cpu",quantile="0.5"} 0.001558533
http_response_time{path="/cpu",quantile="0.9"} 0.001568791
http_response_time{path="/cpu",quantile="0.95"} 0.001575515
http_response_time{path="/cpu",quantile="0.99"} 0.00167129
http_response_time_sum{path="/cpu"} 8.703053113000022
http_response_time_count{path="/cpu"} 5560
```

观察 P95 干扰率：(0.001574852-0.001575515)/0.001575515 = 0.4%，干扰率很低，满足小于 5% 条件。

## 结论
使用阿里内核，只需要设置在线和离线的进程调度优先级 cpu.bvt_warp_ns，就可以解决超线程干扰问题。
