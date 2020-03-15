
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

cfssl gencert \
  -ca=../1-ca/ca.pem \
  -ca-key=../1-ca/ca-key.pem \
  -config=../1-ca/ca-config.json \
  -profile=kubernetes kube-controller-manager-csr.json | \
  cfssljson -bare kube-controller-manager # -> kube-controller-manager-key.pem, kube-controller-manager.pem
