



#ifndef __LIB_MAPS_H_
#define __LIB_MAPS_H_

#include "common.h"
#include "ipv6.h"
#include "ids.h"

#include "bpf/compiler.h"


struct bpf_elf_map __section_maps ENDPOINTS_MAP = {
	.type		= BPF_MAP_TYPE_HASH,
	.size_key	= sizeof(struct endpoint_key),
	.size_value	= sizeof(struct endpoint_info),
	.pinning	= PIN_GLOBAL_NS,
	.max_elem	= ENDPOINTS_MAP_SIZE,
	.flags		= CONDITIONAL_PREALLOC,
};



#ifndef SKIP_CALLS_MAP

// CALLS_MAP 在 xdp.go 里定义为 "cilium_calls_xdp" map

/* Private per EP map for internal tail calls */
struct bpf_elf_map __section_maps CALLS_MAP = {
	.type		= BPF_MAP_TYPE_PROG_ARRAY,
	.id		= CILIUM_MAP_CALLS,
	.size_key	= sizeof(__u32),
	.size_value	= sizeof(__u32),
	.pinning	= PIN_GLOBAL_NS,
	.max_elem	= CILIUM_CALL_SIZE,
};
#endif /* SKIP_CALLS_MAP */



#ifndef SKIP_CALLS_MAP
static __always_inline void ep_tail_call(struct __ctx_buff *ctx, const __u32 index)
{
	tail_call_static(ctx, &CALLS_MAP, index);
}
#endif /* SKIP_CALLS_MAP */



#endif