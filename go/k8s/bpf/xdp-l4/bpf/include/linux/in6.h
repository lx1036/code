


#ifndef _UAPI_LINUX_IN6_H
#define _UAPI_LINUX_IN6_H

#include <linux/types.h>



/*
 *	IPv6 address structure
 */

struct in6_addr {
	union {
		__u8		u6_addr8[16];
		__be16		u6_addr16[8];
		__be32		u6_addr32[4];
	} in6_u;
#define s6_addr			in6_u.u6_addr8
#define s6_addr16		in6_u.u6_addr16
#define s6_addr32		in6_u.u6_addr32
};




#endif /* _UAPI_LINUX_IN6_H */