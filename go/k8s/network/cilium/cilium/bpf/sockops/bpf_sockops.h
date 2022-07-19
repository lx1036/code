//
// Created by 刘祥 on 7/5/22.
//

#ifndef BPF_BPF_SOCKOPS_H
#define BPF_BPF_SOCKOPS_H


#include <linux/bpf.h>
#include "../lib/common.h"

/* Structure representing an L7 sock */
struct sock_key {
    union {
        struct {
            __u32		sip4;
            __u32		pad1;
            __u32		pad2;
            __u32		pad3;
        };
        union v6addr	sip6;
    };
    union {
        struct {
            __u32		dip4;
            __u32		pad4;
            __u32		pad5;
            __u32		pad6;
        };
        union v6addr	dip6;
    };
    __u8 family;
    __u8 pad7;
    __u16 pad8;
    __u32 sport;
    __u32 dport;
} __packed;








#endif //BPF_BPF_SOCKOPS_H
