

/**
 * /root/linux-5.10.142/tools/testing/selftests/bpf/progs/test_xdp_noinline.c
 *
 */

#include <stddef.h>
#include <stdint.h>
#include <stdbool.h>

#include <linux/bpf.h>
//#include <linux/stddef.h>
#include <linux/pkt_cls.h>
#include <linux/if_ether.h>
#include <linux/icmp.h>
#include <linux/ip.h>
#include <linux/in.h>
#include <linux/tcp.h>
#include <linux/udp.h>

// /root/linux-5.10.142/tools/lib/bpf/bpf_helpers.h
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

#define F_ICMP (1 << 0)
#define F_SYN_SET (1 << 1)

struct flow_key {
    union {
        __be32 src;
        __be32 srcv6[4];
    };
    union {
        __be32 dst;
        __be32 dstv6[4];
    };
    union {
        __u32 ports;
        __u16 port16[2];
    };
    __u8 proto;
};

struct packet_description {
    struct flow_key flow;
    __u8 flags;
};

struct vip_definition {
    union {
        __be32 vip;
        __be32 vipv6[4];
    };
    __u16 port;
    __u16 family;
    __u8 proto;
};

struct vip_meta {
    __u32 flags;
    __u32 vip_num;
};

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 512);
    __type(key, struct vip_definition);
    __type(value, struct vip_meta);
} vip_map SEC(".maps");


struct real_pos_lru {
    __u32 pos;
    __u64 atime;
};

struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __uint(max_entries, 300);
    __uint(map_flags, 1U << 1);
    __type(key, struct flow_key);
    __type(value, struct real_pos_lru);
} lru_cache SEC(".maps"); // ct_map

static inline int process_l3_headers_v4(struct packet_description *pckt,
                                        __u8 *protocol,
                                        void *data,
                                        void *data_end) {


    struct iphdr *iph = (struct iphdr *) (data + sizeof(struct ethhdr));
    if ((void *)(iph + 1) > data_end) {
        return XDP_DROP; // 1
    }
    *protocol = iph->protocol;
    pckt->flow.proto = *protocol;
    if (*protocol == IPPROTO_ICMP) {

    } else {
        pckt->flow.src = iph->saddr;
        pckt->flow.dst = iph->daddr;
    }

    return -1;
}

static inline bool parse_tcp(void *data, void *data_end,
                struct packet_description *pckt)
{


     struct tcphdr *tcph = data + sizeof(struct ethhdr) + sizeof(struct iphdr);
    if ((void *)(tcph+1) > data_end) {
        return 0;
    }

    if (tcph->syn) {
        pckt->flags |= F_SYN_SET;//0x2
    }

    if (!(pckt->flags & F_ICMP)) {
        pckt->flow.port16[0] = tcph->source;
        pckt->flow.port16[1] = tcph->dest;
    } else {
        pckt->flow.port16[0] = tcph->dest;
        pckt->flow.port16[1] = tcph->source;
    }

    return true;
}

static inline void connection_table_lookup() {

}

static inline bool get_packet_dst() {



    return 1;
}

static __always_inline int process_packet(void *data, __u64 off, void *data_end, struct xdp_md *ctx)
{
    struct packet_description pckt = {};
    int action;
    __u8 protocol;

    action = process_l3_headers_v4(&pckt, &protocol, data, data_end);
    if (action > 0) {
        return action;
    }

    protocol = pckt.flow.proto;
    if (protocol == IPPROTO_TCP) {
        if (!parse_tcp(data, data_end, &pckt))
            return XDP_DROP;
    } else if (protocol == IPPROTO_UDP) {
//        if (!parse_udp(data, data_end, &pckt))
//            return XDP_DROP;
    } else {
        return XDP_TX;
    }

    struct vip_definition vip = { };
    vip.vip = pckt.flow.dst;
    vip.port = pckt.flow.port16[1]; // dst port
    vip.proto = pckt.flow.proto;
    struct vip_meta *vip_info;
    vip_info = bpf_map_lookup_elem(&vip_map, &vip);
    if (!vip_info) {
        vip.port = 0;
        vip_info = bpf_map_lookup_elem(&vip_map, &vip);
        if (!vip_info)
            return XDP_PASS;
        if (!(vip_info->flags & (1 << 4)))
            pckt.flow.port16[1] = 0;
    }
//    if (data_end - data > 1400)
//        return XDP_DROP;

    struct real_definition *dst = NULL;
    if (!dst) {
        if (vip_info->flags & F_ICMP)
            pckt.flow.port16[0] = 0;
        if (!(pckt.flags & F_SYN_SET) && !(vip_info->flags & F_SYN_SET))
            connection_table_lookup(&dst, &pckt, lru_map);
        if (dst)
            goto out;

        if (!get_packet_dst(&dst, &pckt, vip_info))
            return XDP_DROP;
    }

out:

    return XDP_DROP;
}


SEC("xdp-lbl4")
int balancer_ingress_v4(struct xdp_md *ctx)
{
    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;
    struct ethhdr *eth = data;
    __u32 nh_off;

    nh_off = sizeof(struct ethhdr);
    if (data + nh_off > data_end)
        return XDP_DROP;
    if (eth->h_proto == bpf_htons(ETH_P_IP)) // only ipv4
        return process_packet(data, nh_off, data_end, ctx);
    else
        return XDP_DROP;
}


char _license[] SEC("license") = "GPL";
