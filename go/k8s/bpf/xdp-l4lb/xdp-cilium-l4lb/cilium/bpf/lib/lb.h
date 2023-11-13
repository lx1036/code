



#ifndef __LB_H_
#define __LB_H_

#include "csum.h"
#include "conntrack.h"
#include "ipv4.h"
#include "hash.h"
#include "ids.h"

#ifdef ENABLE_IPV4
struct bpf_elf_map __section_maps LB4_SERVICES_MAP_V2 = {
	.type		= BPF_MAP_TYPE_HASH,
	.size_key	= sizeof(struct lb4_key),
	.size_value	= sizeof(struct lb4_service),
	.pinning	= PIN_GLOBAL_NS,
	.max_elem	= CILIUM_LB_MAP_MAX_ENTRIES,
	.flags		= CONDITIONAL_PREALLOC,
};

#ifdef ENABLE_SESSION_AFFINITY
struct bpf_elf_map __section_maps LB4_AFFINITY_MAP = {
	.type		= BPF_MAP_TYPE_LRU_HASH,
	.size_key	= sizeof(struct lb4_affinity_key),
	.size_value	= sizeof(struct lb_affinity_val),
	.pinning	= PIN_GLOBAL_NS,
	.max_elem	= CILIUM_LB_MAP_MAX_ENTRIES,
};
#endif

#endif /* ENABLE_IPV4 */


static __always_inline 
bool lb4_svc_is_hostport(const struct lb4_service *svc __maybe_unused) {
#ifdef ENABLE_HOSTPORT
	return svc->flags & SVC_FLAG_HOSTPORT;
#else
	return false;
#endif /* ENABLE_HOSTPORT */
}

static __always_inline
bool lb4_svc_is_local_scope(const struct lb4_service *svc)
{
	return svc->flags & SVC_FLAG_LOCAL_SCOPE;
}

static __always_inline
bool lb4_svc_has_src_range_check(const struct lb4_service *svc __maybe_unused)
{
#ifdef ENABLE_SRC_RANGE_CHECK
	return svc->flags & SVC_FLAG_SOURCE_RANGE;
#else
	return false;
#endif /* ENABLE_SRC_RANGE_CHECK */
}



#ifdef ENABLE_IPV4

// add source range check
static __always_inline
bool lb4_src_range_ok(const struct lb4_service *svc __maybe_unused, __u32 saddr __maybe_unused) {
#ifdef ENABLE_SRC_RANGE_CHECK
	struct lb4_src_range_key key;

	if (!lb4_svc_has_src_range_check(svc))
		return true;

	key = (typeof(key)) {
		.lpm_key = { SRC_RANGE_STATIC_PREFIX(key), {} },
		.rev_nat_id = svc->rev_nat_index,
		.addr = saddr,
	};

	if (map_lookup_elem(&LB4_SRC_RANGE_MAP, &key))
		return true;

	return false;
#else
	return true;
#endif /* ENABLE_SRC_RANGE_CHECK */
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




#endif /* ENABLE_IPV4 */


#endif /* __LB_H_ */
