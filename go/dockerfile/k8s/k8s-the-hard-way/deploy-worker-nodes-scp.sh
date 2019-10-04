for instance in worker-0 worker-1 worker-2; do
  gcloud compute scp deploy-worker-nodes.sh ${instance}:~/
done
