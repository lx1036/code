



#ifndef __LIB_COMMON_H_
#define __LIB_COMMON_H_




/* Value of endpoint map */
struct endpoint_info {
	__u32		ifindex;
	__u16		unused; /* used to be sec_label, no longer used */
	__u16           lxc_id;
	__u32		flags;
	mac_t		mac;
	mac_t		node_mac;
	__u32		pad[4];
};





#include "overloadable.h"

#endif /* __LIB_COMMON_H_ */