


# Socket BPF
解决的问题：对于源和目的端都在同一台机器的应用来说，可以使用 Socket BPF 绕过整个 TCP/IP 协议栈，直接将数据发送到 socket 对端。






# 参考文献
**[利用 ebpf sockmap/redirection 提升 socket 性能（2020）](http://arthurchiao.art/blog/socket-acceleration-with-ebpf-zh/)**

**[使用eBPF（绕过 TCP/IP）加速云原生应用程序的经验教训](https://www.cnxct.com/lessons-using-ebpf-accelerating-cloud-native-zh/)**

**[代码1](https://github.com/ArthurChiao/socket-acceleration-with-ebpf)**

**[代码2](https://github.com/cyralinc/os-eBPF)**