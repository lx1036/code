

#ifndef XDP_CILIUM_L4LB_CONNTRACK_H
#define XDP_CILIUM_L4LB_CONNTRACK_H


#include <linux/icmpv6.h>
#include <linux/icmp.h>

//#include <bpf/verifier.h>

#include "common.h"
#include "utils.h"
#include "ipv4.h"
#include "ipv6.h"
#include "dbg.h"
#include "l4.h"
#include "nat46.h"
#include "signal.h"
#include "config.h"



enum {
    ACTION_UNSPEC,
    ACTION_CREATE,
    ACTION_CLOSE,
};

union tcp_flags {
    struct {
        __u8 upper_bits;
        __u8 lower_bits;
        __u16 pad;
    };
    __u32 value;
};

#ifdef HAVE_LRU_HASH_MAP_TYPE
#define CT_MAP_TYPE BPF_MAP_TYPE_LRU_HASH
#else
#define CT_MAP_TYPE BPF_MAP_TYPE_HASH
#endif

struct bpf_elf_map __section_maps CT_MAP_TCP4 = {
        .type		= CT_MAP_TYPE,
        .size_key	= sizeof(struct ipv4_ct_tuple),
        .size_value	= sizeof(struct ct_entry),
        .pinning	= PIN_GLOBAL_NS,
        .max_elem	= CT_MAP_SIZE_TCP,
#ifndef HAVE_LRU_HASH_MAP_TYPE
        .flags		= CONDITIONAL_PREALLOC,
#endif
};

struct bpf_elf_map __section_maps CT_MAP_ANY4 = {
        .type		= CT_MAP_TYPE,
        .size_key	= sizeof(struct ipv4_ct_tuple),
        .size_value	= sizeof(struct ct_entry),
        .pinning	= PIN_GLOBAL_NS,
        .max_elem	= CT_MAP_SIZE_ANY,
#ifndef HAVE_LRU_HASH_MAP_TYPE
        .flags		= CONDITIONAL_PREALLOC,
#endif
};

static __always_inline struct bpf_elf_map *
get_ct_map4(const struct ipv4_ct_tuple *tuple)
{
    if (tuple->nexthdr == IPPROTO_TCP)
        return &CT_MAP_TCP4;

    return &CT_MAP_ANY4;
}

// 从 xdp_md 字节数组里获取 dport 字段值
static __always_inline int ipv4_ct_extract_l4_ports(struct __ctx_buff *ctx, int off, int dir __maybe_unused,
                                                    struct ipv4_ct_tuple *tuple, bool *has_l4_header __maybe_unused) {
//#ifdef ENABLE_IPV4_FRAGMENTS
    void *data, *data_end;
	struct iphdr *ip4;

	/* This function is called from ct_lookup4(), which is sometimes called
	 * after data has been invalidated (see handle_ipv4_from_lxc())
	 */
	if (!revalidate_data(ctx, &data, &data_end, &ip4))
		return DROP_CT_INVALID_HDR;

	return ipv4_handle_fragmentation(ctx, ip4, off, dir,
				    (struct ipv4_frag_l4ports *)&tuple->dport,
				    has_l4_header);
//#else
//    /* load sport + dport into tuple */
//    if (ctx_load_bytes(ctx, off, &tuple->dport, 4) < 0)
//        return DROP_CT_INVALID_HDR;
//#endif

    return CTX_ACT_OK;
}


