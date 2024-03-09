

#ifndef XDP_CILIUM_L4LB_NETDEV_CONFIG_H
#define XDP_CILIUM_L4LB_NETDEV_CONFIG_H

/*
 * This is just a dummy header with dummy values to allow for test
 * compilation without the full code generation engine backend.
 */
#define DROP_NOTIFY
#ifndef SKIP_DEBUG
#define DEBUG
#endif
#define SECLABEL 2
#define SECLABEL_NB 0xfffff
#define CALLS_MAP test_cilium_calls_65535


#endif //XDP_CILIUM_L4LB_NETDEV_CONFIG_H
