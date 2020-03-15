






for instance in worker-0 worker-1 worker-2; do
  gcloud compute scp \
    ../../2-certificates/1-ca/ca.pem \
    ../../2-certificates/3-kubelet/${instance}.pem \
    ../../2-certificates/3-kubelet/${instance}-key.pem \
    ../../3-kubeconfig/1-kubelet/${instance}.kubeconfig \
    ../../3-kubeconfig/2-kube-proxy/kube-proxy.kubeconfig \
    deploy-worker.sh \
    ${instance}:~/
done

