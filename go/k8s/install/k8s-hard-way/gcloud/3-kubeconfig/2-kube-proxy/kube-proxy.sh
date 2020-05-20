
# 查询 kubernetes-the-hard-way 的静态 IP 地址
KUBERNETES_PUBLIC_ADDRESS=$(gcloud compute addresses describe kubernetes-the-hard-way \
  --region $(gcloud config get-value compute/region) \
  --format 'value(address)')


# kubeconfig kube-proxy, 为 kube-proxy 服务生成 kubeconfig 配置文件
# -> kube-proxy.kubeconfig
{
  kubectl config set-cluster kubernetes-the-hard-way \
    --certificate-authority=../../2-certificates/1-ca/ca.pem \
    --embed-certs=true \
    --server=https://"${KUBERNETES_PUBLIC_ADDRESS}":6443 \
    --kubeconfig=kube-proxy.kubeconfig

  kubectl config set-credentials system:kube-proxy \
    --client-certificate=../../2-certificates/5-kube-proxy/kube-proxy.pem \
    --client-key=../../2-certificates/5-kube-proxy/kube-proxy-key.pem \
    --embed-certs=true \
    --kubeconfig=kube-proxy.kubeconfig

  kubectl config set-context default \
    --cluster=kubernetes-the-hard-way \
    --user=system:kube-proxy \
    --kubeconfig=kube-proxy.kubeconfig

  kubectl config use-context default --kubeconfig=kube-proxy.kubeconfig
}
