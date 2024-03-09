
/**
 * https://github.com/y805939188/simple-k8s-cni/blob/master/plugins/vxlan/ebpf/vxlan_ingress.c
 */


// /root/linux-5.10.142/include/uapi/linux/bpf.h
#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/if_packet.h>
#include <linux/ip.h>
#include <linux/pkt_cls.h>

// /root/linux-5.10.142/tools/lib/bpf/bpf_helpers.h
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

#ifndef __section
#define __section(X) __attribute__((section(X), used))
#endif
#ifndef __section_maps
#define __section_maps __section("maps")
#endif

#define PIN_GLOBAL_NS   2
#define DEFAULT_TUNNEL_ID 13190
#define LOCAL_DEV_VXLAN 1;
#define LOCAL_DEV_VETH 2;

struct bpf_elf_map {
    __u32 type;
    __u32 size_key;
    __u32 size_value;
    __u32 max_elem;
    __u32 flags;
    __u32 id;
    __u32 pinning;
    __u32 inner_id;
    __u32 inner_idx;
};

struct endpointKey {
    __u32 ip;
};
struct endpointInfo {
    __u32 ifIndex;
    __u32 lxcIfIndex;
    __u8 mac[8];
    __u8 nodeMac[8];
};

struct bpf_elf_map __section_maps ding_lxc = {
    .type		= BPF_MAP_TYPE_HASH,
    .size_key	= sizeof(struct endpointKey),
    .size_value	= sizeof(struct endpointInfo),
    .pinning	= PIN_GLOBAL_NS, // 设置 LIBBPF_PIN_BY_NAME 导致 pin path 为 /sys/fs/bpf/tc/424e22c8e74276a6484f398886d426f441d9b849/
    .max_elem	= 256,
};

struct podNodeKey {
    __u32 ip;
};

struct podNodeValue {
    __u32 ip;
};

struct bpf_elf_map __section_maps ding_ip = {
    .type		= BPF_MAP_TYPE_HASH,
    .size_key	= sizeof(struct podNodeKey),
    .size_value	= sizeof(struct podNodeValue),
    .pinning	= PIN_GLOBAL_NS, // 设置 LIBBPF_PIN_BY_NAME 导致 pin path 为 /sys/fs/bpf/tc/424e22c8e74276a6484f398886d426f441d9b849/
    .max_elem	= 256,
};

struct localNodeMapKey {
    __u32 type;
};
struct localNodeMapValue {
    __u32 ifIndex;
};
struct bpf_elf_map __section_maps ding_local = {
        .type		= BPF_MAP_TYPE_HASH,
        .size_key	= sizeof(struct localNodeMapKey),
        .size_value	= sizeof(struct localNodeMapValue),
        .pinning	= PIN_GLOBAL_NS, // 设置 LIBBPF_PIN_BY_NAME 导致 pin path 为 /sys/fs/bpf/tc/424e22c8e74276a6484f398886d426f441d9b849/
        .max_elem	= 256,
};

SEC("veth_pair_ingress")
int veth_pair_ingress(struct __sk_buff *skb) {
    void *data = (void *)(long)skb->data;
    void *data_end = (void *)(long)skb->data_end;
    if (data + sizeof(struct ethhdr) + sizeof(struct iphdr) > data_end) {
        return TC_ACT_UNSPEC;
    }

    struct ethhdr  *eth  = data;
    struct iphdr   *ip   = (data + sizeof(struct ethhdr));
    if (eth->h_proto != bpf_htons(ETH_P_IP)) {
        return TC_ACT_UNSPEC;
    }

    // 在 go 那头儿往 ebpf 的 map 里存的时候我这个 arm 是按照小端序存的
    // 这里给转成网络的大端序
    __u32 src_ip = bpf_htonl(ip->saddr);
    __u32 dst_ip = bpf_htonl(ip->daddr);
    // 拿到 mac 地址
    __u8 src_mac[ETH_ALEN];
    __u8 dst_mac[ETH_ALEN];
    struct endpointKey epKey = {};
    epKey.ip = dst_ip;
    // 在 lxc 中查找
    struct endpointInfo *ep = bpf_map_lookup_elem(&ding_lxc, &epKey);
    if (ep) {
        // 如果能找到说明是要发往本机其他 pod 中的
        // 把 mac 地址改成目标 pod 的两对儿 veth 的 mac 地址
        __builtin_memcpy(src_mac, ep->nodeMac, ETH_ALEN);
        __builtin_memcpy(dst_mac, ep->mac, ETH_ALEN);
        bpf_skb_store_bytes(skb, offsetof(struct ethhdr, h_source), dst_mac, ETH_ALEN, 0);
        bpf_skb_store_bytes(skb, offsetof(struct ethhdr, h_dest), src_mac, ETH_ALEN, 0);
        return bpf_redirect_peer(ep->lxcIfIndex, 0);
    }

    struct podNodeKey podNodeKey = {
        .ip = dst_ip,
    };
    struct podNodeValue *podNode = bpf_map_lookup_elem(&ding_ip, &podNodeKey);
    if (podNode) {
        // 进到这里说明该目标 ip 是本集群内的 ip
        struct localNodeMapKey localKey = {};
        localKey.type = LOCAL_DEV_VXLAN;
        struct localNodeMapValue *localValue = bpf_map_lookup_elem(&ding_local, &localKey);
        if (localValue) {
            // redirect 到 vxlan egress
            return bpf_redirect(localValue->ifIndex, 0);
        }
        return TC_ACT_UNSPEC;
    }

    return TC_ACT_UNSPEC;
}