static __always_inline __u8 __ct_lookup(const void *map, struct __ctx_buff *ctx,
                                        const void *tuple, int action, int dir,
                                        struct ct_state *ct_state,
                                        bool is_tcp, union tcp_flags seen_flags,
                                        __u32 *monitor)
{
    struct ct_entry *entry;
    int reopen;

//    relax_verifier();

    entry = map_lookup_elem(map, tuple);
    if (entry) {
        cilium_dbg(ctx, DBG_CT_MATCH, entry->lifetime, entry->rev_nat_index);
        if (ct_entry_alive(entry))
            *monitor = ct_update_timeout(entry, is_tcp, dir, seen_flags);
        if (ct_state) {
            ct_state->rev_nat_index = entry->rev_nat_index;
            ct_state->loopback = entry->lb_loopback;
            ct_state->node_port = entry->node_port;
            ct_state->ifindex = entry->ifindex;
            ct_state->dsr = entry->dsr;
            ct_state->proxy_redirect = entry->proxy_redirect;
            /* See the ct_create4 comments re the rx_bytes hack */
            if (dir == CT_SERVICE)
                ct_state->backend_id = entry->rx_bytes;
        }

#ifdef ENABLE_NAT46
        /* This packet needs nat46 translation */
		if (entry->nat46 && !ctx_load_meta(ctx, CB_NAT46_STATE))
			ctx_store_meta(ctx, CB_NAT46_STATE, NAT46);
#endif
#ifdef CONNTRACK_ACCOUNTING
        /* FIXME: This is slow, per-cpu counters? */
        if (dir == CT_INGRESS) {
            __sync_fetch_and_add(&entry->rx_packets, 1);
            __sync_fetch_and_add(&entry->rx_bytes, ctx_full_len(ctx));
        } else if (dir == CT_EGRESS) {
            __sync_fetch_and_add(&entry->tx_packets, 1);
            __sync_fetch_and_add(&entry->tx_bytes, ctx_full_len(ctx));
        }
#endif
        switch (action) {
            case ACTION_CREATE:
                reopen = entry->rx_closing | entry->tx_closing;
                reopen |= seen_flags.value & TCP_FLAG_SYN;
                if (unlikely(reopen == (TCP_FLAG_SYN|0x1))) {
                    ct_reset_closing(entry);
                    *monitor = ct_update_timeout(entry, is_tcp, dir, seen_flags);
                    return CT_REOPENED;
                }
                break;

            case ACTION_CLOSE:
                /* If we got an RST and have not seen both SYNs,
                 * terminate the connection. (For CT_SERVICE, we do not
                 * see both directions, so flags of established
                 * connections would not include both SYNs.)
                 */
                if (!ct_entry_seen_both_syns(entry) &&
                    (seen_flags.value & TCP_FLAG_RST) &&
                    dir != CT_SERVICE) {
                    entry->rx_closing = 1;
                    entry->tx_closing = 1;
                } else if (dir == CT_INGRESS) {
                    entry->rx_closing = 1;
                } else {
                    entry->tx_closing = 1;
                }

                *monitor = TRACE_PAYLOAD_LEN;
                if (ct_entry_alive(entry))
                    break;
                __ct_update_timeout(entry, bpf_sec_to_mono(CT_CLOSE_TIMEOUT),
                                    dir, seen_flags, CT_REPORT_FLAGS);
                break;
        }

        return CT_ESTABLISHED;
    }

    *monitor = TRACE_PAYLOAD_LEN;
    return CT_NEW;
}

/* Offset must point to IPv4 header */
static __always_inline int ct_lookup4(const void *map,
                                      struct ipv4_ct_tuple *tuple,
                                      struct __ctx_buff *ctx, int off, int dir,
                                      struct ct_state *ct_state, __u32 *monitor)
{
    int err, ret = CT_NEW, action = ACTION_UNSPEC;
    bool is_tcp = tuple->nexthdr == IPPROTO_TCP,
            has_l4_header = true;
    union tcp_flags tcp_flags = { .value = 0 };

    /* The tuple is created in reverse order initially to find a
     * potential reverse flow. This is required because the RELATED
     * or REPLY state takes precedence over ESTABLISHED due to
     * policy requirements.
     *
     * tuple->flags separates entries that could otherwise be overlapping.
     */
    if (dir == CT_INGRESS)
        tuple->flags = TUPLE_F_OUT;
    else if (dir == CT_EGRESS)
        tuple->flags = TUPLE_F_IN;
    else if (dir == CT_SERVICE)
        tuple->flags = TUPLE_F_SERVICE;
    else
        return DROP_CT_INVALID_HDR;

