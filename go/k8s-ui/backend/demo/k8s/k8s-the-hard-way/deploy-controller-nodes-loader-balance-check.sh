KUBERNETES_PUBLIC_ADDRESS=$(gcloud compute addresses describe kubernetes-the-hard-way \
  --region $(gcloud config get-value compute/region) \
  --format 'value(address)')

curl --cacert ca.pem https://${KUBERNETES_PUBLIC_ADDRESS}:6443/version
