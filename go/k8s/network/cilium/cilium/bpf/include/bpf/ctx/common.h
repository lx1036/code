//
// Created by 刘祥 on 7/1/22.
//

#ifndef __BPF_CTX_COMMON_H_
#define __BPF_CTX_COMMON_H_

#include "../compiler.h"
#include "skb.h"


static __always_inline void *ctx_data(const struct __ctx_buff *ctx)
{
    return (void *)(unsigned long)ctx->data;
}

static __always_inline void *ctx_data_end(const struct __ctx_buff *ctx)
{
    return (void *)(unsigned long)ctx->data_end;
}


#endif //__BPF_CTX_COMMON_H_
