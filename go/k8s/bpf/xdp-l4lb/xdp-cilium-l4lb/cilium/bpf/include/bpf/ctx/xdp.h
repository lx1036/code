



#ifndef __BPF_CTX_XDP_H_
#define __BPF_CTX_XDP_H_


#include <linux/if_ether.h>

#define __ctx_buff			xdp_md
#define __ctx_is			__ctx_xdp


#include "common.h"
#include "../helpers_xdp.h"
#include "../builtins.h"
#include "../section.h"
#include "../loader.h"
#include "../csum.h"




#define CTX_ACT_OK			XDP_PASS
#define CTX_ACT_DROP		XDP_DROP
#define CTX_ACT_TX			XDP_TX	/* hairpin only */

#define CTX_DIRECT_WRITE_OK		1

// xdp_md.cb + 4字节*2
/* cb + RECIRC_MARKER + XFER_MARKER */
#define META_PIVOT	((int)(field_sizeof(struct xdp_md, cb) + sizeof(__u32) * 2))

/* This must be a mask and all offsets guaranteed to be less than that. */
#define __CTX_OFF_MAX			0xff

#define ctx_get_tunnel_key		xdp_get_tunnel_key__stub
#define ctx_set_tunnel_key		xdp_set_tunnel_key__stub
#define ctx_event_output		xdp_event_output
#define ctx_adjust_meta			xdp_adjust_meta

#define ctx_pull_data(ctx, ...)		do { /* Already linear. */ } while (0)


static __always_inline __maybe_unused int
xdp_load_bytes(const struct xdp_md *ctx, __u64 off, void *to, const __u64 len)
{
    void *from;
    int ret;
    /* LLVM tends to generate code that verifier doesn't understand,
     * so force it the way we want it in order to open up a range
     * on the reg.
     */
    asm volatile("r1 = *(u32 *)(%[ctx] +0)\n\t"
                 "r2 = *(u32 *)(%[ctx] +4)\n\t"
                 "%[off] &= %[offmax]\n\t"
                 "r1 += %[off]\n\t"
                 "%[from] = r1\n\t"
                 "r1 += %[len]\n\t"
                 "if r1 > r2 goto +2\n\t"
                 "%[ret] = 0\n\t"
                 "goto +1\n\t"
                 "%[ret] = %[errno]\n\t"
            : [ret]"=r"(ret), [from]"=r"(from)
    : [ctx]"r"(ctx), [off]"r"(off), [len]"ri"(len),
    [offmax]"i"(__CTX_OFF_MAX), [errno]"i"(-EINVAL)
    : "r1", "r2");
    if (!ret)
        memcpy(to, from, len);
    return ret;
}

static __always_inline __maybe_unused int
xdp_store_bytes(const struct xdp_md *ctx, __u64 off, const void *from,
                const __u64 len, __u64 flags __maybe_unused)
{
    void *to;
    int ret;
    /* See xdp_load_bytes(). */
    asm volatile("r1 = *(u32 *)(%[ctx] +0)\n\t"
                 "r2 = *(u32 *)(%[ctx] +4)\n\t"
                 "%[off] &= %[offmax]\n\t"
                 "r1 += %[off]\n\t"
                 "%[to] = r1\n\t"
                 "r1 += %[len]\n\t"
                 "if r1 > r2 goto +2\n\t"
                 "%[ret] = 0\n\t"
                 "goto +1\n\t"
                 "%[ret] = %[errno]\n\t"
            : [ret]"=r"(ret), [to]"=r"(to)
    : [ctx]"r"(ctx), [off]"r"(off), [len]"ri"(len),
    [offmax]"i"(__CTX_OFF_MAX), [errno]"i"(-EINVAL)
    : "r1", "r2");
    if (!ret)
        memcpy(to, from, len);
    return ret;
}

#define ctx_load_bytes			xdp_load_bytes
#define ctx_store_bytes			xdp_store_bytes

struct bpf_elf_map __section_maps cilium_xdp_scratch = {
	.type		= BPF_MAP_TYPE_PERCPU_ARRAY,
	.size_key	= sizeof(int),
	.size_value	= META_PIVOT,
	.pinning	= PIN_GLOBAL_NS,
	.max_elem	= 1,
};

static __always_inline __maybe_unused void
ctx_store_meta(struct xdp_md *ctx __maybe_unused, const __u64 off, __u32 datum) {
	__u32 zero = 0;
	__u32 *data_meta = map_lookup_elem(&cilium_xdp_scratch, &zero);

	if (always_succeeds(data_meta))
		data_meta[off] = datum;
	build_bug_on((off + 1) * sizeof(__u32) > META_PIVOT);
}

static __always_inline __maybe_unused __u32
ctx_load_meta(const struct xdp_md *ctx __maybe_unused, const __u64 off) {
	__u32 zero = 0, *data_meta = map_lookup_elem(&cilium_xdp_scratch, &zero);

	if (always_succeeds(data_meta))
		return data_meta[off];
	build_bug_on((off + 1) * sizeof(__u32) > META_PIVOT);
	return 0;
}

static __always_inline __maybe_unused __u32
ctx_get_protocol(const struct xdp_md *ctx) {
	void *data_end = ctx_data_end(ctx);
	struct ethhdr *eth = ctx_data(ctx);

	if (ctx_no_room(eth + 1, data_end))
		return 0;

	return eth->h_proto; // 还是从二层头里读取 protocol
}


#endif /* __BPF_CTX_XDP_H_ */
