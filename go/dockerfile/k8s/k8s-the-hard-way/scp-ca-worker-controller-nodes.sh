for instance in worker-0 worker-1 worker-2; do
  gcloud compute scp ca.pem ${instance}-key.pem ${instance}.pem ${instance}:~/
done

for instance in controller-0 controller-1 controller-2; do
  gcloud compute scp ca.pem ca-key.pem kubernetes-key.pem kubernetes.pem \
    service-account-key.pem service-account.pem ${instance}:~/
done
