



#ifndef __LB_H_
#define __LB_H_

#include "csum.h"
#include "conntrack.h"
#include "ipv4.h"
#include "hash.h"
#include "ids.h"
#include "l4.h"


#ifdef LB_DEBUG
#define cilium_dbg_lb cilium_dbg
#else
#define cilium_dbg_lb(a, b, c, d)
#endif


//#ifdef ENABLE_IPV4
struct bpf_elf_map __section_maps LB4_SERVICES_MAP_V2 = {
	.type		= BPF_MAP_TYPE_HASH,
	.size_key	= sizeof(struct lb4_key),
	.size_value	= sizeof(struct lb4_service),
	.pinning	= PIN_GLOBAL_NS,
	.max_elem	= CILIUM_LB_MAP_MAX_ENTRIES,
	.flags		= CONDITIONAL_PREALLOC,
};

//#ifdef ENABLE_SRC_RANGE_CHECK
struct bpf_elf_map __section_maps LB4_SRC_RANGE_MAP = {
        .type		= BPF_MAP_TYPE_LPM_TRIE,
        .size_key	= sizeof(struct lb4_src_range_key),
        .size_value	= sizeof(__u8),
        .pinning	= PIN_GLOBAL_NS,
        .max_elem	= LB4_SRC_RANGE_MAP_SIZE,
        .flags		= BPF_F_NO_PREALLOC,
};
//#endif

//#ifdef ENABLE_SESSION_AFFINITY
struct bpf_elf_map __section_maps LB4_AFFINITY_MAP = {
	.type		= BPF_MAP_TYPE_LRU_HASH,
	.size_key	= sizeof(struct lb4_affinity_key),
	.size_value	= sizeof(struct lb_affinity_val),
	.pinning	= PIN_GLOBAL_NS,
	.max_elem	= CILIUM_LB_MAP_MAX_ENTRIES,
};
//#endif

struct bpf_elf_map __section_maps LB4_BACKEND_MAP = {
        .type           = BPF_MAP_TYPE_HASH,
        .size_key       = sizeof(__u16),
        .size_value     = sizeof(struct lb4_backend),
        .pinning        = PIN_GLOBAL_NS,
        .max_elem       = CILIUM_LB_MAP_MAX_ENTRIES,
        .flags          = CONDITIONAL_PREALLOC,
};

//#endif /* ENABLE_IPV4 */




static __always_inline bool __lb_svc_is_routable(__u8 flags)
{
    return (flags & SVC_FLAG_ROUTABLE) != 0;
}

static __always_inline
bool lb4_svc_is_routable(const struct lb4_service *svc)
{
    return __lb_svc_is_routable(svc->flags);
}

static __always_inline 
bool lb4_svc_is_hostport(const struct lb4_service *svc __maybe_unused) {
#ifdef ENABLE_HOSTPORT
	return svc->flags & SVC_FLAG_HOSTPORT;
#else
	return false;
#endif /* ENABLE_HOSTPORT */
}

// 实现 k8s service 是不是 external 还是 local: externalTrafficPolicy=Local
static __always_inline
bool lb4_svc_is_local_scope(const struct lb4_service *svc) {
	return svc->flags & SVC_FLAG_LOCAL_SCOPE;
}

static __always_inline
bool lb4_svc_has_src_range_check(const struct lb4_service *svc __maybe_unused)
{
//#ifdef ENABLE_SRC_RANGE_CHECK
	return svc->flags & SVC_FLAG_SOURCE_RANGE;
//#else
//	return false;
//#endif /* ENABLE_SRC_RANGE_CHECK */
}


//#ifdef ENABLE_IPV4

