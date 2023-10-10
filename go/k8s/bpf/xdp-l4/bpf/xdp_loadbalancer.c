
// #include <linux/ipv6.h>
#include <stdbool.h>
#include <stddef.h>
#include <string.h>

#include <linux/bpf.h>
#include <linux/types.h>
#include <linux/in.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>

#include <loadbalancer/balancer_consts.h>
#include <loadbalancer/protocol.h>





__attribute__((__always_inline__)) static inline void connection_table_lookup(struct real_definition** real, 
    struct packet_description* pckt, void* lru_map, bool isGlobalLru) {
    struct real_pos_lru* dst_lru;
    __u64 cur_time;
    __u32 key;
    dst_lru = bpf_map_lookup_elem(lru_map, &pckt->flow);
    if (!dst_lru) {
        return;
    }
    if (!isGlobalLru && pckt->flow.proto == IPPROTO_UDP) {
        cur_time = bpf_ktime_get_ns();
        if (cur_time - dst_lru->atime > LRU_UDP_TIMEOUT) {
            return;
        }
        dst_lru->atime = cur_time;
    }
    key = dst_lru->pos;
    pckt->real_index = key;
    *real = bpf_map_lookup_elem(&reals, &key);
    
    return;
}

// 三层头获取 src/dst ip，需要考虑 icmp 协议数据
__attribute__((__always_inline__)) static inline int process_l3_headers(struct packet_description* pckt,
    __u8* protocol, __u64 off, __u16* pkt_bytes, void* data, void* data_end, bool is_ipv6) {
    __u64 iph_len;
    int action;
    struct iphdr* iph; // ip header

    if (is_ipv6) {
        // TODO
    } else {
        iph = data + off;
        if (iph + 1 > data_end) {
            // bogus packet, len less than minimum ethernet frame size
            return XDP_DROP;
        }
        // ihl contains len of ipv4 header in 32bit words
        if (iph->ihl != 5) {
            // if len of ipv4 hdr is not equal to 20bytes that means that header
            // contains ip options, and we dont support em
            return XDP_DROP;
        }

        pckt->tos = iph->tos;
        *protocol = iph->protocol;
        pckt->flow.proto = *protocol;
        *pkt_bytes = bpf_ntohs(iph->tot_len);
        off += IPV4_HDR_LEN_NO_OPT; // +20bytes
        if (iph->frag_off & PCKT_FRAGMENTED) {
            // we drop fragmented packets.
            return XDP_DROP;
        }
        if (*protocol == IPPROTO_ICMP) {
            action = parse_icmp(data, data_end, off, pckt);
            if (action >= 0) {
                return action;
            }
        } else {
            pckt->flow.src = iph->saddr;
            pckt->flow.dst = iph->daddr;
        }
    }
    
    return FURTHER_PROCESSING;
}

