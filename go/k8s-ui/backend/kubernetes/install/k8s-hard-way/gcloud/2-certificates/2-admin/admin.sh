
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
cfssl gencert \
  -ca=../1-ca/ca.pem \
  -ca-key=../1-ca/ca-key.pem \
  -config=../1-ca/ca-config.json \
  -profile=kubernetes admin-csr.json | cfssljson -bare admin # -> admin-key.pem, admin.pem
