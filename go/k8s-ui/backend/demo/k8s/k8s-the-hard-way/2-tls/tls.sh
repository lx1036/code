# Certificate Authority
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

## ca pem and private key(凭证和私钥)
cfssl gencert -initca ca-csr.json | cfssljson -bare ca # -> ca-key.pem, ca.pem

# Client Credential
## Admin Client Credential
cat > admin-csr.json <<EOF
{
  "CN": "admin",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "US",
      "L": "Portland",
      "O": "system:masters",
      "OU": "Kubernetes The Hard Way",
      "ST": "Oregon"
    }
  ]
}
EOF
### admin pem and private key(Admin客户端的凭证和私钥)
cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json \
    -profile=kubernetes admin-csr.json | cfssljson -bare admin # -> admin-key.pem, admin.pem

## Kubelet Client Credential(https://kubernetes.io/zh/docs/reference/access-authn-authz/node/)
for instance in worker-0 worker-1 worker-2; do
  cat > ${instance}-csr.json <<EOF
  {
    "CN": "system:node:${instance}",
    "key": {
      "algo": "rsa",
      "size": 2048
    },
    "names": [
      {
        "C": "US",
        "L": "Portland",
        "O": "system:nodes",
        "OU": "Kubernetes The Hard Way",
        "ST": "Oregon"
      }
    ]
  }
EOF
EXTERNAL_IP=$(gcloud compute instances describe ${instance} \
  --format 'value(networkInterfaces[0].accessConfigs[0].natIP)')
INTERNAL_IP=$(gcloud compute instances describe ${instance} \
  --format 'value(networkInterfaces[0].networkIP)')
### 给每台 worker 节点创建凭证和私钥
### (worker-0-key.pem/workder-0.pem,worker-1-key.pem/workder-1.pem,worker-2-key.pem/workder-2.pem,)
cfssl gencert \
  -ca=ca.pem \
  -ca-key=ca-key.pem \
  -config=ca-config.json \
  -hostname=${instance},"${EXTERNAL_IP}","${INTERNAL_IP}" \
  -profile=kubernetes \
  ${instance}-csr.json | cfssljson -bare ${instance}
done

## Kube-controller-manager Client Credential
cat > kube-controller-manager-csr.json <<EOF
{
  "CN": "system:kube-controller-manager",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "US",
      "L": "Portland",
      "O": "system:kube-controller-manager",
      "OU": "Kubernetes The Hard Way",
      "ST": "Oregon"
    }
  ]
}
EOF
cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json \
  -profile=kubernetes kube-controller-manager-csr.json | cfssljson -bare kube-controller-manager # -> kube-controller-manager-key.pem, kube-controller-manager.pem

## Kube-proxy Client Credential
cat > kube-proxy-csr.json <<EOF
{
  "CN": "system:kube-proxy",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "US",
      "L": "Portland",
      "O": "system:node-proxier",
      "OU": "Kubernetes The Hard Way",
      "ST": "Oregon"
    }
  ]
}
EOF
cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json \
  -profile=kubernetes kube-proxy-csr.json | cfssljson -bare kube-proxy # -> kube-proxy-key.pem, kube-proxy.pem

# Server Credential
