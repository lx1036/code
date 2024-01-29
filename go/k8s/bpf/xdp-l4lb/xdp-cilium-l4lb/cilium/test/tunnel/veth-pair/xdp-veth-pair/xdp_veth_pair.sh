#!/bin/bash

# Create 3 namespaces with 3 veth peers, and
# forward packets in-between using native XDP
#
#                      XDP_TX
# NS1(veth11)        NS2(veth22)        NS3(veth33)
#      |                  |                  |
#      |                  |                  |
#   (veth1,            (veth2,            (veth3,
#   id:111)            id:122)            id:133)
#     ^ |                ^ |                ^ |
#     | |  XDP_REDIRECT  | |  XDP_REDIRECT  | |
#     | ------------------ ------------------ |
#     -----------------------------------------
#                    XDP_REDIRECT



