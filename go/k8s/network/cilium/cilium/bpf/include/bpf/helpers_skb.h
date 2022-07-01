//
// Created by 刘祥 on 7/1/22.
//

#ifndef BPF_HELPERS_SKB_H
#define BPF_HELPERS_SKB_H

#include <bpf/ctx/skb.h>
#include <bpf/types_mapper.h>

#include "helpers.h"


static int BPF_FUNC(skb_pull_data, struct __sk_buff *skb, __u32 len);


#endif //BPF_HELPERS_SKB_H