    switch (tuple->nexthdr) {
        case IPPROTO_ICMP:
            if (1) {
                __be16 identifier = 0;
                __u8 type;

                if (ctx_load_bytes(ctx, off, &type, 1) < 0)
                    return DROP_CT_INVALID_HDR;
                if ((type == ICMP_ECHO || type == ICMP_ECHOREPLY) &&
                    ctx_load_bytes(ctx, off + offsetof(struct icmphdr, un.echo.id),
                &identifier, 2) < 0)
                return DROP_CT_INVALID_HDR;

                tuple->sport = 0;
                tuple->dport = 0;

                switch (type) {
                    case ICMP_DEST_UNREACH:
                    case ICMP_TIME_EXCEEDED:
                    case ICMP_PARAMETERPROB:
                        tuple->flags |= TUPLE_F_RELATED;
                        break;

                    case ICMP_ECHOREPLY:
                        tuple->sport = identifier;
                        break;
                    case ICMP_ECHO:
                        tuple->dport = identifier;
                        /* fall through */
                    default:
                        action = ACTION_CREATE;
                        break;
                }
            }
            break;

        case IPPROTO_TCP:
            err = ipv4_ct_extract_l4_ports(ctx, off, dir, tuple, &has_l4_header);
            if (err < 0)
                return err;

            action = ACTION_CREATE;

            if (has_l4_header) {
                if (ctx_load_bytes(ctx, off + 12, &tcp_flags, 2) < 0)
                    return DROP_CT_INVALID_HDR;

                if (unlikely(tcp_flags.value & (TCP_FLAG_RST|TCP_FLAG_FIN)))
                    action = ACTION_CLOSE;
            }
            break;

        case IPPROTO_UDP:
            err = ipv4_ct_extract_l4_ports(ctx, off, dir, tuple, NULL);
            if (err < 0)
                return err;

            action = ACTION_CREATE;
            break;

        default:
            /* Can't handle extension headers yet */
            return DROP_CT_UNKNOWN_PROTO;
    }

    /* Lookup the reverse direction
     *
     * This will find an existing flow in the reverse direction.
     */
#ifndef QUIET_CT
    cilium_dbg3(ctx, DBG_CT_LOOKUP4_1, tuple->saddr, tuple->daddr,
                (bpf_ntohs(tuple->sport) << 16) | bpf_ntohs(tuple->dport));
    cilium_dbg3(ctx, DBG_CT_LOOKUP4_2, (tuple->nexthdr << 8) | tuple->flags, 0, 0);
#endif
    ret = __ct_lookup(map, ctx, tuple, action, dir, ct_state, is_tcp, tcp_flags, monitor);
    if (ret != CT_NEW) {
        if (likely(ret == CT_ESTABLISHED || ret == CT_REOPENED)) {
            if (unlikely(tuple->flags & TUPLE_F_RELATED))
                ret = CT_RELATED;
            else
                ret = CT_REPLY;
        }
        goto out;
    }

//    relax_verifier();

    /* Lookup entry in forward direction */
    if (dir != CT_SERVICE) {
        ipv4_ct_tuple_reverse(tuple);
        ret = __ct_lookup(map, ctx, tuple, action, dir, ct_state, is_tcp, tcp_flags, monitor);
    }

out:
    cilium_dbg(ctx, DBG_CT_VERDICT, ret < 0 ? -ret : ret, ct_state->rev_nat_index);
    if (conn_is_dns(tuple->dport))
        *monitor = MTU;
    return ret;
}

