


KUBERNETES_PUBLIC_ADDRESS=$(gcloud compute addresses describe kubernetes-the-hard-way \
  --region $(gcloud config get-value compute/region) \
  --format 'value(address)') # External IP addresses 分配静态IP地址

echo "$KUBERNETES_PUBLIC_ADDRESS"


cat > kubernetes-csr.json <<EOF
{
  "CN": "kubernetes",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "US",
      "L": "Portland",
      "O": "Kubernetes",
      "OU": "Kubernetes The Hard Way",
      "ST": "Oregon"
    }
  ]
}
EOF

# 10.32.0.1 @see https://github.com/kelseyhightower/kubernetes-the-hard-way/blob/master/docs/04-certificate-authority.md
cfssl gencert \
  -ca=../1-ca/ca.pem \
  -ca-key=../1-ca/ca-key.pem \
  -config=../1-ca/ca-config.json \
  -hostname=10.32.0.1,10.240.0.10,10.240.0.11,10.240.0.12,"${KUBERNETES_PUBLIC_ADDRESS}",127.0.0.1,kubernetes.default \
  -profile=kubernetes kubernetes-csr.json | cfssljson -bare kubernetes # -> kubernetes-key.pem, kubernetes.pem
