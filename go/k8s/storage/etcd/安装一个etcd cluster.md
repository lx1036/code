https://juejin.im/post/6881941310358421511

# 背景
etcd是golang写的一个分布式key-value数据存储服务，用来存储配置数据。同时，还可以用来作为分布式锁、分布式队列、消息发布和订阅、服务注册和发现等等功能。

部署节点一般是奇数方便选主，节点之间数据通过raft协议保持强一致性。K8s使用etcd作为数据库，主要是kube-apiserver组件和etcd数据通信。

为了方便深入学习了解etcd，先本地部署一个带有证书认证的etcd cluster，方便调试开发。

# 步骤
(1) 下载etcd和证书工具cfssl
```
# etcdctl是etcd官方提供的CLI工具，方便与etcd交互
wget https://github.com/etcd-io/etcd/releases/download/v3.4.10/etcd-v3.4.10-darwin-amd64.zip
unzip etcd-v3.4.10-darwin-amd64.zip
mv etcd-v3.4.10-darwin-amd64/etcd /usr/local/bin/etcd
mv etcd-v3.4.10-darwin-amd64/etcdctl /usr/local/bin/etcdctl
# 或者
macos的为
brew install etcd

# 证书工具cfssl
curl -o cfssl https://pkg.cfssl.org/R1.2/cfssl_darwin-amd64
curl -o cfssljson https://pkg.cfssl.org/R1.2/cfssljson_darwin-amd64
chmod +x cfssl cfssljson
sudo mv cfssl cfssljson /usr/local/bin/
# 或者macos的为
brew install cfssl

# linux安装cfssl
wget https://pkg.cfssl.org/R1.2/cfssl_linux-amd64
wget https://pkg.cfssl.org/R1.2/cfssljson_linux-amd64
wget https://pkg.cfssl.org/R1.2/cfssl-certinfo_linux-amd64
chmod +x cfssl_linux-amd64 cfssljson_linux-amd64 cfssl-certinfo_linux-amd64
mv cfssl_linux-amd64 /usr/local/bin/cfssl
mv cfssljson_linux-amd64 /usr/local/bin/cfssljson
mv cfssl-certinfo_linux-amd64 /usr/bin/cfssl-certinfo
```

(2) 生成etcd服务端和客户端证书

执行根证书ca.sh脚本：
```

# Docs: https://kubernetes.io/zh/docs/tasks/tls/managing-tls-in-a-cluster/

# Certificate Authority
## cfssl: https://github.com/cloudflare/cfssl
cat > ca-config.json <<EOF
{
  "signing": {
    "default": {
      "expiry": "8760h"
    },
    "profiles": {
      "kubernetes": {
        "usages": ["signing", "key encipherment", "server auth", "client auth"],
        "expiry": "8760h"
      }
    }
  }
}
EOF

```
# 集群初始化时，在没有证书的情况下，使用该命令生成 CA 证书及 CA的私钥
cfssl gencert -initca ca-csr.json | ./cfssljson -bare ca
 
# 如果集群已经在运行，但是 CA 证书即将到期，使用以下命令 复用之前的 私钥 重新生成 CA 证书，此 CA证书与之前的 CA证书可以同时使用，所以可以对线上集群逐步替换此为 此 CA 证书
## 使用 cfssl gencert -h 查看帮助，如何续签证书，有两种方式
##    Re-generate a CA cert with the CA key and CSR:
##        cfssl gencert -initca -ca-key key CSRJSON
##
##    Re-generate a CA cert with the CA key and certificate:
##        cfssl gencert -renewca -ca cert -ca-key key
 
# 根据 issue：https://github.com/cloudflare/cfssl/issues/1034 ，可以在 ca-csr.json 内添加对应的 CA 字段的配置，来设置 CA 证书的有效期，cfssl 工具默认是 5年
  {
    "CA": {
        "expiry": "127200h",
        "pathlen": 0
    },
    "CN": "QIHU 360 SOFTWARE CO. LIMITED",
    "key": {
        "algo": "ecdsa",
        "size": 384
    },
    "names": [
        {
            "C": "CN",
            "ST": "Beijing",
            "L": "Beijing",
            "O": "QIHU 360",
            "OU": "QSSWEB"
        }
    ]
}
```

####
# CN: Common Name
# https://blog.cloudflare.com/introducing-cfssl/
####
cat > ca-csr.json <<EOF
{
  "CN": "Kubernetes",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "CN",
      "L": "BeiJing",
      "O": "Kubernetes",
      "OU": "CA",
      "ST": "BeiJing"
    }
  ]
}
EOF

cfssl gencert -initca ca-csr.json | cfssljson -bare ca
```


执行服务端证书脚本 etcd-server.sh：
```
cat > etcd-csr.json <<EOF
{
  "CN": "etcd",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "CN",
      "L": "BeiJing",
      "O": "system:masters",
      "OU": "etcd",
      "ST": "BeiJing"
    }
  ]
}
EOF

cfssl gencert \
  -ca=./ca.pem \
  -ca-key=./ca-key.pem \
  -config=./ca-config.json \
  -hostname=127.0.0.1,kubernetes.default \
  -profile=kubernetes etcd-csr.json | cfssljson -bare etcd # -> etcd-key.pem, etcd.pem
```


执行客户端证书脚本 etcd-client.sh：
```
cat > etcd-client-csr.json <<EOF
{
  "CN": "etcd-client",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "CN",
      "L": "BeiJing",
      "O": "system:masters",
      "OU": "etcd",
      "ST": "BeiJing"
    }
  ]
}
EOF

cfssl gencert \
  -ca=./ca.pem \
  -ca-key=./ca-key.pem \
  -config=./ca-config.json \
  -hostname=127.0.0.1,kubernetes.default \
  -profile=kubernetes \
  etcd-client-csr.json | cfssljson -bare etcd-client # -> etcd-client-key.pem, etcd-client.pem
```

