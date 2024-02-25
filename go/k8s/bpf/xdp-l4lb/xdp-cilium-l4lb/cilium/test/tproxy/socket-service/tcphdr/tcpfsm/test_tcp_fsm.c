

/**
 * /root/linux-5.10.142/tools/testing/selftests/bpf/progs/test_tcpbpf_kern.c
 * /root/linux-5.10.142/tools/testing/selftests/bpf/test_tcpbpf_user.c
 */

#include <stddef.h>
#include <string.h>

#include <linux/bpf.h>
#include <linux/types.h>
#include <linux/in.h>
#include <linux/ip.h>
#include <linux/tcp.h>
//#include <linux/socket.h> // /root/linux-5.10.142/include/linux/socket.h 不起作用
#include <sys/socket.h>

#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

//#define AF_INET 2 // #include <sys/socket.h> 包含
#define SOL_IP        0 // /root/linux-5.10.142/include/linux/socket.h

struct tcpbpf_globals {
    __u32 event_map;
    __u32 total_retrans;
    __u32 data_segs_in;
    __u32 data_segs_out;
    __u32 bad_cb_test_rv;
    __u32 good_cb_test_rv;
    __u64 bytes_received;
    __u64 bytes_acked;
    __u32 num_listen;
    __u32 num_close_events;
};

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, 4);
    __type(key, __u32);
    __type(value, struct tcpbpf_globals);
} global_map SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, 2);
    __type(key, __u32);
    __type(value, int);
} sockopt_results SEC(".maps");

static inline void update_event_map(int event) {
    __u32 key = 0;
    struct tcpbpf_globals g, *gp;

    gp = bpf_map_lookup_elem(&global_map, &key);
    if (gp == NULL) {
        struct tcpbpf_globals g = {0};
        g.event_map |= (1 << event);
        bpf_map_update_elem(&global_map, &key, &g, BPF_ANY);
    } else {
        g = *gp;
        g.event_map |= (1 << event);
        bpf_map_update_elem(&global_map, &key, &g, BPF_ANY);
    }
}

SEC("sockops")
int bpf_tcp_fsm(struct bpf_sock_ops *skops) {
    int rv = -1;
    int v = 0;
    int save_syn = 1;
    int good_call_rv = 0;
    int bad_call_rv = 0;
    char header[sizeof(struct iphdr) + sizeof(struct tcphdr)];
    struct tcphdr *thdr;

    __u32 op = skops->op;
    update_event_map((int) op);
    switch (op) {
        /* Called when TCP changes state. Arg1: old_state Arg2: new_state */
        case BPF_SOCK_OPS_STATE_CB:
            // 表示关闭 tcp connection
            if (skops->args[1] == BPF_TCP_CLOSE) {
                // INFO: 这里是不是可以不用两次 bpf_map_update_elem()
                __u32 key = 0;
                struct tcpbpf_globals g, *gp;
                gp = bpf_map_lookup_elem(&global_map, &key);
                if (!gp)
                    break;
                g = *gp;
                if (skops->args[0] == BPF_TCP_LISTEN) { // listen->close
                    g.num_listen++;
                } else {
                    g.total_retrans = skops->total_retrans;
                    g.data_segs_in = skops->data_segs_in;
                    g.data_segs_out = skops->data_segs_out;
                    g.bytes_received = skops->bytes_received;
                    g.bytes_acked = skops->bytes_acked;
                }
                g.num_close_events++;
                bpf_map_update_elem(&global_map, &key, &g, BPF_ANY);
            }
            break;

            // 调用 listen() 切换到 listen 状态时调用
        case BPF_SOCK_OPS_TCP_LISTEN_CB: {
            bpf_sock_ops_cb_flags_set(skops, BPF_SOCK_OPS_STATE_CB_FLAG);
            v = (int) bpf_setsockopt(skops, IPPROTO_TCP, TCP_SAVE_SYN, &save_syn, sizeof(save_syn));
            /* Update global map result of setsock opt */
            __u32 key = 0;
            bpf_map_update_elem(&sockopt_results, &key, &v, BPF_ANY);
        }
            break;

            // {} 是为了区分 local 变量 __u32 key
            // active connection，就是 client 主动建立 tcp connection
        case BPF_SOCK_OPS_ACTIVE_ESTABLISHED_CB: {
            /* Test failure to set largest cb flag (assumes not defined) */
            bad_call_rv = (int) bpf_sock_ops_cb_flags_set(skops, 0x80); // 0x80=hex(1<<7)
            /* Set callback */
            good_call_rv = (int) bpf_sock_ops_cb_flags_set(skops, BPF_SOCK_OPS_STATE_CB_FLAG);
            __u32 key = 0;
            /* Update results */
            struct tcpbpf_globals g, *gp;
            gp = bpf_map_lookup_elem(&global_map, &key);
            if (!gp)
                break;
            g = *gp;
            g.bad_cb_test_rv = bad_call_rv;
            g.good_cb_test_rv = good_call_rv;
            bpf_map_update_elem(&global_map, &key, &g, BPF_ANY);
        }
            break;

            // passive connection, 就是 server 被动建立 tcp connection
        case BPF_SOCK_OPS_PASSIVE_ESTABLISHED_CB:
            skops->sk_txhash = 0x12345f;
            v = 0xff;
            rv = (int) bpf_setsockopt(skops, SOL_IP, IP_TOS, &v, sizeof(v));
            if (skops->family == AF_INET) {
                v = (int) bpf_getsockopt(skops, IPPROTO_TCP, TCP_SAVED_SYN, header,
                                         (sizeof(struct iphdr) + sizeof(struct tcphdr))); // IPPROTO_TCP=SOL_TCP
                if (!v) {
                    int offset = sizeof(struct iphdr);
                    thdr = (struct tcphdr *) (header + offset);
                    v = thdr->syn;
                    __u32 key = 1;
                    bpf_map_update_elem(&sockopt_results, &key, &v, BPF_ANY);
                }
            }
            break;

        case BPF_SOCK_OPS_RTO_CB:
        case BPF_SOCK_OPS_RETRANS_CB:
            break;
        default:
            rv = -1;
    }

    skops->reply = rv;
    return 1;
}


char _license[] SEC("license") = "GPL";