// add source range check, Enable check of service source ranges (currently, only for LoadBalancer)
static __always_inline
bool lb4_src_range_ok(const struct lb4_service *svc __maybe_unused, __u32 saddr __maybe_unused) {
//#ifdef ENABLE_SRC_RANGE_CHECK

	if (!lb4_svc_has_src_range_check(svc))
		return true;

    // 指定初始化器（designated initializer）的语法来对结构体进行初始化
    struct lb4_src_range_key key = {
        .addr = saddr,
        .lpm_key = { SRC_RANGE_STATIC_PREFIX(key), {} },
        .rev_nat_id = svc->rev_nat_index,
    };

	if (map_lookup_elem(&LB4_SRC_RANGE_MAP, &key))
		return true;

	return false;
//#else
//	return true;
//#endif /* ENABLE_SRC_RANGE_CHECK */
}


static __always_inline struct lb4_service *lb4_lookup_service(struct lb4_key *key, const bool scope_switch) {
	struct lb4_service *svc;

	key->scope = LB_LOOKUP_SCOPE_EXT;
	key->backend_slot = 0;
	svc = map_lookup_elem(&LB4_SERVICES_MAP_V2, key); // cilium_lb4_services_v2
	if (svc) {
		if (!scope_switch || !lb4_svc_is_local_scope(svc))
			return svc->count ? svc : NULL;
		key->scope = LB_LOOKUP_SCOPE_INT;
		svc = map_lookup_elem(&LB4_SERVICES_MAP_V2, key);
		if (svc && svc->count)
			return svc;
	}

	return NULL;
}

// 获取 tcp header dst port 字段值
static __always_inline int extract_l4_port(struct __ctx_buff *ctx, __u8 nexthdr, int l4_off,
                                           int dir __maybe_unused, __be16 *port, __maybe_unused struct iphdr *ip4) {
    int ret;

    switch (nexthdr) {
        case IPPROTO_TCP:
        case IPPROTO_UDP:
//#ifdef ENABLE_IPV4_FRAGMENTS
            if (ip4) {
                struct ipv4_frag_l4ports ports = {};
                ret = ipv4_handle_fragmentation(ctx, ip4, l4_off, dir, &ports, NULL);
                if (IS_ERR(ret))
                    return ret;
                *port = ports.dport;
                break;
            }
//#endif
            /* Port offsets for UDP and TCP are the same */
            // 从字节数组 bytes 里取 port，可参考!!!
            ret = l4_load_port(ctx, l4_off + TCP_DPORT_OFF, port);
            if (IS_ERR(ret))
                return ret;
            break;

        case IPPROTO_ICMPV6:
        case IPPROTO_ICMP:
            /* No need to perform a service lookup for ICMP packets */
            return DROP_NO_SERVICE;

        default:
            /* Pass unknown L4 to stack */
            return DROP_UNKNOWN_L4;
    }

    return 0;
}

/** Extract IPv4 LB key from packet
 * @arg ctx		Packet
 * @arg ip4		Pointer to L3 header
 * @arg l4_off		Offset to L4 header
 * @arg key		Pointer to store LB key in
 * @arg csum_off	Pointer to store L4 checksum field offset  in
 * @arg dir		Flow direction
 *
 * Returns:
 *   - CTX_ACT_OK on successful extraction
 *   - DROP_UNKNOWN_L4 if packet should be ignore (sent to stack)
 *   - Negative error code
 */
static __always_inline int lb4_extract_key(struct __ctx_buff *ctx __maybe_unused,
                                           struct iphdr *ip4,
                                           int l4_off __maybe_unused,
                                           struct lb4_key *key,
                                           struct csum_offset *csum_off,
                                           int dir)
{
    /* FIXME: set after adding support for different L4 protocols in LB */
    key->proto = 0;
    key->address = (dir == CT_INGRESS) ? ip4->saddr : ip4->daddr;
    if (ipv4_has_l4_header(ip4))
        csum_l4_offset_and_flags(ip4->protocol, csum_off);

    return extract_l4_port(ctx, ip4->protocol, l4_off, dir, &key->dport, ip4);
}

static __always_inline struct lb4_backend *__lb4_lookup_backend(__u16 backend_id)
{
    return map_lookup_elem(&LB4_BACKEND_MAP, &backend_id);
}

