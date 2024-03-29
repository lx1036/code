#include "lib/utils.h"

/*
 cilium.v4.external.str 100.208.40.178
 cilium.v4.internal.str 100.216.152.173
 cilium.v4.nodeport.str [100.208.40.178]

 cilium.v4.internal.raw 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xff, 0xa, 0xd8, 0x98, 0xad
 */

#define NAT46_PREFIX { .addr = { 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xff, 0x0, 0x0, 0x0, 0x0 } }
#define CT_MAP_TCP4 cilium_ct4_global
#define CT_MAP_ANY4 cilium_ct_any4_global
#define CT_MAP_TCP6 cilium_ct6_global
#define CT_MAP_ANY6 cilium_ct_any6_global
#define CT_MAP_SIZE_TCP 589480
#define CT_MAP_SIZE_ANY 294740
#define ALLOW_ICMP_FRAG_NEEDED 1
#define CILIUM_IPV4_FRAG_MAP_MAX_ENTRIES 8192
#define CILIUM_LB_MAP_MAX_ENTRIES 65536
#define CT_CLOSE_TIMEOUT 10
#define CT_CONNECTION_LIFETIME_NONTCP 60
#define CT_CONNECTION_LIFETIME_TCP 21600
#define CT_REPORT_FLAGS 0x00ff
#define CT_REPORT_INTERVAL 5
#define CT_SERVICE_LIFETIME_NONTCP 60
#define CT_SERVICE_LIFETIME_TCP 21600
#define CT_SYN_TIMEOUT 60
#define DIRECT_ROUTING_DEV_IFINDEX 2
#define ENABLE_DSR 1
#define ENABLE_EXTERNAL_IP 1
#define ENABLE_HOSTPORT 1
#define ENABLE_HOST_SERVICES_FULL 1
#define ENABLE_HOST_SERVICES_TCP 1
#define ENABLE_HOST_SERVICES_UDP 1
#define ENABLE_IDENTITY_MARK 1
#define ENABLE_IPV4 1
#define ENABLE_IPV4_FRAGMENTS 1
#define ENABLE_LOADBALANCER 1
#define ENABLE_MASQUERADE 1
#define ENABLE_NODEPORT 1
#define ENABLE_NODEPORT_HAIRPIN 1
#define ENABLE_SERVICES 1
#define ENABLE_SESSION_AFFINITY 1
#define ENCRYPT_MAP cilium_encrypt_state
#define ENDPOINTS_MAP cilium_lxc
#define ENDPOINTS_MAP_SIZE 65535
#define EP_POLICY_MAP cilium_ep_to_policy
#define EVENTS_MAP cilium_events
#define HEALTH_ID 4
#define HOST_ID 1
#define INIT_ID 5
#define IPCACHE_MAP cilium_ipcache
#define IPCACHE_MAP_SIZE 512000
#define IPV4_DIRECT_ROUTING 2989019146
#define IPV4_FRAG_DATAGRAMS_MAP cilium_ipv4_frag_datagrams
#define IPV4_GATEWAY 0xad98d80a
#define IPV4_LOOPBACK 0x12afea9
#define IPV4_MASK 0xc0ffffff
#define IPV4_SNAT_EXCLUSION_DST_CIDR 0x98d80a
#define IPV4_SNAT_EXCLUSION_DST_CIDR_LEN 21
#define KERNEL_HZ 1000
#define LB4_AFFINITY_MAP cilium_lb4_affinity
#define LB4_BACKEND_MAP cilium_lb4_backends
#define LB4_REVERSE_NAT_MAP cilium_lb4_reverse_nat
#define LB4_REVERSE_NAT_SK_MAP cilium_lb4_reverse_sk
#define LB4_REVERSE_NAT_SK_MAP_SIZE 294740
#define LB4_SERVICES_MAP_V2 cilium_lb4_services_v2
#define LB6_BACKEND_MAP cilium_lb6_backends
#define LB6_REVERSE_NAT_MAP cilium_lb6_reverse_nat
#define LB6_REVERSE_NAT_SK_MAP cilium_lb6_reverse_sk
#define LB6_REVERSE_NAT_SK_MAP_SIZE 294740
#define LB6_SERVICES_MAP_V2 cilium_lb6_services_v2
#define LB_AFFINITY_MATCH_MAP cilium_lb_affinity_match
#define METRICS_MAP cilium_metrics
#define METRICS_MAP_SIZE 1024
#define MTU 1500
#define NODEPORT_NEIGH4 cilium_nodeport_neigh4
#define NODEPORT_NEIGH4_SIZE 589480
#define NODEPORT_PORT_MAX 32767
#define NODEPORT_PORT_MAX_NAT 65535
#define NODEPORT_PORT_MIN 30000
#define NODEPORT_PORT_MIN_NAT 32768
#define NO_REDIRECT 1
#define POLICY_CALL_MAP cilium_call_policy
#define POLICY_MAP_SIZE 16384
#define POLICY_PROG_MAP_SIZE 65535
#define REMOTE_NODE_ID 6
#define SIGNAL_MAP cilium_signals
#define SNAT_MAPPING_IPV4 cilium_snat_v4_external
#define SNAT_MAPPING_IPV4_SIZE 589480
#define SOCKOPS_MAP_SIZE 65535
#define TRACE_PAYLOAD_LEN 128ULL
#define TUNNEL_ENDPOINT_MAP_SIZE 65536
#define TUNNEL_MAP cilium_tunnel_map
#define UNMANAGED_ID 3
#define WORLD_ID 2

