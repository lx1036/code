


# syn cookie(TFO?)
代码在:
```md
/root/linux-5.10.142/tools/testing/selftests/bpf/progs/test_tcp_check_syncookie_kern.c
/root/linux-5.10.142/tools/testing/selftests/bpf/test_tcp_check_syncookie_user.c
/root/linux-5.10.142/tools/testing/selftests/bpf/test_tcp_check_syncookie.sh
```

## tcp_syncookies

/proc/sys/net/ipv4/tcp_syncookies 是 Linux 内核中的一个配置选项，用于控制 TCP 协议的 SYN cookies 功能。

SYN cookies 是一种防止 SYN flood 攻击的技术。当攻击者向服务器发送大量的伪造源 IP 地址的 SYN 请求时，服务器会为每个请求分配一个连接队列项，
并等待客户端的确认。由于这些请求都是伪造的，客户端不会发送确认，导致服务器上的连接队列项耗尽，从而拒绝服务。

为了防止这种情况，Linux 内核提供了一个名为 SYN cookies 的功能。当服务器检测到过多的 SYN 请求时，它可以生成一个特殊的 cookie，
并将其包含在 SYN+ACK 响应中。客户端必须在后续的 ACK 报文中正确地返回这个 cookie，才能建立连接。这样可以有效地防止伪造源 IP 地址的 SYN 
请求耗尽服务器的连接队列项。

/proc/sys/net/ipv4/tcp_syncookies 配置选项就是用来控制 SYN cookies 功能的。如果该值设置为 1，则表示启用 SYN cookies 功能；
如果设置为 0，则表示禁用 SYN cookies 功能；如果值为 2 的情况下，服务器可以在某些情况下发送带有时间戳的 SYN cookies。


## tcp_fastopen
/proc/sys/net/ipv4/tcp_fastopen 是 Linux 内核中的一个配置选项，用于控制 TCP Fast Open (TFO) 功能。

TCP Fast Open 是一种优化 TCP 连接建立过程的方法。在传统的 TCP 连接过程中，客户端首先发送一个 SYN 包，然后服务器回复一个 SYN+ACK 包，
最后客户端再发送一个 ACK 包，完成三次握手。而在 TFO 中，客户端可以在第一次发送 SYN 包时就携带数据，从而减少一次网络往返时间，提高连接速度。

/proc/sys/net/ipv4/tcp_fastopen 配置选项就是用来控制 TFO 功能的。
如果该值设置为 1 或更高，则表示启用 TFO 功能；如果设置为 0，则表示禁用 TFO 功能；
如果值为7，表示内核将支持 TFO 功能，并且在同一时间内，最多可以有 65535 个并发的 TFO 连接。



相关文章:
```md
https://cs.pynote.net/net/tcp/202205052/
```





