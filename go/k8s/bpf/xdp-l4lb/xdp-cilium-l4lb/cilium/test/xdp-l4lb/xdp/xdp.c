

#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/pkt_cls.h>
#include <linux/types.h>
#include <linux/ip.h>
#include <linux/in.h>
#include <linux/tcp.h>
#include <linux/udp.h>

#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>


#define MAX_IPTNL_ENTRIES 256U
#define AF_INET 2

struct vip {
    union {
        __u32 v6[4];
        __u32 v4;
    } daddr;
    __u16 dport;
    __u16 family;
    __u8 protocol;
};
struct vip *unused_vip __attribute__((unused));

struct iptnl_info {
    union {
        __u32 v6[4];
        __u32 v4;
    } saddr;
    union {
        __u32 v6[4];
        __u32 v4;
    } daddr;
    __u16 family;
    __u8 dmac[6];
};
struct iptnl_info *unused_iptnl_info __attribute__((unused));

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, MAX_IPTNL_ENTRIES);
    __type(key, struct vip);
    __type(value, struct iptnl_info);
} vip2tnl SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, 256);
    __type(key, __u32);
    __type(value, __u64);
} rxcnt SEC(".maps");

static __always_inline void count_tx(__u32 protocol)
{
    __u64 *rxcnt_count;
    rxcnt_count = bpf_map_lookup_elem(&rxcnt, &protocol);
    if (rxcnt_count)
        *rxcnt_count += 1;
}

static __always_inline void set_ethhdr(struct ethhdr *new_eth, const struct ethhdr *old_eth,
        const struct iptnl_info *tnl, __be16 h_proto) {
    __builtin_memcpy(new_eth->h_source, old_eth->h_dest, sizeof(new_eth->h_source));
    __builtin_memcpy(new_eth->h_dest, tnl->dmac, sizeof(new_eth->h_dest));
    new_eth->h_proto = h_proto;
}

static __always_inline int get_dport(void *trans_data, void *data_end, __u8 protocol) {
    struct tcphdr *th;
    struct udphdr *uh;
    switch (protocol) {
        case IPPROTO_TCP:
            th = (struct tcphdr *)trans_data;
            if ((void*)th + 1 > data_end)
                return -1;
            return th->dest;
        case IPPROTO_UDP:
            uh = (struct udphdr *)trans_data;
            if ((void*)uh + 1 > data_end)
                return -1;
            return uh->dest;
        default:
            return 0;
    }
}

static __always_inline int handle_ipv4(struct xdp_md *xdp) {
    void *data = (void *)(long)xdp->data;
    void *data_end = (void *)(long)xdp->data_end;
    struct iptnl_info *tnl;
    struct ethhdr *new_eth;
    struct ethhdr *old_eth;
    struct iphdr *iph = data + sizeof(struct ethhdr);
    __u16 *next_iph;
    __u16 payload_len;
    struct vip vip = {};
    int dport;
    __u32 csum = 0;
    int i;

    if ((void*)iph + 1 > data_end)
        return XDP_DROP;

    dport = get_dport(iph + 1, data_end, iph->protocol); // 注意这里的 iphdr+1
    if (dport == -1)
        return XDP_DROP;

    vip.protocol = iph->protocol;
    vip.family = AF_INET;
    vip.daddr.v4 = iph->daddr;
    vip.dport = dport; // 可能为 0
    payload_len = bpf_ntohs(iph->tot_len);
    tnl = bpf_map_lookup_elem(&vip2tnl, &vip);
    /* It only does v4-in-v4 */
    if (!tnl || tnl->family != AF_INET)
        return XDP_PASS;

    // 腾出 iphdr 字节大小空间，下面使用 ipip 协议
    if (bpf_xdp_adjust_head(xdp, 0 - (int)sizeof(struct iphdr)))
        return XDP_DROP;

    data = (void *)(long)xdp->data;
    data_end = (void *)(long)xdp->data_end;
    new_eth = data;
    iph = data + sizeof(*new_eth);
    old_eth = data + sizeof(*iph); // ???, 不太对吧，应该是 l4 hdr
    if ((void*)new_eth + 1 > data_end ||
        (void*)old_eth + 1 > data_end ||
        (void*)iph + 1 > data_end)
        return XDP_DROP;

    set_ethhdr(new_eth, old_eth, tnl, bpf_htons(ETH_P_IP));
    iph->version = 4;
    iph->ihl = sizeof(*iph) >> 2; // ipip 有两个 iphdr, ihl(ip header length) 必须是 5(0101) *4
    iph->frag_off =	0;
    iph->protocol = IPPROTO_IPIP;
    iph->check = 0;
    iph->tos = 0;
    iph->tot_len = bpf_htons(payload_len + sizeof(*iph));
    iph->daddr = tnl->daddr.v4;
    iph->saddr = tnl->saddr.v4;
    iph->ttl = 8;
    next_iph = (__u16 *)iph;

#pragma clang loop unroll(disable)
    for (i = 0; i < sizeof(*iph) >> 1; i++)
        csum += *next_iph++;
    iph->check = ~((csum & 0xffff) + (csum >> 16));

    count_tx(vip.protocol);

    return XDP_TX;
}

SEC("xdp/xdp_tx_iptunnel")
int xdp_tx_iptunnel(struct xdp_md *xdp)
{
    void *data_end = (void *)(long)xdp->data_end;
    void *data = (void *)(long)xdp->data;
    struct ethhdr *eth = data;
    __u16 h_proto;

    if ((void *)eth + 1 > data_end)
        return XDP_DROP;

    h_proto = eth->h_proto;
    if (h_proto == bpf_htons(ETH_P_IP))
        return handle_ipv4(xdp);
    else if (h_proto == bpf_htons(ETH_P_IPV6))
        return XDP_PASS;
//        return handle_ipv6(xdp);
    else
        return XDP_DROP;
}

char _license[] SEC("license") = "GPL";