static __always_inline int ct_create4(const void *map_main,
                                      const void *map_related,
                                      struct ipv4_ct_tuple *tuple,
                                      struct __ctx_buff *ctx, const int dir,
                                      const struct ct_state *ct_state,
                                      bool proxy_redirect)
{
    /* Create entry in original direction */
    struct ct_entry entry = { };
    bool is_tcp = tuple->nexthdr == IPPROTO_TCP;
    union tcp_flags seen_flags = { .value = 0 };

    /* Note if this is a proxy connection so that replies can be redirected
     * back to the proxy.
     */
    entry.proxy_redirect = proxy_redirect;
    entry.lb_loopback = ct_state->loopback;
    entry.node_port = ct_state->node_port;
//    relax_verifier();
    entry.dsr = ct_state->dsr;
    entry.ifindex = ct_state->ifindex;

    /* Previously, the rx_bytes field was not used for entries with
     * the dir=CT_SERVICE (see GH#7060). Therefore, we can safely abuse
     * this field to save the backend_id.
     */
    if (dir == CT_SERVICE)
        entry.rx_bytes = ct_state->backend_id;
    entry.rev_nat_index = ct_state->rev_nat_index;
    seen_flags.value |= is_tcp ? TCP_FLAG_SYN : 0;
    ct_update_timeout(&entry, is_tcp, dir, seen_flags);

    if (dir == CT_INGRESS) {
        entry.rx_packets = 1;
        entry.rx_bytes = ctx_full_len(ctx);
    } else if (dir == CT_EGRESS) {
        entry.tx_packets = 1;
        entry.tx_bytes = ctx_full_len(ctx);
    }

#ifdef ENABLE_NAT46
    if (ctx_load_meta(ctx, CB_NAT46_STATE) == NAT64)
		entry.nat46 = dir == CT_EGRESS;
#endif

    cilium_dbg3(ctx, DBG_CT_CREATED4, entry.rev_nat_index,
                ct_state->src_sec_id, ct_state->addr);

    entry.src_sec_id = ct_state->src_sec_id;
    if (map_update_elem(map_main, tuple, &entry, 0) < 0) {
        send_signal_ct_fill_up(ctx, SIGNAL_PROTO_V4);
        return DROP_CT_CREATE_FAILED;
    }

    if (ct_state->addr && ct_state->loopback) {
        __u8 flags = tuple->flags;
        __be32 saddr, daddr;

        saddr = tuple->saddr;
        daddr = tuple->daddr;

        /* We are looping back into the origin endpoint through a
         * service, set up a conntrack tuple for the reply to ensure we
         * do rev NAT before attempting to route the destination
         * address which will not point back to the right source.
         */
        tuple->flags = TUPLE_F_IN;
        if (dir == CT_INGRESS) {
            tuple->saddr = ct_state->addr;
            tuple->daddr = ct_state->svc_addr;
        } else {
            tuple->saddr = ct_state->svc_addr;
            tuple->daddr = ct_state->addr;
        }

        if (map_update_elem(map_main, tuple, &entry, 0) < 0) {
            send_signal_ct_fill_up(ctx, SIGNAL_PROTO_V4);
            return DROP_CT_CREATE_FAILED;
        }

        tuple->saddr = saddr;
        tuple->daddr = daddr;
        tuple->flags = flags;
    }

    if (map_related != NULL) {
        /* Create an ICMP entry to relate errors */
        struct ipv4_ct_tuple icmp_tuple = {
                .daddr = tuple->daddr,
                .saddr = tuple->saddr,
                .nexthdr = IPPROTO_ICMP,
                .sport = 0,
                .dport = 0,
                .flags = tuple->flags | TUPLE_F_RELATED,
        };

        entry.seen_non_syn = true; /* For ICMP, there is no SYN. */
        /* Previous map update succeeded, we could delete it in case
         * the below throws an error, but we might as well just let
         * it time out.
         */
        if (map_update_elem(map_related, &icmp_tuple, &entry, 0) < 0) {
            send_signal_ct_fill_up(ctx, SIGNAL_PROTO_V4);
            return DROP_CT_CREATE_FAILED;
        }
    }
    return 0;
}



#endif //XDP_CILIUM_L4LB_CONNTRACK_H
