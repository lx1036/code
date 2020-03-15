
## Kubelet Client Credential(https://kubernetes.io/zh/docs/reference/access-authn-authz/node/)
for instance in worker-0 worker-1 worker-2; do
  cat > ${instance}-csr.json <<EOF
  {
    "CN": "system:node:${instance}",
    "key": {
      "algo": "rsa",
      "size": 2048
    },
    "names": [
      {
        "C": "US",
        "L": "Portland",
        "O": "system:nodes",
        "OU": "Kubernetes The Hard Way",
        "ST": "Oregon"
      }
    ]
  }
EOF

EXTERNAL_IP=$(gcloud compute instances describe ${instance} \
  --format 'value(networkInterfaces[0].accessConfigs[0].natIP)')

INTERNAL_IP=$(gcloud compute instances describe ${instance} \
  --format 'value(networkInterfaces[0].networkIP)')

### 给每台 worker 节点创建凭证和私钥
### (worker-0-key.pem/workder-0.pem,worker-1-key.pem/workder-1.pem,worker-2-key.pem/workder-2.pem,)
cfssl gencert \
  -ca=../1-ca/ca.pem \
  -ca-key=../1-ca/ca-key.pem \
  -config=../1-ca/ca-config.json \
  -hostname=${instance},"${EXTERNAL_IP}","${INTERNAL_IP}" \
  -profile=kubernetes \
  ${instance}-csr.json | cfssljson -bare ${instance}
done
