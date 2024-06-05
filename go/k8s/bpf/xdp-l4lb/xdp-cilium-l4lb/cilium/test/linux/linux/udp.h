



#ifndef _UAPI_LINUX_UDP_H
#define _UAPI_LINUX_UDP_H

#include <linux/types.h>


struct udphdr {
	__be16	source;
	__be16	dest;
	__be16	len;
	__sum16	check;
};


#endif /* _UAPI_LINUX_UDP_H */