

## Kube-scheduler Client Credential
cat > kube-scheduler-csr.json <<EOF
{
  "CN": "system:kube-scheduler",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "US",
      "L": "Portland",
      "O": "system:node-scheduler",
      "OU": "Kubernetes The Hard Way",
      "ST": "Oregon"
    }
  ]
}
EOF
cfssl gencert \
  -ca=../1-ca/ca.pem \
  -ca-key=../1-ca/ca-key.pem \
  -config=../1-ca/ca-config.json \
  -profile=kubernetes kube-scheduler-csr.json | \
  cfssljson -bare kube-scheduler # -> kube-scheduler-key.pem, kube-scheduler.pem
