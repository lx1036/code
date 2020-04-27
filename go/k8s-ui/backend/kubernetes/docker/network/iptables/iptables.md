
# iptables

![iptables 工作流程图](./img/tables_traverse.jpg)


**[(1)iptables 防火墙-四表/五链、数据包匹配流程、编写 iptables 规则](https://zhuanlan.zhihu.com/p/84432006)**
(1) Linux 防火墙基础
iptables: 属于`用户空间`的防火墙命令工具，来和内核空间的 netfilter 交互，在目录 `/sbin/iptables` 中。
netfilter: Linux 内核中实现包过滤防火墙的内部结构，属于`内核空间`的防火墙工具。

(2)iptables 4种规则表、5种链结构
规则表：对数据包进行过滤或处理
链：容纳多种防火墙规则
![iptables](./img/iptables.jpg)

四种规则表：
iptables 管理四个不同的规则表，分别由独立的内核模块管理：
* filter 表: 用来过滤数据包，每一个rule会决定丢弃或处理这个数据包，
内核模块为 iptable_filter，表内包括 `input`、`forward` 和 `output` 三个链。

* nat 表(network address transfer): 主要用来修改数据包的IP地址、端口号信息。
内核模块为 iptable_nat，表包括 `prerouting`、`postrouting` 和 `output` 三个链。

* mangle 表：主要用来修改数据包的`服务类型`、`生存周期`来标记数据包，实现流量整形、策略路由等等功能。
内核模块为 iptable_mangle，表包括 `prerouting`、`postrouting`、`input`、`forward` 和 `output` 五个链。

* raw 表：主要用来是否对数据包进行状态跟踪。
内核模块为 iptable_raw，表包括 `prerouting` 和 `output` 两个链。


五种链：
input 链：访问本地地址的数据包时，使用 input 链规则。
output 链：本机访问外网数据包时，使用 output 链规则。
forward 链：当需要本机防火墙转发数据包到其他地址时，使用 forward 链规则。
prerouting 链：对数据包做路由选择之前，使用 prerouting 链规则。
postrouting 链：对数据包做路由选择之后，使用 postrouting 链规则。

![iptables 数据包流向图，此图也很清晰](./img/iptables-2.jpg)

数据包进来时，防火墙开始工作，对应的每一个链内的规则，依次按照规则表顺序为：raw -> mangle -> nat -> filter 表内顺序来处理数据包。

(3)编写防火墙规则(iptables 命令行手动执行)
```shell script
# iptables [-t 表名] 管理选项 [链名] [匹配条件] [-j 控制类型]
# 在 filter 表 INPUT 链中追加(-A: append)一个必须是 TCP 包的 ACCEPT 规则:
iptables -t filter -A INPUT -p TCP -j ACCEPT
# 使用 -I 来管理可以指定规则的顺序号，默认作为第一条规则
# 容许转发 192.168.1.0/24 网段的数据包，但不包括 192.168.123.123
iptables -I FORWARD -s 192.168.123.123 -j REJECT
iptables -I FORWARD -s 192.168.1.0/24 -j ACCEPT
# 来自 192.168.123.0/24 网段IP的数据包，拒绝转发和处理，即直接封锁
iptables -I INPUT -s 192.168.123.0/24 -j DROP
iptables -I FORWARD -s 192.168.123.0/24 -j DROP
```
* 表名、链名：指定iptables命令所操作的表和链，未指定表名时将默认使用filter表；
* 管理选项：表示iptables规则的操作方式，比如：插入、增加、删除、查看等；
* 匹配条件：指定要处理的数据包的特征，不符合指定条件的数据包不在处理；
* 控制类型：指数据包的处理方式，比如：允许、拒绝、丢弃等

(4)编写防火墙规则(go代码执行)


