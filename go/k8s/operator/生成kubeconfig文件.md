

# 背景
kubeconfig文件是与k8s api进行restful api操作的凭证，里面包含api-server地址和用户证书。

# 步骤
(1)生成ca文件
1.1 ca配置文件
```shell script
cat > ca-config.json <<EOF
{
  "signing": {
    "default": {
      "expiry": "87600h"
    },
    "profiles": {
      "kubernetes": {
        "usages": ["signing", "key encipherment", "server auth", "client auth"],
        "expiry": "87600h"
      }
    }
  }
}
EOF
```
新建 CA 凭证签发请求文件:
```shell script
cat > ca-csr.json <<EOF
{
  "CN": "Kubernetes",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "US",
      "L": "Portland",
      "O": "Kubernetes",
      "OU": "CA",
      "ST": "Oregon"
    }
  ]
}
EOF
```
生成 CA 凭证和私钥:
```shell script
cfssl gencert -initca ca-csr.json | cfssljson -bare ca
```
最后会生成ca.pem和ca-key.pem文件

(2)admin用户的kubeconfig文件
```shell script
cat > cluster-admin-csr.json <<EOF
{
  "CN": "cluster-admin",
  "key": {
    "algo": "ecdsa",
    "size": 384
  },
  "names": [
    {
      "C": "CN",
      "L": "Beijing",
      "O": "system:masters",
      "OU": "kube",
      "ST": "Beijing"
    }
  ]
}
EOF
 
cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json \
    -profile=kubernetes cluster-admin-csr.json | cfssljson -bare cluster-admin
 
{
  kubectl config set-cluster shyc \
    --certificate-authority=ca.pem \
    --embed-certs=true \
    --server=https://1.2.3.4 \
    --kubeconfig=shyc.kubeconfig
 
  kubectl config set-credentials shyc \
    --client-certificate=cluster-admin.pem \
    --client-key=cluster-admin-key.pem \
    --embed-certs=true \
    --kubeconfig=shyc.kubeconfig
 
  kubectl config set-context shyc \
    --cluster=shyc \
    --user=shyc \
    --kubeconfig=shyc.kubeconfig
 
  kubectl config use-context shyc --kubeconfig=shyc.kubeconfig
}
```
最后，会生成 cluster-admin 用户的证书和私钥文件，以及私钥签名 shyc 的 kubeconfig 文件。

(3)node节点证书生成
node证书请求文件和生成证书文件执行脚本：

```shell script
# ./node_cert.sh ${hostname}，如 ./node_cert.sh docker1234.tec.net

ip=$(dig +short $1)

echo "{ \
  \"CN\": \"$1\", \
  \"hosts\": [
    \"$1\", \
    \"$ip\" \
  ], \
  \"key\": { \
    \"algo\": \"ecdsa\", \
    \"size\": 384 \
  }, \
  \"names\": [ \
    { \
      \"C\": \"CN\", \
      \"L\": \"Beijing\", \
      \"O\": \"Technology Co. Ltd.\", \
      \"OU\":\"Kube\" \
    } \
  ] \
}"

cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=kubernetes \
  $1-csr.json | cfssljson -bare $1
```
最终，会生成由ca根证书签发的当前node节点的证书文件，如docker1234.tec.net.pem和docker1234.tec.net-key.pem