static __always_inline struct lb4_backend *
lb4_lookup_backend(struct __ctx_buff *ctx __maybe_unused, __u16 backend_id)
{
    struct lb4_backend *backend;

    backend = __lb4_lookup_backend(backend_id);
    if (!backend)
        cilium_dbg_lb(ctx, DBG_LB4_LOOKUP_BACKEND_FAIL, backend_id, 0);

    return backend;
}

static __always_inline
struct lb4_service *__lb4_lookup_backend_slot(struct lb4_key *key)
{
    return map_lookup_elem(&LB4_SERVICES_MAP_V2, key);
}

static __always_inline
struct lb4_service *lb4_lookup_backend_slot(struct __ctx_buff *ctx __maybe_unused,
                                            struct lb4_key *key, __u16 slot)
{
    struct lb4_service *svc;

    key->backend_slot = slot;
    cilium_dbg_lb(ctx, DBG_LB4_LOOKUP_BACKEND_SLOT, key->backend_slot, key->dport);
    svc = __lb4_lookup_backend_slot(key);
    if (svc)
        return svc;

    cilium_dbg_lb(ctx, DBG_LB4_LOOKUP_BACKEND_SLOT_V2_FAIL, key->backend_slot, key->dport);

    return NULL;
}

/* Backend slot 0 is always reserved for the service frontend. */
#if LB_SELECTION == LB_SELECTION_RANDOM
// 随机数，还不是 rr 负载均衡算法
static __always_inline __u16
lb4_select_backend_id(struct __ctx_buff *ctx,
                      struct lb4_key *key,
                      const struct ipv4_ct_tuple *tuple __maybe_unused,
                      const struct lb4_service *svc)
{
    __u32 slot = (get_prandom_u32() % svc->count) + 1;
    struct lb4_service *be = lb4_lookup_backend_slot(ctx, key, slot);

    return be ? be->backend_id : 0;
}

// TODO: 添加 maglev 算法

#endif

