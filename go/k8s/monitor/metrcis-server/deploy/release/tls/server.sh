


cat > metrics-server-csr.json <<EOF
{
  "CN": "metrics-server.kube-system.svc",
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
  -hostname="metrics-server,metrics-server.kube-system,metrics-server.kube-system.svc" \
  -profile=kubernetes metrics-server-csr.json | cfssljson -bare metrics-server # -> metrics-server-key.pem, metrics-server.pem

