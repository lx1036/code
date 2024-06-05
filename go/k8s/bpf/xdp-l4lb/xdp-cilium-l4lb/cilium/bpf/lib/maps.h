



#ifndef __LIB_MAPS_H_
#define __LIB_MAPS_H_

#include <stddef.h>
#include <linux/bpf.h>
#include <linux/types.h>
#include <bpf/bpf_helpers.h>


#include "common.h"


struct {
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
    __uint(max_entries, 128);
//    __uint(max_entries, __NR_CPUS__);
    __uint(key_size, sizeof(__u32));
    __uint(value_size, sizeof(__u32));
    __uint(pinning, LIBBPF_PIN_BY_NAME);
} cilium_events SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 4096);
    __uint(key_size, sizeof(struct ipv4_ct_tuple));
    __uint(value_size, sizeof(struct ct_entry));
    __uint(pinning, LIBBPF_PIN_BY_NAME);
#ifndef HAVE_LRU_HASH_MAP_TYPE
    __uint(map_flags, BPF_F_NO_PREALLOC);
#endif
} cilium_ct_tcp4 SEC(".maps");


//struct bpf_elf_map __section_maps ENDPOINTS_MAP = {
//	.type		= BPF_MAP_TYPE_HASH,
//	.size_key	= sizeof(struct endpoint_key),
//	.size_value	= sizeof(struct endpoint_info),
//	.pinning	= PIN_GLOBAL_NS,
//	.max_elem	= ENDPOINTS_MAP_SIZE,
//	.flags		= CONDITIONAL_PREALLOC,
//};
//
//
//
//#ifndef SKIP_CALLS_MAP
//
//// CALLS_MAP 在 xdp.go 里定义为 "cilium_calls_xdp" map
//
///* Private per EP map for internal tail calls */
//struct bpf_elf_map __section_maps CALLS_MAP = {
//	.type		= BPF_MAP_TYPE_PROG_ARRAY,
//	.id		= CILIUM_MAP_CALLS,
//	.size_key	= sizeof(__u32),
//	.size_value	= sizeof(__u32),
//	.pinning	= PIN_GLOBAL_NS,
//	.max_elem	= CILIUM_CALL_SIZE,
//};
//#endif /* SKIP_CALLS_MAP */
//
//
//struct ipcache_key {
//	struct bpf_lpm_trie_key lpm_key;
//	__u16 pad1;
//	__u8 pad2;
//	__u8 family;
//	union {
//		struct {
//			__u32		ip4;
//			__u32		pad4;
//			__u32		pad5;
//			__u32		pad6;
//		};
//		union v6addr	ip6;
//	};
//} __packed;
//
///* Global IP -> Identity map for applying egress label-based policy */
//// 实际上在用户态里定义为 "cilium_ipcache" bpf map
//struct bpf_elf_map __section_maps IPCACHE_MAP = {
//	.type		= LPM_MAP_TYPE,
//	.size_key	= sizeof(struct ipcache_key),
//	.size_value	= sizeof(struct remote_endpoint_info),
//	.pinning	= PIN_GLOBAL_NS,
//	.max_elem	= IPCACHE_MAP_SIZE,
//	.flags		= BPF_F_NO_PREALLOC,
//};



#ifndef SKIP_CALLS_MAP
static __always_inline void 
ep_tail_call(struct __ctx_buff *ctx, const __u32 index) {
	tail_call_static(ctx, &CALLS_MAP, index);
}
#endif /* SKIP_CALLS_MAP */



#endif
