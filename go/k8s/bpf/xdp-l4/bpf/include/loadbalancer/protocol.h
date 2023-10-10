

#ifndef __PROTOCOL_H
#define __PROTOCOL_H

#include <stddef.h>

#include <linux/bpf.h>
#include <linux/types.h>
#include <linux/ip.h>
#include <linux/ipv6.h>
#include <linux/icmp.h>
#include <linux/icmpv6.h>
#include <linux/tcp.h>
#include <linux/udp.h>
#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>

#include <loadbalancer/balancer_consts.h>
#include <loadbalancer/csum_helpers.h>

/*
 * This file contains description of all structs which has been used both by
 * balancer and by packet's parsing routines
 */

// flow metadata
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

// client's packet metadata
struct packet_description {
  struct flow_key flow;
  __u32 real_index;
  __u8 flags;
  // dscp / ToS value in client's packet
  __u8 tos;
};

// value for ctl array, could contain e.g. mac address of default router
// or other flags
struct ctl_value {
  union {
    __u64 value;
    __u32 ifindex;
    __u8 mac[6];
  };
};

// vip's definition for lookup
struct vip_definition {
  union {
    __be32 vip;
    __be32 vipv6[4];
  };
  __u16 port;
  __u8 proto;
};

// result of vip's lookup
struct vip_meta {
  __u32 flags;
  __u32 vip_num;
};

// where to send client's packet from LRU_MAP
struct real_pos_lru {
  __u32 pos;
  __u64 atime;
};

// where to send client's packet from lookup in ch ring.
struct real_definition {
  union {
    __be32 dst;
    __be32 dstv6[4];
  };
  __u8 flags;
};

// per vip statistics
struct lb_stats {
  __u64 v1;
  __u64 v2;
};

// key for ipv4 lpm lookups
struct v4_lpm_key {
  __u32 prefixlen;
  __be32 addr;
};

// key for ipv6 lpm lookups
struct v6_lpm_key {
  __u32 prefixlen;
  __be32 addr[4];
};

struct address {
  union {
    __be32 addr;
    __be32 addrv6[4];
  };
};

#ifdef KATRAN_INTROSPECTION
// metadata about packet, copied to the userspace through event pipe
struct event_metadata {
  __u32 event;
  __u32 pkt_size;
  __u32 data_len;
} __attribute__((__packed__));

#endif

#ifdef RECORD_FLOW_INFO
// Route information saved from GUE packets
struct flow_debug_info {
  union {
    __be32 l4_hop;
    __be32 l4_hopv6[4];
  };
  union {
    __be32 this_hop;
    __be32 this_hopv6[4];
  };
};
#endif // of RECORD_FLOW_INFO

// struct for quic packets statistics counters
struct lb_quic_packets_stats {
  __u64 ch_routed;
  __u64 cid_initial;
  __u64 cid_invalid_server_id;
  __u64 cid_invalid_server_id_sample;
  __u64 cid_routed;
  __u64 cid_unknown_real_dropped;
  __u64 cid_v0;
  __u64 cid_v1;
  __u64 cid_v2;
  __u64 cid_v3;
  __u64 dst_match_in_lru;
  __u64 dst_mismatch_in_lru;
  __u64 dst_not_found_in_lru;
};


// map, which contains all the vips for which we are doing load balancing，这个是关键 map!!!
struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(key_size, sizeof(struct vip_definition));
	__uint(value_size, sizeof(struct vip_meta));
	__uint(max_entries, MAX_VIPS);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
  __uint(map_flags, NO_FLAGS);
} vip_map SEC(".maps");

// control array. contains metadata such as default router mac
// and/or interfaces ifindexes
// indexes:
// 0 - default's mac
struct {
	__uint(type, BPF_MAP_TYPE_ARRAY);
	__uint(key_size, sizeof(__u32));
	__uint(value_size, sizeof(struct ctl_value));
	__uint(max_entries, CTL_MAP_SIZE);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
  __uint(map_flags, NO_FLAGS);
} ctl_array SEC(".maps");

