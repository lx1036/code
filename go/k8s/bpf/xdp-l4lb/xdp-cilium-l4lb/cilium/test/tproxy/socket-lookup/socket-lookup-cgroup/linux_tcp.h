
#ifndef XDP_CILIUM_L4LB_LINUX_TCP_H
#define XDP_CILIUM_L4LB_LINUX_TCP_H


// /root/linux-5.10.142/include/linux/tcp.h


#include <stdbool.h>
#include <linux/types.h>
#include <bpf/bpf_helpers.h>


struct sock_common {
    unsigned char skc_state;
    __u16 skc_num;
} __attribute__((preserve_access_index));

enum sk_pacing {
    SK_PACING_NONE = 0,
    SK_PACING_NEEDED = 1,
    SK_PACING_FQ = 2,
};

struct sock {
    struct sock_common __sk_common;
    unsigned long sk_pacing_rate;
    __u32 sk_pacing_status; /* see enum sk_pacing */
} __attribute__((preserve_access_index));

struct inet_sock {
    struct sock sk;
} __attribute__((preserve_access_index));

struct inet_connection_sock {
    struct inet_sock icsk_inet;
    __u8 icsk_ca_state: 6,
            icsk_ca_setsockopt: 1,
            icsk_ca_dst_locked: 1;
    struct {
        __u8 pending;
    } icsk_ack;
    __u64 icsk_ca_priv[104 / sizeof(__u64)];
} __attribute__((preserve_access_index));

struct tcp_sock {
    struct inet_connection_sock inet_conn;

    __u32 rcv_nxt;
    __u32 snd_nxt;
    __u32 snd_una;
    __u8 ecn_flags;
    __u32 delivered;
    __u32 delivered_ce;
    __u32 snd_cwnd;
    __u32 snd_cwnd_cnt;
    __u32 snd_cwnd_clamp;
    __u32 snd_ssthresh;
    __u8 syn_data: 1,    /* SYN includes data */
    syn_fastopen: 1,    /* SYN includes Fast Open option */
    syn_fastopen_exp: 1,/* SYN includes Fast Open exp. option */
    syn_fastopen_ch: 1, /* Active TFO re-enabling probe */
    syn_data_acked: 1,/* data in SYN is acked by SYN-ACK */
    save_syn: 1,    /* Save headers of SYN packet */
    is_cwnd_limited: 1,/* forward progress limited by snd_cwnd? */
    syn_smc: 1;    /* SYN includes SMC */
    __u32 max_packets_out;
    __u32 lsndtime; /* timestamp of last sent data packet (for restart window) */
    __u32 prior_cwnd;
    __u64 tcp_mstamp;    /* most recent packet received/sent */
} __attribute__((preserve_access_index));


#endif //XDP_CILIUM_L4LB_LINUX_TCP_H
