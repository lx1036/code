
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
