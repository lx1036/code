
# Client Credential
## Admin Client Credential

# CN => Common Name
# C => Country
# L => Locality
# O => Organization
# OU => Organization Unit
# ST => State/Province

cat > admin-csr.json <<EOF
{
  "CN": "nginx.lx1036.com",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "China",
      "L": "China",
      "O": "system:masters",
      "OU": "Kubernetes The Hard Way",
      "ST": "BeiJing"
    }
  ]
}
EOF

### admin pem and private key(Admin客户端的凭证和私钥)
cfssl gencert \
  -ca=./ca.pem \
  -ca-key=./ca-key.pem \
  -config=./ca-config.json \
  -profile=kubernetes admin-csr.json | cfssljson -bare admin # -> admin-key.pem, admin.pem
