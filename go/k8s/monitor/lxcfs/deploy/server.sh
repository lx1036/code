


cat > lxcfs-webhook-csr.json <<EOF
{
  "CN": "lxcfs-webhook.kube-system.svc",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "CN",
      "L": "BeiJing",
      "O": "QC",
      "ST": "BeiJing"
    }
  ]
}
EOF

cfssl gencert \
  -ca=./ca.pem \
  -ca-key=./ca-key.pem \
  -config=./ca-config.json \
  -hostname="lxcfs-webhook,lxcfs-webhook.kube-system,lxcfs-webhook.kube-system.svc" \
  -profile=kubernetes lxcfs-webhook-csr.json | cfssljson -bare lxcfs-webhook # -> lxcfs-webhook-key.pem, lxcfs-webhook.pem

