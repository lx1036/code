

# TCP INFO
TCP_INFO(https://man7.org/linux/man-pages/man7/tcp.7.html):Used to collect information about this socket.
代码在: /root/linux-5.10.142/include/uapi/linux/tcp.h



# tcp rtt
Trace TCP round trip time.

可以直接使用命令获取 TCP_INFO 信息，ss 命令可以显示 Linux 系统中 Socket 的统计信息。
例如，使用以下命令可以列出所有 TCP 连接以及其详细状态：

```shell
ss -it
```

代码在：
```md
示例一:
用户态: /root/linux-5.10.142/tools/testing/selftests/bpf/prog_tests/tcp_rtt.c
内核态: /root/linux-5.10.142/tools/testing/selftests/bpf/progs/tcp_rtt.c

示例二:
用户态: https://github.com/cilium/ebpf/blob/main/examples/tcprtt_sockops/main.go
内核态: https://github.com/cilium/ebpf/blob/main/examples/tcprtt_sockops/tcprtt_sockops.c

示例三:
用户态: https://github.com/cilium/ebpf/blob/main/examples/tcprtt/main.go
内核态: https://github.com/cilium/ebpf/blob/main/examples/tcprtt/tcprtt.c

示例四:
用户态: https://github.com/mozillazg/libbpfgo-tools/blob/master/tools/tcprtt/main.go
内核态: https://github.com/iovisor/bcc/blob/master/libbpf-tools/tcprtt.bpf.c

示例五:
https://github.com/iovisor/bcc/blob/master/tools/tcprtt.py
https://github.com/iovisor/bcc/blob/master/tools/tcprtt_example.txt

```

RTT（Round Trip Time）由三部分组成：链路的传播时间（propagation delay）+ 末端系统的处理时间 + 路由器缓存中的排队和处理时间（queuing delay）
* 前两个部分的值对于一个TCP连接相对固定，路由器缓存中的排队和处理时间会随着整个网络拥塞程度的变化而变化。所以RTT的变化在一定程度上反应了网络的拥塞程度。


tcpinfo 结构体代码在 /root/linux-5.10.142/include/uapi/linux/tcp.h :

```c

struct tcp_info {
	__u8	tcpi_state;
	__u8	tcpi_ca_state;
	__u8	tcpi_retransmits;
	__u8	tcpi_probes;
	__u8	tcpi_backoff;
	__u8	tcpi_options;
	__u8	tcpi_snd_wscale : 4, tcpi_rcv_wscale : 4;
	__u8	tcpi_delivery_rate_app_limited:1, tcpi_fastopen_client_fail:2;

	__u32	tcpi_rto;
	__u32	tcpi_ato;
	__u32	tcpi_snd_mss;
	__u32	tcpi_rcv_mss;

	__u32	tcpi_unacked;
	__u32	tcpi_sacked;
	__u32	tcpi_lost;
	__u32	tcpi_retrans;
	__u32	tcpi_fackets;

	/* Times. */
	__u32	tcpi_last_data_sent;
	__u32	tcpi_last_ack_sent;     /* Not remembered, sorry. */
	__u32	tcpi_last_data_recv;
	__u32	tcpi_last_ack_recv;

	/* Metrics. */
	__u32	tcpi_pmtu;
	__u32	tcpi_rcv_ssthresh;
	__u32	tcpi_rtt;
	__u32	tcpi_rttvar;
	__u32	tcpi_snd_ssthresh;
	__u32	tcpi_snd_cwnd;
	__u32	tcpi_advmss;
	__u32	tcpi_reordering;

	__u32	tcpi_rcv_rtt;
	__u32	tcpi_rcv_space;

	__u32	tcpi_total_retrans;

	__u64	tcpi_pacing_rate;
	__u64	tcpi_max_pacing_rate;
	__u64	tcpi_bytes_acked;    /* RFC4898 tcpEStatsAppHCThruOctetsAcked */
	__u64	tcpi_bytes_received; /* RFC4898 tcpEStatsAppHCThruOctetsReceived */
	__u32	tcpi_segs_out;	     /* RFC4898 tcpEStatsPerfSegsOut */
	__u32	tcpi_segs_in;	     /* RFC4898 tcpEStatsPerfSegsIn */

	__u32	tcpi_notsent_bytes;
	__u32	tcpi_min_rtt;
	__u32	tcpi_data_segs_in;	/* RFC4898 tcpEStatsDataSegsIn */
	__u32	tcpi_data_segs_out;	/* RFC4898 tcpEStatsDataSegsOut */

	__u64   tcpi_delivery_rate;

	__u64	tcpi_busy_time;      /* Time (usec) busy sending data */
	__u64	tcpi_rwnd_limited;   /* Time (usec) limited by receive window */
	__u64	tcpi_sndbuf_limited; /* Time (usec) limited by send buffer */

	__u32	tcpi_delivered;
	__u32	tcpi_delivered_ce;

	__u64	tcpi_bytes_sent;     /* RFC4898 tcpEStatsPerfHCDataOctetsOut */
	__u64	tcpi_bytes_retrans;  /* RFC4898 tcpEStatsPerfOctetsRetrans */
	__u32	tcpi_dsack_dups;     /* RFC4898 tcpEStatsStackDSACKDups */
	__u32	tcpi_reord_seen;     /* reordering events seen */

	__u32	tcpi_rcv_ooopack;    /* Out-of-order packets received */

	__u32	tcpi_snd_wnd;	     /* peer's advertised receive window after
				      * scaling (bytes)
				      */
};

```

