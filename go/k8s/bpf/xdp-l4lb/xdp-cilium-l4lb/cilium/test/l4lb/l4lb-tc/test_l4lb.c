
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


#define PCKT_FRAGMENTED 65343
#define MAX_VIPS 12
#define MAX_REALS 5
#define CTL_MAP_SIZE 16
#define F_ICMP (1 << 0)
#define F_SYN_SET (1 << 1)
#define RING_SIZE 2
#define CH_RINGS_SIZE (MAX_VIPS * RING_SIZE)
#define F_HASH_NO_SRC_PORT (1 << 0)
#define IPV4_HDR_LEN_NO_OPT 20

typedef __u32 u32;


struct packet_description {
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
    __u8 flags;
};

struct vip {
    union {
        __u32 v6[4];
        __u32 v4;
    } daddr;
    __u16 dport;
    __u16 family;
    __u8 protocol;
};

struct vip_meta {
    __u32 flags;
    __u32 vip_num;
};

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, MAX_VIPS);
    __type(key, struct vip);
    __type(value, struct vip_meta);
} vip_map SEC(".maps");

struct ctl_value {
    union {
        __u64 value;
        __u32 ifindex;
        __u8 mac[6];
    };
};

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, CTL_MAP_SIZE);
    __type(key, __u32);
    __type(value, struct ctl_value);
} ctl_array SEC(".maps");

struct real_definition {
    union {
        __be32 dst;
        __be32 dstv6[4];
    };
    __u8 flags;
};

struct vip_stats {
    __u64 bytes;
    __u64 pkts;
};

struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, MAX_VIPS);
    __type(key, __u32);
    __type(value, struct vip_stats);
} stats SEC(".maps");

#define __jhash_final(a, b, c)			\
{						\
	c ^= b; c -= rol32(b, 14);		\
	a ^= c; a -= rol32(c, 11);		\
	b ^= a; b -= rol32(a, 25);		\
	c ^= b; c -= rol32(b, 16);		\
	a ^= c; a -= rol32(c, 4);		\
	b ^= a; b -= rol32(a, 14);		\
	c ^= b; c -= rol32(b, 24);		\
}
#define JHASH_INITVAL		0xdeadbeef
static inline u32 __jhash_nwords(u32 a, u32 b, u32 c, u32 initval)
{
    a += initval;
    b += initval;
    c += initval;
    __jhash_final(a, b, c);
    return c;
}

static inline u32 jhash_2words(u32 a, u32 b, u32 initval)
{
    return __jhash_nwords(a, b, 0, initval + JHASH_INITVAL + (2 << 2));
}

static __always_inline bool parse_tcp(void *data, __u64 off, void *data_end, struct packet_description *pckt)
{
    struct tcphdr *tcp;
    tcp = data + off;
    if ((void *)(tcp + 1) > data_end)
        return false;

    // syn 包
    if (tcp->syn)
        pckt->flags |= F_SYN_SET;
    if (!(pckt->flags & F_ICMP)) {
        pckt->port16[0] = tcp->source;
        pckt->port16[1] = tcp->dest;
    } else {
        pckt->port16[0] = tcp->dest;
        pckt->port16[1] = tcp->source;
    }

    return true;
}

static __always_inline bool parse_udp(void *data, __u64 off, void *data_end, struct packet_description *pckt)
{
    struct udphdr *udp;
    udp = data + off;

    if ((void *)(udp + 1) > data_end)
        return false;

    if (!(pckt->flags & F_ICMP)) {
        pckt->port16[0] = udp->source;
        pckt->port16[1] = udp->dest;
    } else {
        pckt->port16[0] = udp->dest;
        pckt->port16[1] = udp->source;
    }

    return true;
}

static __always_inline __u32 get_packet_hash(struct packet_description *pckt)
{
    return jhash_2words(pckt->src, pckt->ports, CH_RINGS_SIZE); // 12*2
}

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, CH_RINGS_SIZE);
    __type(key, __u32);
    __type(value, __u32);
} ch_rings SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, MAX_REALS);
    __type(key, __u32);
    __type(value, struct real_definition);
} reals SEC(".maps");

