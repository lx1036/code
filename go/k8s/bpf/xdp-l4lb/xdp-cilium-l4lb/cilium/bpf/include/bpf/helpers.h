



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

            





#endif /* __BPF_HELPERS__ */