// JSON_OUTPUT: eyJBTExPV19JQ01QX0ZSQUdfTkVFREVEIjoiMSIsIkNJTElVTV9JUFY0X0ZSQUdfTUFQX01BWF9FTlRSSUVTIjoiODE5MiIsIkNJTElVTV9MQl9NQVBfTUFYX0VOVFJJRVMiOiI2NTUzNiIsIkNUX0NMT1NFX1RJTUVPVVQiOiIxMCIsIkNUX0NPTk5FQ1RJT05fTElGRVRJTUVfTk9OVENQIjoiNjAiLCJDVF9DT05ORUNUSU9OX0xJRkVUSU1FX1RDUCI6IjIxNjAwIiwiQ1RfUkVQT1JUX0ZMQUdTIjoiMHgwMGZmIiwiQ1RfUkVQT1JUX0lOVEVSVkFMIjoiNSIsIkNUX1NFUlZJQ0VfTElGRVRJTUVfTk9OVENQIjoiNjAiLCJDVF9TRVJWSUNFX0xJRkVUSU1FX1RDUCI6IjIxNjAwIiwiQ1RfU1lOX1RJTUVPVVQiOiI2MCIsIkRJUkVDVF9ST1VUSU5HX0RFVl9JRklOREVYIjoiMiIsIkVOQUJMRV9EU1IiOiIxIiwiRU5BQkxFX0VYVEVSTkFMX0lQIjoiMSIsIkVOQUJMRV9IT1NUUE9SVCI6IjEiLCJFTkFCTEVfSE9TVF9TRVJWSUNFU19GVUxMIjoiMSIsIkVOQUJMRV9IT1NUX1NFUlZJQ0VTX1RDUCI6IjEiLCJFTkFCTEVfSE9TVF9TRVJWSUNFU19VRFAiOiIxIiwiRU5BQkxFX0lERU5USVRZX01BUksiOiIxIiwiRU5BQkxFX0lQVjQiOiIxIiwiRU5BQkxFX0lQVjRfRlJBR01FTlRTIjoiMSIsIkVOQUJMRV9MT0FEQkFMQU5DRVIiOiIxIiwiRU5BQkxFX01BU1FVRVJBREUiOiIxIiwiRU5BQkxFX05PREVQT1JUIjoiMSIsIkVOQUJMRV9OT0RFUE9SVF9IQUlSUElOIjoiMSIsIkVOQUJMRV9TRVJWSUNFUyI6IjEiLCJFTkFCTEVfU0VTU0lPTl9BRkZJTklUWSI6IjEiLCJFTkNSWVBUX01BUCI6ImNpbGl1bV9lbmNyeXB0X3N0YXRlIiwiRU5EUE9JTlRTX01BUCI6ImNpbGl1bV9seGMiLCJFTkRQT0lOVFNfTUFQX1NJWkUiOiI2NTUzNSIsIkVQX1BPTElDWV9NQVAiOiJjaWxpdW1fZXBfdG9fcG9saWN5IiwiRVZFTlRTX01BUCI6ImNpbGl1bV9ldmVudHMiLCJIRUFMVEhfSUQiOiI0IiwiSE9TVF9JRCI6IjEiLCJJTklUX0lEIjoiNSIsIklQQ0FDSEVfTUFQIjoiY2lsaXVtX2lwY2FjaGUiLCJJUENBQ0hFX01BUF9TSVpFIjoiNTEyMDAwIiwiSVBWNF9ESVJFQ1RfUk9VVElORyI6IjI5ODkwMTkxNDYiLCJJUFY0X0ZSQUdfREFUQUdSQU1TX01BUCI6ImNpbGl1bV9pcHY0X2ZyYWdfZGF0YWdyYW1zIiwiSVBWNF9HQVRFV0FZIjoiMHhhZDk4ZDgwYSIsIklQVjRfTE9PUEJBQ0siOiIweDEyYWZlYTkiLCJJUFY0X01BU0siOiIweGMwZmZmZmZmIiwiSVBWNF9TTkFUX0VYQ0xVU0lPTl9EU1RfQ0lEUiI6IjB4OThkODBhIiwiSVBWNF9TTkFUX0VYQ0xVU0lPTl9EU1RfQ0lEUl9MRU4iOiIyMSIsIktFUk5FTF9IWiI6IjEwMDAiLCJMQjRfQUZGSU5JVFlfTUFQIjoiY2lsaXVtX2xiNF9hZmZpbml0eSIsIkxCNF9CQUNLRU5EX01BUCI6ImNpbGl1bV9sYjRfYmFja2VuZHMiLCJMQjRfUkVWRVJTRV9OQVRfTUFQIjoiY2lsaXVtX2xiNF9yZXZlcnNlX25hdCIsIkxCNF9SRVZFUlNFX05BVF9TS19NQVAiOiJjaWxpdW1fbGI0X3JldmVyc2Vfc2siLCJMQjRfUkVWRVJTRV9OQVRfU0tfTUFQX1NJWkUiOiIyOTQ3NDAiLCJMQjRfU0VSVklDRVNfTUFQX1YyIjoiY2lsaXVtX2xiNF9zZXJ2aWNlc192MiIsIkxCNl9CQUNLRU5EX01BUCI6ImNpbGl1bV9sYjZfYmFja2VuZHMiLCJMQjZfUkVWRVJTRV9OQVRfTUFQIjoiY2lsaXVtX2xiNl9yZXZlcnNlX25hdCIsIkxCNl9SRVZFUlNFX05BVF9TS19NQVAiOiJjaWxpdW1fbGI2X3JldmVyc2Vfc2siLCJMQjZfUkVWRVJTRV9OQVRfU0tfTUFQX1NJWkUiOiIyOTQ3NDAiLCJMQjZfU0VSVklDRVNfTUFQX1YyIjoiY2lsaXVtX2xiNl9zZXJ2aWNlc192MiIsIkxCX0FGRklOSVRZX01BVENIX01BUCI6ImNpbGl1bV9sYl9hZmZpbml0eV9tYXRjaCIsIk1FVFJJQ1NfTUFQIjoiY2lsaXVtX21ldHJpY3MiLCJNRVRSSUNTX01BUF9TSVpFIjoiMTAyNCIsIk1UVSI6IjE1MDAiLCJOT0RFUE9SVF9ORUlHSDQiOiJjaWxpdW1fbm9kZXBvcnRfbmVpZ2g0IiwiTk9ERVBPUlRfTkVJR0g0X1NJWkUiOiI1ODk0ODAiLCJOT0RFUE9SVF9QT1JUX01BWCI6IjMyNzY3IiwiTk9ERVBPUlRfUE9SVF9NQVhfTkFUIjoiNjU1MzUiLCJOT0RFUE9SVF9QT1JUX01JTiI6IjMwMDAwIiwiTk9ERVBPUlRfUE9SVF9NSU5fTkFUIjoiMzI3NjgiLCJOT19SRURJUkVDVCI6IjEiLCJQT0xJQ1lfQ0FMTF9NQVAiOiJjaWxpdW1fY2FsbF9wb2xpY3kiLCJQT0xJQ1lfTUFQX1NJWkUiOiIxNjM4NCIsIlBPTElDWV9QUk9HX01BUF9TSVpFIjoiNjU1MzUiLCJSRU1PVEVfTk9ERV9JRCI6IjYiLCJTSUdOQUxfTUFQIjoiY2lsaXVtX3NpZ25hbHMiLCJTTkFUX01BUFBJTkdfSVBWNCI6ImNpbGl1bV9zbmF0X3Y0X2V4dGVybmFsIiwiU05BVF9NQVBQSU5HX0lQVjRfU0laRSI6IjU4OTQ4MCIsIlNPQ0tPUFNfTUFQX1NJWkUiOiI2NTUzNSIsIlRSQUNFX1BBWUxPQURfTEVOIjoiMTI4VUxMIiwiVFVOTkVMX0VORFBPSU5UX01BUF9TSVpFIjoiNjU1MzYiLCJUVU5ORUxfTUFQIjoiY2lsaXVtX3R1bm5lbF9tYXAiLCJVTk1BTkFHRURfSUQiOiIzIiwiV09STERfSUQiOiIyIn0=
#ifndef CILIUM_NET_MAC
#define CILIUM_NET_MAC { .addr = {0x72,0x18,0xaa,0x3a,0x11,0x4c}} // 这个是 cilium_net 网卡 mac
#endif /* CILIUM_NET_MAC */
#define HOST_IFINDEX 10
#define HOST_IFINDEX_MAC { .addr = {0xd6,0xd7,0xc7,0x77,0xfb,0x9b}} // 这个是 cilium_host 网卡 mac
#define CILIUM_IFINDEX 11
#define EPHEMERAL_MIN 32768
#define NATIVE_DEV_MAC_BY_IFINDEX(IFINDEX) ({ \
        union macaddr __mac = {.addr = {0x0, 0x0, 0x0, 0x0, 0x0, 0x0}}; \
        switch (IFINDEX) { \
        case 2: {union macaddr __tmp = {.addr = {0x0c,0xc4,0x7a,0x57,0x90,0x5c}}; __mac=__tmp;} break; \ // 这个是 eth0 网卡 mac
        } \
        __mac; })