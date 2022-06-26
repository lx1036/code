

```shell
iptables-save -t filter

*filter
:INPUT ACCEPT [17:1523]
:FORWARD ACCEPT [0:0]
:OUTPUT ACCEPT [18:3591]
:CILIUM_FORWARD - [0:0]
:CILIUM_INPUT - [0:0]
:CILIUM_OUTPUT - [0:0]
:DOCKER - [0:0]
:DOCKER-ISOLATION-STAGE-1 - [0:0]
:DOCKER-ISOLATION-STAGE-2 - [0:0]
:DOCKER-USER - [0:0]
:KUBE-EXTERNAL-SERVICES - [0:0]
:KUBE-FIREWALL - [0:0]
:KUBE-FORWARD - [0:0]
:KUBE-KUBELET-CANARY - [0:0]
:KUBE-NWPLCY-DEFAULT - [0:0]
:KUBE-PROXY-CANARY - [0:0]
:KUBE-SERVICES - [0:0]
-A INPUT -m comment --comment "cilium-feeder: CILIUM_INPUT" -j CILIUM_INPUT
-A INPUT -d 192.168.0.2/32 -p udp -m udp --dport 53 -j ACCEPT
-A INPUT -d 192.168.0.2/32 -p tcp -m tcp --dport 53 -j ACCEPT
-A INPUT -d 169.254.20.10/32 -p udp -m udp --dport 53 -j ACCEPT
-A INPUT -d 169.254.20.10/32 -p tcp -m tcp --dport 53 -j ACCEPT
-A INPUT -m conntrack --ctstate NEW -m comment --comment "kubernetes service portals" -j KUBE-SERVICES
-A INPUT -m conntrack --ctstate NEW -m comment --comment "kubernetes externally-visible service portals" -j KUBE-EXTERNAL-SERVICES
-A INPUT -j KUBE-FIREWALL
-A FORWARD -m comment --comment "cilium-feeder: CILIUM_FORWARD" -j CILIUM_FORWARD
-A FORWARD -m comment --comment "kubernetes forwarding rules" -j KUBE-FORWARD
-A FORWARD -m conntrack --ctstate NEW -m comment --comment "kubernetes service portals" -j KUBE-SERVICES
-A FORWARD -j DOCKER-USER
-A FORWARD -j DOCKER-ISOLATION-STAGE-1
-A FORWARD -o docker0 -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
-A FORWARD -o docker0 -j DOCKER
-A FORWARD -i docker0 ! -o docker0 -j ACCEPT
-A FORWARD -i docker0 -o docker0 -j ACCEPT
-A OUTPUT -m comment --comment "cilium-feeder: CILIUM_OUTPUT" -j CILIUM_OUTPUT
-A OUTPUT -s 192.168.0.2/32 -p udp -m udp --sport 53 -j ACCEPT
-A OUTPUT -s 192.168.0.2/32 -p tcp -m tcp --sport 53 -j ACCEPT
-A OUTPUT -s 169.254.20.10/32 -p udp -m udp --sport 53 -j ACCEPT
-A OUTPUT -s 169.254.20.10/32 -p tcp -m tcp --sport 53 -j ACCEPT
-A OUTPUT -m conntrack --ctstate NEW -m comment --comment "kubernetes service portals" -j KUBE-SERVICES
-A OUTPUT -j KUBE-FIREWALL
-A CILIUM_FORWARD -o cilium_host -m comment --comment "cilium: any->cluster on cilium_host forward accept" -j ACCEPT
-A CILIUM_FORWARD -i cilium_host -m comment --comment "cilium: cluster->any on cilium_host forward accept (nodeport)" -j ACCEPT
-A CILIUM_FORWARD -i lxc+ -m comment --comment "cilium: cluster->any on lxc+ forward accept" -j ACCEPT
-A CILIUM_FORWARD -i cilium_net -m comment --comment "cilium: cluster->any on cilium_net forward accept (nodeport)" -j ACCEPT
-A CILIUM_FORWARD -o lxc+ -m comment --comment "cilium: any->cluster on lxc+ forward accept" -j ACCEPT
-A CILIUM_FORWARD -i lxc+ -m comment --comment "cilium: cluster->any on lxc+ forward accept (nodeport)" -j ACCEPT
-A CILIUM_INPUT -m mark --mark 0x200/0xf00 -m comment --comment "cilium: ACCEPT for proxy traffic" -j ACCEPT
-A CILIUM_OUTPUT -m mark --mark 0xa00/0xfffffeff -m comment --comment "cilium: ACCEPT for proxy return traffic" -j ACCEPT
-A CILIUM_OUTPUT -m mark ! --mark 0xe00/0xf00 -m mark ! --mark 0xd00/0xf00 -m mark ! --mark 0xa00/0xe00 -m comment --comment "cilium: host->any mark as from host" -j MARK --set-xmark 0xc00/0xf00
-A DOCKER-ISOLATION-STAGE-1 -i docker0 ! -o docker0 -j DOCKER-ISOLATION-STAGE-2
-A DOCKER-ISOLATION-STAGE-1 -j RETURN
-A DOCKER-ISOLATION-STAGE-2 -o docker0 -j DROP
-A DOCKER-ISOLATION-STAGE-2 -j RETURN
-A DOCKER-USER -j RETURN
-A KUBE-FIREWALL -m comment --comment "kubernetes firewall for dropping marked packets" -m mark --mark 0x8000/0x8000 -j DROP
-A KUBE-FIREWALL ! -s 127.0.0.0/8 -d 127.0.0.0/8 -m comment --comment "block incoming localnet connections" -m conntrack ! --ctstate RELATED,ESTABLISHED,DNAT -j DROP
-A KUBE-FORWARD -m comment --comment "kubernetes forwarding rules" -m mark --mark 0x4000/0x4000 -j ACCEPT
-A KUBE-FORWARD -m comment --comment "kubernetes forwarding conntrack pod source rule" -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
-A KUBE-FORWARD -m comment --comment "kubernetes forwarding conntrack pod destination rule" -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT
-A KUBE-NWPLCY-DEFAULT -m comment --comment "rule to mark traffic matching a network policy" -j MARK --set-xmark 0x10000/0x10000
-A KUBE-NWPLCY-DEFAULT -d 100.123.24.0/24 -m comment --comment "allow traffic to cluster IP" -j RETURN
-A KUBE-NWPLCY-DEFAULT -d 200.123.24.0/24 -m comment --comment "allow traffic to cluster IP" -j RETURN
-A KUBE-SERVICES -d 192.168.26.207/32 -p tcp -m comment --comment "default/nginx-demo has no endpoints" -m tcp --dport 80 -j REJECT --reject-with icmp-port-unreachable
-A KUBE-SERVICES -d 100.20.30.43/32 -p tcp -m comment --comment "default/nginx-demo has no endpoints" -m tcp --dport 80 -j REJECT --reject-with icmp-port-unreachable
-A KUBE-SERVICES -d 192.168.230.204/32 -p tcp -m comment --comment "cattle-prometheus/access-prometheus:nginx-http has no endpoints" -m tcp --dport 80 -j REJECT --reject-with icmp-port-unreachable
-A KUBE-SERVICES -d 192.168.22.201/32 -p tcp -m comment --comment "cattle-prometheus/access-grafana:http has no endpoints" -m tcp --dport 80 -j REJECT --reject-with icmp-port-unreachable
COMMIT
# Completed on Mon May 30 23:36:52 2022
```


