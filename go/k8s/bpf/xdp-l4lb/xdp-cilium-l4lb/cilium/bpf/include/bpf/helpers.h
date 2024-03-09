



#ifndef __BPF_HELPERS__
#define __BPF_HELPERS__


#include <linux/bpf.h>

#include "ctx/ctx.h"
#include "compiler.h"

#ifndef BPF_FUNC
# define BPF_FUNC(NAME, ...)						\
	(* NAME)(__VA_ARGS__) __maybe_unused = (void *)BPF_FUNC_##NAME
#endif

#ifndef BPF_STUB
# define BPF_STUB(NAME, ...)						\
	(* NAME##__stub)(__VA_ARGS__) __maybe_unused = (void *)((__u32)-1)
#endif

#ifndef BPF_FUNC_REMAP
# define BPF_FUNC_REMAP(NAME, ...)					\
	(* NAME)(__VA_ARGS__) __maybe_unused
#endif


#if __ctx_is == __ctx_skb
# include "helpers_skb.h"
#else
# include "helpers_xdp.h"
#endif

/* Padded struct so the dmac at the end can be passed to another helper
 * e.g. as a map value buffer. Otherwise verifier will trip over it with
 * 'invalid indirect read from stack off'.
 */
struct bpf_fib_lookup_padded {
    struct bpf_fib_lookup l;
    __u8 pad[2];
};

/* Map access/manipulation */
static void *BPF_FUNC(map_lookup_elem, const void *map, const void *key);
static int BPF_FUNC(map_update_elem, const void *map, const void *key, const void *value, __u32 flags);
static int BPF_FUNC(map_delete_elem, const void *map, const void *key);

/* Tail calls */
static void BPF_FUNC(tail_call, void *ctx, const void *map, __u32 index);

/* Random numbers */
// u32 bpf_get_prandom_u32(void)
static __u32 BPF_FUNC(get_prandom_u32);

/* Routing helpers */
static int BPF_FUNC(fib_lookup, void *ctx, struct bpf_fib_lookup *params, __u32 plen, __u32 flags);

/* Checksumming */
static int BPF_FUNC_REMAP(csum_diff_external, const void *from, __u32 size_from,
                          const void *to, __u32 size_to, __u32 seed) = (void *)BPF_FUNC_csum_diff;




#endif /* __BPF_HELPERS__ */
