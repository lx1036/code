

#ifndef __PROTOCOL_H
#define __PROTOCOL_H

#include <stddef.h>

#include <linux/bpf.h>
#include <linux/types.h>
#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>

#include <loadbalancer/balancer_consts.h>

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


// map vip stats
struct {
	__uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
	__uint(key_size, sizeof(__u32));
	__uint(value_size, sizeof(struct lb_stats));
	__uint(max_entries, STATS_MAP_SIZE);
	__uint(pinning, LIBBPF_PIN_BY_NAME);
    __uint(map_flags, NO_FLAGS);
} stats SEC(".maps");



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


#endif