// map which contains opaque real's id to real mapping, 这个是关键 map!!!
struct {
	__uint(type, BPF_MAP_TYPE_ARRAY);
	__uint(key_size, sizeof(__u32));
	__uint(value_size, sizeof(struct real_definition));
	__uint(max_entries, MAX_REALS);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
  __uint(map_flags, NO_FLAGS);
} reals SEC(".maps");
// map with per real pps/bps statistic
struct {
	__uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
	__uint(key_size, sizeof(__u32));
	__uint(value_size, sizeof(struct lb_stats));
	__uint(max_entries, MAX_REALS);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
  __uint(map_flags, NO_FLAGS);
} reals_stats SEC(".maps");


// map vip stats
struct {
	__uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
	__uint(key_size, sizeof(__u32));
	__uint(value_size, sizeof(struct lb_stats));
	__uint(max_entries, STATS_MAP_SIZE);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
  __uint(map_flags, NO_FLAGS);
} stats SEC(".maps");

// fallback lru. we should never hit this one outside of unittests
struct {
	__uint(type, BPF_MAP_TYPE_LRU_HASH);
	__uint(key_size, sizeof(struct flow_key));
	__uint(value_size, sizeof(struct real_pos_lru));
	__uint(max_entries, DEFAULT_LRU_SIZE);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
  __uint(map_flags, NO_FLAGS);
} fallback_cache SEC(".maps");

// map which contains cpu core to lru mapping
struct {
	__uint(type, BPF_MAP_TYPE_ARRAY_OF_MAPS);
	__uint(key_size, sizeof(__u32));
	__uint(value_size, sizeof(__u32));
	__uint(max_entries, MAX_SUPPORTED_CPUS);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
  __uint(map_flags, NO_FLAGS);
  __array(
      values,
      struct {
        __uint(type, BPF_MAP_TYPE_LRU_HASH);
        __uint(key_size, sizeof(struct flow_key));
	      __uint(value_size, sizeof(struct real_pos_lru));
        // __type(key, struct flow_key);
        // __type(value, struct real_pos_lru);
        __uint(max_entries, DEFAULT_LRU_SIZE);
      });
} lru_mapping SEC(".maps");



__attribute__((__always_inline__)) static inline int send_icmp_reply(void* data, void* data_end) {
    struct iphdr* iph;
    struct icmphdr* icmp_hdr;
    __u32 tmp_addr = 0;
    __u64 csum = 0;
    __u64 off = 0;

    if ((data + sizeof(struct ethhdr) + sizeof(struct iphdr) +
        sizeof(struct icmphdr)) > data_end) {
        return XDP_DROP;
    }
    off += sizeof(struct ethhdr);
    iph = data + off;
    off += sizeof(struct iphdr);
    icmp_hdr = data + off;
    icmp_hdr->type = ICMP_ECHOREPLY;
    // the only diff between icmp echo and reply hdrs is type;
    // in first case it's 8; in second it's 0; so instead of recalc
    // checksum from ground up we will just adjust it.
    icmp_hdr->checksum += 0x0008;
    iph->ttl = DEFAULT_TTL;
    tmp_addr = iph->daddr;
    iph->daddr = iph->saddr;
    iph->saddr = tmp_addr;
    iph->check = 0;
    ipv4_csum_inline(iph, &csum);
    iph->check = csum;
    return swap_mac_and_send(data, data_end);
}


__attribute__((__always_inline__)) static inline int parse_icmp(void* data, void* data_end, 
    __u64 off, struct packet_description* pckt) {
    struct icmphdr* icmp_hdr;
    struct iphdr* iph;
    icmp_hdr = data + off;
    if (icmp_hdr + 1 > data_end) {
        return XDP_DROP;
    }

    if (icmp_hdr->type == ICMP_ECHO) {
        return send_icmp_reply(data, data_end);
    }
    if (icmp_hdr->type != ICMP_DEST_UNREACH) {
        return XDP_PASS;
    }
    if (icmp_hdr->code == ICMP_FRAG_NEEDED) {
        __u32 stats_key = MAX_VIPS + ICMP_PTB_V4_STATS;
        struct lb_stats* icmp_ptb_v4_stats = bpf_map_lookup_elem(&stats, &stats_key);
        if (!icmp_ptb_v4_stats) {
          return XDP_DROP;
        }
        icmp_ptb_v4_stats->v1 += 1;
        __u16 mtu = bpf_ntohs(icmp_hdr->un.frag.mtu);
        if (mtu < MAX_MTU_IN_PTB_TO_DROP) {
            icmp_ptb_v4_stats->v2 += 1;
        }
    }

    off += sizeof(struct icmphdr);
    iph = data + off;
    if (iph + 1 > data_end) {
        return XDP_DROP;
    }
    if (iph->ihl != 5) {
        return XDP_DROP;
    }
    pckt->flow.proto = iph->protocol;
    pckt->flags |= F_ICMP;
    pckt->flow.src = iph->daddr;
    pckt->flow.dst = iph->saddr;
    return FURTHER_PROCESSING;
}


