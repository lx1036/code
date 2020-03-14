

# 查询 kubernetes-the-hard-way 的静态 IP 地址
KUBERNETES_PUBLIC_ADDRESS=$(gcloud compute addresses describe kubernetes-the-hard-way \
  --region $(gcloud config get-value compute/region) \
  --format 'value(address)')


# kubeconfig kube-scheduler 配置文件
# -> kube-scheduler.kubeconfig
{
  kubectl config set-cluster kubernetes-the-hard-way \
    --certificate-authority=../../2-certificates/1-ca/ca.pem \
    --embed-certs=true \
    --server=https://127.0.0.1:6443 \
    --kubeconfig=kube-scheduler.kubeconfig

  kubectl config set-credentials system:kube-scheduler \
    --client-certificate=../../2-certificates/6-kube-scheduler/kube-scheduler.pem \
    --client-key=../../2-certificates/6-kube-scheduler/kube-scheduler-key.pem \
    --embed-certs=true \
    --kubeconfig=kube-scheduler.kubeconfig

  kubectl config set-context default \
    --cluster=kubernetes-the-hard-way \
    --user=system:kube-scheduler \
    --kubeconfig=kube-scheduler.kubeconfig

  kubectl config use-context default --kubeconfig=kube-scheduler.kubeconfig
}
