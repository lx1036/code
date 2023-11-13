



#ifndef __BPF_CTX_XDP_H_
#define __BPF_CTX_XDP_H_


#include <linux/if_ether.h>

#define __ctx_buff			xdp_md
#define __ctx_is			__ctx_xdp


#include "common.h"
#include "../helpers_xdp.h"





#define CTX_ACT_OK			XDP_PASS
#define CTX_ACT_DROP		XDP_DROP
#define CTX_ACT_TX			XDP_TX	/* hairpin only */

#define CTX_DIRECT_WRITE_OK		1

					/* cb + RECIRC_MARKER + XFER_MARKER */
#define META_PIVOT			((int)(field_sizeof(struct __sk_buff, cb) + \
					       sizeof(__u32) * 2))

/* This must be a mask and all offsets guaranteed to be less than that. */
#define __CTX_OFF_MAX			0xff

#define ctx_get_tunnel_key		xdp_get_tunnel_key__stub
#define ctx_set_tunnel_key		xdp_set_tunnel_key__stub
#define ctx_event_output		xdp_event_output
#define ctx_adjust_meta			xdp_adjust_meta


struct bpf_elf_map __section_maps cilium_xdp_scratch = {
	.type		= BPF_MAP_TYPE_PERCPU_ARRAY,
	.size_key	= sizeof(int),
	.size_value	= META_PIVOT,
	.pinning	= PIN_GLOBAL_NS,
	.max_elem	= 1,
};

static __always_inline __maybe_unused void
ctx_store_meta(struct xdp_md *ctx __maybe_unused, const __u64 off, __u32 datum)
{
	__u32 zero = 0;
	__u32 *data_meta = map_lookup_elem(&cilium_xdp_scratch, &zero);

	if (always_succeeds(data_meta))
		data_meta[off] = datum;
	build_bug_on((off + 1) * sizeof(__u32) > META_PIVOT);
}

static __always_inline __maybe_unused __u32
ctx_load_meta(const struct xdp_md *ctx __maybe_unused, const __u64 off)
{
	__u32 zero = 0, *data_meta = map_lookup_elem(&cilium_xdp_scratch, &zero);

	if (always_succeeds(data_meta))
		return data_meta[off];
	build_bug_on((off + 1) * sizeof(__u32) > META_PIVOT);
	return 0;
}




#endif /* __BPF_CTX_XDP_H_ */