
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
#include <loadbalancer/balancer_structs.h>









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

    action = process_l3_headers(&pckt, &protocol, ethhdr_off, &pkt_bytes, data, data_end, is_ipv6);
    if (action >= 0) {
        return action;
    }
    protocol = pckt.flow.proto;

#ifdef INLINE_DECAP_IPIP
    // TODO
#endif // INLINE_DECAP_IPIP

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