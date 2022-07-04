//
// Created by 刘祥 on 7/1/22.
//

#ifndef BPF_IP_H
#define BPF_IP_H

#include <linux/types.h>


struct iphdr {
#if defined(__LITTLE_ENDIAN_BITFIELD)
    __u8	ihl:4,
		version:4;
#elif defined (__BIG_ENDIAN_BITFIELD)
    __u8	version:4,
  		ihl:4;
#else
#endif

    __u8	tos;
    __be16	tot_len;
    __be16	id;
    __be16	frag_off;
    __u8	ttl;
    __u8	protocol;
    __sum16	check;
    __be32	saddr;
    __be32	daddr;
    /*The options start here. */
};


#endif //BPF_IP_H
