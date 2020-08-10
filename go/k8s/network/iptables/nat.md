
```shell script
iptables-save -t nat > iptables-nat.txt
cat iptables-nat.txt | grep "abc-production" > abc-production.txt # abc-production为业务名
```

# SNAT
```shell script
iptables -t nat -A POSTROUTING -s 10.10.0.0/16 -j SNAT --to-source 公网IP
```
这条命令的意思是将来自 10.10.0.0/16 网段的报文的源地址改为公司的公网 IP 地址。
* -t nat：表示 NAT 表
* -A POSTROUTING：表示将该条规则添加到 POSTROUTING 链的末尾，A 就是 append。
* -j SNAT：表示使用 SNAT 动作
* --to-source：表示将报文的源 IP 修改为哪个公网 IP 地址

# DNAT
```shell script
iptables -t nat -I PREROUTING -d 公网IP -p tcp --dport 公网端口 -j DNAT --to-destination 私网IP:端口号
```
这条命令的意思是将来自公网IP:端口号的报文的目的地址改为私网IP:端口，可以看到这里多了端口的信息。
原因是要区分公网访问的是私网的那个服务，所以需要明确到端口层级，才能精确送到客户端。而SNAT不需要端口信息也可以完成正确转发。
* -I PREROUTING：表示将该条规则插入到 PREROUTING 的首部，I 就是 insert
* --to-destination：表示将报文的目的 IP：端口修改为哪个私网IP：端口
