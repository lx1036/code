



/* Simple NAT engine in BPF. */
#ifndef __LIB_NAT__
#define __LIB_NAT__


#include <linux/icmp.h>
#include <linux/tcp.h>
#include <linux/udp.h>
#include <linux/ip.h>
// #include <linux/icmpv6.h>
// #include <linux/ipv6.h>

#include <bpf/helpers_skb.h>

#include "common.h"
#include "drop.h"
#include "signal.h"
#include "conntrack.h"
#include "conntrack_map.h"
// #include "icmp6.h"

#ifdef HAVE_LRU_HASH_MAP_TYPE
# define NAT_MAP_TYPE BPF_MAP_TYPE_LRU_HASH
#else
# define NAT_MAP_TYPE BPF_MAP_TYPE_HASH
#endif

enum {
    NAT_DIR_EGRESS  = TUPLE_F_OUT,
    NAT_DIR_INGRESS = TUPLE_F_IN,
};

struct nat_entry {
    __u64 created;
    __u64 host_local;	/* Only single bit used. */
    __u64 pad1;		/* Future use. */
    __u64 pad2;		/* Future use. */
};

struct ipv4_nat_entry {
    struct nat_entry common;
    union {
        struct {
            __be32 to_saddr;
            __be16 to_sport;
        };
        struct {
            __be32 to_daddr;
            __be16 to_dport;
        };
    };
};

struct ipv4_nat_target {
	__be32 addr;
	const __u16 min_port; /* host endianness */
	const __u16 max_port; /* host endianness */
	bool src_from_world;
};

struct bpf_elf_map __section_maps SNAT_MAPPING_IPV4 = {
        .type		= NAT_MAP_TYPE,
        .size_key	= sizeof(struct ipv4_ct_tuple),
        .size_value	= sizeof(struct ipv4_nat_entry),
        .pinning	= PIN_GLOBAL_NS,
        .max_elem	= SNAT_MAPPING_IPV4_SIZE,
#ifndef HAVE_LRU_HASH_MAP_TYPE
        .flags		= CONDITIONAL_PREALLOC,
#endif
};

static __always_inline __maybe_unused void *
__snat_lookup(const void *map, const void *tuple)
{
    return map_lookup_elem(map, tuple);
}

static __always_inline
struct ipv4_nat_entry *snat_v4_lookup(const struct ipv4_ct_tuple *tuple)
{
    return __snat_lookup(&SNAT_MAPPING_IPV4, tuple);
}

/**
 *
 */
static __always_inline int snat_v4_rewrite_ingress(struct __ctx_buff *ctx, struct ipv4_ct_tuple *tuple, struct ipv4_nat_entry *state, __u32 off) {
    int ret, flags = BPF_F_PSEUDO_HDR;
    struct csum_offset csum = {};
    __be32 sum_l4 = 0, sum;

    if (state->to_daddr == tuple->daddr &&
        state->to_dport == tuple->dport)
        return 0;
    sum = csum_diff(&tuple->daddr, 4, &state->to_daddr, 4, 0);
    csum_l4_offset_and_flags(tuple->nexthdr, &csum);

    if (state->to_dport != tuple->dport) {
        switch (tuple->nexthdr) {
            case IPPROTO_TCP:
            case IPPROTO_UDP:
                ret = l4_modify_port(ctx, off, offsetof(struct tcphdr, dest), &csum, state->to_dport, tuple->dport);
                if (ret < 0)
                    return ret;
                break;
            case IPPROTO_ICMP: {
                __be32 from, to;

                if (ctx_store_bytes(ctx, off + offsetof(struct icmphdr, un.echo.id), &state->to_dport, sizeof(state->to_dport), 0) < 0)
                    return DROP_WRITE_ERROR;
                from = tuple->dport;
                to = state->to_dport;
                flags = 0; /* ICMPv4 has no pseudo-header */
                sum_l4 = csum_diff(&from, 4, &to, 4, 0);
                csum.offset = offsetof(struct icmphdr, checksum);
                break;
            }}
    }

    if (ctx_store_bytes(ctx, ETH_HLEN + offsetof(struct iphdr, daddr), &state->to_daddr, 4, 0) < 0)
        return DROP_WRITE_ERROR;

    if (l3_csum_replace(ctx, ETH_HLEN + offsetof(struct iphdr, check), 0, sum, 0) < 0)
        return DROP_CSUM_L3;

    if (tuple->nexthdr == IPPROTO_ICMP)
        sum = sum_l4;

    if (csum.offset && csum_l4_replace(ctx, off, &csum, 0, sum, flags) < 0)
        return DROP_CSUM_L4;

    return 0;
}

