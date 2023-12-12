
//go:build ignore


// /root/linux-5.10.142/include/uapi/linux/bpf.h
#include <linux/bpf.h>
// /root/linux-5.10.142/include/uapi/linux/types.h
#include <linux/types.h>
// /root/linux-5.10.142/tools/lib/bpf/bpf_tracing.h, 定义 BPF_PROG()
#include <bpf/bpf_tracing.h>
// /root/linux-5.10.142/tools/lib/bpf/bpf_endian.h
#include <bpf/bpf_endian.h>
// /root/linux-5.10.142/tools/lib/bpf/bpf_helpers.h
#include <bpf/bpf_helpers.h>

#define AF_INET 2
#define SOCKOPS_MAP_SIZE 65535
#define NULL    ((void *)0)

enum {
    SOCK_TYPE_ACTIVE  = 0,
    SOCK_TYPE_PASSIVE = 1,
};

struct sk_key {
    __u32 local_ip4;
    __u32 remote_ip4;
    __u32 local_port;
    __u32 remote_port;
};

struct sk_info {
    struct sk_key sk_key;
    __u8 sk_type;
};

struct rtt_event {
    __u16 sport;
    __u16 dport;
    __u32 saddr;
    __u32 daddr;
    __u32 srtt;
};
struct rtt_event *unused_event __attribute__((unused));

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, SOCKOPS_MAP_SIZE);
    __type(key, struct sk_key);
    __type(value, struct sk_info);
} map_estab_sk SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1 << 24);
} rtt_events SEC(".maps");

static inline void init_sk_key(struct bpf_sock_ops *skops, struct sk_key *sk_key) {
    sk_key->local_ip4   = bpf_ntohl(skops->local_ip4);
    sk_key->remote_ip4  = bpf_ntohl(skops->remote_ip4);
    sk_key->local_port  = skops->local_port;
    sk_key->remote_port = bpf_ntohl(skops->remote_port);
}

static inline void bpf_sock_ops_establish_cb(struct bpf_sock_ops *skops, __u8 sock_type) {
    int err;
    struct sk_info sk_info = {};
    // Only process IPv4 sockets
    if (skops == NULL || skops->family != AF_INET)
        return;

    // Initialize the 4-tuple key
    init_sk_key(skops, &sk_info.sk_key);
    sk_info.sk_type = sock_type;

    // Store the socket info in map using the 4-tuple as key
    // We keep track of TCP connections in 'established' state
    err = bpf_map_update_elem(&map_estab_sk, &sk_info.sk_key, &sk_info, BPF_NOEXIST); // 这里选择 BPF_NOEXIST
    if (err != 0) {
        // Storing the 4-tuple in map has failed, return early.
        // This can happen in case the 4-tuple already exists in the map (i.e. BPF_NOEXIST flag)
        return;
    }

    // Enable sockops callbacks for RTT and TCP state change
    bpf_sock_ops_cb_flags_set(skops, BPF_SOCK_OPS_RTT_CB_FLAG | BPF_SOCK_OPS_STATE_CB_FLAG);
}

static inline void bpf_sock_ops_state_cb(struct bpf_sock_ops *skops) {
    struct sk_key sk_key = {};

    // Socket changed state. args[0] stores the previous state.
    // Perform cleanup of map entry if socket is exiting
    // the 'established' state,
    if (skops->args[0] == BPF_TCP_ESTABLISHED) {
        init_sk_key(skops, &sk_key);
        bpf_map_delete_elem(&map_estab_sk, &sk_key);
    }
}

static inline void bpf_sock_ops_rtt_cb(struct bpf_sock_ops *skops) {
    struct sk_key sk_key = {};
    struct sk_info *sk_info;
    struct rtt_event *rtt_event;

    // Initialize the 4-tuple key
    init_sk_key(skops, &sk_key);

    // Retrieve the socket info from map of established connections
    sk_info = bpf_map_lookup_elem(&map_estab_sk, &sk_key);
    if (!sk_info)
        return;

    rtt_event = bpf_ringbuf_reserve(&rtt_events, sizeof(struct rtt_event), 0);
    if (!rtt_event) {
        return;
    }

    switch (sk_info->sk_type) {
        case SOCK_TYPE_ACTIVE:
            // If socket is 'active', 'local' means 'source'
            // and 'remote' means 'destination'
            rtt_event->saddr = sk_info->sk_key.local_ip4;
            rtt_event->daddr = sk_info->sk_key.remote_ip4;
            rtt_event->sport = sk_info->sk_key.local_port;
            rtt_event->dport = sk_info->sk_key.remote_port;
            break;
        case SOCK_TYPE_PASSIVE:
            // If socket is 'passive', 'local' means 'destination'
            // and 'remote' means 'source'
            rtt_event->saddr = sk_info->sk_key.remote_ip4;
            rtt_event->daddr = sk_info->sk_key.local_ip4;
            rtt_event->sport = sk_info->sk_key.remote_port;
            rtt_event->dport = sk_info->sk_key.local_port;
            break;
    }

    // Extract smoothed RTT
    rtt_event->srtt = skops->srtt_us >> 3;
    rtt_event->srtt /= 1000; // ms

    // Send RTT event data to userspace app via ring buffer
    bpf_ringbuf_submit(rtt_event, 0);
}

SEC("sockops")
int bpf_sockops_cb(struct bpf_sock_ops *skops) {
    __u32 op;
    op = skops->op;

    switch (op) {
        case BPF_SOCK_OPS_ACTIVE_ESTABLISHED_CB:
            // client<-server, 回包
            bpf_sock_ops_establish_cb(skops, SOCK_TYPE_ACTIVE);
            break;
        case BPF_SOCK_OPS_PASSIVE_ESTABLISHED_CB:
            // client->server
            bpf_sock_ops_establish_cb(skops, SOCK_TYPE_PASSIVE);
            break;
        case BPF_SOCK_OPS_RTT_CB:
            bpf_sock_ops_rtt_cb(skops);
            break;
        case BPF_SOCK_OPS_STATE_CB:
            bpf_sock_ops_state_cb(skops);
            break;
    }

    return 0;
}

char _license[] SEC("license") = "GPL";
