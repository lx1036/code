
#include <errno.h>
#include <stdbool.h>

#include <linux/if_ether.h>
#include <linux/in.h>
#include <linux/ip.h>
#include <linux/tcp.h>
#include <linux/bpf.h>
#include <sys/socket.h>

// 顺序 <bpf/xxx> 在 <linux/xxx> 后面
#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>


#include "linux_tcp.h"

static volatile const __u16 g_serv_port = 0;


static inline void set_tuple(struct bpf_sock_tuple *tuple, const struct iphdr *iphdr, const struct tcphdr *tcph) {
    tuple->ipv4.saddr = iphdr->saddr;
    tuple->ipv4.daddr = iphdr->daddr;
    tuple->ipv4.sport = tcph->source;
    tuple->ipv4.dport = tcph->dest;
}

static inline int is_allowed_peer_cg(struct __sk_buff *skb, const struct iphdr *iphdr, const struct tcphdr *tcph) {
    __u64 cgid, acgid, peer_cgid, peer_acgid;
    struct bpf_sock_tuple tuple;
    size_t tuple_len = sizeof(tuple.ipv4);
    struct bpf_sock *peer_sk;

    set_tuple(&tuple, iphdr, tcph);
    // 根据五元组 lookup established connection socket
    peer_sk = bpf_sk_lookup_tcp(skb, &tuple, tuple_len, BPF_F_CURRENT_NETNS, 0);
    if (!peer_sk)
        return 0;

    cgid = bpf_skb_cgroup_id(skb);
    peer_cgid = bpf_sk_cgroup_id(peer_sk);
    acgid = bpf_skb_ancestor_cgroup_id(skb, 2);
    peer_acgid = bpf_sk_ancestor_cgroup_id(peer_sk, 2);

    bpf_sk_release(peer_sk);

    return cgid && cgid == peer_cgid && acgid && acgid == peer_acgid;
}

// /root/linux-5.10.142/tools/testing/selftests/bpf/progs/cgroup_skb_sk_lookup_kern.c
SEC("cgroup_skb/ingress")
int ingress_lookup(struct __sk_buff *skb) {
    __u32 serv_port_key = 0;
    struct iphdr iphdr;
    struct tcphdr tcph;

    if (skb->protocol != bpf_htons(ETH_P_IP)) // only ipv4
        return 1;

    /** For SYN packets coming to listening socket skb->remote_port will be
     * zero, so IPv4/TCP headers are loaded to identify remote peer
     * instead.
     */
    if (bpf_skb_load_bytes(skb, 0, &iphdr, sizeof(iphdr)))
        return 1;

    if (iphdr.protocol != IPPROTO_TCP)
        return 1;

    if (bpf_skb_load_bytes(skb, sizeof(iphdr), &tcph, sizeof(tcph)))
        return 1;

    if (!g_serv_port)
        return 0;

    if (tcph.dest != g_serv_port)
        return 1;

    return is_allowed_peer_cg(skb, &iphdr, &tcph);
}


struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, 1);
    __type(key, __u32);
    __type(value, __u32);
} test_result SEC(".maps");

// /root/linux-5.10.142/tools/testing/selftests/bpf/progs/load_bytes_relative.c
// /root/linux-5.10.142/tools/testing/selftests/bpf/prog_tests/load_bytes_relative.c
SEC("cgroup_skb/egress")
int load_bytes_relative(struct __sk_buff *skb) {
    struct ethhdr eth;
    struct iphdr iph;

    __u32 map_key = 0;
    __u32 test_passed = 0;

    /* MAC header is not set by the time cgroup_skb/egress triggers */
    if (bpf_skb_load_bytes_relative(skb, 0, &eth, sizeof(eth), BPF_HDR_START_MAC) != -EFAULT)
        goto fail;

    if (bpf_skb_load_bytes_relative(skb, 0, &iph, sizeof(iph), BPF_HDR_START_NET))
        goto fail;

    if (bpf_skb_load_bytes_relative(skb, 0xffff, &iph, sizeof(iph), BPF_HDR_START_NET) != -EFAULT) // 0xffff???
        goto fail;

//    if (bpf_skb_load_bytes_relative(skb, 0xffffffff, &iph, sizeof(iph), BPF_HDR_START_NET)) // 0xffff???
//        goto fail;

    test_passed = 1;

    fail:
    bpf_map_update_elem(&test_result, &map_key, &test_passed, BPF_ANY);

    return 1;
}


// /root/linux-5.10.142/tools/testing/selftests/bpf/progs/test_sock_fields.c
// /root/linux-5.10.142/tools/testing/selftests/bpf/prog_tests/sock_fields.c

/* Always return CG_OK so that no pkt will be filtered out */
#define CG_OK 1

