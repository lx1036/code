

#ifndef XDP_CILIUM_L4LB_FEATURES_SKB_H
#define XDP_CILIUM_L4LB_FEATURES_SKB_H



#include "features.h"

/* Only skb related features. */

#if HAVE_PROG_TYPE_HELPER(sched_cls, bpf_skb_change_tail)
# define BPF_HAVE_CHANGE_TAIL 1
#endif

#if HAVE_PROG_TYPE_HELPER(sched_cls, bpf_fib_lookup)
# define BPF_HAVE_FIB_LOOKUP 1
#endif







#endif //XDP_CILIUM_L4LB_FEATURES_SKB_H
