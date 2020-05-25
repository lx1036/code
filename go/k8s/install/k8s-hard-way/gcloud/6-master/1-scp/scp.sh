

for instance in controller-0 controller-1 controller-2; do
  gcloud compute scp \
    ../../2-certificates/1-ca/ca.pem \
    ../../2-certificates/1-ca/ca-key.pem \
    ../../2-certificates/7-kube-apiserver/kubernetes.pem \
    ../../2-certificates/7-kube-apiserver/kubernetes-key.pem \
    ../../2-certificates/8-service-account/service-account.pem \
    ../../2-certificates/8-service-account/service-account-key.pem \
    ../../4-encryption/encryption-config.yaml \
    ../2-deploy/master.sh \
    ${instance}:~/
done

