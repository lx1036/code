


#ifndef __CSUM_HELPERS_H
#define __CSUM_HELPERS_H

#include <stddef.h>

#include <linux/bpf.h>
#include <linux/types.h>
#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>

#include <loadbalancer/balancer_consts.h>



__attribute__((__always_inline__)) static inline __u16 csum_fold_helper(__u64 csum) {
    int i;
#pragma unroll
    for (i = 0; i < 4; i++) {
        if (csum >> 16)
        csum = (csum & 0xffff) + (csum >> 16);
    }
    return ~csum;
}

__attribute__((__always_inline__)) static inline void ipv4_csum_inline(void* iph, __u64* csum) {
    __u16* next_iph_u16 = (__u16*)iph;
#pragma clang loop unroll(full)
    for (int i = 0; i < sizeof(struct iphdr) >> 1; i++) {
        *csum += *next_iph_u16++;
    }
    *csum = csum_fold_helper(*csum);
}


#endif // of __CSUM_HELPERS_H
