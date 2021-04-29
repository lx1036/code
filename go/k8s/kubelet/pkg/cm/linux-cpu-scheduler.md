



# linux cpu scheduler
https://www.kernel.org/doc/Documentation/scheduler/sched-bwc.txt

关键文件：cpu.stat。kubelet -> cadvisor -> runc cgroup 读取的就是容器的 cpu.stat 文件，知晓该容器cpu资源实际使用率。

Statistics
----------
A group's bandwidth statistics are exported via 3 fields in cpu.stat.

cpu.stat:
- nr_periods: Number of enforcement intervals that have elapsed.
- nr_throttled: Number of times the group has been throttled/limited.
- throttled_time: The total time duration (in nanoseconds) for which entities
  of the group have been throttled.

Examples
--------
1. Limit a group to 1 CPU worth of runtime.
   If period is 250ms and quota is also 250ms, the group will get
   1 CPU worth of runtime every 250ms.
  ```shell
     # echo 250000 > cpu.cfs_quota_us /* quota = 250ms */
     # echo 250000 > cpu.cfs_period_us /* period = 250ms */
  ```

2. Limit a group to 2 CPUs worth of runtime on a multi-CPU machine.
   With 500ms period and 1000ms quota, the group can get 2 CPUs worth of
   runtime every 500ms.
  ```shell
    # echo 1000000 > cpu.cfs_quota_us /* quota = 1000ms */
    # echo 500000 > cpu.cfs_period_us /* period = 500ms */
  ```
   The larger period here allows for increased burst capacity.

3. Limit a group to 20% of 1 CPU.
   With 50ms period, 10ms quota will be equivalent to 20% of 1 CPU.
  ```shell
    # echo 10000 > cpu.cfs_quota_us /* quota = 10ms */
     # echo 50000 > cpu.cfs_period_us /* period = 50ms */
  ```
   By using a small period here we are ensuring a consistent latency response at the expense of burst capacity.



