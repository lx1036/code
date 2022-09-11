

# CPU 隔离
单机CPU隔离方面主要分为以下几个部分：
1. 同一个逻辑CPU上运行的进程优先级问题
2. 同一个物理core上的超线程干扰问题
3. 将进程绑定在不同的逻辑CPU上的能力
4. NUMA架构下的CPU、内存绑定逻辑

## 同一逻辑CPU上运行的进程优先级问题 GroupIdentity(cpu/cpu.bvt_warp_ns)
这个问题的解决办法在社区版内核是无法解决的，阿里内核通过更改CFS调度算法，实现了不同进程具有不同的优先级。

linux 内核进程调度优先级分类，即 aliyun linux 内核的 Group Identity 功能，见 @see https://mp.weixin.qq.com/s/y8k_q6rhTIubQ-lqvDp2hw
在阿里的内核中通过修改CFS调度算法，实现将进程分为4个优先级进行调度，通过 cgroup 进行管理，管理文件为 cpu.bvt_warp_ns，
修改cpu.bvt_warp_ns 会改变 cpu.identity 的值，级别和对应的 cpu.identity 的值如下表：

| 级别(bvt) | 说明 | 对应 identity_group 值，二进制参考如下图 |
| --- | --- |------------------------------|
| 2 | 该cgroup中的任务为最高优先级，具备SMT驱逐能力 | 22(10110)                    |
| 1 | 该cgroup中的任务为高优先级 | 18(10010)                    |
| 0 | 该cgroup中的任务为普通系统任务 | 0（00000）|
| -1 | 该cgroup中的任务为低优先级任务 | 9（01010）|

```shell
# 该 pod 是 guaranteed
[root@docker39 cpu]# cat kubepods/podef0329ae-4e2f-41b4-a2ee-ed7c51023ee7/2e520b146d322f54c8c256e6369cd735ff0bc056c86069a734a095e6567688cb/cpu.bvt_warp_ns # sandbox 容器
0
[root@docker39 cpu]# cat kubepods/podef0329ae-4e2f-41b4-a2ee-ed7c51023ee7/edbf780d09f1318317f8a7b609352d69e9ec9989f6dcf1c9c2562c0fb699c04b/cpu.bvt_warp_ns # 业务容器
2
[root@docker39 cpu]# cat kubepods/podef0329ae-4e2f-41b4-a2ee-ed7c51023ee7/edbf780d09f1318317f8a7b609352d69e9ec9989f6dcf1c9c2562c0fb699c04b/cpu.identity
22
```


## 同一个物理core上的超线程干扰问题




## 将进程绑定在不同的逻辑CPU上的能力




## NUMA架构下的CPU、内存绑定逻辑


