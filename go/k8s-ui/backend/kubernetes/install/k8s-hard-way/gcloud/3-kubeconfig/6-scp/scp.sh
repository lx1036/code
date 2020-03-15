



# 将 kubelet 与 kube-proxy kubeconfig 配置文件复制到每个 worker 节点上
for instance in worker-0 worker-1 worker-2; do
  gcloud compute scp ../1-kubelet/${instance}.kubeconfig ../2-kube-proxy/kube-proxy.kubeconfig ${instance}:~/
done

# 将 admin、kube-controller-manager 与 kube-scheduler kubeconfig 配置文件复制到每个 master 节点上
for instance in controller-0 controller-1 controller-2; do
  gcloud compute scp \
    ../5-admin/admin.kubeconfig \
    ../3-kube-controller-manager/kube-controller-manager.kubeconfig \
    ../4-kube-scheduler/kube-scheduler.kubeconfig ${instance}:~/
done
