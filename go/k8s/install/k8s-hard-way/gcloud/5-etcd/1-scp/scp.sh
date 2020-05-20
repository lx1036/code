

for instance in controller-0 controller-1 controller-2; do
  gcloud compute scp \
    ../../2-certificates/1-ca/ca.pem \
    ../../2-certificates/7-kube-apiserver/kubernetes-key.pem \
    ../../2-certificates/7-kube-apiserver/kubernetes.pem ${instance}:~/
done
