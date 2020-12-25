

# BGP
BGP is a standard protocol for exchanging routing information between two routers in a network.

Calico can run BGP in three modes:
1. Full mesh - where each node talks BGP to each other, easily scaling to 100 nodes, 
   on top of an underlying L2 network or using IPIP overlay 
2. With route reflectors - where each node talks to one or more BGP route reflectors, scaling beyond 100 nodes, 
   on top of an underlying L2 network or using IPIP overlay
3. Peered with TOR (Top of Rack) routers - in a physical data center where each node talks to routers in the top of the corresponding rack, 
   scaling to the limits of your physical data center.
   
> 目前我们用的是第三种方式，每一个worker node和Tor交换机建立bgp peer。这样，pod ip可以在cluster外部被路由routable:
> https://docs.projectcalico.org/networking/determine-best-networking#pod-ip-routability-outside-of-the-cluster
> https://docs.projectcalico.org/networking/bgp#top-of-rack-tor
> https://docs.projectcalico.org/reference/architecture/design/l2-interconnect-fabric

> ToR Switch: Top of Rack，机顶交换机；Spine Switch: 机柜间交换机，在上一层可以认为是机房交换机。


交换机开启BGP过程：
如果是一台新的交换机的话 它上面是没有开启BGP的 我得手动给他弄一下
1. 在TOR交换机（直连服务器的那台）开启BGP
2. 把TOR和核心建立BGP邻居
3. 把新TOR交换机加到NOSA(Network Operation Standard API)数据库里，然后同步数据
