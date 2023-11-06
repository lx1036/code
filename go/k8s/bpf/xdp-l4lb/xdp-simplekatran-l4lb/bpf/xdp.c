
// #include <stdbool.h>

#define bool	_Bool
# undef false
# define false 0
# undef true
# define true 1

#include <linux/bpf.h>
#include <linux/in.h>
#include <linux/if_ether.h>
#include <linux/ip.h>

#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>

#define MAX_MAP_ENTRIES 1
#define BE_ETH_P_IP 8
#define BE_ETH_P_IPV6 56710

#define NO_FLAGS 0

#ifndef memcpy
#define memcpy(dest, src, n) __builtin_memcpy((dest), (src), (n))
#endif

struct arguments {
  __u8 dst_mac[6];
  __u32 daddr;
  __u32 saddr;
  __u32 vip;
};

struct {
	__uint(type, BPF_MAP_TYPE_ARRAY);
	__uint(key_size, sizeof(__u32));
	__uint(value_size, sizeof(struct arguments));
	__uint(max_entries, MAX_MAP_ENTRIES);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
    __uint(map_flags, NO_FLAGS);
} xdp_params_array SEC(".maps");

// checksum 函数

__attribute__((__always_inline__)) static inline 
__u16 csum_fold_helper(__u64 csum) {
    int i;
#pragma unroll
    for (i = 0; i < 4; i++) {
        if (csum >> 16) {
            csum = (csum & 0xffff) + (csum >> 16);
        }
    }
    return ~csum;
}

__attribute__((__always_inline__)) static inline 
void ipv4_csum_inline(void *iph, __u64 *csum) {
    __u16 *next_iph_u16 = (__u16 *)iph; // 
#pragma clang loop unroll(full)
    for (int i = 0; i < sizeof(struct iphdr) >> 1; i++) {
        *csum += *next_iph_u16++;
    }
    *csum = csum_fold_helper(*csum);    
}

static inline __attribute__((__always_inline__))
void create_v4_hdr(struct iphdr *iph, __u32 saddr, __u32 daddr, __u16 pkt_bytes, __u8 proto) {
    // IPIP 封装这么简单!!!

    // 2.赋值空的 iphdr
    __u64 csum = 0;
    iph->version = 4; // ipv4
    iph->ihl = 5;
    iph->frag_off = 0;
    iph->protocol = proto; // IPPROTO_IPIP = 4, /* IPIP tunnels (older KA9Q tunnels use 94) */
    iph->check = 0;
    iph->tot_len = bpf_htons(pkt_bytes + sizeof(struct iphdr));
    iph->daddr = daddr; // rs_ip
    iph->saddr = saddr; // node_ip
    iph->ttl = 64;
    ipv4_csum_inline(iph, &csum);
    iph->check = csum;
}

static inline __attribute__((__always_inline__))
bool encap_v4(struct xdp_md *xdp, __u8 dst_mac[], __u32 saddr, __u32 daddr, __u32 pkt_bytes) {
    // 1.ipip encap, 包头向前移动 iphdr 字节，给 ipip 留位置
    if (bpf_xdp_adjust_head(xdp, 0 - (int)sizeof(struct iphdr))) {
        return false;
    }

    void *data;
    void *data_end;
    struct iphdr *iph;
    struct ethhdr *new_eth;
    struct ethhdr *old_eth;
    data = (void *)(long)xdp->data;
    data_end = (void *)(long)xdp->data_end;
    new_eth = data;
    iph = data + sizeof(struct ethhdr);
    old_eth = data + sizeof(struct iphdr); // 不应该是 tcphdr 么???
    if ((void *)new_eth + 1 > data_end 
    || (void *)old_eth + 1 > data_end 
    || (void *)iph + (sizeof(struct iphdr)) > data_end) {
        return false;
    }

    // 修改二层头
    memcpy(new_eth->h_dest, dst_mac, 6); // 网关 mac 赋值给新的包
    if (old_eth->h_dest + 6 > data_end) {
        return false;
    }
    memcpy(new_eth->h_source, old_eth->h_dest, 6); // dest mac -> src mac
    new_eth->h_proto = BE_ETH_P_IP;

    // ipip 是三层封装，修改三层头
    create_v4_hdr(iph, saddr, daddr, pkt_bytes, IPPROTO_IPIP);

    return true;
}


SEC("xdp")
int bpf_xdp_entry(struct xdp_md* ctx) {
    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;
    struct ethhdr *eth = data;
    __u32 eth_proto;
    __u32 ethhdr_off;
    ethhdr_off = sizeof(struct ethhdr);

    if (data + ethhdr_off + 1 > data_end) {
        return XDP_DROP;
    }

    // skip ipv6
    eth_proto = eth->h_proto;
    if (eth_proto == BE_ETH_P_IPV6) { // ipv4/ipv6
        return XDP_PASS;
    }

    bpf_printk("got a packet");

    struct iphdr *iph = data;
    // ip_header = data + sizeof(struct ethhdr);
    __u32 iphdr_off;
    iphdr_off = sizeof(struct iphdr);
    if ((data + iphdr_off + 1) > data_end) {
        bpf_printk("malformed");
        return XDP_DROP;
    }
    // only tcp
    if (iph->protocol != IPPROTO_TCP) {
        bpf_printk("not tcp");
        return XDP_PASS;
    }

    __u32 payload_len = bpf_ntohs(iph->tot_len);

    struct arguments *args;
    __u32 key = 0;
    args = (struct arguments *)bpf_map_lookup_elem(&xdp_params_array, &key);
    if (!args) {
        bpf_printk("no args");
        return XDP_PASS;
    }

    // 只能三个参数
    bpf_printk("Args: dstmac[%d]", args->dst_mac);
    bpf_printk("Args: daddr[%u] saddr[%u] vip [%u]", args->daddr, args->saddr, args->vip);
    if (args->vip != bpf_htonl(iph->daddr)) {
        bpf_printk("Not vip addr %u %u %u", bpf_ntohl(iph->daddr), iph->daddr, bpf_htonl(iph->daddr));
        return XDP_PASS;
    }

    // ipip 封装, dst_mac:网关mac, saddr:node_ip, daddr:rs_ip
    bool res = encap_v4(ctx, args->dst_mac, bpf_htonl(args->saddr), bpf_htonl(args->daddr), payload_len);
    bpf_printk("sending back %d", res);
    return XDP_TX;
}


char _license[] SEC("license") = "GPL";
