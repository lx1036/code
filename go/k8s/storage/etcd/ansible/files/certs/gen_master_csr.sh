
ip=$(dig +short $1)

cat > $1-csr.json <<EOF
{
  "CN": "$1",
  "hosts": [
    "$1",
    "$ip"
  ],
  "key": {
    "algo": "ecdsa",
    "size": 384
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