// 注意这里的 icmp 还需要再加上 sizeof(struct iphdr)
__attribute__((__always_inline__)) static inline __u64 calc_offset(bool is_ipv6, bool is_icmp) {
  __u64 off = sizeof(struct ethhdr);
  if (is_ipv6) {
    off += sizeof(struct ipv6hdr);
    if (is_icmp) {
      off += (sizeof(struct icmp6hdr) + sizeof(struct ipv6hdr));
    }
  } else {
    off += sizeof(struct iphdr);
    if (is_icmp) {
      off += (sizeof(struct icmphdr) + sizeof(struct iphdr));
    }
  }
  return off;
}

// 获取 tcp 的 port
__attribute__((__always_inline__)) static inline bool parse_tcp(void* data, 
  void* data_end, bool is_ipv6, struct packet_description* pckt) {
    bool is_icmp = !((pckt->flags & F_ICMP) == 0);
    __u64 off = calc_offset(is_ipv6, is_icmp);
    struct tcphdr* tcp;
    tcp = data + off;
    if (tcp + 1 > data_end) {
      return false;
    }

    if (tcp->syn) {
      pckt->flags |= F_SYN_SET;
    }

    if (!is_icmp) {
    pckt->flow.port16[0] = tcp->source;
    pckt->flow.port16[1] = tcp->dest;
    } else {
      // packet_description was created from icmp "packet too big". hence
      // we need to invert src/dst ports
      pckt->flow.port16[0] = tcp->dest;
      pckt->flow.port16[1] = tcp->source;
    }

    return true;
}

// 获取 udp 的 port
__attribute__((__always_inline__)) static inline bool parse_udp(void* data, 
  void* data_end,bool is_ipv6,struct packet_description* pckt) {
    bool is_icmp = !((pckt->flags & F_ICMP) == 0);
    __u64 off = calc_offset(is_ipv6, is_icmp);
    struct udphdr* udp;
    udp = data + off;
    if (udp + 1 > data_end) {
      return false;
    }

    if (!is_icmp) {
      pckt->flow.port16[0] = udp->source;
      pckt->flow.port16[1] = udp->dest;
    } else {
      // packet_description was created from icmp "packet too big". hence
      // we need to invert src/dst ports
      pckt->flow.port16[0] = udp->dest;
      pckt->flow.port16[1] = udp->source;
    }

    return true;
}


__attribute__((__always_inline__)) static inline void create_v4_hdr(struct iphdr* iph, __u8 tos, 
    __u32 saddr, __u32 daddr, __u16 pkt_bytes, __u8 proto) {
    __u64 csum = 0;
    iph->version = 4;
    iph->ihl = 5;
    iph->frag_off = 0;
    iph->protocol = proto;
    iph->check = 0;
  #ifdef COPY_INNER_PACKET_TOS
    iph->tos = tos;
  #else
    iph->tos = DEFAULT_TOS;
  #endif
    iph->tot_len = bpf_htons(pkt_bytes + sizeof(struct iphdr));
    iph->daddr = daddr;
    iph->saddr = saddr;
    iph->ttl = DEFAULT_TTL;
    ipv4_csum_inline(iph, &csum);
    iph->check = csum;
}

