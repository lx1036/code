

# linux numa 架构
socket: 表示 cpu 插槽个数，比如一台物理机有两个 cpu 插槽，就是两个 socket。
numa node: 一个 socket 可以被划分为多个 numa node。

```shell
yum install -y numactl
numactl -H

lscpu # 查看 cpu 拓扑，或者 `cat /proc/cpuinfo` 查看具体信息

Architecture:        x86_64
CPU op-mode(s):      32-bit, 64-bit
Byte Order:          Little Endian
CPU(s):              80
On-line CPU(s) list: 0-79
Thread(s) per core:  2
Core(s) per socket:  20
Socket(s):           2
NUMA node(s):        2
Vendor ID:           GenuineIntel
CPU family:          6
Model:               85
Model name:          Intel(R) Xeon(R) Gold 5218R CPU @ 2.10GHz
Stepping:            7
CPU MHz:             2100.000
CPU max MHz:         2100.0000
CPU min MHz:         800.0000
BogoMIPS:            4200.00
Virtualization:      VT-x
L1d cache:           32K
L1i cache:           32K
L2 cache:            1024K
L3 cache:            28160K
NUMA node0 CPU(s):   0-19,40-59
NUMA node1 CPU(s):   20-39,60-79
```
