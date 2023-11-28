

#ifndef __LIB_ETH__
#define __LIB_ETH__

#include <linux/if_ether.h>

#ifndef ETH_HLEN
#define ETH_HLEN __ETH_HLEN
#endif


union macaddr {
    struct {
        __u32 p1;
        __u16 p2;
    };
    __u8 addr[6];
};




#endif /* __LIB_ETH__ */