static volatile const struct sockaddr_in srv_sa = {};
struct bpf_sock listen_sk = {};
struct bpf_sock srv_sk = {};
struct bpf_sock cli_sk = {};
struct bpf_tcp_sock cli_tp = {};
struct bpf_tcp_sock srv_tp = {};
struct bpf_tcp_sock listen_tp = {};
__u64 parent_cg_id = 0;
__u64 child_cg_id = 0;
__u64 lsndtime = 0;

enum bpf_linum_array_idx {
    EGRESS_LINUM_IDX,
    INGRESS_LINUM_IDX,
    READ_SK_DST_PORT_LINUM_IDX,
    __NR_BPF_LINUM_ARRAY_IDX,
};

struct bpf_spinlock_cnt {
    struct bpf_spin_lock lock;
    __u32 cnt;
};

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, __NR_BPF_LINUM_ARRAY_IDX);
    __type(key, __u32);
    __type(value, __u32);
} linum_map SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_SK_STORAGE);
    __uint(map_flags, BPF_F_NO_PREALLOC);
    __type(key, int);
    __type(value, struct bpf_spinlock_cnt);
} sk_pkt_out_cnt SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_SK_STORAGE);
    __uint(map_flags, BPF_F_NO_PREALLOC);
    __type(key, int);
    __type(value, struct bpf_spinlock_cnt);
} sk_pkt_out_cnt10 SEC(".maps");

static int ret_log(__u32 linum_idx) {
    __u32 linum = __LINE__;
    bpf_map_update_elem(&linum_map, &linum_idx, &linum, BPF_ANY);
    return CG_OK;
}

static bool is_loopback(__u32 ip) {
    return ip == bpf_htonl(0x7f000001);
}

static void skcpy(struct bpf_sock *dst, const struct bpf_sock *src)
{
    dst->bound_dev_if = src->bound_dev_if;
    dst->family = src->family;
    dst->type = src->type;
    dst->protocol = src->protocol;
    dst->mark = src->mark;
    dst->priority = src->priority;
    dst->src_ip4 = src->src_ip4;
    dst->src_port = src->src_port;
    dst->dst_ip4 = src->dst_ip4;
    dst->dst_port = src->dst_port;
    dst->state = src->state;
}

static void tpcpy(struct bpf_tcp_sock *dst, const struct bpf_tcp_sock *src)
{
    dst->snd_cwnd = src->snd_cwnd;
    dst->srtt_us = src->srtt_us;
    dst->rtt_min = src->rtt_min;
    dst->snd_ssthresh = src->snd_ssthresh;
    dst->rcv_nxt = src->rcv_nxt;
    dst->snd_nxt = src->snd_nxt;
    dst->snd_una = src->snd_una;
    dst->mss_cache = src->mss_cache;
    dst->ecn_flags = src->ecn_flags;
    dst->rate_delivered = src->rate_delivered;
    dst->rate_interval_us = src->rate_interval_us;
    dst->packets_out = src->packets_out;
    dst->retrans_out = src->retrans_out;
    dst->total_retrans = src->total_retrans;
    dst->segs_in = src->segs_in;
    dst->data_segs_in = src->data_segs_in;
    dst->segs_out = src->segs_out;
    dst->data_segs_out = src->data_segs_out;
    dst->lost_out = src->lost_out;
    dst->sacked_out = src->sacked_out;
    dst->bytes_received = src->bytes_received;
    dst->bytes_acked = src->bytes_acked;
}

static __noinline bool sk_dst_port__load_word(struct bpf_sock *sk)
{
    __u32 *word = (__u32 *)&sk->dst_port;
    return word[0] == bpf_htonl(0xcafe0000);
}

static __noinline bool sk_dst_port__load_half(struct bpf_sock *sk)
{
    __u16 *half = (__u16 *)&sk->dst_port;
    return half[0] == bpf_htons(0xcafe);
}

static __noinline bool sk_dst_port__load_byte(struct bpf_sock *sk)
{
    __u8 *byte = (__u8 *)&sk->dst_port;
    return byte[0] == 0xca && byte[1] == 0xfe;
}

SEC("cgroup_skb/ingress")
int ingress_read_sock_fields(struct __sk_buff *skb) {
    struct bpf_tcp_sock *tp;
    __u32 linum, linum_idx;
    struct bpf_sock *sk;

    linum_idx = INGRESS_LINUM_IDX;

    sk = skb->sk;
    if (!sk)
        return ret_log(linum_idx);

    /* Not the testing ingress traffic to the server */
    if (sk->family != AF_INET || !is_loopback(sk->src_ip4) ||
        sk->src_port != bpf_ntohs(srv_sa.sin_port))
        return CG_OK;

    /* Only interested in TCP_LISTEN */
    if (sk->state != 10)
        return CG_OK;

    /* It must be a fullsock for cgroup_skb/ingress prog */
    sk = bpf_sk_fullsock(sk);
    if (!sk)
        return ret_log(linum_idx);

    tp = bpf_tcp_sock(sk);
    if (!tp)
        return ret_log(linum_idx);

    skcpy(&listen_sk, sk);
    tpcpy(&listen_tp, tp);

    return CG_OK;
}

