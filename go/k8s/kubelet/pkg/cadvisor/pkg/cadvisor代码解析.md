



# cadvisor
https://github.com/google/cadvisor





## cadvisor安装时所需mount目录
* /proc: 









## 参考文献
**[cAdvisor源码分析](https://cloud.tencent.com/developer/article/1096375)**



## TroubleShooting
(1)cadvisor是如何拿到container cpu_usage memory_usage的？
!!! 注意：没有调用 docker api，而是直接读取 cgroup stats 文件。docker 也是直接读取 cgroup stats 文件。cadvisor 这里绕过 docker，踢掉 docker。
最后会通过 CgroupManager 读取每一个 subsystem stats 文件，比如 cpu，去读取该容器的 cpu.stat 文件：

```shell
cat cpu.stat
#nr_periods 196329
#nr_throttled 1
#throttled_time 53507508

cat memory.usage_in_bytes
#8683520
```

**[cpu.stat](https://www.kernel.org/doc/Documentation/scheduler/sched-bwc.txt)**:
- nr_periods: Number of enforcement intervals that have elapsed.
- nr_throttled: Number of times the group has been throttled/limited.
- throttled_time: The total time duration (in nanoseconds) for which entities of the group have been throttled.
