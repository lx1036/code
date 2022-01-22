
# wireshark 本地抓包 BGP 协议数据

(1) 本地开启 wireshark，filter 中输入 bgp，然后 Analyze -> Decode As -> 加上 tcp port 1790(route server tcp.port==1790), Current 选择 BGP

(2) 开启 route client(tcp.port==1791) 和 route server(tcp.port==1790)

(3) 查看 BGP 建立过程中发送的 TCP 信息：OPEN Message -> OPEN Message -> KEEPALIVE Message -> KEEPALIVE Message，然后 Established 状态
