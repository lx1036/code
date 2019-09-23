for instance in controller-0 controller-1 controller-2; do
  gcloud compute scp deploy-controller-nodes.sh deploy-controller-nodes-health-check.sh deploy-controller-nodes-kubelet-rbac.sh ${instance}:~/
done
