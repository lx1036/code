//
// Created by 刘祥 on 7/1/22.
//

#ifndef __LIB_OVERLOADABLE_SKB_H_
#define __LIB_OVERLOADABLE_SKB_H_



static __always_inline __maybe_unused void bpf_clear_meta(struct __sk_buff *ctx) {
    __u32 zero = 0;

    ctx->cb[0] = zero;
    ctx->cb[1] = zero;
    ctx->cb[2] = zero;
    ctx->cb[3] = zero;
    ctx->cb[4] = zero;
}






#endif //__LIB_OVERLOADABLE_SKB_H_
