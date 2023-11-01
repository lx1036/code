

// /root/linux-5.10.142/include/uapi/linux/tcp.h

#ifndef _UAPI_LINUX_TCP_H
#define _UAPI_LINUX_TCP_H

#include <linux/types.h>

struct tcphdr {
	__be16	source;
	__be16	dest;
	__be32	seq;
	__be32	ack_seq;

	__u16	doff:4,
		res1:4,
		cwr:1,
		ece:1,
		urg:1,
		ack:1,
		psh:1,
		rst:1,
		syn:1,
		fin:1;
        
	__be16	window;
	__sum16	check;
	__be16	urg_ptr;
};




#endif /* _UAPI_LINUX_TCP_H */