static __always_inline int lb4_local(const void *map, struct __ctx_buff *ctx,
                                     int l3_off, int l4_off,
                                     struct csum_offset *csum_off,
                                     struct lb4_key *key,
                                     struct ipv4_ct_tuple *tuple,
                                     const struct lb4_service *svc,
                                     struct ct_state *state, __be32 saddr,
                                     bool has_l4_header,
                                     const bool skip_l3_xlate)
{
    __u32 monitor; /* Deliberately ignored; regular CT will determine monitoring. */
    __be32 new_saddr = 0, new_daddr;
    __u8 flags = tuple->flags;
    struct lb4_backend *backend;
    __u32 backend_id = 0;
    int ret;
//#ifdef ENABLE_SESSION_AFFINITY
    union lb4_affinity_client_id client_id = {
		.client_ip = saddr,
	};
//#endif
    ret = ct_lookup4(map, tuple, ctx, l4_off, CT_SERVICE, state, &monitor);
    switch (ret) {
        case CT_NEW:
//#ifdef ENABLE_SESSION_AFFINITY
            if (lb4_svc_is_affinity(svc)) {
			backend_id = lb4_affinity_backend_id_by_addr(svc, &client_id);
			if (backend_id != 0) {
				backend = lb4_lookup_backend(ctx, backend_id);
				if (backend == NULL)
					backend_id = 0;
			}
		}
//#endif
            if (backend_id == 0) {
                /* No CT entry has been found, so select a svc endpoint */
                backend_id = lb4_select_backend_id(ctx, key, tuple, svc);
                backend = lb4_lookup_backend(ctx, backend_id);
                if (backend == NULL)
                    goto drop_no_service;
            }

            state->backend_id = backend_id;
            state->rev_nat_index = svc->rev_nat_index;

            ret = ct_create4(map, NULL, tuple, ctx, CT_SERVICE, state, false);
            /* Fail closed, if the conntrack entry create fails drop
             * service lookup.
             */
            if (IS_ERR(ret))
                goto drop_no_service;
            goto update_state;
        case CT_REOPENED:
        case CT_ESTABLISHED:
        case CT_RELATED:
        case CT_REPLY:
            /* For backward-compatibility we need to update reverse NAT
             * index in the CT_SERVICE entry for old connections, as later
             * in the code we check whether the right backend is used.
             * Having it set to 0 would trigger a new backend selection
             * which would in many cases would pick a different backend.
             */
            if (unlikely(state->rev_nat_index == 0)) {
                state->rev_nat_index = svc->rev_nat_index;
                ct_update4_rev_nat_index(map, tuple, state);
            }
            break;
        default:
            goto drop_no_service;
    }

    /* If the CT_SERVICE entry is from a non-related connection (e.g.
     * endpoint has been removed, but its CT entries were not (it is
     * totally possible due to the bug in DumpReliablyWithCallback)),
     * then a wrong (=from unrelated service) backend can be selected.
     * To avoid this, check that reverse NAT indices match. If not,
     * select a new backend.
     */
    if (state->rev_nat_index != svc->rev_nat_index) {
//#ifdef ENABLE_SESSION_AFFINITY
        if (lb4_svc_is_affinity(svc))
			backend_id = lb4_affinity_backend_id_by_addr(svc,
								     &client_id);
//#endif
        if (!backend_id) {
            backend_id = lb4_select_backend_id(ctx, key, tuple, svc);
            if (!backend_id)
                goto drop_no_service;
        }

        state->backend_id = backend_id;
        ct_update4_backend_id(map, tuple, state);
        state->rev_nat_index = svc->rev_nat_index;
        ct_update4_rev_nat_index(map, tuple, state);
    }
    /* If the lookup fails it means the user deleted the backend out from
     * underneath us. To resolve this fall back to hash. If this is a TCP
     * session we are likely to get a TCP RST.
     */
    backend = lb4_lookup_backend(ctx, state->backend_id);
    if (!backend) {
        key->backend_slot = 0;
        svc = lb4_lookup_service(key, false);
        if (!svc)
            goto drop_no_service;
        backend_id = lb4_select_backend_id(ctx, key, tuple, svc);
        backend = lb4_lookup_backend(ctx, backend_id);
        if (!backend)
            goto drop_no_service;
        state->backend_id = backend_id;
        ct_update4_backend_id(map, tuple, state);
    }

    update_state:
    /* Restore flags so that SERVICE flag is only used in used when the
     * service lookup happens and future lookups use EGRESS or INGRESS.
     */
    tuple->flags = flags;
    state->rev_nat_index = svc->rev_nat_index;
    state->addr = new_daddr = backend->address;

//#ifdef ENABLE_SESSION_AFFINITY
    if (lb4_svc_is_affinity(svc))
		lb4_update_affinity_by_addr(svc, &client_id,
					    state->backend_id);
//#endif

#ifndef DISABLE_LOOPBACK_LB
    /* Special loopback case: The origin endpoint has transmitted to a
     * service which is being translated back to the source. This would
     * result in a packet with identical source and destination address.
     * Linux considers such packets as martian source and will drop unless
     * received on a loopback device. Perform NAT on the source address
     * to make it appear from an outside address.
     */
    if (saddr == backend->address) {
        new_saddr = IPV4_LOOPBACK;
        state->loopback = 1;
        state->addr = new_saddr;
        state->svc_addr = saddr;
    }

    if (!state->loopback)
#endif
        tuple->daddr = backend->address;

    return lb_skip_l4_dnat() ? CTX_ACT_OK :
           lb4_xlate(ctx, &new_daddr, &new_saddr, &saddr,
                     tuple->nexthdr, l3_off, l4_off, csum_off, key,
                     backend, has_l4_header, skip_l3_xlate);
    drop_no_service:
    tuple->flags = flags;
    return DROP_NO_SERVICE;
}

//#endif /* ENABLE_IPV4 */
#endif /* __LB_H_ */