__attribute__((__always_inline__)) static inline bool encap_v4(struct xdp_md* xdp, struct ctl_value* cval,
    struct packet_description* pckt, struct real_definition* dst, __u32 pkt_bytes) {
    void* data;
    void* data_end;
    struct iphdr* iph;
    struct ethhdr* new_eth;
    struct ethhdr* old_eth;
    __u32 ip_suffix = bpf_htons(pckt->flow.port16[0]); // src port
    ip_suffix <<= 16;
    ip_suffix ^= pckt->flow.src;
    __u64 csum = 0;
    // ipip encap
    if (bpf_xdp_adjust_head(xdp, 0 - (int)sizeof(struct iphdr))) {
      return false;
    }

    data = (void*)(long)xdp->data;
    data_end = (void*)(long)xdp->data_end;
    new_eth = data;
    iph = data + sizeof(struct ethhdr);
    old_eth = data + sizeof(struct iphdr);
    if (new_eth + 1 > data_end || old_eth + 1 > data_end || iph + 1 > data_end) {
      return false;
    }
    memcpy(new_eth->h_dest, cval->mac, 6);
    memcpy(new_eth->h_source, old_eth->h_dest, 6);
    new_eth->h_proto = BE_ETH_P_IP;

    create_v4_hdr(iph, pckt->tos, ((0xFFFF0000 & ip_suffix) | IPIP_V4_PREFIX),
      dst->dst, pkt_bytes, IPPROTO_IPIP);

    return true;
}


__attribute__((__always_inline__)) static inline void create_v6_hdr(struct ipv6hdr* ip6h, __u8 tc, 
__u32* saddr, __u32* daddr, __u16 payload_len, __u8 proto) {
    ip6h->version = 6;
    memset(ip6h->flow_lbl, 0, sizeof(ip6h->flow_lbl));
  #ifdef COPY_INNER_PACKET_TOS
    ip6h->priority = (tc & 0xF0) >> 4;
    ip6h->flow_lbl[0] = (tc & 0x0F) << 4;
  #else
    ip6h->priority = DEFAULT_TOS;
  #endif
    ip6h->nexthdr = proto;
    ip6h->payload_len = bpf_htons(payload_len);
    ip6h->hop_limit = DEFAULT_TTL;
    memcpy(ip6h->saddr.s6_addr32, saddr, 16);
    memcpy(ip6h->daddr.s6_addr32, daddr, 16);
}

__attribute__((__always_inline__)) static inline bool encap_v6(struct xdp_md* xdp, struct ctl_value* cval, 
    bool is_ipv6, struct packet_description* pckt, struct real_definition* dst, __u32 pkt_bytes) {
    void* data;
    void* data_end;
    struct ipv6hdr* ip6h;
    struct ethhdr* new_eth;
    struct ethhdr* old_eth;
    __u16 payload_len;
    __u32 ip_suffix;
    __u32 saddr[4];
    __u8 proto;
    // ip(6)ip6 encap
    if (bpf_xdp_adjust_head(xdp, 0 - (int)sizeof(struct ipv6hdr))) {
      return false;
    }
    data = (void*)(long)xdp->data;
    data_end = (void*)(long)xdp->data_end;
    new_eth = data;
    ip6h = data + sizeof(struct ethhdr);
    old_eth = data + sizeof(struct ipv6hdr);
    if (new_eth + 1 > data_end || old_eth + 1 > data_end || ip6h + 1 > data_end) {
      return false;
    }
    memcpy(new_eth->h_dest, cval->mac, 6);
    memcpy(new_eth->h_source, old_eth->h_dest, 6);
    new_eth->h_proto = BE_ETH_P_IPV6;

    if (is_ipv6) {
      proto = IPPROTO_IPV6;
      ip_suffix = pckt->flow.srcv6[3] ^ pckt->flow.port16[0];
      payload_len = pkt_bytes + sizeof(struct ipv6hdr);
    } else {
      proto = IPPROTO_IPIP;
      ip_suffix = pckt->flow.src ^ pckt->flow.port16[0];
      payload_len = pkt_bytes;
    }

    saddr[0] = IPIP_V6_PREFIX1;
    saddr[1] = IPIP_V6_PREFIX2;
    saddr[2] = IPIP_V6_PREFIX3;
    saddr[3] = ip_suffix;

    create_v6_hdr(ip6h, pckt->tos, saddr, dst->dstv6, payload_len, proto);

    return true;
}







#endif // __PROTOCOL_H
