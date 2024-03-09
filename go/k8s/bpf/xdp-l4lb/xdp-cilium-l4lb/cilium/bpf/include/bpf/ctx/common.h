



#ifndef __BPF_CTX_COMMON_H_
#define __BPF_CTX_COMMON_H_

#include <linux/types.h>
#include <linux/bpf.h>

#include "../compiler.h"
#include "../errno.h"


#define __ctx_skb		1
#define __ctx_xdp		2



static __always_inline void *ctx_data(const struct xdp_md *ctx)
{
	return (void *)(unsigned long)ctx->data;
}

static __always_inline void *ctx_data_meta(const struct xdp_md *ctx)
{
	return (void *)(unsigned long)ctx->data_meta;
}

static __always_inline void *ctx_data_end(const struct xdp_md *ctx)
{
	return (void *)(unsigned long)ctx->data_end;
}

static __always_inline bool ctx_no_room(const void *needed, const void *limit)
{
	return unlikely(needed > limit);
}



/////////////////////补充定义///////////////////////////

/*
 * Helper macro to place programs, maps, license in
 * different sections in elf_bpf file. Section names
 * are interpreted by elf_bpf loader
 */
#define SEC(NAME) __attribute__((section(NAME), used))



/////////////////////补充定义///////////////////////////


#endif /* __BPF_CTX_COMMON_H_ */