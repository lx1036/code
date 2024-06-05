
cmake_minimum_required(VERSION 3.26)
project(cilium_l4lb C)

set(CMAKE_C_STANDARD 11)

include_directories(linux)




add_executable(xdp_cilium_l4lb_test
        l4lb/l4lb-tc-tunnel-cilium/bpf_xdp_veth_host.c
        l4lb/l4lb-tc-tunnel-cilium/test_tc_tunnel.c
        tunnel/test_tunnel_kern.c
        linux/linux/ip.h
        tcp/tcp-syncookie/tcp_syncookie.c
        linux/bpf/bpf_helper_defs.h
        tproxy/socket-service/tcphdr/tcpnotify/tcpnotify.c
        tcp/tcpconnect/tcp_connect.c
        tcp/tcpconnect/tcp_connect.c
        tcp/tcpconnect/tcp_connect.h
        xdp-l4lb/xdp-vlan/test_xdp_vlan.c
        linux/linux/if_vlan.h
        xdp-l4lb/xdp-ping/xdp_ping.c
        linux/linux/if_packet.h
        tunnel/veth-pair/xdp-veth-pair/xdp_veth_pair.c
        xdp-l4lb/xdp-redirect/xdp_redirect.c
        xdp-l4lb/xdp-meta/xdp_meta.c
        xdp-l4lb/xdp/xdp.c
        tunnel/ipip/tc_l2_redirect_kern.c
        tunnel/ipip/ipip-tunnel/test_tunnel_kern.c
        tunnel/vxlan/vxlan-tunnel/test_tunnel_kern.c
        tunnel/veth-pair/pod-to-pod/veth_ingress.c
        tunnel/veth-pair/tc-veth-pair/test_tc_peer.c
        tunnel/veth-pair/tc-veth-pair/test_tc_neigh.c
        tunnel/veth-pair/tc-veth-pair/test_tc_neigh_fib.c
        tproxy/socket-service/tcphdr/tcprtt/demo1/tcprtt_sockops.c
        tproxy/socket-service/tcphdr/tcprtt/demo2/tcprtt_sockops.c
        tunnel/vxlan/pod-to-pod/vxlan.c
        l4lb/l4lb-tc/test_l4lb.c
        l4lb/l4lb-nodeport/tc_nodeport.c
        l4lb/slb-xdp/slb.bpf.c
        tproxy/socket-service/tcphdr/test_tcp_hdr_options.c
        xdp-acl/acl.c
        tproxy/socket-assign/test_sk_assign.c
        tproxy/socket-lookup/test_sk_lookup.c
        tproxy/socket-lookup/main_test.c
        tproxy/socket-reuseport/test_sk_reuseport.c
        xdp-l4lb/xdp-socket/dumpframes/protocol_filter.c
        tproxy/socket-reuseport/epoll_test.c
        tproxy/bpf-map-in-map/test_map_in_map.c
        tproxy/bpf-tailcall/tailcall1.c
        tproxy/socket-service/connect-force-port/connect_force_port4.c
        tproxy/socket-service/connect-sendmsg/connect_send_recvmsg.c
        tproxy/socket-service/sockopt/test_sockopt.c
        tproxy/socket-service/tcphdr/rewrite-constant/test_rewrite_const.c
        tproxy/socket-service/udp-sk-limit/test_udp_sk_limit.c
        tproxy/socket-service/tcphdr/tcp-socket-cookie/test_tcp_socket_cookie.c
        tproxy/socket-service/tcphdr/tcpfsm/test_tcp_fsm.c
        tproxy/socket-lookup/socket-lookup-cgroup/socket_lookup_cgroup.c
        tproxy/socket-lookup/socket-lookup-cgroup/linux_tcp.h
        tproxy/socket-lookup/socket-lookup/test_sk_lookup.c
        tproxy/socket-bypass-tcpip/test_socket_bypass_tcpip.c
        tproxy/socket-filter/test_socket_filter.c


)