SEC("cgroup_skb/egress")
int read_sk_dst_port(struct __sk_buff *skb) {
    __u32 linum, linum_idx;
    struct bpf_sock *sk;

    linum_idx = READ_SK_DST_PORT_LINUM_IDX;
    sk = skb->sk;
    if (!sk)
        return ret_log(linum_idx);

    /* Ignore everything but the SYN from the client socket */
    if (sk->state != BPF_TCP_SYN_SENT)
        return CG_OK;

    if (!sk_dst_port__load_word(sk))
        return ret_log(linum_idx);
    if (!sk_dst_port__load_half(sk))
        return ret_log(linum_idx);
    if (!sk_dst_port__load_byte(sk))
        return ret_log(linum_idx);

    return CG_OK;
}

SEC("cgroup_skb/egress")
int egress_read_sock_fields(struct __sk_buff *skb)
{
    struct bpf_spinlock_cnt cli_cnt_init = { .lock = 0, .cnt = 0xeB9F };
    struct bpf_spinlock_cnt *pkt_out_cnt, *pkt_out_cnt10;
    struct bpf_tcp_sock *tp, *tp_ret;
    struct bpf_sock *sk, *sk_ret;
    __u32 linum, linum_idx;
    struct tcp_sock *ktp;

    linum_idx = EGRESS_LINUM_IDX;

    sk = skb->sk;
    if (!sk)
        return ret_log(linum_idx);

    /* Not the testing egress traffic or
     * TCP_LISTEN (10) socket will be copied at the ingress side.
     */
    if (sk->family != AF_INET || !is_loopback(sk->src_ip4) || sk->state == BPF_TCP_LISTEN)
        return CG_OK;

    if (sk->src_port == bpf_ntohs(srv_sa.sin_port)) {
        /* Server socket */
        sk_ret = &srv_sk;
        tp_ret = &srv_tp;
    } else if (sk->dst_port == bpf_ntohs(srv_sa.sin_port)) {
        /* Client socket */
        sk_ret = &cli_sk;
        tp_ret = &cli_tp;
    } else {
        /* Not the testing egress traffic */
        return CG_OK;
    }

    /* It must be a fullsock for cgroup_skb/egress prog */
    sk = bpf_sk_fullsock(sk);
    if (!sk)
        return ret_log(linum_idx);

    /* Not the testing egress traffic */
    if (sk->protocol != IPPROTO_TCP)
        return CG_OK;

    tp = bpf_tcp_sock(sk);
    if (!tp)
        return ret_log(linum_idx);

    skcpy(sk_ret, sk);
    tpcpy(tp_ret, tp);

    if (sk_ret == &srv_sk) {
        ktp = bpf_skc_to_tcp_sock(sk);
        if (!ktp)
            return ret_log(linum_idx);

        lsndtime = ktp->lsndtime;

        child_cg_id = bpf_sk_cgroup_id(ktp);
        if (!child_cg_id)
            return ret_log(linum_idx);

        parent_cg_id = bpf_sk_ancestor_cgroup_id(ktp, 2);
        if (!parent_cg_id)
            return ret_log(linum_idx);

        /* The userspace has created it for srv sk */
        pkt_out_cnt = bpf_sk_storage_get(&sk_pkt_out_cnt, ktp, 0, 0);
        pkt_out_cnt10 = bpf_sk_storage_get(&sk_pkt_out_cnt10, ktp, 0, 0);
    } else {
        pkt_out_cnt = bpf_sk_storage_get(&sk_pkt_out_cnt, sk, &cli_cnt_init, BPF_SK_STORAGE_GET_F_CREATE);
        pkt_out_cnt10 = bpf_sk_storage_get(&sk_pkt_out_cnt10, sk, &cli_cnt_init, BPF_SK_STORAGE_GET_F_CREATE);
    }

    if (!pkt_out_cnt || !pkt_out_cnt10)
        return ret_log(linum_idx);

    /* Even both cnt and cnt10 have lock defined in their BTF,
     * intentionally one cnt takes lock while one does not
     * as a test for the spinlock support in BPF_MAP_TYPE_SK_STORAGE.
     */
    pkt_out_cnt->cnt += 1;
    bpf_spin_lock(&pkt_out_cnt10->lock);
    pkt_out_cnt10->cnt += 10;
    bpf_spin_unlock(&pkt_out_cnt10->lock);

    return CG_OK;
}


char _license[] SEC("license") = "GPL";