SEC("vxlan_ingress")
int vxlan_ingress(struct __sk_buff *skb) {
    void *data = (void *)(long)skb->data;
    void *data_end = (void *)(long)skb->data_end;
    if (data + sizeof(struct ethhdr) + sizeof(struct iphdr) > data_end) {
        return TC_ACT_UNSPEC;
    }

    struct ethhdr *eth = data;
    struct iphdr *ip = (data + sizeof(struct ethhdr));
    if (eth->h_proto != bpf_htons(ETH_P_IP)) {
        return TC_ACT_UNSPEC;
    }

    __u32 src_ip = bpf_htonl(ip->saddr);
    __u32 dst_ip = bpf_htonl(ip->daddr);
    bpf_printk("the dst_ip is: %d", dst_ip);
    bpf_printk("the ip->daddr is: %d", ip->daddr);

    struct endpointKey epKey = {
            .ip = dst_ip,
    };
    struct endpointInfo *ep = bpf_map_lookup_elem(&ding_lxc, &epKey);
    if (!ep) {
        return TC_ACT_OK;
    }
    // 找到的话说明是发往本机 pod 中的流量
    // 此时需要做 stc mac 和 dst mac 的更新

    // 拿到 mac 地址
    __u8 src_mac[ETH_ALEN];
    __u8 dst_mac[ETH_ALEN];
    // 将 mac 改成本机 pod 的那对儿 veth pair 的 mac
    __builtin_memcpy(src_mac, ep->nodeMac, ETH_ALEN);
    __builtin_memcpy(dst_mac, ep->mac, ETH_ALEN);
    // 将 mac 更新到 skb 中
    bpf_skb_store_bytes(skb, offsetof(struct ethhdr, h_dest), dst_mac, ETH_ALEN, 0);
    bpf_skb_store_bytes(skb, offsetof(struct ethhdr, h_source), src_mac, ETH_ALEN, 0);

    return bpf_redirect(ep->lxcIfIndex, 0);
}

SEC("vxlan_egress")
int vxlan_egress(struct __sk_buff *skb) {
    void *data = (void *)(long)skb->data;
    void *data_end = (void *)(long)skb->data_end;
    if (data + sizeof(struct ethhdr) + sizeof(struct iphdr) > data_end) {
        return TC_ACT_UNSPEC;
    }

    struct ethhdr  *eth  = data;
    struct iphdr   *ip   = (data + sizeof(struct ethhdr));
    if (eth->h_proto != bpf_htons(ETH_P_IP)) {
        return TC_ACT_UNSPEC;
    }

    __u32 src_ip = bpf_htonl(ip->saddr);
    __u32 dst_ip = bpf_htonl(ip->daddr);
    bpf_printk("the dst_ip is: %d", dst_ip);
    bpf_printk("the ip->daddr is: %d", ip->daddr);

    // 获取目标 ip 所在的 node ip
    struct podNodeKey podNodeKey = {};
    podNodeKey.ip = dst_ip;
    struct podNodeValue *podNode = bpf_map_lookup_elem(&ding_ip, &podNodeKey);
    if (podNode) {
        __u32 dst_node_ip = podNode->ip;
        // 准备一个 tunnel
        struct bpf_tunnel_key key;
        int ret;
        __builtin_memset(&key, 0x0, sizeof(key));
        key.remote_ipv4 = podNode->ip;
        key.tunnel_id = DEFAULT_TUNNEL_ID;
        key.tunnel_tos = 0;
        key.tunnel_ttl = 64;
        // 添加外头的隧道 udp
        ret = bpf_skb_set_tunnel_key(skb, &key, sizeof(key), BPF_F_ZERO_CSUM_TX);
        if (ret < 0) {
            bpf_printk("bpf_skb_set_tunnel_key failed");
            return TC_ACT_SHOT;
        }
        return TC_ACT_OK;
    }

    return TC_ACT_OK;
}


char _license[] SEC("license") = "GPL";
