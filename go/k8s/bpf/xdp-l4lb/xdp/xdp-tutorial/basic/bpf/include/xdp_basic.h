


#ifndef __XDP_BASIC_H
#define __XDP_BASIC_H


#include <linux/types.h>
#include <linux/bpf.h>

struct datarec {
    __u64 rx_packets;
};



#ifndef XDP_ACTION_MAX
#define XDP_ACTION_MAX (XDP_REDIRECT + 1) // 4+1
#endif

#endif



