
# HTTP 长链接
k8s informer 机制使用了 HTTP/1.1 版本，该版本默认通过 keep-alive 字段实现长链接，并通过 Header 中 Transfer-Encoding=chunked 分块
传输编码来实现动态内容分块传输，这样就实现了服务端向客户端推送增量消息，即每当 k8s 资源对象发生 Create/Update/Delete 事件时，客户端都会
收到这个增量事件数据。
代码如下：
https://github.com/kubernetes/kubernetes/blob/1dd5338295409edcfff11505e7bb246f0d325d15/staging/src/k8s.io/apiserver/pkg/endpoints/handlers/watch.go#L199-L203

问题：如果 client 和 server 保持长链接，但是很久时间内没有数据报文传输，这时候server不确定client是在线还是下线了，怎么解决？
TCP 协议设计中考虑了这个问题，当超过一段时间后(默认2小时)，server 会发送空报文给对方并有重试机制，来确认client是否在线。如果没有收到client的
ack包后，就认为client已经下线。
tcp keep-alive 内核中关于这个心跳检测机制的三个参数：
```shell
net.ipv4.tcp_keepalive_intvl = 15 # 时间间隔 interval
net.ipv4.tcp_keepalive_probes = 5 # 一共尝试次数
net.ipv4.tcp_keepalive_time = 1800 # 30 min
```
