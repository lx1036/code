# Docs: https://kubernetes.io/zh/docs/tasks/tls/managing-tls-in-a-cluster/

# Certificate Authority
## cfssl: https://github.com/cloudflare/cfssl
cat > ca-config.json <<EOF
{
  "signing": {
    "default": {
      "expiry": "87600h"
    },
    "profiles": {
      "kubernetes": {
        "usages": [
            "signing",
            "key encipherment",
            "server auth",
            "client auth"
        ],
        "expiry": "87600h"
      },
      "kubelet": {
        "usages": [
            "signing",
            "key encipherment",
            "client auth"
        ],
        "expiry": "87600h"
      },
      "kube-service-account": {
        "usages": [
            "signing",
            "key encipherment",
            "server auth",
            "client auth"
        ],
        "expiry": "87600h"
      }
    }
  }
}
EOF

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