// __attribute__((__always_inline__)) 属性用于指示编译器始终内联函数，以提高代码的执行效率
__attribute__((__always_inline__)) static inline int
process_packet(struct xdp_md* xdp, __u64 ethhdr_off, bool is_ipv6) {
    void* data = (void*)(long)xdp->data;
    void* data_end = (void*)(long)xdp->data_end;
    struct ctl_value* cval;
    struct real_definition* dst = NULL;
    struct packet_description pckt = {};
    struct vip_definition vip = {};
    struct vip_meta* vip_info;
    struct lb_stats* data_stats;
    __u64 iph_len;
    __u8 protocol;
    __u16 original_sport;

    int action;
    __u32 vip_num;
    __u32 mac_addr_pos = 0;
    __u16 pkt_bytes;

    // 1. 获取 vip 目的地址 ip
    action = process_l3_headers(&pckt, &protocol, ethhdr_off, &pkt_bytes, data, data_end, is_ipv6);
    if (action >= 0) {
        return action;
    }
    protocol = pckt.flow.proto;

#ifdef INLINE_DECAP_IPIP
    // TODO
#endif // INLINE_DECAP_IPIP

    // 2. 获取端口 port
    if (protocol == IPPROTO_TCP) {
        if (!parse_tcp(data, data_end, is_ipv6, &pckt)) {
            return XDP_DROP;
        }
    } else if (protocol == IPPROTO_UDP) {
        if (!parse_udp(data, data_end, is_ipv6, &pckt)) {
            return XDP_DROP;
        }

#ifdef INLINE_DECAP_GUE
    // TODO
#endif // of INLINE_DECAP_GUE
    } else {
        // send to tcp/ip stack
        return XDP_PASS;
    }

    // 获取 vip 的目的地址 ip:port
    if (is_ipv6) {
        memcpy(vip.vipv6, pckt.flow.dstv6, 16);
    } else {
        vip.vip = pckt.flow.dst;
    }
    vip.port = pckt.flow.port16[1]; // port
    vip.proto = pckt.flow.proto;
    // 3. 查 vip 表，如果没有，则根据 ip 查找
    vip_info = bpf_map_lookup_elem(&vip_map, &vip);
    if (!vip_info) {
        vip.port = 0;
        vip_info = bpf_map_lookup_elem(&vip_map, &vip);
        if (!vip_info) {
            return XDP_PASS;
        }

        if (!(vip_info->flags & F_HASH_DPORT_ONLY)) {
            // VIP, which doesnt care about dst port (all packets to this VIP w/ diff
            // dst port but from the same src port/ip must go to the same real
            pckt.flow.port16[1] = 0;
        }
    }

    // TODO: 处理 大包
    if (data_end - data > MAX_PCKT_SIZE) {

    }

    __u32 stats_key = MAX_VIPS + LRU_CNTRS;
    data_stats = bpf_map_lookup_elem(&stats, &stats_key);
    if (!data_stats) {
        return XDP_DROP;
    }
    // total packets
    data_stats->v1 += 1;

    if ((vip_info->flags & F_HASH_NO_SRC_PORT)) {
        // service, where diff src port, but same ip must go to the same real,
        // e.g. gfs
        pckt.flow.port16[0] = 0;
    }

    __u32 cpu_num = bpf_get_smp_processor_id();
    void* lru_map = bpf_map_lookup_elem(&lru_mapping, &cpu_num);
    if (!lru_map) {
        lru_map = &fallback_cache;
        __u32 lru_stats_key = MAX_VIPS + FALLBACK_LRU_CNTR;
        struct lb_stats* lru_stats = bpf_map_lookup_elem(&stats, &lru_stats_key);
        if (!lru_stats) {
        return XDP_DROP;
        }
        // We were not able to retrieve per cpu/core lru and falling back to
        // default one. This counter should never be anything except 0 in prod.
        // We are going to use it for monitoring.
        lru_stats->v1 += 1;
    }

    // 4. 根据 vip, 查 real server 表
    // Lookup dst based on id in packet
    if ((vip_info->flags & F_QUIC_VIP)) {
        // TODO: quic
    }

    // save the original sport before making real selection, possibly changing its value.
    original_sport = pckt.flow.port16[0]; // src port
    if (!dst) {
#ifdef TCP_SERVER_ID_ROUTING
        // TODO
#endif // TCP_SERVER_ID_ROUTING
        
        // Next, try to lookup dst in the lru_cache
        if (!dst && !(pckt.flags & F_SYN_SET) && !(vip_info->flags & F_LRU_BYPASS)) {
            connection_table_lookup(&dst, &pckt, lru_map, /*isGlobalLru=*/false);
        }

#ifdef GLOBAL_LRU_LOOKUP
        // TODO
#endif // GLOBAL_LRU_LOOKUP

        // if dst is not found, route via consistent-hashing of the flow.
        if (!dst) {
            if (pckt.flow.proto == IPPROTO_TCP) {
                __u32 lru_stats_key = MAX_VIPS + LRU_MISS_CNTR;
                struct lb_stats* lru_stats = bpf_map_lookup_elem(&stats, &lru_stats_key);
                if (!lru_stats) {
                    return XDP_DROP;
                }
                if (pckt.flags & F_SYN_SET) {
                    // miss because of new tcp session
                    lru_stats->v1 += 1;
                } else {
                    // miss of non-syn tcp packet. could be either because of LRU
                    // trashing or because another katran is restarting and all the
                    // sessions have been reshuffled
                    REPORT_TCP_NONSYN_LRUMISS(xdp, data, data_end - data, false);
                    lru_stats->v2 += 1;
                }
            }
            if (!get_packet_dst(&dst, &pckt, vip_info, is_ipv6, lru_map)) {
                return XDP_DROP;
            }

            // track the lru miss counter of vip in lru_miss_stats_vip
            if (update_vip_lru_miss_stats(&vip, &pckt, vip_info, is_ipv6) >= 0) {
                return XDP_DROP;
            }

            // lru misses (either new connection or lru is full and starts to trash)
            data_stats->v2 += 1;
        }
    }

    cval = bpf_map_lookup_elem(&ctl_array, &mac_addr_pos);
    if (!cval) {
        return XDP_DROP;
    }

    vip_num = vip_info->vip_num;
    data_stats = bpf_map_lookup_elem(&stats, &vip_num);
    if (!data_stats) {
        return XDP_DROP;
    }
    data_stats->v1 += 1;
    data_stats->v2 += pkt_bytes;
    // per real statistics
    data_stats = bpf_map_lookup_elem(&reals_stats, &pckt.real_index);
    if (!data_stats) {
        return XDP_DROP;
    }
    data_stats->v1 += 1;
    data_stats->v2 += pkt_bytes;
#ifdef LOCAL_DELIVERY_OPTIMIZATION
    if ((vip_info->flags & F_LOCAL_VIP) && (dst->flags & F_LOCAL_REAL)) {
        return XDP_PASS;
    }
#endif

    // restore the original sport value to use it as a seed for the GUE sport
    pckt.flow.port16[0] = original_sport;
    if (dst->flags & F_IPV6) {
        if (!PCKT_ENCAP_V6(xdp, cval, is_ipv6, &pckt, dst, pkt_bytes)) {
            return XDP_DROP;
        }
    } else {
        if (!PCKT_ENCAP_V4(xdp, cval, &pckt, dst, pkt_bytes)) {
            return XDP_DROP;
        }
    }

    return XDP_TX;
}


SEC("xdp")
int balancer_ingress(struct xdp_md* ctx) {
  void* data = (void*)(long)ctx->data;
  void* data_end = (void*)(long)ctx->data_end;
  struct ethhdr* eth = data; // 二层头
  __u32 eth_proto;
  __u32 ethhdr_off;
  ethhdr_off = sizeof(struct ethhdr);
  if (data + ethhdr_off > data_end) {
    // bogus packet, len less than minimum ethernet frame size
    return XDP_DROP;
  }

  eth_proto = eth->h_proto;
  if (eth_proto == BE_ETH_P_IP) { // ipv4
    return process_packet(ctx, ethhdr_off, false);
  } else if (eth_proto == BE_ETH_P_IPV6) { // ipv6
    return process_packet(ctx, ethhdr_off, true);
  } else {
    // pass to tcp/ip stack
    return XDP_PASS;
  }
}

char _license[] SEC("license") = "GPL";