static __always_inline bool get_packet_dst(struct real_definition **real, struct packet_description *pckt, struct vip_meta *vip_info)
{
    __u32 hash = get_packet_hash(pckt) % RING_SIZE;
    __u32 key = RING_SIZE * vip_info->vip_num + hash;
    __u32 *real_pos;

    real_pos = bpf_map_lookup_elem(&ch_rings, &key);
    if (!real_pos)
        return false;
    key = *real_pos;
    *real = bpf_map_lookup_elem(&reals, &key);
    if (!(*real))
        return false;
    return true;
}

static __always_inline int process_packet(void *data, __u64 off, void *data_end, struct __sk_buff *skb)
{
    struct iphdr *iph;
    __u8 protocol;
    struct packet_description pckt = {};
    __u16 pkt_bytes;

    iph = data + off;
    if ((void*)(iph + 1) > data_end)
        return TC_ACT_SHOT;
    if (iph->ihl != 5) // ???
        return TC_ACT_SHOT;

    protocol = iph->protocol;
    pckt.proto = protocol;
    pkt_bytes = bpf_ntohs(iph->tot_len);
    off += IPV4_HDR_LEN_NO_OPT; // *tcp_option field

    // 1.获取 src/dst ip
    if (iph->frag_off & PCKT_FRAGMENTED)
        return TC_ACT_SHOT;
    if (protocol == IPPROTO_ICMP) {
//        action = parse_icmp(data, data_end, off, &pckt);
//        if (action >= 0)
//            return action;
//        off += IPV4_PLUS_ICMP_HDR;
    } else {
        pckt.src = iph->saddr;
        pckt.dst = iph->daddr;
    }

    // 2.获取 tcp/udp port
    if (protocol == IPPROTO_TCP) {
        if (!parse_tcp(data, off, data_end, &pckt))
            return TC_ACT_SHOT;
    } else if (protocol == IPPROTO_UDP) {
        if (!parse_udp(data, off, data_end, &pckt))
            return TC_ACT_SHOT;
    } else {
        return TC_ACT_SHOT;
    }

    struct vip vip = {};
    struct vip_meta *vip_info;
    vip.daddr.v4 = pckt.dst;
    vip.dport = pckt.port16[1];
    vip.protocol = pckt.proto;
    vip_info = bpf_map_lookup_elem(&vip_map, &vip);
    if (!vip_info) {
        vip.dport = 0;
        vip_info = bpf_map_lookup_elem(&vip_map, &vip);
        if (!vip_info)
            return TC_ACT_SHOT;
        pckt.port16[1] = 0;
    }

    if (vip_info->flags & F_HASH_NO_SRC_PORT)
        pckt.port16[0] = 0;

    struct real_definition *dst;
    if (!get_packet_dst(&dst, &pckt, vip_info))
        return TC_ACT_SHOT;

    __u32 v4_intf_pos = 1;
    struct ctl_value *cval;
    struct bpf_tunnel_key tkey = {};
    struct vip_stats *data_stats;
//    struct ethhdr *eth = (void *)(long)skb->data;
    __u32 ifindex;
    __u32 vip_num;
    cval = bpf_map_lookup_elem(&ctl_array, &v4_intf_pos);
    if (!cval)
        return TC_ACT_SHOT;
    ifindex = cval->ifindex;
    tkey.remote_ipv4 = dst->dst;

    vip_num = vip_info->vip_num;
    data_stats = bpf_map_lookup_elem(&stats, &vip_num);
    if (!data_stats)
        return TC_ACT_SHOT;
    data_stats->pkts++;
    data_stats->bytes += pkt_bytes;
    bpf_skb_set_tunnel_key(skb, &tkey, sizeof(tkey), 0);
//    *(__u32 *)eth->h_dest = tkey.remote_ipv4; // ??? 不太对吧，eth->h_dest 是 mac 地址

    return bpf_redirect(ifindex, 0);
}


SEC("l4lb-demo")
int balancer_ingress(struct __sk_buff *ctx)
{
    void *data_end = (void *)(long)ctx->data_end;
    void *data = (void *)(long)ctx->data;
    struct ethhdr *eth = data;
    __u32 nh_off;

    nh_off = sizeof(struct ethhdr);
    if (data + nh_off > data_end)
        return TC_ACT_SHOT;
    if (eth->h_proto == bpf_htons(ETH_P_IP))
        return process_packet(data, nh_off, data_end, ctx);
    else
        return TC_ACT_SHOT;
}


char _license[] SEC("license") = "GPL";