static __always_inline int snat_v4_handle_mapping(struct __ctx_buff *ctx,
                                                  struct ipv4_ct_tuple *tuple,
                                                  struct ipv4_nat_entry **state,
                                                  struct ipv4_nat_entry *tmp,
                                                  int dir, __u32 off,
                                                  const struct ipv4_nat_target *target)
{
    int ret;

    *state = snat_v4_lookup(tuple);
    ret = snat_v4_track_local(ctx, tuple, *state, dir, off, target);
    if (ret < 0)
        return ret;
    else if (*state)
        return NAT_CONTINUE_XLATE;
    else if (dir == NAT_DIR_INGRESS)
        return tuple->nexthdr != IPPROTO_ICMP &&
               bpf_ntohs(tuple->dport) < target->min_port ?
               NAT_PUNT_TO_STACK : DROP_NAT_NO_MAPPING;
    else
        return snat_v4_new_mapping(ctx, tuple, (*state = tmp), target);
}

static __always_inline __maybe_unused int snat_v4_process(struct __ctx_buff *ctx, int dir,
                                                          const struct ipv4_nat_target *target,
                                                          bool from_endpoint)
{
    struct icmphdr icmphdr __align_stack_8;
    struct ipv4_nat_entry *state, tmp;
    struct ipv4_ct_tuple tuple = {};
    void *data, *data_end;
    struct iphdr *ip4;
    struct {
        __be16 sport;
        __be16 dport;
    } l4hdr;
    bool icmp_echoreply = false;
    __u64 off;
    int ret;

    build_bug_on(sizeof(struct ipv4_nat_entry) > 64);

    if (!revalidate_data(ctx, &data, &data_end, &ip4))
        return DROP_INVALID;

    tuple.nexthdr = ip4->protocol;
    tuple.daddr = ip4->daddr;
    tuple.saddr = ip4->saddr;
    tuple.flags = dir;
    off = ((void *)ip4 - data) + ipv4_hdrlen(ip4);
    switch (tuple.nexthdr) {
        case IPPROTO_TCP:
        case IPPROTO_UDP:
            if (ctx_load_bytes(ctx, off, &l4hdr, sizeof(l4hdr)) < 0)
                return DROP_INVALID;
            tuple.dport = l4hdr.dport;
            tuple.sport = l4hdr.sport;
            break;
        case IPPROTO_ICMP:
            if (ctx_load_bytes(ctx, off, &icmphdr, sizeof(icmphdr)) < 0)
                return DROP_INVALID;
            if (icmphdr.type != ICMP_ECHO &&
                icmphdr.type != ICMP_ECHOREPLY)
                return DROP_NAT_UNSUPP_PROTO;
            if (icmphdr.type == ICMP_ECHO) {
                tuple.dport = 0;
                tuple.sport = icmphdr.un.echo.id;
            } else {
                tuple.dport = icmphdr.un.echo.id;
                tuple.sport = 0;
                icmp_echoreply = true;
            }
            break;
        default:
            return NAT_PUNT_TO_STACK;
    };

    if (snat_v4_can_skip(target, &tuple, dir, from_endpoint, icmp_echoreply))
        return NAT_PUNT_TO_STACK;
    ret = snat_v4_handle_mapping(ctx, &tuple, &state, &tmp, dir, off, target);
    if (ret > 0)
        return CTX_ACT_OK;
    if (ret < 0)
        return ret;

    return dir == NAT_DIR_EGRESS ?
           snat_v4_rewrite_egress(ctx, &tuple, state, off, ipv4_has_l4_header(ip4)) :
           snat_v4_rewrite_ingress(ctx, &tuple, state, off);
}




#endif /* __LIB_NAT__ */
