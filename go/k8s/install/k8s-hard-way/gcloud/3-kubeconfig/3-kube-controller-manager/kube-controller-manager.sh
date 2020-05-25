
# 查询 kubernetes-the-hard-way 的静态 IP 地址
KUBERNETES_PUBLIC_ADDRESS=$(gcloud compute addresses describe kubernetes-the-hard-way \
  --region $(gcloud config get-value compute/region) \
  --format 'value(address)')


# kubeconfig kube-controller-manager 配置文件
# -> kube-controller-manager.kubeconfig
{
  kubectl config set-cluster kubernetes-the-hard-way \
    --certificate-authority=../../2-certificates/1-ca/ca.pem \
    --embed-certs=true \
    --server=https://127.0.0.1:6443 \
    --kubeconfig=kube-controller-manager.kubeconfig

  kubectl config set-credentials system:kube-controller-manager \
    --client-certificate=../../2-certificates/4-kube-controller-manager/kube-controller-manager.pem \
    --client-key=../../2-certificates/4-kube-controller-manager/kube-controller-manager-key.pem \
    --embed-certs=true \
    --kubeconfig=kube-controller-manager.kubeconfig

  kubectl config set-context default \
    --cluster=kubernetes-the-hard-way \
    --user=system:kube-controller-manager \
    --kubeconfig=kube-controller-manager.kubeconfig

  kubectl config use-context default --kubeconfig=kube-controller-manager.kubeconfig
}
