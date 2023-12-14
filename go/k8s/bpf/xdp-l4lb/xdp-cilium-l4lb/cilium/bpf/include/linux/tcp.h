

// /root/linux-5.10.142/include/uapi/linux/tcp.h

#ifndef _UAPI_LINUX_TCP_H
#define _UAPI_LINUX_TCP_H

#include <linux/types.h>

// https://datatracker.ietf.org/doc/html/rfc9293#name-header-format
// 2+2+4+4+2+2+2+2=20
struct tcphdr {
	__be16	source; // src port
	__be16	dest; // dst port
	__be32	seq; // Sequence Number
	__be32	ack_seq; // Acknowledgment Number

	__u16	doff:4, // Data Offset
		res1:4, // Rsrvd
		cwr:1, //
		ece:1,
		urg:1,
		ack:1,
		psh:1,
		rst:1,
		syn:1,
		fin:1;
        
	__be16	window; // Window
	__sum16	check; // Checksum
	__be16	urg_ptr; // Urgent Pointer
};




#endif /* _UAPI_LINUX_TCP_H */


