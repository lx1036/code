


# 问题
本 demo 来自于内核 patch 文档 https://www.spinics.net/lists/netdev/msg403579.html, 用来验证 bpf ipip 的示例。

代码在内核里: /root/linux-5.10.142/samples/bpf/tc_l2_redirect.sh

(1)为方便使用，用户态直接使用 bpftool 来 update map, 参考文档 https://manpages.ubuntu.com/manpages/focal/en/man8/bpftool-map.8.html :

```shell
# 安装 bpftool 工具
apt install -y linux-tools-5.4.0-164-generic jq

root@iZj6ch9k9rqv9n8ab77e0wZ:~/ipip# cat /sys/class/net/tun1/ifindex
45
# 注意这里的 45 值在最前面，而不是 "0 0 0 45", 754974720=int("0x2d000000", 16) [python3]
# 另外 key 和 value 都是四字节大小
root@iZj6ch9k9rqv9n8ab77e0wZ:~/ipip# bpftool map show pinned /sys/fs/bpf/tc/globals/tun_iface
140: array  flags 0x0
        key 4B  value 4B  max_entries 1  memlock 4096B
root@iZj6ch9k9rqv9n8ab77e0wZ:~/ipip# bpftool map update pinned /sys/fs/bpf/tc/globals/tun_iface key 0 0 0 0 value 45 0 0 0
root@iZj6ch9k9rqv9n8ab77e0wZ:~/ipip# bpftool map dump pinned /sys/fs/bpf/tc/globals/tun_iface -j | jq
[
  {
    "key": [
      "0x00",
      "0x00",
      "0x00",
      "0x00"
    ],
    "value": [
      "0x2d",
      "0x00",
      "0x00",
      "0x00"
    ]
  }
]

```

(2)查看 bpf_trace_printk() 函数的日志
```shell
tail -n 100 /sys/kernel/debug/tracing/trace
# tracer: nop
#
# entries-in-buffer/entries-written: 5/5   #P:4
#
#                                _-----=> irqs-off
#                               / _----=> need-resched
#                              | / _---=> hardirq/softirq
#                              || / _--=> preempt-depth
#                              ||| /     delay
#           TASK-PID     CPU#  ||||   TIMESTAMP  FUNCTION
#              | |         |   ||||      |         |
            ping-1137009 [001] ..s1 150505.398215: 0: e/ingress redirect daddr4:a0a0166 to ifindex:754974720
            ping-1137062 [001] ..s1 150515.532237: 0: e/ingress redirect daddr4:a0a0166 to ifindex:754974720
            ping-1137518 [002] ..s1 150769.460319: 0: e/ingress redirect daddr4:a0a0166 to ifindex:754974720
            ping-1137844 [003] ..s1 150923.706842: 0: e/ingress redirect daddr4:a0a0166 to ifindex:45
     ksoftirqd/3-30      [003] ..s. 150923.706959: 0: ingress forward to ifindex:45 daddr4:a020101
```


# 参考文献

