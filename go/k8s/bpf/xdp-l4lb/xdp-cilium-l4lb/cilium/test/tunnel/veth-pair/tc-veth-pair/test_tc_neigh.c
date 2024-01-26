
#include <stddef.h>
#include <stdint.h>
#include <stdbool.h>

#include <linux/bpf.h>
#include <linux/stddef.h>
#include <linux/pkt_cls.h>
#include <linux/if_ether.h>
#include <linux/in.h>
#include <linux/ip.h>
#include <linux/ipv6.h>

#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

#define ip4_src 0xad100164 /* 173.16.1.100 */
#define ip4_dst 0xad100264 /* 173.16.2.100 */

enum {
    dev_src,
    dev_dst,
};

struct bpf_map_def SEC("maps") ifindex_map = {
    .type        = BPF_MAP_TYPE_ARRAY,
    .key_size    = sizeof(int),
    .value_size    = sizeof(int),
    .max_entries    = 2,
};

static __always_inline bool is_remote_ep_v4(struct __sk_buff *skb, __be32 addr)
{
    void *data_end = (void *)(long)(skb->data_end);
    void *data = (void *)(long)(skb->data);
    struct iphdr *ip4h;

    if (data + sizeof(struct ethhdr) > data_end)
        return false;

    ip4h = (struct iphdr *)(data + sizeof(struct ethhdr));
    if ((void *)(ip4h + 1) > data_end)
        return false;

    bpf_printk("expect %x, actual %x, result %d", addr, ip4h->daddr, ip4h->daddr == addr);
    return ip4h->daddr == addr;
}

static __always_inline int get_dev_ifindex(int which)
{
    int *ifindex = bpf_map_lookup_elem(&ifindex_map, &which);
    return ifindex ? *ifindex : 0;
}

SEC("chk_egress")
int tc_chk(struct __sk_buff *skb)
{
    void *data_end = (void *)(long)(skb->data_end);
    void *data = (void *)(long)(skb->data);
    __u32 *raw = data;

    if (data + sizeof(struct ethhdr) > data_end)
        return TC_ACT_SHOT;

    return !raw[0] && !raw[1] && !raw[2] ? TC_ACT_SHOT : TC_ACT_OK; // ???
}

SEC("dst_ingress")
int tc_dst(struct __sk_buff *skb)
{
    __u8 zero[ETH_ALEN * 2];
    bool redirect = false;

//    switch (skb->protocol) {
//    case __bpf_constant_htons(ETH_P_IP):
//        redirect = is_remote_ep_v4(skb, __bpf_constant_htonl(ip4_src)); // 173.16.2.100->173.16.1.100
//        bpf_printk("[dst_ingress]redirect: %d", redirect);
//        break;
//    default:
//        break;
//    }

    if (skb->protocol != __bpf_constant_htons(ETH_P_IP)) { // arp
        return TC_ACT_OK;
    }

    if (skb->protocol == __bpf_constant_htons(ETH_P_IP)) {
        redirect = is_remote_ep_v4(skb, __bpf_constant_htonl(ip4_src)); // 173.16.2.100->173.16.1.100
        bpf_printk("[dst_ingress]redirect: %d", redirect);
    }

    bpf_printk("[dst_ingress]redirect: %d", redirect);

    if (!redirect)
        bpf_printk("[dst_ingress]can not redirect to %x, redirect: %d, redirect2: %d", ip4_src, !redirect, redirect);
        return TC_ACT_OK;

    __builtin_memset(&zero, 0, sizeof(zero));
    if (bpf_skb_store_bytes(skb, 0, &zero, sizeof(zero), 0) < 0)
        return TC_ACT_SHOT;

    int if_index = get_dev_ifindex(dev_src);
    bpf_printk("[dst_ingress]bpf_redirect_neigh if_index: %d", if_index);
    return bpf_redirect_neigh(if_index, NULL, 0, 0);
}

SEC("src_ingress")
int tc_src(struct __sk_buff *skb)
{
    __u8 zero[ETH_ALEN * 2]; // *2 是表示 src_mac 和 dst_mac
    bool redirect = false;

//    switch (skb->protocol) {
//    case __bpf_constant_htons(ETH_P_IP):
//        redirect = is_remote_ep_v4(skb, __bpf_constant_htonl(ip4_dst)); // 173.16.1.100->173.16.2.100
//        bpf_printk("[src_ingress]redirect: %d", redirect);
//        break;
//    default:
//        break;
//    }

    if (skb->protocol != __bpf_constant_htons(ETH_P_IP)) { // arp
        return TC_ACT_OK;
    }

    if (skb->protocol == __bpf_constant_htons(ETH_P_IP)) {
        redirect = is_remote_ep_v4(skb, __bpf_constant_htonl(ip4_dst)); // 173.16.1.100->173.16.2.100
        bpf_printk("[src_ingress]redirect: %d", redirect);
    }

    bpf_printk("[src_ingress]redirect: %d", redirect);

    if (!redirect)
        bpf_printk("[src_ingress]can not redirect to %x, redirect: %d, redirect2: %d", ip4_dst, !redirect, redirect);
        return TC_ACT_OK;

    __builtin_memset(&zero, 0, sizeof(zero));
    if (bpf_skb_store_bytes(skb, 0, &zero, sizeof(zero), 0) < 0) // 表示把 skb 的 src_mac 和 dst_mac 置空
        return TC_ACT_SHOT;

    // 填充 src_mac=mac(veth_src_fwd),dst_mac=mac(veth_dst_fwd)
    int if_index = get_dev_ifindex(dev_dst);
    bpf_printk("[src_ingress]bpf_redirect_neigh if_index: %d", if_index);
    return bpf_redirect_neigh(if_index, NULL, 0, 0);
}


char __license[] SEC("license") = "GPL";
