

# CPU Cgroup 压制水位线测试(阿里内核)
阿里 linux 内核必然包含 cpu.bvt_warp_ns cpu.identity 这些 cgroup，否则就是社区 linux 内核。

## 背景
所有离线低优先级 pod 放在 /sys/fs/cgroup/cpu/offline cgroup 下，通过设置整个 offline cpu.cfs_quota_us 就可以管理该 offline cgroup
下的所有 pod 的压制水位线。cpu.cfs_quota_us 计算公式为：

```shell
# 假定机器有 0-23 共 24 个 cpu
# 比如压制水位线为整机 cpu 20% 就开始压制，但是只会压制 offline cgroup，不会压制 online cgroup，即不会设置 online 的 cpu.cfs_quota_us
cpu.cfs_quota_us = 20% * 24 * cpu.cfs_period_us = 480000
```

这样所有在 offline cgroup 的 pods 锁使用的资源不会超过整机的 20%。
同时还可以动态设置，比如现在动态设置成 10%，然后就会立刻被压制成只能消费 10% cpu 资源。

## 验证

```shell
cd /sys/fs/cgroup/cpu
mkdir online
mkdir offline

# 参考 https://www.kernel.org/doc/html/latest/admin-guide/cgroup-v1/cgroups.html#what-does-clone-children-do
# "This flag only affects the cpuset controller. If the clone_children flag is enabled (1) in a cgroup, 
# a new cpuset cgroup will copy its configuration from the parent during initialization."
cd cd /sys/fs/cgroup/cpuset
echo 1 > cgroup.clone_children
mkdir online
mkdir offline

# 这时没设置压制水位线，所有 cpu 都是消费 100%
docker run -it --cgroup-parent=/offline lx1036/ubuntu:stress-ng bash
stress-ng -c 24 # 打满 24 个 cpu

# 压制水位线设置成 20%
cd /sys/fs/cgroup/cpu/offline
echo 480000 > cpu.cfs_quota_us
# 压制水位线设置成 10%
echo 240000 > cpu.cfs_quota_us
```

![cpu-suppress](./imgs/cpu-suppress.png)


## 结论
使用阿里内核，整机水位线压制只需要动态设置 cpu.cfs_quota_us cgroup，就可以使得离线所有 pod 的 cpu 资源总和不会超过设定值。
在混部时，cpu.cfs_quota_us = node.Total * SLOPercent - pod(LS).Used - system.Used, LS 表示在线 pod 实际使用 cpu 资源