```shell
iptables-save -t nat

*nat
:PREROUTING ACCEPT [0:0]
:INPUT ACCEPT [0:0]
:OUTPUT ACCEPT [11:2640]
:POSTROUTING ACCEPT [11:2640]
:CILIUM_OUTPUT_nat - [0:0]
:CILIUM_POST_nat - [0:0]
:CILIUM_PRE_nat - [0:0]
:DOCKER - [0:0]
:KUBE-FIREWALL - [0:0]
:KUBE-KUBELET-CANARY - [0:0]
:KUBE-LOAD-BALANCER - [0:0]
:KUBE-MARK-DROP - [0:0]
:KUBE-MARK-MASQ - [0:0]
:KUBE-NODE-PORT - [0:0]
:KUBE-NODEPORTS - [0:0]
:KUBE-POSTROUTING - [0:0]
:KUBE-PROXY-CANARY - [0:0]
:KUBE-SEP-5JL6OY6WVFSZ67SX - [0:0]
:KUBE-SEP-D4JHMZ4ZLPL2ZYZI - [0:0]
:KUBE-SEP-DSNLGRBAF2JMR7ZU - [0:0]
:KUBE-SEP-MU7PO7VX34QPS5HP - [0:0]
:KUBE-SEP-UJOTF6NNN2ZLINHJ - [0:0]
:KUBE-SEP-XDAKEI2XLX7RCSFE - [0:0]
:KUBE-SEP-Y4ZYYNHOT4MGAPXB - [0:0]
:KUBE-SERVICES - [0:0]
:KUBE-SVC-BRK3P4PPQWCLKOAN - [0:0]
:KUBE-SVC-ERIFXISQEP7F7OF4 - [0:0]
:KUBE-SVC-FXR4M2CWOGAZGGYD - [0:0]
:KUBE-SVC-JD5MR3NA4I4DYORP - [0:0]
:KUBE-SVC-NPX46M4PTMTKRN6Y - [0:0]
:KUBE-SVC-QMWWTXBG7KFJQKLO - [0:0]
:KUBE-SVC-TCOU7JCQXEZGVUNU - [0:0]
-A PREROUTING -m comment --comment "cilium-feeder: CILIUM_PRE_nat" -j CILIUM_PRE_nat
-A PREROUTING -m comment --comment "kubernetes service portals" -j KUBE-SERVICES
-A PREROUTING -m addrtype --dst-type LOCAL -j DOCKER
-A OUTPUT -m comment --comment "cilium-feeder: CILIUM_OUTPUT_nat" -j CILIUM_OUTPUT_nat
-A OUTPUT -m comment --comment "kubernetes service portals" -j KUBE-SERVICES
-A OUTPUT ! -d 127.0.0.0/8 -m addrtype --dst-type LOCAL -j DOCKER
-A POSTROUTING -m comment --comment "cilium-feeder: CILIUM_POST_nat" -j CILIUM_POST_nat
-A POSTROUTING -m comment --comment "kubernetes postrouting rules" -j KUBE-POSTROUTING
-A POSTROUTING -s 172.17.0.0/16 ! -o docker0 -j MASQUERADE
-A DOCKER -i docker0 -j RETURN
-A KUBE-FIREWALL -j KUBE-MARK-DROP
-A KUBE-LOAD-BALANCER -j KUBE-MARK-MASQ
-A KUBE-MARK-MASQ -j MARK --set-xmark 0x4000/0x4000
-A KUBE-POSTROUTING -m comment --comment "Kubernetes endpoints dst ip:port, source ip for solving hairpin purpose" -m set --match-set KUBE-LOOP-BACK dst,dst,src -j MASQUERADE
-A KUBE-POSTROUTING -m mark ! --mark 0x4000/0x4000 -j RETURN
-A KUBE-POSTROUTING -j MARK --set-xmark 0x4000/0x0
-A KUBE-POSTROUTING -m comment --comment "kubernetes service traffic requiring SNAT" -j MASQUERADE
-A KUBE-SEP-5JL6OY6WVFSZ67SX -s 20.225.0.9/32 -m comment --comment "kube-system/kube-dns:metrics" -j KUBE-MARK-MASQ
-A KUBE-SEP-5JL6OY6WVFSZ67SX -p tcp -m comment --comment "kube-system/kube-dns:metrics" -m tcp -j DNAT --to-destination  --random --to-destination  --random --to-destination 0.0.0.0:0
-A KUBE-SEP-D4JHMZ4ZLPL2ZYZI -s 20.225.0.148/32 -m comment --comment "kube-system/metrics-server" -j KUBE-MARK-MASQ
-A KUBE-SEP-D4JHMZ4ZLPL2ZYZI -p tcp -m comment --comment "kube-system/metrics-server" -m tcp -j DNAT --to-destination  --random --to-destination  --random --to-destination 0.0.0.0
-A KUBE-SEP-DSNLGRBAF2JMR7ZU -s 20.225.0.9/32 -m comment --comment "kube-system/kube-dns-upstream:dns-tcp" -j KUBE-MARK-MASQ
-A KUBE-SEP-DSNLGRBAF2JMR7ZU -p tcp -m comment --comment "kube-system/kube-dns-upstream:dns-tcp" -m tcp -j DNAT --to-destination  --random --to-destination  --random --to-destination
-A KUBE-SEP-MU7PO7VX34QPS5HP -s 20.225.0.9/32 -m comment --comment "kube-system/kube-dns:dns" -j KUBE-MARK-MASQ
-A KUBE-SEP-MU7PO7VX34QPS5HP -p udp -m comment --comment "kube-system/kube-dns:dns" -m udp -j DNAT --to-destination  --random --to-destination  --random --to-destination
-A KUBE-SEP-UJOTF6NNN2ZLINHJ -s 20.225.0.9/32 -m comment --comment "kube-system/kube-dns:dns-tcp" -j KUBE-MARK-MASQ
-A KUBE-SEP-UJOTF6NNN2ZLINHJ -p tcp -m comment --comment "kube-system/kube-dns:dns-tcp" -m tcp -j DNAT --to-destination  --random --to-destination  --random --to-destination
-A KUBE-SEP-XDAKEI2XLX7RCSFE -s 10.208.40.179/32 -m comment --comment "default/kubernetes:https" -j KUBE-MARK-MASQ
-A KUBE-SEP-XDAKEI2XLX7RCSFE -p tcp -m comment --comment "default/kubernetes:https" -m tcp -j DNAT --to-destination :0 --persistent --to-destination :0 --persistent --to-destination 0.0.0.0 --persistent
-A KUBE-SEP-Y4ZYYNHOT4MGAPXB -s 20.225.0.9/32 -m comment --comment "kube-system/kube-dns-upstream:dns" -j KUBE-MARK-MASQ
-A KUBE-SEP-Y4ZYYNHOT4MGAPXB -p udp -m comment --comment "kube-system/kube-dns-upstream:dns" -m udp -j DNAT --to-destination  --random --to-destination  --random --to-destination
-A KUBE-SERVICES -m comment --comment "Kubernetes service lb portal" -m set --match-set KUBE-LOAD-BALANCER dst,dst -j KUBE-LOAD-BALANCER
-A KUBE-SERVICES ! -s 10.42.0.0/16 -m comment --comment "Kubernetes service cluster ip + port for masquerade purpose" -m set --match-set KUBE-CLUSTER-IP dst,dst -j KUBE-MARK-MASQ
-A KUBE-SERVICES -m addrtype --dst-type LOCAL -j KUBE-NODE-PORT
-A KUBE-SERVICES -m set --match-set KUBE-CLUSTER-IP dst,dst -j ACCEPT
-A KUBE-SERVICES -m set --match-set KUBE-LOAD-BALANCER dst,dst -j ACCEPT
-A KUBE-SVC-BRK3P4PPQWCLKOAN -m comment --comment "kube-system/kube-dns-upstream:dns-tcp" -j KUBE-SEP-DSNLGRBAF2JMR7ZU
-A KUBE-SVC-ERIFXISQEP7F7OF4 -m comment --comment "kube-system/kube-dns:dns-tcp" -j KUBE-SEP-UJOTF6NNN2ZLINHJ
-A KUBE-SVC-FXR4M2CWOGAZGGYD -m comment --comment "kube-system/kube-dns-upstream:dns" -j KUBE-SEP-Y4ZYYNHOT4MGAPXB
-A KUBE-SVC-JD5MR3NA4I4DYORP -m comment --comment "kube-system/kube-dns:metrics" -j KUBE-SEP-5JL6OY6WVFSZ67SX
-A KUBE-SVC-NPX46M4PTMTKRN6Y -m comment --comment "default/kubernetes:https" -j KUBE-SEP-XDAKEI2XLX7RCSFE
-A KUBE-SVC-QMWWTXBG7KFJQKLO -m comment --comment "kube-system/metrics-server" -j KUBE-SEP-D4JHMZ4ZLPL2ZYZI
-A KUBE-SVC-TCOU7JCQXEZGVUNU -m comment --comment "kube-system/kube-dns:dns" -j KUBE-SEP-MU7PO7VX34QPS5HP
COMMIT
```