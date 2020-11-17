
bird进程作用(C语言写的)：https://github.com/projectcalico/bird/tree/master

bird项目介绍：https://github.com/projectcalico/bird/blob/feature-ipinip/BIRD-README
The BIRD project aims to develop a dynamic IP routing daemon with full support
of all modern routing protocols.

What do we support
==================

	o  Both IPv4 and IPv6 (use --enable-ipv6 when configuring)
	o  Multiple routing tables
	o  Border Gateway Protocol (BGPv4)
	o  Routing Information Protocol (RIPv2, RIPng)
	o  Open Shortest Path First protocol (OSPFv2, OSPFv3)
	o  Babel Routing Protocol (Babel)
	o  Bidirectional Forwarding Detection (BFD)
	o  IPv6 router advertisements
	o  Static routes
	o  Inter-table protocol
	o  Command-line interface allowing on-line control and inspection of
	   status of the daemon
	o  Soft reconfiguration, no need to use complex online commands to
	   change the configuration, just edit the configuration file and notify
	   BIRD to re-read it and it will smoothly switch itself to the new
	   configuration, not disturbing routing protocols unless they are
	   affected by the configuration changes
	o  Powerful language for route filtering, see doc/bird.conf.example
	o  Linux, FreeBSD, NetBSD and OpenBSD ports