最后会生成根证书、etcd启动时需要的服务端证书和etcdctl交互时需要的客户端证书：

![](https://p1-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/c8936843313245b489e38f3672b2df7e~tplv-k3u1fbpfcp-watermark.webp)




其中，客户端证书也可以直接复用服务端证书，所以可以不用再生成个客户端证书。为了更安全，这里还是生成个客户端证书。



(3)开启一个etcd cluster

在Makefile文件内写上：
```
PWD := $(shell pwd)

etcd1-auth:
	etcd \
      --name=infra1 \
      --client-cert-auth=true \
      --cert-file=$(PWD)/tls/etcd.pem \
      --key-file=$(PWD)/tls/etcd-key.pem \
      --peer-client-cert-auth=true \
      --peer-cert-file=$(PWD)/tls/etcd.pem \
      --peer-key-file=$(PWD)/tls/etcd-key.pem \
      --trusted-ca-file=$(PWD)/tls/ca.pem \
      --peer-trusted-ca-file=$(PWD)/tls/ca.pem \
      --initial-advertise-peer-urls=https://127.0.0.1:42380 \
      --listen-peer-urls=https://127.0.0.1:42380 \
      --listen-client-urls=https://127.0.0.1:42379 \
      --advertise-client-urls=https://127.0.0.1:42379 \
      --initial-cluster-token=etcd-cluster-0 \
      --initial-cluster 'infra1=https://127.0.0.1:42380,infra2=https://127.0.0.1:52380,infra3=https://127.0.0.1:62380' \
      --initial-cluster-state=new \
      --data-dir=$(PWD)/infra1.etcd \
      --logger=zap \
      --log-outputs=stderr

etcd2-auth:
	etcd \
      --name=infra2 \
      --client-cert-auth=true \
      --cert-file=$(PWD)/tls/etcd.pem \
      --key-file=$(PWD)/tls/etcd-key.pem \
      --peer-client-cert-auth=true \
      --peer-cert-file=$(PWD)/tls/etcd.pem \
      --peer-key-file=$(PWD)/tls/etcd-key.pem \
      --trusted-ca-file=$(PWD)/tls/ca.pem \
      --peer-trusted-ca-file=$(PWD)/tls/ca.pem \
      --initial-advertise-peer-urls=https://127.0.0.1:52380 \
      --listen-peer-urls=https://127.0.0.1:52380 \
      --listen-client-urls=https://127.0.0.1:52379 \
      --advertise-client-urls=https://127.0.0.1:52379 \
      --initial-cluster-token=etcd-cluster-0 \
      --initial-cluster 'infra1=https://127.0.0.1:42380,infra2=https://127.0.0.1:52380,infra3=https://127.0.0.1:62380' \
      --initial-cluster-state=new \
      --data-dir=$(PWD)/infra2.etcd \
      --logger=zap \
      --log-outputs=stderr

etcd3-auth:
	etcd \
      --name=infra3 \
      --client-cert-auth=true \
      --cert-file=$(PWD)/tls/etcd.pem \
      --key-file=$(PWD)/tls/etcd-key.pem \
      --peer-client-cert-auth=true \
      --peer-cert-file=$(PWD)/tls/etcd.pem \
      --peer-key-file=$(PWD)/tls/etcd-key.pem \
      --trusted-ca-file=$(PWD)/tls/ca.pem \
      --peer-trusted-ca-file=$(PWD)/tls/ca.pem \
      --initial-advertise-peer-urls=https://127.0.0.1:62380 \
      --listen-peer-urls=https://127.0.0.1:62380 \
      --listen-client-urls=https://127.0.0.1:62379 \
      --advertise-client-urls=https://127.0.0.1:62379 \
      --initial-cluster-token=etcd-cluster-0 \
      --initial-cluster 'infra1=https://127.0.0.1:42380,infra2=https://127.0.0.1:52380,infra3=https://127.0.0.1:62380' \
      --initial-cluster-state=new \
      --data-dir=$(PWD)/infra3.etcd \
      --logger=zap \
      --log-outputs=stderr


list:
	ETCDCTL_API=3 etcdctl member list \
		--write-out=table \
		--endpoints=https://127.0.0.1:42379 \
		--cacert $(PWD)/tls/ca.pem \
		--cert $(PWD)/tls/etcd-client.pem \
		--key $(PWD)/tls/etcd-client-key.pem
```

在三个终端分别执行 `make etcd1-auth`、`make etcd2-auth`和`make etcd3-auth`，就搭建了一个3节点的etcd cluster了，并且etcd1/etcd2/etcd3的监听在客户端的端口分别是42379/52379/62379。并且能正常读写数据：

![](https://p3-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/e90da89cf174475ea864d81cee465b4d~tplv-k3u1fbpfcp-watermark.webp)




并且，etcd支持数据持久化，k-v数据保存在b+tree里。etcd数据存储在当前目录下，有snap和wal文件。snap文件是快照文件，存的是全量数据；wal是增量数据，当wal里的key达到一定数量时会自动生成快照。这种设计可以提高性能减少内存压力：

![](https://p6-juejin.byteimg.com/tos-cn-i-k3u1fbpfcp/19e11a231cbe4bbb9acbd5b76d58c67d~tplv-k3u1fbpfcp-watermark.webp)



总之，etcd作为分布式KV存储服务，小巧性能好，维护也方便，是个相当优秀的工具。

# 参考文献
https://etcd.io/docs/v3.4.0/dev-guide/local_cluster/









