# check
for instance in controller-0 controller-1 controller-2; do
  gcloud compute ssh ${instance} --command "kubectl get nodes --kubeconfig admin.kubeconfig"
done
