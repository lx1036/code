

KUBERNETES_PUBLIC_ADDRESS=$(gcloud compute addresses describe kubernetes-the-hard-way \
    --region $(gcloud config get-value compute/region) \
    --format 'value(address)')

kubectl config set-cluster kubernetes-the-hard-way \
  --certificate-authority=../2-certificates/1-ca/ca.pem \
  --embed-certs=true \
  --server=https://"${KUBERNETES_PUBLIC_ADDRESS}":6443

kubectl config set-credentials admin \
  --client-certificate=../2-certificates/2-admin/admin.pem \
  --client-key=../2-certificates/2-admin/admin-key.pem

kubectl config set-context kubernetes-the-hard-way \
  --cluster=kubernetes-the-hard-way \
  --user=admin

kubectl config use-context kubernetes-the-hard-way


# validation
kubectl get componentstatuses -o json
kubectl get nodes
